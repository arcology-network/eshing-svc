package workers

import (
	"fmt"

	ethCommon "github.com/arcology/3rd-party/eth/common"
	"github.com/arcology/common-lib/merkle"
	"github.com/arcology/common-lib/types"
	"github.com/arcology/component-lib/actor"
	"github.com/arcology/component-lib/log"
	urlcommon "github.com/arcology/concurrenturl/v2/common"
	"go.uber.org/zap"
)

type HpmtWorker struct {
	actor.WorkerThread
	acctBaseLength int
}

var (
	nilHash = ethCommon.Hash{}
)

//return a Subscriber struct
func NewHpmtWorker(concurrency int, groupid string) *HpmtWorker {
	hw := HpmtWorker{}
	hw.Set(concurrency, groupid)
	return &hw
}

func (hw *HpmtWorker) OnStart() {
	hw.acctBaseLength = len(urlcommon.NewPlatform().Eth10Account())
}

func (hw *HpmtWorker) Stop() {
}

func (hw *HpmtWorker) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgExecuted:
			euresults := v.Data.(*[]*types.EuResult)
			hw.AddLog(log.LogLevel_Info, "received euresults", zap.Int("size", len(*euresults)))
			roothash := hw.calculateRoothash(*euresults)

			if v.Height == 0 {
				//hash is created with genesis data,it is not nessary for make block
				pareninfo := types.ParentInfo{
					ParentHash: ethCommon.Hash{},
					ParentRoot: roothash,
				}
				hw.AddLog(log.LogLevel_Info, "getRoothash for Parentinfo", zap.String("roothash", fmt.Sprintf("%x", roothash)))
				hw.MsgBroker.Send(actor.MsgInitParentInfo, &pareninfo)
				hw.MsgBroker.Send(actor.MsgClearCompletedEuresults, "true")
			} else {
				hw.MsgBroker.Send(actor.MsgAcctHash, roothash)
			}
		}
	}

	return nil
}

func (hw *HpmtWorker) calculateRoothash(euresults []*types.EuResult) ethCommon.Hash {
	accts := []string{}
	datas := map[string][][]byte{}
	acctLength := hw.acctBaseLength + ethCommon.AddressLength*2
	for _, euresult := range euresults {
		hw.AddLog(log.LogLevel_Info, "calculateRoothash", zap.String("euresult", euresult.H), zap.Int("size", len(euresult.Transitions)))
		for _, transition := range euresult.Transitions {
			key := transition.GetPath()
			if len(key) < acctLength {
				continue
			} else {
				key = key[:acctLength]
			}
			if data, ok := datas[key]; ok {
				data = append(data, transition.Encode())
				datas[key] = data
			} else {
				accts = append(accts, key)
				datas[key] = [][]byte{transition.Encode()}
			}
		}
	}
	if len(accts) == 0 {
		return nilHash
	}
	rootDatas := make([][]byte, len(accts))
	for i, key := range accts {
		data := datas[key]
		in := merkle.NewMerkle(len(data), merkle.Sha256)
		in.Init(data)
		rootDatas[i] = in.GetRoot()

	}

	all := merkle.NewMerkle(len(rootDatas), merkle.Sha256)
	all.Init(rootDatas)
	root := all.GetRoot()

	return ethCommon.BytesToHash(root)
}
