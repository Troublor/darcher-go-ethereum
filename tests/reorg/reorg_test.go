package reorg

import (
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"testing"
)

func TestMakeSure(t *testing.T) {
	db, blockchain, err := core.NewCanonical(ethash.NewFaker(), 0, true)
	if err != nil {
		t.Fatalf("failed to create pristine chain: %v", err)
	}
	defer blockchain.Stop()
	first := make([]int64, 10)
	for i := 0; i < len(first); i++ {
		first[i] = 10
	}
	blocks, _ := core.GenerateChain(params.TestChainConfig, blockchain.CurrentBlock(), ethash.NewFaker(), db, len(first), func(i int, b *core.BlockGen) {
		b.OffsetTime(first[i])
	})
	_, _ = blockchain.InsertChain(blocks)
	currentTd := blockchain.GetTdByHash(blockchain.CurrentBlock().Hash())
	td := new(big.Int)
	for _, block := range blocks {
		td.Add(td, block.Difficulty())
	}
	//want := new(big.Int).Add(blockchain.genesisBlock.Difficulty(), big.NewInt(12615120))
	if td.Cmp(currentTd) != 0 {
		t.Fatal("not sum")
	}
	//if have.Cmp(want) != 0 {
	//	t.Fatal("td not match")
	//}
}

//func TestFakeMinedBlockInsert(t *testing.T) {
//	var (
//		engine      consensus.Engine
//		chainConfig *params.ChainConfig
//		db          = rawdb.NewMemoryDatabase()
//	)
//
//	chainConfig = params.AllEthashProtocolChanges
//	engine = ethash.NewFaker()
//
//
//	w, b := newTestWorker(t, chainConfig, engine, db, 0)
//	defer w.Close()
//
//	db2 := rawdb.NewMemoryDatabase()
//	b.genesis.MustCommit(db2)
//	chain, _ := core.NewBlockChain(db2, nil, b.chain.Config(), engine, vm.Config{}, nil)
//	defer chain.Stop()
//
//	loopErr := make(chan error)
//	newBlock := make(chan struct{})
//	listenNewBlock := func() {
//		sub := w.GetMux().Subscribe(core.NewMinedBlockEvent{})
//		defer sub.Unsubscribe()
//
//		for item := range sub.Chan() {
//			block := item.Data.(core.NewMinedBlockEvent).Block
//			_, err := chain.InsertChain([]*types.Block{block})
//			if err != nil {
//				loopErr <- fmt.Errorf("failed to insert new mined block:%d, error:%v", block.NumberU64(), err)
//			}
//			newBlock <- struct{}{}
//		}
//	}
//	// Ignore empty commit here for less noise
//	w.skipSealHook = func(task *task) bool {
//		return len(task.receipts) == 0
//	}
//	w.start() // Start mining!
//	go listenNewBlock()
//
//	for i := 0; i < 5; i++ {
//		b.txPool.AddLocal(b.newRandomTx(true))
//		b.txPool.AddLocal(b.newRandomTx(false))
//		w.postSideBlock(core.ChainSideEvent{Block: b.newRandomUncle()})
//		w.postSideBlock(core.ChainSideEvent{Block: b.newRandomUncle()})
//		select {
//		case e := <-loopErr:
//			t.Fatal(e)
//		case <-newBlock:
//		case <-time.NewTimer(3 * time.Second).C: // Worker needs 1s to include new changes.
//			t.Fatalf("timeout")
//		}
//	}
//}
