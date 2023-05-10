package mahjong

type CallType int

const (
	Get CallType = -1 + iota
	Skip
	Discard
	Chi
	Pon
	DaiMinKan
	ShouMinKan
	AnKan
	Riichi
	Ron
	Tsumo
	KyuShuKyuHai
	ChanKan
)

type Calls []*Call

type Call struct {
	CallType         CallType `json:"type"`
	CallTiles        Tiles    `json:"tiles"`
	CallTilesFromWho []int    `json:"who"`
}

func NewCall(meldType CallType, CallTiles Tiles, CallTilesFromWho []int) *Call {
	return &Call{
		CallType:         meldType,
		CallTiles:        CallTiles,
		CallTilesFromWho: CallTilesFromWho,
	}
}

func CallEqual(call1 *Call, call2 *Call) bool {
	if call1.CallType != call2.CallType {
		return false
	}
	if TilesEqual(call1.CallTiles, call2.CallTiles) {
		return true
	}
	return false
}

func (call *Call) Copy() *Call {
	tilesFromWho := make([]int, len(call.CallTilesFromWho))
	copy(tilesFromWho, call.CallTilesFromWho)
	return &Call{
		CallType:         call.CallType,
		CallTiles:        call.CallTiles.Copy(),
		CallTilesFromWho: tilesFromWho,
	}
}

func (calls *Calls) Copy() Calls {
	callsCopy := make(Calls, len(*calls), cap(*calls))
	copy(callsCopy, *calls)
	return callsCopy
}

func (calls *Calls) Append(call *Call) {
	*calls = append(*calls, call)
}

func (calls *Calls) Remove(call *Call) {
	idx := calls.Index(call)
	*calls = append((*calls)[:idx], (*calls)[idx+1:]...)
}

func (calls *Calls) Index(call *Call) int {
	for idx, c := range *calls {
		if CallEqual(c, call) {
			return idx
		}
	}
	panic("call not in calls!")
}
