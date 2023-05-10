package tenhou

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"github.com/hphphp123312/mahjong-datapreprocess/mahjong"
	"github.com/hphphp123321/go-common"
	"regexp"
	"strconv"
	"strings"
)

func DecompressBzipBytes(content []byte) string {
	bz2Reader := bzip2.NewReader(bytes.NewReader(content))
	logBytes := make([]byte, 1024*100)
	_, err := bz2Reader.Read(logBytes)
	ErrPanic(err)
	logBytes = bytes.Trim(logBytes, "\x00")
	return string(logBytes)
}

func ProcessGame(logs string) (retBoardStates []*mahjong.BoardState, dan []int) {
	retBoardStates = []*mahjong.BoardState{}
	defer func() {
		if err := recover(); err != nil { // 如果recover返回为空，说明没报错
			fmt.Println(err)
			retBoardStates = []*mahjong.BoardState{}
		}
	}()

	allInfo := compileInitInfo(logs)
	gameInfo := allInfo[0]
	gameType, startOya, dan := compileGameInfo(gameInfo)
	if gameType != 169 || startOya != 0 {
		return
	}
	p0 := mahjong.NewMahjongPlayer()
	p1 := mahjong.NewMahjongPlayer()
	p2 := mahjong.NewMahjongPlayer()
	p3 := mahjong.NewMahjongPlayer()
	playerSlices := []*mahjong.MahjongPlayer{p0, p1, p2, p3}
	game := mahjong.NewMahjongGame(playerSlices)
	game.NumGame = 1
	finalReward := compileFinalReward(allInfo[len(allInfo)-1])
	p0.FinalReward = int(finalReward[1])
	p1.FinalReward = int(finalReward[3])
	p2.FinalReward = int(finalReward[5])
	p3.FinalReward = int(finalReward[7])

	// round info
	for idx := 1; idx < len(allInfo); idx++ {
		var rSlice []*mahjong.BoardState
		roundInfo := allInfo[idx]
		windRound, numHonba, numRiichi, doraIndicators, points, rewards := compileRoundInfo(roundInfo)

		game.NewGameRound(windRound)
		for j, point := range points {
			playerSlices[j].Points = point * 100
			playerSlices[j].Reward = rewards[2*j+1] * 100
		}
		hai0, hai1, hai2, hai3 := compileInitTiles(roundInfo)
		p0.InitTilesWind(hai0, (16-windRound)%4)
		p1.InitTilesWind(hai1, (17-windRound)%4)
		p2.InitTilesWind(hai2, (18-windRound)%4)
		p3.InitTilesWind(hai3, (19-windRound)%4)

		ops := compileOps(roundInfo)
		for i := 0; i < len(ops)-1; i++ {
			// main loop
			boardState := mahjong.NewBoardState()
			boardState.RoundWind = windRound / 4
			boardState.NumHonba = numHonba
			boardState.NumRiichi = numRiichi
			boardState.DoraIndicators = doraIndicators

			var playerWinds []int
			var validActionsSlices []mahjong.Calls
			op := ops[i]
			opType := compileOpType(op)
			var player *mahjong.MahjongPlayer

			if playerIdx, ok := getTileMap[opType]; ok {
				// Get Tile Stage

				player = playerSlices[playerIdx]
				playerWinds = append(playerWinds, player.Wind)
				tileID := compileOpTile(op)
				game.DealTile()
				game.GetTileProcess(player, tileID)
				validActions := game.JudgeSelfCalls(player)
				validActionsSlices = append(validActionsSlices, validActions)
				if i != 0 && len(rSlice) != 0 {
					for j := len(rSlice) - 1; rSlice[j].RealActionIdx == -1; j-- {
						realCall := &mahjong.Call{
							CallType:         mahjong.Skip,
							CallTiles:        mahjong.Tiles{-1, -1, -1, -1},
							CallTilesFromWho: []int{-1, -1, -1, -1},
						}
						rSlice[j].RealActionIdx = rSlice[j].ValidActions.Index(realCall)
						if j == 0 {
							break
						}
					}
				}
			} else if playerIdx, ok := discardTileMap[opType]; ok {
				// Discard Tile Stage

				player = playerSlices[playerIdx]
				tileID := compileOpTile(op)
				game.DiscardTileProcess(player, tileID)
				for _, wind := range GetOtherWinds(player.Wind) {
					validActions := game.JudgeOtherCalls(game.PosPlayer[wind], tileID)
					if len(validActions) > 1 {
						playerWinds = append(playerWinds, wind)
						validActionsSlices = append(validActionsSlices, validActions)
					}
				}
				for j := len(rSlice) - 1; rSlice[j].RealActionIdx == -1; j-- {
					if player.RiichiStep == 1 {
						realCall := &mahjong.Call{
							CallType:         mahjong.Riichi,
							CallTiles:        mahjong.Tiles{tileID, -1, -1, -1},
							CallTilesFromWho: []int{player.Wind, -1, -1, -1},
						}
						rSlice[j].RealActionIdx = rSlice[j].ValidActions.Index(realCall)
					} else {
						realCall := &mahjong.Call{
							CallType:         mahjong.Discard,
							CallTiles:        mahjong.Tiles{tileID, -1, -1, -1},
							CallTilesFromWho: []int{player.Wind, -1, -1, -1},
						}
						rSlice[j].RealActionIdx = rSlice[j].ValidActions.Index(realCall)
					}
					if j == 0 {
						break
					}
				}
			} else if opType == "N" {

				// Player Call
				//lastOp := ops[i-1]
				//lastOpType := compileOpType(lastOp)
				playerIdx := compileOpWho(op)
				player = playerSlices[playerIdx]
				playerWinds = append(playerWinds, player.Wind)
				meldBytes := compileMeld(op)
				meldCall := ProcessMeld(meldBytes, player.Wind)
				if common.SliceContain([]mahjong.CallType{mahjong.AnKan, mahjong.ShouMinKan}, meldCall.CallType) {
					game.ProcessSelfCall(player, meldCall)
				} else {
					game.ProcessOtherCall(player, meldCall)
				}
				if common.SliceContain([]mahjong.CallType{mahjong.AnKan, mahjong.ShouMinKan, mahjong.DaiMinKan}, meldCall.CallType) {
					playerWinds = []int{}
					player.ShantenNum = player.GetShantenNum()
				}
				validActions := game.JudgeDiscardCall(player)
				validActionsSlices = append(validActionsSlices, validActions)
				var delSlice []int
				for j := len(rSlice) - 1; rSlice[j].RealActionIdx == -1; j-- {
					if rSlice[j].PlayerWind == player.Wind {
						rSlice[j].RealActionIdx = rSlice[j].ValidActions.Index(meldCall)
					} else {
						delSlice = append(delSlice, j)
					}
					if j == 0 {
						break
					}
				}
				for j := 0; j < len(delSlice); j++ {
					delIdx := delSlice[j]
					rSlice = append(rSlice[:delIdx], rSlice[delIdx+1:]...)
				}
			} else if opType == "REACH" {

				// reach
				playerIdx := compileOpWho(op)
				player = playerSlices[playerIdx]
				step := compileReachStep(op)
				player.RiichiStep = step
				if step == 2 {
					player.IsRiichi = true
				}
			} else if opType == "DORA" {

				// DORA after kan
				tileID := compileOpTile(op)
				doraIndicators.Append(tileID)
			}

			if len(playerWinds) == 0 {
				continue
			}
			for wind, p := range []*mahjong.PlayerState{&boardState.P0, &boardState.P1, &boardState.P2, &boardState.P3} {
				p.Points = game.PosPlayer[wind].Points
				p.Melds = game.PosPlayer[wind].Melds[:]
				p.DiscardTiles = game.PosPlayer[wind].DiscardTiles[:]
				p.TilesTsumoGiri = game.PosPlayer[wind].TilesTsumoGiri[:]
				p.IsRiichi = game.PosPlayer[wind].IsRiichi
				p.PointsReward = game.PosPlayer[wind].Reward
				p.FinalReward = game.PosPlayer[wind].FinalReward
			}
			for k, w := range playerWinds {
				boardStateCopy := mahjong.BoardStateCopy(boardState)
				boardStateCopy.HandTiles = game.PosPlayer[w].HandTiles.Copy()
				boardStateCopy.PlayerWind = w
				boardStateCopy.Position = player.Wind
				boardStateCopy.ValidActions = validActionsSlices[k].Copy()
				boardStateCopy.RealActionIdx = -1
				boardStateCopy.NumRemainTiles = game.Tiles.NumRemainTiles
				rSlice = append(rSlice, boardStateCopy)
			}
		}
		if len(rSlice) == 0 {
			continue
		}
		var delSlice []int
		for j := len(rSlice) - 1; rSlice[j].RealActionIdx == -1; j-- {
			if rSlice[j].RealActionIdx != -1 {
				break
			} else {
				delSlice = append(delSlice, j)
			}
			if j == 0 {
				break
			}
		}
		for j := 0; j < len(delSlice); j++ {
			delIdx := delSlice[j]
			rSlice = append(rSlice[:delIdx], rSlice[delIdx+1:]...)
		}
		retBoardStates = append(retBoardStates, rSlice...)
	}
	for _, boardState := range retBoardStates {
		if boardState.RealActionIdx < 0 {
			panic("no real action")
		}
	}
	return retBoardStates, dan
}

func compileReachStep(op string) (meld int) {
	re := regexp.MustCompile("step=\"(.*?)\"")
	var err error
	meld, err = strconv.Atoi(re.FindAllStringSubmatch(op, -1)[0][1])
	ErrPanic(err)
	return meld
}

func compileMeld(op string) (meld int) {
	re := regexp.MustCompile("m=\"(.*?)\"")
	var err error
	meld, err = strconv.Atoi(re.FindAllStringSubmatch(op, -1)[0][1])
	ErrPanic(err)
	return meld
}

func compileOpWho(op string) (who int) {
	re := regexp.MustCompile("who=\"(.*?)\"")
	var err error
	who, err = strconv.Atoi(re.FindAllStringSubmatch(op, -1)[0][1])
	ErrPanic(err)
	return who
}

func compileOpTile(op string) (tileID int) {
	re := regexp.MustCompile("[0-9]+")
	var err error
	tileID, err = strconv.Atoi(re.FindString(op))
	ErrPanic(err)
	return tileID
}

func compileOpType(op string) (opType string) {
	re := regexp.MustCompile("^[A-Za-z]+")
	opType = re.FindString(op)
	return opType
}

func compileOps(roundInfo string) (ops []string) {
	re := regexp.MustCompile("<(.*?)/>")
	opsSlices := re.FindAllStringSubmatch(roundInfo, -1)
	for _, op := range opsSlices {
		ops = append(ops, op[1])
	}
	return ops
}

func compileInitTiles(roundInfo string) (t0 mahjong.Tiles, t1 mahjong.Tiles, t2 mahjong.Tiles, t3 mahjong.Tiles) {
	t0, t1, t2, t3 = make(mahjong.Tiles, 13, 14), make(mahjong.Tiles, 13, 14), make(mahjong.Tiles, 13, 14), make(mahjong.Tiles, 13, 14)

	re := regexp.MustCompile("hai0=\"(.*?)\"")
	tiles := re.FindAllStringSubmatch(roundInfo, -1)
	tileSplit := strings.Split(tiles[0][1], ",")
	copy(t0, StringSlices2IntSlices(tileSplit))

	re = regexp.MustCompile("hai1=\"(.*?)\"")
	tiles = re.FindAllStringSubmatch(roundInfo, -1)
	tileSplit = strings.Split(tiles[0][1], ",")
	copy(t1, StringSlices2IntSlices(tileSplit))

	re = regexp.MustCompile("hai2=\"(.*?)\"")
	tiles = re.FindAllStringSubmatch(roundInfo, -1)
	tileSplit = strings.Split(tiles[0][1], ",")
	copy(t2, StringSlices2IntSlices(tileSplit))

	re = regexp.MustCompile("hai3=\"(.*?)\"")
	tiles = re.FindAllStringSubmatch(roundInfo, -1)
	tileSplit = strings.Split(tiles[0][1], ",")
	copy(t3, StringSlices2IntSlices(tileSplit))

	return t0, t1, t2, t3
}

func compileRoundInfo(roundInfo string) (windRound int, numHonba int, numRiichi int, doraIndicators mahjong.Tiles, points []int, rewards []int) {
	re := regexp.MustCompile("seed=\"(.*?)\"")
	seed := re.FindAllStringSubmatch(roundInfo, -1)
	seedSplit := strings.Split(seed[0][1], ",")
	v, err := strconv.Atoi(seedSplit[0])
	ErrPanic(err)
	windRound = v
	v, err = strconv.Atoi(seedSplit[1])
	ErrPanic(err)
	numHonba = v
	v, err = strconv.Atoi(seedSplit[2])
	ErrPanic(err)
	numRiichi = v
	v, err = strconv.Atoi(seedSplit[len(seedSplit)-1])
	doraIndicators = append(doraIndicators, v)

	re = regexp.MustCompile("ten=\"(.*?)\"")
	playerPoints := re.FindAllStringSubmatch(roundInfo, -1)
	playerPointsSplit := strings.Split(playerPoints[0][1], ",")
	points = StringSlices2IntSlices(playerPointsSplit)

	re = regexp.MustCompile("sc=\"(.*?)\"")
	playerRewards := re.FindAllStringSubmatch(roundInfo, -1)
	playerRewardsSplit := strings.Split(playerRewards[0][1], ",")
	rewards = StringSlices2IntSlices(playerRewardsSplit)

	return windRound, numHonba, numRiichi, doraIndicators, points, rewards
}

func compileGameInfo(gameInfo string) (gameType int, startOya int, dan []int) {
	re := regexp.MustCompile("type=\"(.*?)\"")
	s := re.FindAllStringSubmatch(gameInfo, -1)
	g, err := strconv.Atoi(s[0][1])
	ErrPanic(err)
	gameType = g

	re = regexp.MustCompile("oya=\"(.*?)\"")
	s = re.FindAllStringSubmatch(gameInfo, -1)
	o, err := strconv.Atoi(s[0][1])
	ErrPanic(err)
	startOya = o

	re = regexp.MustCompile("dan=\"(.*?)\"")
	s = re.FindAllStringSubmatch(gameInfo, -1)
	dan = StringSlices2IntSlices(strings.Split(s[0][1], ","))
	return gameType, startOya, dan
}

func compileInitInfo(logs string) (allInfo []string) {
	re := regexp.MustCompile("<INIT ")
	allInfo = re.Split(logs, -1)
	return allInfo
}

func compileFinalReward(logs string) (rewards []float32) {
	re := regexp.MustCompile("owari=\"(.*?)\"")
	playerRewards := re.FindAllStringSubmatch(logs, -1)
	playerRewardsSplit := strings.Split(playerRewards[0][1], ",")
	rewards = StringSlices2StringSlices(playerRewardsSplit)
	return rewards
}
