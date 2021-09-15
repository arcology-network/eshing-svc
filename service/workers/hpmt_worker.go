package workers

import (
	"fmt"
	"time"

	ethCommon "github.com/arcology-network/3rd-party/eth/common"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/merkle"
	"github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"github.com/arcology-network/component-lib/storage"
	ccurl "github.com/arcology-network/concurrenturl/v2"
	urlcommon "github.com/arcology-network/concurrenturl/v2/common"
	ccurltype "github.com/arcology-network/concurrenturl/v2/type"
	"github.com/arcology-network/concurrenturl/v2/type/commutative"
	"go.uber.org/zap"
)

type HpmtWorker struct {
	actor.WorkerThread
	DB        urlcommon.DB
	savedRoot ethCommon.Hash
	tmpRoot   ethCommon.Hash

	url    *ccurl.ConcurrentUrl
	merkle *ccurltype.AccountMerkle

	executed *[]*types.EuResult
	paths    []string
	states   interface{}
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

	//hw.acctBaseLength = len(urlcommon.NewPlatform().Eth10Account())

	platform := urlcommon.NewPlatform()
	persistentDB := urlcommon.NewDataStore()
	meta, _ := commutative.NewMeta(platform.Eth10Account())
	persistentDB.Save(platform.Eth10Account(), meta)
	hw.DB = persistentDB

	hw.savedRoot = nilHash
	hw.tmpRoot = nilHash

	hw.url = ccurl.NewConcurrentUrl(hw.DB)
	hw.merkle = ccurltype.NewAccountMerkle(platform)

	hw.paths = []string{}

}

func (hw *HpmtWorker) Stop() {
}

func (hw *HpmtWorker) OnMessageArrived(msgs []*actor.Message) error {
	for _, v := range msgs {
		switch v.Name {
		case actor.MsgEuResults:
			data := msgs[0].Data.(*types.Euresults)
			hw.AddLog(log.LogLevel_Info, "import euresults", zap.Int("size", len(*data)))
			hw.importEuresults(*data)
		case actor.MsgBlockCompleted:
			result := v.Data.(string)
			if actor.MsgBlockCompleted_Success == result {
				hw.savedRoot = hw.tmpRoot
				if len(hw.paths) > 0 {
					hw.save(hw.paths, hw.states)
					hw.AddLog(log.LogLevel_Debug, "save states", zap.Int("size", len(*hw.executed)))
				}
			}
			hw.executed = &[]*types.EuResult{}
			hw.merkle = ccurltype.NewAccountMerkle(urlcommon.NewPlatform())
		case actor.MsgExecuted:
			hw.executed = v.Data.(*[]*types.EuResult)
			//hw.MsgBroker.Send(actor.MsgGc, "true")
			hw.AddLog(log.LogLevel_Info, "received euresults", zap.Int("size", len(*hw.executed)))

			if v.Height == 0 {
				hw.importEuresults(*hw.executed)
			}
			begintime := time.Now()
			roothash, paths, states := hw.getRoothashV3(*hw.executed)
			hw.AddLog(log.LogLevel_Info, "calculate eshing roothash", zap.Duration("time", time.Since(begintime)))
			hw.paths = paths
			hw.states = states

			if v.Height == 0 {
				//hash is created with genesis data,it is not nessary for make block
				pareninfo := types.ParentInfo{
					ParentHash: ethCommon.Hash{},
					ParentRoot: roothash,
				}
				hw.AddLog(log.LogLevel_Info, "getRoothash for Parentinfo", zap.String("roothash", fmt.Sprintf("%x", roothash)))
				hw.MsgBroker.Send(actor.MsgInitParentInfo, &pareninfo)
				hw.MsgBroker.Send(actor.MsgClearCompletedEuresults, "true")

				hw.savedRoot = hw.tmpRoot

				if len(hw.paths) > 0 {
					hw.save(hw.paths, hw.states)
					hw.AddLog(log.LogLevel_Debug, "save states", zap.Int("size", len(*hw.executed)))
				}
				hw.executed = &[]*types.EuResult{}
				hw.merkle = ccurltype.NewAccountMerkle(urlcommon.NewPlatform())
			} else {
				hw.MsgBroker.Send(actor.MsgAcctHash, roothash)
			}
		}
	}

	return nil
}

func (hw *HpmtWorker) importEuresults(euresults []*types.EuResult) {
	_, transitions := storage.GetTransitions(euresults)
	//go func() {
	hw.url.Indexer().Import(transitions)
	//}()
	hw.merkle.Import(transitions)
}

func (hw *HpmtWorker) getRoothashV3(euresults []*types.EuResult) (ethCommon.Hash, []string, interface{}) {
	if len(euresults) == 0 {
		return hw.savedRoot, []string{}, nil
	}
	tims := make([]time.Duration, 6)
	t := time.Now()
	txIds := storage.GetTransitionIds(euresults)
	tims[0] = time.Since(t)
	t = time.Now()
	paths, states, errs := hw.url.Indexer().Commit(txIds)
	tims[1] = time.Since(t)
	if errs != nil && len(errs) > 0 {
		hw.AddLog(log.LogLevel_Error, "index commit err", zap.Errors("err msg", errs))
		return nilHash, []string{}, nil
	}
	t = time.Now()
	sortedKeys := hw.merkle.Build(8, hw.url.Indexer().ByPath())
	//modify v1.1 merkle.Build + (*stateDict)[k.(string)+"/"]
	tims[2] = time.Since(t)
	t = time.Now()
	merkles := hw.merkle.GetMerkles()
	tims[3] = time.Since(t)
	t = time.Now()
	rootDatas := make([][]byte, len(sortedKeys))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if merkle, ok := merkles[sortedKeys[i]]; ok {
				rootDatas[i] = merkle.GetRoot()
			}
		}
	}
	common.ParallelWorker(len(sortedKeys), 6, worker)
	tims[4] = time.Since(t)
	t = time.Now()
	all := merkle.NewMerkle(len(rootDatas), merkle.Sha256)
	all.Init(rootDatas)
	root := all.GetRoot()
	hw.tmpRoot = ethCommon.BytesToHash(root)

	tims[5] = time.Since(t)
	hw.AddLog(log.LogLevel_Info, "merkle root completed", zap.Durations("tims", tims))
	return hw.tmpRoot, paths, states
}
func (hw *HpmtWorker) save(paths []string, states interface{}) {
	store := hw.url.Indexer().Store()
	(*store).BatchSave(paths, states) // Commit to the state store
	(*store).Clear()
	hw.url.Indexer().Clear()
}
