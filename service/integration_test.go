package service

import (
	"testing"
	"time"

	ethcmn "github.com/arcology-network/3rd-party/eth/common"
	cmntyp "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/component-lib/actor"
	"github.com/arcology-network/component-lib/log"
	"github.com/arcology-network/component-lib/mock/kafka"
)

func TestNew(t *testing.T) {
	log.InitLog("testlegacy.log", "./log.toml", "eshing", "node1", 0)

	cfg := NewConfigV2(
		4,
		"doesn't matter",
		"doesn't matter",
		"doesn't matter",
		"doesn't matter",
		"doesn't matter",
		"doesn't matter",
		kafka.NewDownloaderCreator(t),
		kafka.NewUploaderCreator(t),
	)
	cfg.Start()

	downloader1 := cfg.downloader1.(*kafka.Downloader)
	downloader2 := cfg.downloader2.(*kafka.Downloader)
	uploader := cfg.uploader.(*kafka.Uploader)

	downloader1.Receive(&actor.Message{
		Name: actor.MsgInitExecuted,
		// Data: &[]*cmntyp.EuResult{},
	})

	h1 := ethcmn.BytesToHash([]byte("h1"))
	h2 := ethcmn.BytesToHash([]byte("h2"))
	downloader2.Receive(&actor.Message{
		Name:   actor.MsgEuResults,
		Height: 1,
		Data: &cmntyp.Euresults{
			{
				H: string(h1.Bytes()),
			},
		},
	})
	downloader2.Receive(&actor.Message{
		Name:   actor.MsgEuResults,
		Height: 1,
		Data: &cmntyp.Euresults{
			{
				H: string(h2.Bytes()),
			},
		},
	})
	downloader1.Receive(&actor.Message{
		Name:   actor.MsgInclusive,
		Height: 1,
		Data: &cmntyp.InclusiveList{
			HashList:   []*ethcmn.Hash{&h1, &h2},
			Successful: []bool{true, true},
		},
	})
	downloader1.Receive(&actor.Message{
		Name:   actor.MsgBlockCompleted,
		Height: 1,
	})

	time.Sleep(time.Second * 3)
	t.Log(uploader.GetCounter())
}
