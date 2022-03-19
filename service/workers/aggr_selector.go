package workers

import (
	"time"

	ethcmn "github.com/arcology-network/3rd-party/eth/common"
	cmncmn "github.com/arcology-network/common-lib/common"
	cmntyp "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/aggregator/aggregator"
)

const (
	aggrStateInit = iota
	aggrStateCollecting
	aggrStateDone
)

type AggrSelector struct {
	actor.WorkerThread

	aggregator *aggregator.Aggregator
	state      int
	height     uint64
}

func NewAggrSelector(concurrency int, groupId string) *AggrSelector {
	aggr := &AggrSelector{
		aggregator: aggregator.NewAggregator(),
		state:      aggrStateDone,
	}
	aggr.Set(concurrency, groupId)
	return aggr
}

func (aggr *AggrSelector) OnStart() {}

func (aggr *AggrSelector) OnMessageArrived(msgs []*actor.Message) error {
	msg := msgs[0]
	switch aggr.state {
	case aggrStateInit:
		if msg.Name == actor.MsgEuResults {
			aggr.onDataReceived(msg)
		} else if msg.Name == actor.MsgInclusive {
			collectStart = time.Now()
			if aggr.onListReceived(msg) {
				aggr.state = aggrStateDone
			} else {
				aggr.state = aggrStateCollecting
			}
		}
	case aggrStateCollecting:
		if msg.Name == actor.MsgEuResults {
			if aggr.onDataReceived(msg) {
				aggr.state = aggrStateDone
			}
		}
	case aggrStateDone:
		if msg.Name == actor.MsgBlockCompleted {
			aggr.aggregator.OnClearInfoReceived()
			aggr.state = aggrStateInit
			aggr.height = msg.Height + 1
		}
	}
	return nil
}

func (aggr *AggrSelector) GetStateDefinitions() map[int][]string {
	return map[int][]string{
		aggrStateInit:       {actor.MsgEuResults, actor.MsgInclusive},
		aggrStateCollecting: {actor.MsgEuResults},
		aggrStateDone:       {actor.MsgBlockCompleted},
	}
}

func (aggr *AggrSelector) GetCurrentState() int {
	return aggr.state
}

func (aggr *AggrSelector) Height() uint64 {
	return aggr.height
}

func (aggr *AggrSelector) onDataReceived(msg *actor.Message) bool {
	data := msg.Data.(*cmntyp.Euresults)
	if data != nil && len(*data) > 0 {
		for _, euResult := range *data {
			list := aggr.aggregator.OnDataReceived(
				cmncmn.ToNewHash(
					ethcmn.BytesToHash([]byte(euResult.H)),
					msg.Height,
					msg.Round),
				euResult,
			)
			if aggr.sendIfDone(list) {
				return true
			}
		}
	}
	return false
}

func (aggr *AggrSelector) onListReceived(msg *actor.Message) bool {
	inclusiveList := msg.Data.(*cmntyp.InclusiveList)
	inclusiveList.Mode = cmntyp.InclusiveMode_Results
	for i := range inclusiveList.HashList {
		newHash := cmncmn.ToNewHash(*inclusiveList.HashList[i], msg.Height, msg.Round)
		inclusiveList.HashList[i] = &newHash
	}
	list, _ := aggr.aggregator.OnListReceived(inclusiveList)
	return aggr.sendIfDone(list)
}

func (aggr *AggrSelector) sendIfDone(list *[]*interface{}) bool {
	if list != nil {
		euResults := make([]*cmntyp.EuResult, len(*list))
		for i, euResult := range *list {
			euResults[i] = (*euResult).(*cmntyp.EuResult)
		}

		calcStart = time.Now()
		CollectTime.Observe(calcStart.Sub(collectStart).Seconds())
		CollectTimeGauge.Set(calcStart.Sub(collectStart).Seconds())
		aggr.MsgBroker.Send(actor.MsgExecuted, &euResults)
		return true
	}
	return false
}
