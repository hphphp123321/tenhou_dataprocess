package tenhou

import "testing"

func TestGetRanksByRewards(t *testing.T) {
	rewards := []int{30, 0, 20, 0, 40, 0, 10, 0}
	ranks := GetRanksByRewards(rewards)
	if ranks[0] != 1 || ranks[1] != 2 || ranks[2] != 0 || ranks[3] != 3 {
		t.Errorf("GetRanksByRewards failed")
	}
}
