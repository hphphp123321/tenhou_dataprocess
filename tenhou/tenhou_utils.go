package tenhou

import (
	"github.com/hphphp123312/mahjong-datapreprocess/mahjong"
	"strconv"
)

var getTileMap = map[string]int{"T": 0, "U": 1, "V": 2, "W": 3}
var discardTileMap = map[string]int{"D": 0, "E": 1, "F": 2, "G": 3}

func ProcessMeld(bytes int, who int) *mahjong.Call {
	var meldType mahjong.CallType
	var tenhouMeldTiles mahjong.Tiles
	var tilesFromWho []int
	var tenhouCalledTile int
	var call = new(mahjong.Call)
	offsetFromWho := bytes & 0x3
	fromWho := (who + offsetFromWho) % 4
	if bytes&0x4 > 0 {
		// chi
		t0, t1, t2 := (bytes>>3)&0x3, (bytes>>5)&0x3, (bytes>>7)&0x3
		baseAndCalled := bytes >> 10
		base, called := baseAndCalled/3, baseAndCalled%3
		base = (base/7)*9 + (base % 7)
		tenhouMeldTiles = mahjong.Tiles{t0 + 4*base, t1 + 4*(base+1), t2 + 4*(base+2)}
		tenhouCalledTile = tenhouMeldTiles[called]
		meldType = mahjong.Chi
		tilesFromWho = []int{who, who, fromWho, -1}
	} else if bytes&0x18 > 0 {
		// pon and addKan
		t4 := (bytes >> 5) & 0x3
		t := [][]int{{1, 2, 3}, {0, 2, 3}, {0, 1, 3}, {0, 1, 2}}[t4]
		t0, t1, t2 := t[0], t[1], t[2]
		baseAndCalled := bytes >> 9
		base, called := baseAndCalled/3, baseAndCalled%3
		if bytes&0x8 > 0 {
			// pon
			meldType = mahjong.Pon
			tenhouMeldTiles = mahjong.Tiles{t0 + 4*base, t1 + 4*base, t2 + 4*base}
			tenhouCalledTile = tenhouMeldTiles[called]
			tilesFromWho = []int{who, who, fromWho, -1}
		} else {
			// addKan
			meldType = mahjong.ShouMinKan
			tenhouMeldTiles = mahjong.Tiles{t0 + 4*base, t1 + 4*base, t2 + 4*base, t4 + 4*base}
			tenhouCalledTile = tenhouMeldTiles[3]
			tilesFromWho = []int{who, who, fromWho, who}
		}
	} else if bytes&0x20 > 0 {
		panic("拔北")
	} else {
		// daiMinKan anKan
		baseAndCalled := bytes >> 8
		base, called := baseAndCalled/4, baseAndCalled%4
		tenhouMeldTiles = mahjong.Tiles{4 * base, 4*base + 1, 4*base + 2, 4*base + 3}
		tenhouCalledTile = tenhouMeldTiles[called]
		if offsetFromWho == 0 {
			// anKan
			meldType = mahjong.AnKan
			tilesFromWho = []int{who, who, who, who}
		} else {
			// daiMinKan
			meldType = mahjong.DaiMinKan
			tilesFromWho = []int{who, who, who, fromWho}
		}
	}
	if meldType != mahjong.AnKan {
		tenhouMeldTiles.Remove(tenhouCalledTile)
		tenhouMeldTiles.Append(tenhouCalledTile)
	}
	if len(tenhouMeldTiles) == 3 {
		tenhouMeldTiles.Append(-1)
	}
	call.CallType = meldType
	call.CallTiles = tenhouMeldTiles
	call.CallTilesFromWho = tilesFromWho
	return call
}

func ErrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func StringSlices2IntSlices(string []string) (ints []int) {
	for _, s := range string {
		i, err := strconv.Atoi(s)
		ErrPanic(err)
		ints = append(ints, i)
	}
	return ints
}

func StringSlices2StringSlices(string []string) (floats []float32) {
	for _, s := range string {
		i, err := strconv.ParseFloat(s, 32)
		s32 := float32(i)
		ErrPanic(err)
		floats = append(floats, s32)
	}
	return floats
}

func GetOtherWinds(wind int) []int {
	otherWinds := []int{0, 1, 2, 3}
	for i, v := range otherWinds {
		if v == wind {
			otherWinds = append(otherWinds[:i], otherWinds[i+1:]...)
			break
		}
	}
	return otherWinds
}
