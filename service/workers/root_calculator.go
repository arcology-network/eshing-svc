package workers

import (
	"time"

	ethcmn "github.com/arcology-network/3rd-party/eth/common"
	cmncmn "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/mempool"
	cmnmkl "github.com/arcology-network/common-lib/merkle"
	"github.com/arcology-network/common-lib/mhasher"
	cmntyp "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/storage"
	urlcmn "github.com/arcology-network/concurrenturl/v2/common"
	urltyp "github.com/arcology-network/concurrenturl/v2/type"
)

type RootCalculator struct {
	storage.BasicDBOperation

	merkle   *urltyp.AccountMerkle
	lastRoot ethcmn.Hash
}

func NewRootCalculator() *RootCalculator {
	return &RootCalculator{
		merkle: urltyp.NewAccountMerkle(urlcmn.NewPlatform()),
	}
}

func (rc *RootCalculator) Import(transitions []urlcmn.UnivalueInterface) {
	cmncmn.ParallelExecute(
		func() {
			rc.BasicDBOperation.Import(transitions)
		},
		func() {
			rc.merkle.Import(transitions)
		})
}

func (rc *RootCalculator) PostImport(euResults []*cmntyp.EuResult, height uint64) {
	if height == 0 {
		_, transitions := storage.GetTransitions(euResults)
		rc.merkle.Import(transitions)
	}
	rc.BasicDBOperation.PostImport(euResults, height)
}

func (rc *RootCalculator) PreCommit(euResults []*cmntyp.EuResult, height uint64) {
	if len(euResults) == 0 {
		return
	}

	rc.BasicDBOperation.PreCommit(euResults, height)
	rc.lastRoot = rc.calculate()
}

func (rc *RootCalculator) PostCommit(euResults []*cmntyp.EuResult, height uint64) {
	rc.BasicDBOperation.PostCommit(euResults, height)
	if height == 0 {
		rc.MsgBroker.Send(actor.MsgInitParentInfo, &cmntyp.ParentInfo{
			ParentHash: ethcmn.Hash{},
			ParentRoot: rc.lastRoot,
		})
	} else {
		rc.MsgBroker.Send(actor.MsgAcctHash, rc.lastRoot)
		CalcTime.Observe(time.Since(calcStart).Seconds())
		CalcTimeGauge.Set(time.Since(calcStart).Seconds())
	}
}

func (rc *RootCalculator) Finalize() {
	rc.BasicDBOperation.Finalize()
	rc.merkle.Clear()
}

func (rc *RootCalculator) calculate() ethcmn.Hash {
	k, v := rc.URL.Indexer().KVs()
	rc.merkle.Build(k, v)

	merkles := rc.merkle.GetMerkles()
	if len(*merkles) == 0 {
		return rc.lastRoot
	}

	keys := make([]string, 0, len(*merkles))
	for p := range *merkles {
		keys = append(keys, p)
	}
	sortedKeys, err := mhasher.SortStrings(keys)
	if err != nil {
		panic(err)
	}

	rootDatas := make([][]byte, len(sortedKeys))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			if merkle, ok := (*merkles)[sortedKeys[i]]; ok {
				rootDatas[i] = merkle.GetRoot()
			}
		}
	}
	cmncmn.ParallelWorker(len(sortedKeys), 6, worker)

	all := cmnmkl.NewMerkle(len(rootDatas), cmnmkl.Sha256)
	nodePool := mempool.NewMempool("nodes", func() interface{} {
		return cmnmkl.NewNode()
	})
	all.Init(rootDatas, nodePool)
	root := all.GetRoot()
	return ethcmn.BytesToHash(root)
}
