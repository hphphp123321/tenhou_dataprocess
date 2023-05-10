package mahjong

import (
	"github.com/dnovikoff/tempai-core/compact"
	"github.com/dnovikoff/tempai-core/hand/calc"
	"github.com/dnovikoff/tempai-core/hand/shanten"
	"github.com/dnovikoff/tempai-core/hand/tempai"
	"github.com/dnovikoff/tempai-core/tile"
)

var YaoKyuTiles = [...]int{0, 8, 9, 17, 18, 26, 27, 28, 29, 30, 31, 32, 33}

func IntToInstance(t int) tile.Instance {
	return tile.Instance(t + 1)
}

func IntsToInstances(tiles Tiles) tile.Instances {
	instances := tile.Instances{}
	for _, num := range tiles {
		instances = append(instances, IntToInstance(num))
	}
	return instances
}

func TilesCallsToCalc(tiles Tiles, calls Calls) (compact.Instances, calc.Option) {
	hand := IntsToInstances(tiles)
	instances := compact.NewInstances()
	instances.Add(hand)

	var meldsOpt calc.Option = nil
	if calls != nil {
		meldsT := CallsToMelds(calls)
		meldsOpt = calc.Declared(meldsT)
	}
	return instances, meldsOpt
}

func CalculateShantenNum(handTiles Tiles, melds Calls) int {
	instances, meldsOpt := TilesCallsToCalc(handTiles, melds)
	res := shanten.Calculate(instances, meldsOpt)
	return res.Total.Value
}

func GetTenhaiSlice(handTiles Tiles, melds Calls) []int {
	var tenhaiSlice []int

	instances, meldsOpt := TilesCallsToCalc(handTiles, melds)
	res := tempai.Calculate(instances, meldsOpt)
	tiles := tempai.GetWaits(res).Tiles()
	for _, t := range tiles {
		tenhaiSlice = append(tenhaiSlice, int(t)-1)
	}
	return tenhaiSlice
}

//func Contain(obj interface{}, target interface{}) bool {
//	targetValue := reflect.ValueOf(target)
//	switch reflect.TypeOf(target).Kind() {
//	case reflect.Slice, reflect.Array:
//		for i := 0; i < targetValue.Len(); i++ {
//			if targetValue.Index(i).Interface() == obj {
//				return true
//			}
//		}
//	case reflect.Map:
//		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
//			return true
//		}
//	}
//
//	return false
//}
