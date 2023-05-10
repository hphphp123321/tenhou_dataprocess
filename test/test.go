package test

import (
	"fmt"
	"github.com/dnovikoff/tempai-core/compact"
	"github.com/dnovikoff/tempai-core/hand/calc"
	"github.com/dnovikoff/tempai-core/hand/shanten"
	"github.com/dnovikoff/tempai-core/hand/tempai"
	"github.com/dnovikoff/tempai-core/tile"
	"github.com/dnovikoff/tempai-core/yaku"
	"github.com/hphphp123312/mahjong-datapreprocess/mahjong"
)

func ExampleCalculate() {
	p1 := mahjong.NewMahjongPlayer()
	p1.ResetForGame()
	p1.HandTiles = mahjong.Tiles{14, 19, 21, 55, 57, 63, 64, 65, 88, 129, 131}
	p1.Melds = mahjong.Calls{mahjong.Call{
		CallType:         mahjong.AnKan,
		CallTiles:        mahjong.Tiles{80, 81, 82, 83},
		CallTilesFromWho: []int{3, 3, 3, 3},
	}}
	p1.GetRiichiTiles()

	generator := compact.NewTileGenerator()
	a := compact.NewInstances()
	handTiles := tile.Tiles{5, 6, 14, 15, 17, 17, 33, 33, 4, 16}
	hand := generator.Tiles(handTiles)
	a.Add(hand)

	declared := []calc.Meld{calc.Kan(tile.Tile(21))}
	melds := calc.Melds{}
	melds = append(melds, declared...)
	cal := calc.Declared(melds)

	res := shanten.Calculate(a, cal)

	results := tempai.Calculate(a, cal)

	fmt.Printf("Total shanten value is: %v\n", res.Total.Value)
	fmt.Printf("Waits are %s\n", tempai.GetWaits(results).Tiles())

	var tenhaiSlice []int
	tiles := tempai.GetWaits(results).Tiles()
	for _, tileID := range tiles {
		tenhaiSlice = append(tenhaiSlice, int(tileID))
	}
	//handTiles = tile.Tiles{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5}
	//hand = generator.Tiles(handTiles)
	//a = compact.NewInstances()
	//a.Add(hand)

	winTile := generator.Instance(tile.Tile(6))
	ctx := &yaku.Context{
		Tile:      winTile,
		Rules:     yaku.RulesTenhouRed(),
		IsTsumo:   true,
		IsChankan: false,
	}

	yakuResult := yaku.Win(results, ctx, nil)
	fmt.Printf("%v\n", yakuResult.String())

	// Output:
	// Hand is 3567m5677p268s77z
	// Regular shanten value is: 2
	// Pairs shanten value is: 4
	// Kokushi shanten value is: 11
	// Total shanten value is: 2
	// Total uke ire: 18/63
	// Hand improves: 123458m456789p12347s7z
}
