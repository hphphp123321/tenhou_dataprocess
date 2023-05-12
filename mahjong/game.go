package mahjong

import (
	"github.com/hphphp123321/go-common"
	"sort"
)

type MahjongGame struct {
	P0               *MahjongPlayer
	P1               *MahjongPlayer
	P2               *MahjongPlayer
	P3               *MahjongPlayer
	Tiles            *MahjongTiles
	WindRound        int
	NumGame          int
	NumRiichi        int
	NumHonba         int
	CurrentRiichiNum int
	Position         int
	PosPlayer        map[int]*MahjongPlayer
}

func NewMahjongGame(playerSlice []*MahjongPlayer) *MahjongGame {
	game := MahjongGame{Tiles: &MahjongTiles{}}
	game.Reset(playerSlice)
	return &game
}

func (game *MahjongGame) NewGameRound(windRound int) {
	game.WindRound = windRound
	game.NumGame += 1
	game.Tiles.Reset()
	game.Position = 0
	game.PosPlayer[(16-windRound)%4] = game.P0
	game.PosPlayer[(17-windRound)%4] = game.P1
	game.PosPlayer[(18-windRound)%4] = game.P2
	game.PosPlayer[(19-windRound)%4] = game.P3
	game.P0.ResetForRound()
	game.P1.ResetForRound()
	game.P2.ResetForRound()
	game.P3.ResetForRound()
}

func (game *MahjongGame) Reset(playerSlice []*MahjongPlayer) {
	game.Tiles.Reset()
	game.NumGame = 0
	game.WindRound = 0
	game.NumRiichi = 0
	game.NumHonba = 0
	game.CurrentRiichiNum = 0
	game.Position = 0

	if playerSlice != nil {
		game.P0 = playerSlice[0]
		game.P1 = playerSlice[1]
		game.P2 = playerSlice[2]
		game.P3 = playerSlice[3]
	}
	game.P0.ResetForGame()
	game.P1.ResetForGame()
	game.P2.ResetForGame()
	game.P3.ResetForGame()
	game.PosPlayer = map[int]*MahjongPlayer{0: playerSlice[0], 1: playerSlice[1], 2: playerSlice[2], 3: playerSlice[3]}
}

func (game *MahjongGame) ProcessOtherCall(pMain *MahjongPlayer, call *Call) {
	if call.CallType == Skip {
		return
	} else if call.CallType == Chi {
		game.processChi(pMain, call)
		game.Position = pMain.Wind
		game.breakIppatsu()
		game.breakRyuukyoku()
	} else if call.CallType == Pon {
		game.processPon(pMain, call)
		game.Position = pMain.Wind
		game.breakIppatsu()
		game.breakRyuukyoku()
	} else if call.CallType == DaiMinKan {
		game.processDaiMinKan(pMain, call)
		game.Position = pMain.Wind
		game.breakIppatsu()
		game.breakRyuukyoku()
	}
}

func (game *MahjongGame) ProcessSelfCall(pMain *MahjongPlayer, call *Call) {
	if call.CallType == Discard {
		game.DiscardTileProcess(pMain, call.CallTiles[0])
	} else if call.CallType == ShouMinKan {
		game.processShouMinKan(pMain, call)
		game.breakIppatsu()
		game.breakRyuukyoku()
	} else if call.CallType == AnKan {
		game.processAnKan(pMain, call)
		game.breakIppatsu()
		game.breakRyuukyoku()
	} else if call.CallType == Riichi {
		game.processRiichi(pMain, call)
		game.breakIppatsu()
	}
}

func (game *MahjongGame) GetTileProcess(pMain *MahjongPlayer, tileID int) {
	if pMain.IsRiichi {
		for _, tile := range pMain.HandTiles {
			game.Tiles.allTiles[tile].discardable = false
		}
	} else {
		for _, tile := range pMain.HandTiles {
			game.Tiles.allTiles[tile].discardable = true
		}
	}
	pMain.HandTiles = append(pMain.HandTiles, tileID)
	pMain.JunNum++
	pMain.JunFuriten = false
}

func (game *MahjongGame) DealTile() {
	game.Tiles.DealTile()
}

func (game *MahjongGame) DiscardTileProcess(pMain *MahjongPlayer, tileID int) {
	if !game.Tiles.allTiles[tileID].discardable {
		panic("Illegal Discard ID")
	}
	if tileID == pMain.HandTiles[len(pMain.HandTiles)-1] {
		pMain.TilesTsumoGiri = append(pMain.TilesTsumoGiri, 1)
	} else {
		pMain.TilesTsumoGiri = append(pMain.TilesTsumoGiri, 0)
	}
	pMain.HandTiles.Remove(tileID)
	pMain.DiscardTiles.Append(tileID)
	pMain.BoardTiles.Append(tileID)
	sort.Ints(pMain.HandTiles)
	pMain.IppatsuStatus = false
	game.Tiles.allTiles[tileID].discardWind = pMain.Wind
	pMain.ShantenNum = pMain.GetShantenNum()
	if pMain.ShantenNum == 0 {
		pMain.TenhaiSlice = pMain.GetTenhaiSlice()
		flag := false
		for _, tile := range pMain.DiscardTiles {
			if common.SliceContain(pMain.TenhaiSlice, tile/4) {
				flag = true
			}
		}
		if flag {
			pMain.DiscardFuriten = true
		} else {
			pMain.DiscardFuriten = false
		}
	}
	otherWinds := game.getOtherWinds()
	for _, wind := range otherWinds {
		if common.SliceContain(game.PosPlayer[wind].TenhaiSlice, tileID/4) {
			game.PosPlayer[wind].FuritenStatus = true
		} else {
			game.PosPlayer[wind].FuritenStatus = false
		}
	}
}

func (game *MahjongGame) GetDiscardableSlice(handTiles Tiles) Tiles {
	tiles := Tiles{}
	for _, tile := range handTiles {
		if game.Tiles.allTiles[tile].discardable {
			tiles.Append(tile)
		}
	}
	return tiles
}

func (game *MahjongGame) JudgeDiscardCall(pMain *MahjongPlayer) Calls {
	var validCalls = make(Calls, 0)
	discardableSlice := game.GetDiscardableSlice(pMain.HandTiles)
	for _, tile := range discardableSlice {
		call := &Call{
			CallType:         Discard,
			CallTiles:        Tiles{tile, -1, -1, -1},
			CallTilesFromWho: []int{pMain.Wind, -1, -1, -1},
		}
		validCalls = append(validCalls, call)
	}
	return validCalls
}

func (game *MahjongGame) JudgeSelfCalls(pMain *MahjongPlayer) Calls {
	var validCalls Calls
	riichi := game.judgeRiichi(pMain)
	shouMinKan := game.judgeShouMinKan(pMain)
	anKan := game.judgeAnKan(pMain)
	discard := game.JudgeDiscardCall(pMain)
	validCalls = append(validCalls, riichi...)
	validCalls = append(validCalls, shouMinKan...)
	validCalls = append(validCalls, anKan...)
	validCalls = append(validCalls, discard...)
	return validCalls
}

func (game *MahjongGame) JudgeOtherCalls(pMain *MahjongPlayer, tileID int) Calls {
	validCalls := Calls{&Call{
		CallType:         Skip,
		CallTiles:        Tiles{-1, -1, -1, -1},
		CallTilesFromWho: []int{-1, -1, -1, -1},
	}}
	daiMinKan := game.judgeDaiMinKan(pMain, tileID)
	pon := game.judgePon(pMain, tileID)
	chi := game.judgeChi(pMain, tileID)
	validCalls = append(validCalls, daiMinKan...)
	validCalls = append(validCalls, pon...)
	validCalls = append(validCalls, chi...)
	return validCalls
}

func (game *MahjongGame) processRiichi(pMain *MahjongPlayer, call *Call) {
	if pMain.JunNum == 1 && pMain.IppatsuStatus {
		pMain.IsDaburuRiichi = true
	}
	pMain.IsRiichi = true
	riichiTile := call.CallTiles[0]
	pMain.Points -= 1000
	game.DiscardTileProcess(pMain, riichiTile)
	pMain.IppatsuStatus = true
}

func (game *MahjongGame) processChi(pMain *MahjongPlayer, call *Call) {
	pMain.HandTiles.Remove(call.CallTiles[0])
	pMain.HandTiles.Remove(call.CallTiles[1])
	tileID := call.CallTiles[2]
	subWind := call.CallTilesFromWho[2]
	game.PosPlayer[subWind].BoardTiles.Remove(tileID)
	pMain.Melds = append(pMain.Melds, call)

	// 食替
	tileClass := tileID / 4
	tile1Class := call.CallTiles[0] / 4
	tile2Class := call.CallTiles[1] / 4
	posClass := make(Tiles, 0, 4)
	if tile1Class-tile2Class == 1 || tile2Class-tile1Class == 1 {
		if !common.SliceContain([]int{0, 9, 18, 8, 17, 26}, tile1Class) && !common.SliceContain([]int{0, 9, 18, 8, 17, 26}, tile2Class) {
			posClass = Tiles{tile1Class - 1, tile1Class + 1, tile2Class - 1, tile2Class + 1}
			posClass.Remove(tileClass)
			posClass.Remove(tile1Class)
			posClass.Remove(tile2Class)
		}
	}
	posClass.Append(tileClass)
	for _, tile := range pMain.HandTiles {
		if common.SliceContain(posClass, tile/4) {
			game.Tiles.allTiles[tile].discardable = false
		} else {
			game.Tiles.allTiles[tile].discardable = true
		}
	}
	pMain.JunNum++
}

func (game *MahjongGame) processPon(pMain *MahjongPlayer, call *Call) {
	pMain.HandTiles.Remove(call.CallTiles[0])
	pMain.HandTiles.Remove(call.CallTiles[1])
	tileID := call.CallTiles[2]
	subWind := call.CallTilesFromWho[2]
	game.PosPlayer[subWind].BoardTiles.Remove(tileID)
	pMain.Melds = append(pMain.Melds, call)
	// 食替
	tileClass := tileID / 4
	for _, tile := range pMain.HandTiles {
		if tile/4 == tileClass {
			game.Tiles.allTiles[tile].discardable = false
		} else {
			game.Tiles.allTiles[tile].discardable = true
		}
	}
	pMain.JunNum++
}

func (game *MahjongGame) processDaiMinKan(pMain *MahjongPlayer, call *Call) {
	pMain.HandTiles.Remove(call.CallTiles[0])
	pMain.HandTiles.Remove(call.CallTiles[1])
	pMain.HandTiles.Remove(call.CallTiles[2])
	tileID := call.CallTiles[3]
	subWind := call.CallTilesFromWho[3]
	game.PosPlayer[subWind].BoardTiles.Remove(tileID)

	pMain.Melds = append(pMain.Melds, call)
}

func (game *MahjongGame) processAnKan(pMain *MahjongPlayer, call *Call) {
	pMain.HandTiles.Remove(call.CallTiles[0])
	pMain.HandTiles.Remove(call.CallTiles[1])
	pMain.HandTiles.Remove(call.CallTiles[2])
	pMain.HandTiles.Remove(call.CallTiles[3])
	pMain.Melds = append(pMain.Melds, call)
}

func (game *MahjongGame) processShouMinKan(pMain *MahjongPlayer, call *Call) {
	tileID := call.CallTiles[3]
	pMain.HandTiles.Remove(tileID)
	for i, meld := range pMain.Melds {
		if meld.CallType != Pon {
			continue
		}
		if meld.CallTiles[0]/4 != tileID/4 {
			continue
		}
		pMain.Melds[i] = call
		return
	}
	panic("ShouMinKan not success!")
}

func (game *MahjongGame) judgeRiichi(pMain *MahjongPlayer) Calls {
	if pMain.IsRiichi || (pMain.ShantenNum > 1 && pMain.JunNum > 1) || pMain.Points < 1000 || game.GetNumRemainTiles() < 4 {
		return make(Calls, 0)
	}
	if len(pMain.HandTiles) != 14 {
		for _, meld := range pMain.Melds {
			if meld.CallType != AnKan {
				return make(Calls, 0)
			}
		}
	}
	tiles := pMain.GetRiichiTiles()
	riichiCalls := make(Calls, 0, len(tiles))
	for _, tileID := range tiles {
		riichiCalls = append(riichiCalls, &Call{
			CallType:         Riichi,
			CallTiles:        Tiles{tileID, -1, -1, -1},
			CallTilesFromWho: []int{pMain.Wind, -1, -1, -1},
		})
	}
	return riichiCalls
}

func (game *MahjongGame) judgeChi(pMain *MahjongPlayer, tileID int) Calls {
	discardWind := game.Tiles.allTiles[tileID].discardWind
	chiClass := tileID / 4
	if pMain.IsRiichi || (pMain.Wind-discardWind+4)%4 != 1 || chiClass > 27 || game.Tiles.allTiles[tileID].isLast || game.GetNumRemainTiles() == 0 {
		return make(Calls, 0)
	}
	handTilesClass := Tiles(pMain.GetHandTilesClass())
	if !(common.SliceContain(handTilesClass, chiClass-1) ||
		common.SliceContain(handTilesClass, chiClass-2) ||
		common.SliceContain(handTilesClass, chiClass+1) ||
		common.SliceContain(handTilesClass, chiClass+2)) {
		return make(Calls, 0)
	}
	var posCombinations [][]int
	if common.SliceContain([]int{0, 9, 18}, chiClass) {
		posCombinations = append(posCombinations, []int{chiClass + 1, chiClass + 2})
	} else if common.SliceContain([]int{1, 10, 19}, chiClass) {
		posCombinations = append(posCombinations, []int{chiClass - 1, chiClass + 1})
		posCombinations = append(posCombinations, []int{chiClass + 1, chiClass + 2})
	} else if common.SliceContain([]int{7, 16, 25}, chiClass) {
		posCombinations = append(posCombinations, []int{chiClass - 1, chiClass + 1})
		posCombinations = append(posCombinations, []int{chiClass - 2, chiClass - 1})
	} else if common.SliceContain([]int{8, 17, 26}, chiClass) {
		posCombinations = append(posCombinations, []int{chiClass - 2, chiClass - 1})
	} else {
		posCombinations = append(posCombinations, []int{chiClass - 1, chiClass + 1})
		posCombinations = append(posCombinations, []int{chiClass + 1, chiClass + 2})
		posCombinations = append(posCombinations, []int{chiClass - 2, chiClass - 1})
	}
	var posCalls Calls
	for _, posCom := range posCombinations {
		tile1Class := posCom[0]
		tile2Class := posCom[1]
		if !(common.SliceContain(handTilesClass, tile1Class) && common.SliceContain(handTilesClass, tile2Class)) {
			continue
		}
		tile1Idx1 := handTilesClass.Index(tile1Class, 0)
		tile1ID := pMain.HandTiles[tile1Idx1]
		tile2Idx1 := handTilesClass.Index(tile2Class, 0)
		tile2ID := pMain.HandTiles[tile2Idx1]
		posCall := &Call{
			CallType:         Chi,
			CallTiles:        Tiles{tile1ID, tile2ID, tileID, -1},
			CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
		}
		posCalls.Append(posCall)
		if common.SliceContain([]int{16, 52, 88}, tile1ID) {
			tile1Idx2 := handTilesClass.Index(tile1Class, tile1Idx1+1)
			if tile1Idx2 != -1 {
				tile1ID = pMain.HandTiles[tile1Idx2]
				posCall = &Call{
					CallType:         Chi,
					CallTiles:        Tiles{tile1ID, tile2ID, tileID, -1},
					CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
				}
				posCalls.Append(posCall)
			}
		} else if common.SliceContain([]int{16, 52, 88}, tile2ID) {
			tile2Idx2 := handTilesClass.Index(tile2Class, tile2Idx1+1)
			if tile2Idx2 != -1 {
				tile2ID = pMain.HandTiles[tile2Idx2]
				posCall = &Call{
					CallType:         Chi,
					CallTiles:        Tiles{tile1ID, tile2ID, tileID, -1},
					CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
				}
				posCalls.Append(posCall)
			}
		}
	}
	// 食替
	if len(pMain.HandTiles) > 7 {
		return posCalls
	}
	delIdxSlice := make([]int, 0, len(posCalls))
	for i, call := range posCalls {
		tile1ID := call.CallTiles[0]
		tile2ID := call.CallTiles[1]
		tile3ID := call.CallTiles[2]
		handTilesCopy := pMain.HandTiles.Copy()
		handTilesCopy.Remove(tile1ID)
		handTilesCopy.Remove(tile2ID)
		tileClass := tile3ID / 4
		tile1Class := tile1ID / 4
		tile2Class := tile2ID / 4
		posClass := make(Tiles, 0, 4)
		if tile1Class-tile2Class == 1 || tile2Class-tile1Class == 1 {
			if !common.SliceContain([]int{0, 9, 18, 8, 17, 26}, tile1Class) &&
				!common.SliceContain([]int{0, 9, 18, 8, 17, 26}, tile2Class) {
				posClass = Tiles{tile1Class - 1, tile1Class + 1, tile2Class - 1, tile2Class + 1}
				posClass.Remove(tileClass)
				posClass.Remove(tile1Class)
				posClass.Remove(tile2Class)
			}
		}
		posClass.Append(tileClass)
		flag := true
		for _, handTilesID := range handTilesCopy {
			if !common.SliceContain(posClass, handTilesID/4) {
				flag = false
				break
			}
		}
		if flag {
			delIdxSlice = append(delIdxSlice, i)
		}
	}
	if len(delIdxSlice) > 0 {
		posCalls = common.RemoveIndex(posCalls, delIdxSlice...)
	}
	return posCalls
}

func (game *MahjongGame) judgePon(pMain *MahjongPlayer, tileID int) Calls {
	if pMain.IsRiichi {
		return make(Calls, 0)
	}
	discardWind := game.Tiles.allTiles[tileID].discardWind
	ponClass := tileID / 4
	tilesClass := Tiles(pMain.GetHandTilesClass())
	tileCount := tilesClass.Count(ponClass)
	if tileCount < 2 || game.Tiles.allTiles[tileID].isLast || game.GetNumRemainTiles() == 0 {
		return make(Calls, 0)
	}
	var posCalls Calls
	tile1Idx := tilesClass.Index(ponClass, 0)
	tile1ID := pMain.HandTiles[tile1Idx]
	tile2Idx := tilesClass.Index(ponClass, tile1Idx+1)
	tile2ID := pMain.HandTiles[tile2Idx]
	posCall := &Call{
		CallType:         Pon,
		CallTiles:        Tiles{tile1ID, tile2ID, tileID, -1},
		CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
	}
	posCalls.Append(posCall)
	if tileCount == 3 {
		tile3Idx := tilesClass.Index(ponClass, tile2Idx+1)
		if tile3Idx == -1 {
			panic("no tile3")
		}
		tile3ID := pMain.HandTiles[tile3Idx]
		if common.SliceContain([]int{16, 52, 88}, tile1ID) {
			posCall = &Call{
				CallType:         Pon,
				CallTiles:        Tiles{tile2ID, tile3ID, tileID, -1},
				CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
			}
			posCalls.Append(posCall)
		} else if common.SliceContain([]int{16, 52, 88}, tile2ID) {
			posCall = &Call{
				CallType:         Pon,
				CallTiles:        Tiles{tile1ID, tile3ID, tileID, -1},
				CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
			}
			posCalls.Append(posCall)
		} else if common.SliceContain([]int{16, 52, 88}, tile3ID) {
			posCall = &Call{
				CallType:         Pon,
				CallTiles:        Tiles{tile1ID, tile3ID, tileID, -1},
				CallTilesFromWho: []int{pMain.Wind, pMain.Wind, discardWind, -1},
			}
			posCalls.Append(posCall)
		}
	}
	return posCalls
}

func (game *MahjongGame) judgeDaiMinKan(pMain *MahjongPlayer, tileID int) Calls {
	discardWind := game.Tiles.allTiles[tileID].discardWind
	if pMain.IsRiichi || game.Tiles.allTiles[tileID].isLast || game.GetNumRemainTiles() == 0 {
		return make(Calls, 0)
	}
	kanClass := tileID / 4
	class := Tiles(pMain.GetHandTilesClass())
	tileCount := class.Count(kanClass)
	if tileCount < 2 {
		return make(Calls, 0)
	}
	posKanTiles := Tiles{kanClass * 4, kanClass*4 + 1, kanClass*4 + 2, kanClass*4 + 3}
	posKanTiles.Remove(tileID)
	tile0 := posKanTiles[0]
	tile1 := posKanTiles[1]
	tile2 := posKanTiles[2]
	if !common.SliceContain(pMain.HandTiles, tile0) ||
		!common.SliceContain(pMain.HandTiles, tile1) ||
		!common.SliceContain(pMain.HandTiles, tile2) {
		return make(Calls, 0)
	}
	var posCalls Calls
	posCall := &Call{
		CallType:         DaiMinKan,
		CallTiles:        Tiles{tile0, tile1, tile2, tileID},
		CallTilesFromWho: []int{pMain.Wind, pMain.Wind, pMain.Wind, discardWind},
	}
	posCalls.Append(posCall)
	return posCalls
}

func (game *MahjongGame) judgeAnKan(pMain *MahjongPlayer) Calls {
	if len(pMain.HandTiles) == 2 || game.GetNumRemainTiles() == 0 {
		return make(Calls, 0)
	}
	tilesClass := Tiles(pMain.GetHandTilesClass())
	var posClass Tiles
	var posCalls = make(Calls, 0)
	for _, tileClass := range tilesClass {
		if common.SliceContain(posClass, tileClass) {
			continue
		}
		if tilesClass.Count(tileClass) == 4 {
			posClass.Append(tileClass)
			a := tilesClass.Index(tileClass, 0)
			b := tilesClass.Index(tileClass, a+1)
			c := tilesClass.Index(tileClass, b+1)
			d := tilesClass.Index(tileClass, c+1)
			if a == -1 || b == -1 || c == -1 || d == -1 {
				panic("index error")
			}
			posCall := &Call{
				CallType:         AnKan,
				CallTiles:        Tiles{pMain.HandTiles[a], pMain.HandTiles[b], pMain.HandTiles[c], pMain.HandTiles[d]},
				CallTilesFromWho: []int{pMain.Wind, pMain.Wind, pMain.Wind, pMain.Wind},
			}
			if pMain.IsRiichi {
				// judge ankan when player riichi
				// if a riichi player has 4 same tiles in hand but not draw the 4th tile, then this ankan is not valid
				if d != len(pMain.HandTiles)-1 {
					continue
				}
				// if a riichi player's tenhai changed after ankan, then this ankan is not valid
				tenhaiSlice := pMain.TenhaiSlice
				tmpHandTiles := common.RemoveIndex(pMain.HandTiles, a, b, c, d)
				melds := pMain.Melds.Copy()
				melds.Append(posCall)
				tenhaiSliceAfterKan := GetTenhaiSlice(tmpHandTiles, melds)
				if !common.SliceEqual(tenhaiSlice, tenhaiSliceAfterKan) {
					continue
				}
			}
			posCalls.Append(posCall)
		}
	}
	return posCalls
}

func (game *MahjongGame) judgeShouMinKan(pMain *MahjongPlayer) Calls {
	if len(pMain.Melds) == 0 || game.GetNumRemainTiles() == 0 {
		return make(Calls, 0)
	}
	var posCalls Calls
	for _, call := range pMain.Melds {
		if call.CallType != Pon {
			continue
		}
		ponClass := call.CallTiles[0] / 4
		for _, tileID := range pMain.HandTiles {
			if tileID/4 == ponClass && game.Tiles.allTiles[tileID].discardable {
				c := call.Copy()
				c.CallType = ShouMinKan
				c.CallTiles = append(c.CallTiles[:3], tileID)
				c.CallTilesFromWho = append(c.CallTilesFromWho[:3], pMain.Wind)
				posCalls = append(posCalls, c)
			}
		}
	}
	return posCalls
}

func (game *MahjongGame) GetNumRemainTiles() int {
	return game.Tiles.NumRemainTiles
}

func (game *MahjongGame) getOtherWinds() []int {
	otherWinds := []int{0, 1, 2, 3}
	for i, v := range otherWinds {
		if v == game.Position {
			otherWinds = append(otherWinds[:i], otherWinds[i+1:]...)
			break
		}
	}
	return otherWinds
}

func (game *MahjongGame) breakIppatsu() {
	for wind, player := range game.PosPlayer {
		if wind == game.Position {
			continue
		}
		player.IppatsuStatus = false
	}
}

func (game *MahjongGame) breakRyuukyoku() {
	for _, player := range game.PosPlayer {
		player.RyuukyokuStatus = false
	}
}
