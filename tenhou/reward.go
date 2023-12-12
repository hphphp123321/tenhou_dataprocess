package tenhou

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

type GlobalReward struct {
	NumGame   int   `json:"num_game"`
	WindRound int   `json:"wind_round"`
	NumHonba  int   `json:"num_honba"`
	NumRiiChi int   `json:"num_riichi"`
	Dealer    int   `json:"dealer"`
	Points    []int `json:"points"`
	FinalRank []int `json:"final_rank"`
}

func ProcessReward(logs string) (globalReward []*GlobalReward, dan []int) {
	globalReward = []*GlobalReward{}
	defer func() {
		if err := recover(); err != nil { // 如果recover返回为空，说明没报错
			fmt.Println(err)
			globalReward = []*GlobalReward{}
		}
	}()

	allInfo := compileInitInfo(logs)
	gameInfo := allInfo[0]
	gameType, startOya, dan := compileGameInfo(gameInfo)
	if gameType != 169 || startOya != 0 {
		return
	}

	// round info
	for idx := 1; idx < len(allInfo); idx++ {
		roundInfo := allInfo[idx]
		windRound, numHonba, numRiichi, _, points, rewards := compileRoundInfo(roundInfo)
		oya := compileOyaInfo(roundInfo)
		if windRound%4 != oya {
			panic("oya error")
		}
		var rReward = &GlobalReward{
			NumGame:   idx,
			WindRound: windRound,
			NumHonba:  numHonba,
			NumRiiChi: numRiichi,
			Points:    points,
			Dealer:    windRound % 4,
			FinalRank: GetRanksByRewards(rewards),
		}
		globalReward = append(globalReward, rReward)
	}
	return globalReward, dan
}

func compileOyaInfo(roundInfo string) (oya int) {
	re := regexp.MustCompile("oya=\"(.*?)\"")
	s := re.FindAllStringSubmatch(roundInfo, -1)
	oya, err := strconv.Atoi(s[0][1])
	ErrPanic(err)
	return oya
}

func GetRanksByRewards(rewards []int) []int {
	nums := make([]int, 4)
	for idx, reward := range rewards {
		if idx%2 == 0 {
			nums[idx/2] = reward
		}
	}

	// 复制原始数组
	original := make([]int, len(nums))
	copy(original, nums)

	// 对复制的数组进行排序
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] > nums[j] // 降序排序
	})

	// 创建排名映射
	rankMap := make(map[int]int)
	for rank, num := range nums {
		rankMap[num] = rank
	}

	// 生成输出数组
	ranks := make([]int, len(nums))
	for i, num := range original {
		ranks[i] = rankMap[num]
	}

	return ranks
}
