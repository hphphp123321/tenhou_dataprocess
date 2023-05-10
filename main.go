package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/hphphp123312/mahjong-datapreprocess/mahjong"
	"github.com/hphphp123312/mahjong-datapreprocess/tenhou"
	"github.com/hphphp123321/go-common"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
	"time"
)

type dataSet struct {
	ID     int    `gorm:"column:ID;primaryKey;autoIncrement;type:int"`
	Data   []byte `gorm:"column:Data;type:bytes"`
	MaxDan int    `gorm:"column:MaxDan;type:int"`
	MinDan int    `gorm:"column:MinDan;type:int"`
}

var callTypeMap = map[int]string{
	0: "Skip",
	1: "Discard",
	2: "Chi",
	3: "Pon",
	4: "DaiMinKan",
	5: "ShouMinKan",
	6: "AnKan",
	7: "Riichi",
}

var datasetsSize = map[string]int{
	"Discard":    30000000,
	"Riichi":     15000000,
	"Chi":        15000000,
	"Pon":        15000000,
	"DaiMinKan":  3000000,
	"ShouMinKan": 3000000,
	"AnKan":      3000000,
	"Skip":       20000000,
}

var datasetsNum = map[string]int{
	"Discard":    0,
	"Riichi":     0,
	"Chi":        0,
	"Pon":        0,
	"DaiMinKan":  0,
	"ShouMinKan": 0,
	"AnKan":      0,
	"Skip":       0,
}

func tenhouGo(bss [][]byte, dbDst *gorm.DB, goNum int) {
	//shanten_test.ExampleCalculate()
	var wg sync.WaitGroup
	var lock sync.Mutex
	var channels = make(chan int, goNum)
	allLen := len(bss)
	var counter = 0

	for idx, bs := range bss {
		wg.Add(1)
		channels <- idx
		go func(b []byte) {
			defer wg.Done()
			log := tenhou.DecompressBzipBytes(b)
			boardStates, dan := tenhou.ProcessGame(log)
			if len(boardStates) == 0 {
				<-channels
				return
			}
			maxDan := common.MaxNum(dan)
			minDan := common.MinNum(dan)
			if minDan <= 15 {
				<-channels
				return
			}
			var dataTypeMap = map[string][]*dataSet{
				"Skip":       {},
				"Discard":    {},
				"Chi":        {},
				"Pon":        {},
				"DaiMinKan":  {},
				"ShouMinKan": {},
				"AnKan":      {},
				"Riichi":     {},
			}

			for _, boardState := range boardStates {
				var player mahjong.PlayerState
				switch boardState.PlayerWind {
				case 0:
					player = boardState.P0
				case 1:
					player = boardState.P1
				case 2:
					player = boardState.P2
				case 3:
					player = boardState.P3
				}
				if player.PointsReward < 0 {
					continue
				}
				callType := callTypeMap[int(boardState.ValidActions[boardState.RealActionIdx].CallType)]
				lock.Lock()
				flag := false
				if callType == "Discard" {
					if datasetsNum["Discard"]%2 == 1 && player.PointsReward <= 8000 {
						flag = true
					}
					if minDan >= 17 {
						flag = false
					}
					if player.PointsReward == 0 && player.FinalReward < 0 {
						flag = true
					}
				} else if callType == "Skip" {
					for _, validAction := range boardState.ValidActions {
						if (validAction.CallType == mahjong.DaiMinKan || validAction.CallType == mahjong.ShouMinKan || validAction.CallType == mahjong.AnKan) && datasetsNum["Skip"]%5 == 1 {
							flag = true
						}
					}
				}
				if datasetsNum[callType] >= datasetsSize[callType] {
					flag = true
				}
				if flag {
					lock.Unlock()
					continue
				}
				lock.Unlock()
				boardBytes, err := json.Marshal(boardState)
				if err != nil {
					panic(err)
				}
				buf := &bytes.Buffer{}
				w := gzip.NewWriter(buf)
				_, err = w.Write(boardBytes)
				if err != nil {
					panic(err)
				}
				err = w.Close()
				if err != nil {
					panic(err)
				}

				compressBytes := buf.Bytes()
				data := &dataSet{
					Data:   compressBytes,
					MaxDan: maxDan,
					MinDan: minDan,
				}
				dataTypeMap[callType] = append(dataTypeMap[callType], data)
			}

			lock.Lock()
			for k, v := range dataTypeMap {
				if len(v) == 0 {
					continue
				}
				dbDst.Table(k).Select("Data", "MaxDan", "MinDan").Create(v)
			}

			for k, v := range dataTypeMap {
				datasetsNum[k] += len(v)
			}
			counter++
			if (counter+1)%1000 == 0 {
				fmt.Printf("file %v/%v done! Dataset num: %v\n", counter+1, allLen, datasetsNum)
			}
			lock.Unlock()
			<-channels
		}(bs)
	}
	wg.Wait()
	close(channels)
}

func main() {
	dbDst, err := gorm.Open(sqlite.Open("dst/datasets.db"), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	err = dbDst.Table("Skip").AutoMigrate(&dataSet{})
	err = dbDst.Table("Discard").AutoMigrate(&dataSet{})
	err = dbDst.Table("Chi").AutoMigrate(&dataSet{})
	err = dbDst.Table("Pon").AutoMigrate(&dataSet{})
	err = dbDst.Table("DaiMinKan").AutoMigrate(&dataSet{})
	err = dbDst.Table("ShouMinKan").AutoMigrate(&dataSet{})
	err = dbDst.Table("AnKan").AutoMigrate(&dataSet{})
	err = dbDst.Table("Riichi").AutoMigrate(&dataSet{})
	if err != nil {
		panic(err)
	}
	for _, dbName := range []string{
		//"src/2021.db",
		//"src/2020.db",
		//"src/2019.db",
		//"src/2018.db",
		//"src/2017.db",
		//"src/2016.db",
		//"src/2015.db",
		//"src/2014.db",
		//"src/2013.db",
		//"src/2012.db",
		//"src/2011.db",
		"src/2010.db",
	} {
		db, _ := gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		var bss [][]byte
		db.Table("logs").Where("is_tonpusen=0 AND is_sanma=0 AND is_processed=1 AND was_error=0").Select("log_content").Find(&bss)
		var a int64
		for k := range datasetsNum {
			dbDst.Table(k).Count(&a)
			datasetsNum[k] = int(a)
		}
		t := time.Now()
		//tenhouGo(bss, dbDst, 32)
		tenhouProcess(bss, dbDst)
		fmt.Printf(dbName+" Done! "+"used time: %v\n", time.Since(t))
	}
}

func tenhouProcess(bss [][]byte, dbDst *gorm.DB) {
	for i, bs := range bss {
		log := tenhou.DecompressBzipBytes(bs)
		boardStates, dan := tenhou.ProcessGame(log)
		maxDan := common.MaxNum(dan)
		minDan := common.MinNum(dan)
		var dataTypeMap = map[string][]*dataSet{
			"Skip":       []*dataSet{},
			"Discard":    []*dataSet{},
			"Chi":        []*dataSet{},
			"Pon":        []*dataSet{},
			"DaiMinKan":  []*dataSet{},
			"ShouMinKan": []*dataSet{},
			"AnKan":      []*dataSet{},
			"Riichi":     []*dataSet{},
		}
		for _, boardState := range boardStates {
			callType := callTypeMap[int(boardState.ValidActions[boardState.RealActionIdx].CallType)]
			json.Marshal(boardState)
			boardBytes, _ := json.Marshal(boardState)
			buf := &bytes.Buffer{}
			w := gzip.NewWriter(buf)
			w.Write(boardBytes)
			w.Close()

			compressBytes := buf.Bytes()
			data := &dataSet{
				Data:   compressBytes,
				MaxDan: maxDan,
				MinDan: minDan,
			}
			dataTypeMap[callType] = append(dataTypeMap[callType], data)
		}
		for k, v := range dataTypeMap {
			if len(v) == 0 {
				continue
			}
			dbDst.Table(k).Select("Data", "MaxDan", "MinDan").Create(v)
		}
		if (i+1)%100 == 0 {
			fmt.Printf("file %v/%v done!\n", i+1, i)
		}
	}
}
