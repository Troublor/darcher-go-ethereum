package miner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"testing"
	"time"
)

/*
{
	"linkReferences": {},
	"object": "6080604052348015600f57600080fd5b506001600081905550603e8060256000396000f3fe6080604052600080fdfea265627a7a723158204ee2aaa6d786d55ae5016c75ed8b5eb654f06ff7c68ac44f4872564b5d52c00764736f6c634300050c0032",
	"opcodes": "PUSH1 0x80 PUSH1 0x40 MSTORE CALLVALUE DUP1 ISZERO PUSH1 0xF JUMPI PUSH1 0x0 DUP1 REVERT JUMPDEST POP PUSH1 0x1 PUSH1 0x0 DUP2 SWAP1 SSTORE POP PUSH1 0x3E DUP1 PUSH1 0x25 PUSH1 0x0 CODECOPY PUSH1 0x0 RETURN INVALID PUSH1 0x80 PUSH1 0x40 MSTORE PUSH1 0x0 DUP1 REVERT INVALID LOG2 PUSH6 0x627A7A723158 KECCAK256 0x4e 0xe2 0xaa 0xa6 0xd7 DUP7 0xd5 GAS 0xe5 ADD PUSH13 0x75ED8B5EB654F06FF7C68AC44F 0x48 PUSH19 0x564B5D52C00764736F6C634300050C00320000 ",
	"sourceMap": "0:82:0:-;;;34:46;8:9:-1;5:2;;;30:1;27;20:12;5:2;34:46:0;72:1;65:4;:8;;;;0:82;;;;;;"
}
*/
var (
	simpleContractByteCode = "6080604052348015600f57600080fd5b506001600081905550603e8060256000396000f3fe6080604052600080fdfea265627a7a723158204ee2aaa6d786d55ae5016c75ed8b5eb654f06ff7c68ac44f4872564b5d52c00764736f6c634300050c0032"
)

var (
	myMinerConfig = &Config{
		Recommit: time.Second,
		GasFloor: params.GenesisGasLimit,
		GasCeil:  params.GenesisGasLimit,
	}
)
var (
	engine      consensus.Engine
	chainConfig *params.ChainConfig
)

var (
	fakeConfig = &Config{
		Recommit: time.Second,
		GasFloor: params.GenesisGasLimit,
		GasCeil:  params.GenesisGasLimit,
	}
	// Test accounts
	fakeBankKey, _  = crypto.GenerateKey()
	fakeBankAddress = crypto.PubkeyToAddress(testBankKey.PublicKey)
	fakeBankFunds   = big.NewInt(1000000000000000000)
)

func prepareBlock(chain *core.BlockChain, parent *types.Block, t *testing.T) {
	num := parent.Number()
	timestamp := time.Now().Unix()
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   core.CalcGasLimit(parent, fakeConfig.GasFloor, fakeConfig.GasCeil),
		Extra:      make([]byte, 0),
		Time:       uint64(timestamp),
		Coinbase:   fakeBankAddress,
	}
	if err := engine.Prepare(chain, header); err != nil {
		t.Error("Failed to prepare header for mining", "err", err)
		return
	}

}

type Worker struct {
	worker
}

func NewFakeWorker(config *Config, chainConfig *params.ChainConfig, engine consensus.Engine, eth Backend, mux *event.TypeMux, isLocalBlock func(*types.Block) bool, init bool) *Worker {
	worker := &worker{
		config:             config,
		chainConfig:        chainConfig,
		engine:             engine,
		eth:                eth,
		mux:                mux,
		chain:              eth.BlockChain(),
		isLocalBlock:       isLocalBlock,
		localUncles:        make(map[common.Hash]*types.Block),
		remoteUncles:       make(map[common.Hash]*types.Block),
		unconfirmed:        newUnconfirmedBlocks(eth.BlockChain(), miningLogAtDepth),
		pendingTasks:       make(map[common.Hash]*task),
		txsCh:              make(chan core.NewTxsEvent, txChanSize),
		chainHeadCh:        make(chan core.ChainHeadEvent, chainHeadChanSize),
		chainSideCh:        make(chan core.ChainSideEvent, chainSideChanSize),
		newWorkCh:          make(chan *newWorkReq),
		taskCh:             make(chan *task),
		resultCh:           make(chan *types.Block, resultQueueSize),
		exitCh:             make(chan struct{}),
		startCh:            make(chan struct{}, 1),
		resubmitIntervalCh: make(chan time.Duration),
		resubmitAdjustCh:   make(chan *intervalAdjust, resubmitAdjustChanSize),
	}
	// Subscribe NewTxsEvent for tx pool
	worker.txsSub = eth.TxPool().SubscribeNewTxsEvent(worker.txsCh)
	// Subscribe events for blockchain
	worker.chainHeadSub = eth.BlockChain().SubscribeChainHeadEvent(worker.chainHeadCh)
	worker.chainSideSub = eth.BlockChain().SubscribeChainSideEvent(worker.chainSideCh)

	// Sanitize recommit interval if the user-specified one is too short.
	recommit := worker.config.Recommit
	if recommit < minRecommitInterval {
		log.Warn("Sanitizing miner recommit interval", "provided", recommit, "updated", minRecommitInterval)
		recommit = minRecommitInterval
	}

	//go worker.mainLoop()
	//go worker.newWorkLoop(recommit)
	//go worker.resultLoop()
	//go worker.taskLoop()
	//
	//// Submit first work to initialize pending state.
	//if init {
	//	worker.startCh <- struct{}{}
	//}
	return &Worker{*worker}
}

func (w *Worker) packBlock(txs map[common.Address]types.Transactions, parent *types.Block, header *types.Header, t *testing.T) *types.Block {
	err := w.makeCurrent(parent, header)
	if err != nil {
		t.Error("Failed to create mining context", "err", err)
		return nil
	}

	// TODO possibly deal with uncles
	var uncles []*types.Header

	txsSorted := types.NewTransactionsByPriceAndNonce(w.current.signer, txs)
	if w.commitTransactions(txsSorted, w.coinbase, nil) {
		t.Error("commit transaction failed")
		return nil
	}
	receipts := make([]*types.Receipt, len(w.current.receipts))
	for i, l := range w.current.receipts {
		receipts[i] = new(types.Receipt)
		*receipts[i] = *l
	}
	s := w.current.state.Copy()
	block, err := w.engine.FinalizeAndAssemble(w.chain, w.current.header, s, w.current.txs, uncles, w.current.receipts)
	if err != nil {
		t.Error("finalize and assemble block failed")
		return nil
	}

	//createdAt := time.Now()

	resultCh := make(chan *types.Block)

	if err := w.engine.Seal(w.chain, block, resultCh, nil); err != nil {
		t.Error("Block sealing failed", "err", err)
		return nil
	}
	sealedBlock := <-resultCh
	if sealedBlock == nil {
		t.Error("sealed block is empty")
		return nil
	}
	return nil
}

func TestGenerateBlockAndImport(t *testing.T) {
	var (
		db = rawdb.NewMemoryDatabase()
	)

	chainConfig = params.AllEthashProtocolChanges
	engine = ethash.NewFaker()

	w, b := newTestWorker(t, chainConfig, engine, db, 0)
	defer w.close()

	db2 := rawdb.NewMemoryDatabase()
	b.genesis.MustCommit(db2)
	chain, _ := core.NewBlockChain(db2, nil, b.chain.Config(), engine, vm.Config{}, nil)
	defer chain.Stop()

	//loopErr := make(chan error)
	//newBlock := make(chan struct{})
	//listenNewBlock := func() {
	//	sub := w.mux.Subscribe(core.NewMinedBlockEvent{})
	//	defer sub.Unsubscribe()
	//
	//	for item := range sub.Chan() {
	//		block := item.Data.(core.NewMinedBlockEvent).Block
	//		_, err := chain.InsertChain([]*types.Block{block})
	//		state, _ := chain.State()
	//		receipts := chain.GetReceiptsByHash(block.Hash())
	//		fmt.Println(state, receipts)
	//		//val := state.GetState(contract.Address(), common.BigToHash(loc))
	//		if err != nil {
	//			loopErr <- fmt.Errorf("failed to insert new mined block:%d, error:%v", block.NumberU64(), err)
	//		}
	//		newBlock <- struct{}{}
	//	}
	//}

	sub := w.mux.Subscribe(core.NewMinedBlockEvent{})
	defer sub.Unsubscribe()

	tx, _ := types.SignTx(types.NewContractCreation(b.txPool.Nonce(testBankAddress), big.NewInt(0), testGas, nil, common.FromHex(simpleContractByteCode)), types.HomesteadSigner{}, testBankKey)
	b.txPool.AddLocal(tx)

	w.start() // Start mining!

	for item := range sub.Chan() {
		block := item.Data.(core.NewMinedBlockEvent).Block
		_, err := chain.InsertChain([]*types.Block{block})
		receipts := chain.GetReceiptsByHash(block.Hash())
		fmt.Println("mined block", block.Number(), "transactions", len(block.Transactions()))
		if len(receipts) > 1 {
			contractAddr := receipts[1].ContractAddress
			//fmt.Println(contractAddr)
			state, _ := chain.State()
			locZero := common.BigToHash(big.NewInt(0))
			val := state.GetState(contractAddr, locZero)
			fmt.Println(val)
			w.stop()
			break
		}
		//fmt.Println(b.txPool.Stats())

		//val := state.GetState(contract.Address(), common.BigToHash(loc))
		if err != nil {
			t.Errorf("failed to insert new mined block:%d, error:%v", block.NumberU64(), err)
		}
	}
}

type FakeNode struct {
	miner       *Miner
	newBlockSub *event.TypeMuxSubscription
	t           *testing.T
}

func NewFakeNode(t *testing.T, minerConfig *Config, chainConfig *params.ChainConfig, engine consensus.Engine, db ethdb.Database) *FakeNode {
	eth := newTestWorkerBackend(t, chainConfig, engine, db, 0)
	mux := new(event.TypeMux)
	miner := New(eth, minerConfig, chainConfig, mux, engine, nil)
	return &FakeNode{
		miner:       miner,
		newBlockSub: nil,
		t:           t,
	}
}

func (node *FakeNode) BlockChain() *core.BlockChain {
	return node.miner.eth.BlockChain()
}

func (node *FakeNode) TxPool() *core.TxPool {
	return node.miner.eth.TxPool()
}

func (node *FakeNode) AddTx(tx *types.Transaction, key *ecdsa.PrivateKey) error {
	signer := types.MakeSigner(node.BlockChain().Config(), node.BlockChain().CurrentBlock().Number())
	signedTx, err := types.SignTx(tx, signer, key)
	if err != nil {
		return err
	}
	err = node.miner.eth.TxPool().AddLocal(signedTx)
	if err != nil {
		return err
	}
	return nil
}

// Deprecated
func (node *FakeNode) MineOneBlock(coinbase common.Address) (*types.Block, types.Receipts) {
	if node.newBlockSub == nil {
		node.newBlockSub = node.miner.mux.Subscribe(core.NewMinedBlockEvent{})
	}
	blockChan := make(chan *types.Block, 1)
	go func() {
		item := <-node.newBlockSub.Chan()
		node.miner.Stop()
		fmt.Printf("get stop signal: number=%d\n", item.Data.(core.NewMinedBlockEvent).Block.NumberU64())
		blockChan <- item.Data.(core.NewMinedBlockEvent).Block
	}()
	node.miner.Start(coinbase)
	block := <-blockChan
	receipts := node.BlockChain().GetReceiptsByHash(block.Hash())
	return block, receipts
}

func (node *FakeNode) MineUntilCommitAllTxs(coinbase common.Address) ([]*types.Block, error) {
	if node.newBlockSub == nil {
		node.newBlockSub = node.miner.mux.Subscribe(core.NewMinedBlockEvent{})
	}
	startBlock := node.BlockChain().CurrentBlock()
	// collect hashes of current pending txs
	pending, queued := node.TxPool().Content()
	txs := make(map[common.Hash]*types.Transaction)
	for _, _txs := range pending {
		for _, tx := range _txs {
			txs[tx.Hash()] = tx
		}
	}
	for _, _txs := range queued {
		for _, tx := range _txs {
			txs[tx.Hash()] = tx
		}
	}
	if len(txs) == 0 {
		return nil, errors.New("no pending transactions")
	}
	eliminateTxs := func(block *types.Block) {
		receipts := node.BlockChain().GetReceiptsByHash(block.Hash())
		for _, receipt := range receipts {
			delete(txs, receipt.TxHash)
		}
	}

	blocksChan := make(chan []*types.Block, 1)
	go func() {
		var blocks []*types.Block
		item := <-node.newBlockSub.Chan()
		for {
			blocksCache := make([]*types.Block, 0)
			block := item.Data.(core.NewMinedBlockEvent).Block
			current := block
			for current.Hash() != startBlock.Hash() {
				blocksCache = append(blocksCache, current)
				eliminateTxs(current)
				current = node.BlockChain().GetBlock(current.ParentHash(), current.NumberU64()-1)
			}
			if len(txs) == 0 {
				node.miner.Stop()
				blocks = blocksCache
				break
			}
			item = <-node.newBlockSub.Chan()
		}
		blocksChan <- blocks
	}()
	node.miner.Start(coinbase)
	blocks := <-blocksChan
	return blocks, nil
}

func (node *FakeNode) ImportBlock(block *types.Block) error {
	_, err := node.BlockChain().InsertChain([]*types.Block{block})
	return err
}

func (node *FakeNode) ImportBlocks(blocks []*types.Block) error {
	_, err := node.BlockChain().InsertChain(blocks)
	return err
}

func TestFakeNodeMinesOnlyOneBlock(t *testing.T) {
	chainConfig = params.AllEthashProtocolChanges
	engine := ethash.NewFaker()
	defer engine.Close()
	db := rawdb.NewMemoryDatabase()
	node := NewFakeNode(t, testConfig, chainConfig, engine, db)
	defer node.miner.worker.close()
	node.MineOneBlock(testBankAddress)
	node.MineOneBlock(testBankAddress)
	node.MineOneBlock(testBankAddress)
	time.Sleep(10 * time.Second)
	n := node.BlockChain().CurrentBlock().Number()
	if n.Cmp(big.NewInt(3)) != 0 {
		t.Fatal("not mined one block")
	}
}

func TestFakeNodeCrossImportBlocks(t *testing.T) {
	chainConfig = params.AllEthashProtocolChanges
	engine1 := ethash.NewFaker()
	defer engine1.Close()
	db1 := rawdb.NewMemoryDatabase()
	node1 := NewFakeNode(t, testConfig, chainConfig, engine1, db1)
	defer node1.miner.worker.close()
	block, _ := node1.MineOneBlock(testBankAddress)

	engine2 := ethash.NewFaker()
	defer engine2.Close()
	db2 := rawdb.NewMemoryDatabase()
	node2 := NewFakeNode(t, testConfig, chainConfig, engine2, db2)
	defer node2.miner.worker.close()
	err := node2.ImportBlock(block)
	if err != nil {
		t.Fatalf("Import block error in node2: %s", err.Error())
	}

	block, _ = node2.MineOneBlock(testUserAddress)
	err = node1.ImportBlock(block)
	if err != nil {
		t.Fatalf("Import block error in node1: %s", err.Error())
	}
}

func TestFakeNodeImportSideChain(t *testing.T) {
	chainConfig = params.AllEthashProtocolChanges
	engine1 := ethash.NewFaker()
	defer engine1.Close()
	db1 := rawdb.NewMemoryDatabase()
	node1 := NewFakeNode(t, testConfig, chainConfig, engine1, db1)
	defer node1.miner.worker.close()
	block, _ := node1.MineOneBlock(testBankAddress)
	fmt.Println("difficulty: ", block.Difficulty())

	engine2 := ethash.NewFaker()
	defer engine2.Close()
	db2 := rawdb.NewMemoryDatabase()
	node2 := NewFakeNode(t, testConfig, chainConfig, engine2, db2)
	defer node2.miner.worker.close()
	block2, _ := node2.MineOneBlock(testUserAddress)
	fmt.Println("difficulty: ", block2.Difficulty())
	// import side chain
	err := node2.ImportBlock(block)
	if err != nil {
		t.Fatalf("import side chain failed")
	}
	if node2.BlockChain().CurrentBlock().Number().Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("bad side chain import")
	}

	if node2.BlockChain().HasBlock(block.Hash(), block.NumberU64()) {
		fmt.Println("node1 wins")
	} else if node2.BlockChain().HasBlock(block2.Hash(), block2.NumberU64()) {
		fmt.Println("node2 wins")
	} else {
		t.Fatal("what the fvck?")
	}
}

func TestContractCreationOld(t *testing.T) {
	chainConfig = params.AllEthashProtocolChanges
	engine1 := ethash.NewFaker()
	defer engine1.Close()
	db1 := rawdb.NewMemoryDatabase()
	node1 := NewFakeNode(t, myMinerConfig, chainConfig, engine1, db1)
	fmt.Println("initialize")
	defer node1.miner.worker.close()

	if node1.BlockChain().CurrentBlock().NumberU64() != 0 {
		t.Fatalf("init block is not 0")
	}
	//node1.MineOneBlock(testBankAddress)
	time.Sleep(2 * time.Second)
	fmt.Println()

	const gasLimit = 1153150
	simpleContractByteCode := "6080604052348015600f57600080fd5b5060016000819055506097806100266000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063da358a3c14602d575b600080fd5b605660048036036020811015604157600080fd5b81019080803590602001909291905050506058565b005b806000819055505056fea265627a7a72315820f61be1cc0b7d1e12225e53bd4ee8e5fe505bf75ae07db0e75b27c4754061d44b64736f6c634300050c0032"
	tx := types.NewContractCreation(node1.TxPool().Nonce(testBankAddress), big.NewInt(0), gasLimit, nil, common.FromHex(simpleContractByteCode))
	fmt.Println("add tx")
	err := node1.AddTx(tx, testBankKey)
	if err != nil {
		t.Fatalf("add tx failed")
	}

	fmt.Println()

	fmt.Printf("start mine one, %d =>\n", node1.BlockChain().CurrentBlock().NumberU64())
	block1, _ := node1.MineOneBlock(testBankAddress)
	//for len(receipts) == 0 {
	//	block1, receipts = node1.MineOneBlock(testBankAddress)
	//}
	//fmt.Println(len(receipts), len(block1.Transactions()))
	time.Sleep(2 * time.Second)
	fmt.Println()

	fmt.Printf("start mine one, %d =>\n", node1.BlockChain().CurrentBlock().NumberU64())
	block1, _ = node1.MineOneBlock(testBankAddress)
	//for len(receipts) == 0 {
	//	block1, receipts = node1.MineOneBlock(testBankAddress)
	//}
	//fmt.Println(len(receipts), len(block1.Transactions()))
	time.Sleep(2 * time.Second)
	fmt.Println()

	fmt.Printf("start mine one, %d =>\n", node1.BlockChain().CurrentBlock().NumberU64())
	block1, _ = node1.MineOneBlock(testBankAddress)
	//for len(receipts) == 0 {
	//	block1, receipts = node1.MineOneBlock(testBankAddress)
	//}
	//fmt.Println(len(receipts), len(block1.Transactions()))
	//time.Sleep(2* time.Second)
	fmt.Println()
	fmt.Println("block", node1.BlockChain().CurrentBlock().NumberU64())

	fmt.Println(block1)
}

func TestMineBlocks(t *testing.T) {
	chainConfig = params.AllEthashProtocolChanges
	engine1 := ethash.NewFaker()
	defer engine1.Close()
	db1 := rawdb.NewMemoryDatabase()
	node1 := NewFakeNode(t, myMinerConfig, chainConfig, engine1, db1)
	fmt.Println("initialize")
	defer node1.miner.worker.close()

	if node1.BlockChain().CurrentBlock().NumberU64() != 0 {
		t.Fatalf("init block is not 0")
	}
	//node1.MineOneBlock(testBankAddress)
	time.Sleep(2 * time.Second)
	fmt.Println()

	const gasLimit = 1153150
	simpleContractByteCode := "6080604052348015600f57600080fd5b5060016000819055506097806100266000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063da358a3c14602d575b600080fd5b605660048036036020811015604157600080fd5b81019080803590602001909291905050506058565b005b806000819055505056fea265627a7a72315820f61be1cc0b7d1e12225e53bd4ee8e5fe505bf75ae07db0e75b27c4754061d44b64736f6c634300050c0032"
	tx := types.NewContractCreation(node1.TxPool().Nonce(testBankAddress), big.NewInt(0), gasLimit, nil, common.FromHex(simpleContractByteCode))
	fmt.Println("add tx")
	err := node1.AddTx(tx, testBankKey)
	if err != nil {
		t.Fatalf("add tx failed")
	}

	fmt.Println()

	fmt.Printf("start mine, %d =>\n", node1.BlockChain().CurrentBlock().NumberU64())
	blocks, _ := node1.MineUntilCommitAllTxs(testBankAddress)
	fmt.Println("blocks", len(blocks))
	fmt.Println("current", node1.BlockChain().CurrentBlock().NumberU64())
	fmt.Println("chain")
	for i := node1.BlockChain().CurrentBlock().NumberU64(); i > 0; i-- {
		b := node1.BlockChain().GetBlockByNumber(i)
		fmt.Printf("%d:%d <- ", len(b.Transactions()), i)
	}
	fmt.Println()
	time.Sleep(time.Second)
}

func TestContractTransactionReorg(t *testing.T) {
	// node 1
	chainConfig = params.AllEthashProtocolChanges
	engine1 := ethash.NewFaker()
	defer engine1.Close()
	db1 := rawdb.NewMemoryDatabase()
	node1 := NewFakeNode(t, myMinerConfig, chainConfig, engine1, db1)
	defer node1.miner.worker.close()

	// node 2
	engine2 := ethash.NewFaker()
	defer engine2.Close()
	db2 := rawdb.NewMemoryDatabase()
	node2 := NewFakeNode(t, testConfig, chainConfig, engine2, db2)
	defer node2.miner.worker.close()

	// node 1 commit a tx, in which a contract is created
	const gasLimit = 1153150
	simpleContractByteCode := "6080604052348015600f57600080fd5b5060016000819055506097806100266000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063da358a3c14602d575b600080fd5b605660048036036020811015604157600080fd5b81019080803590602001909291905050506058565b005b806000819055505056fea265627a7a72315820f61be1cc0b7d1e12225e53bd4ee8e5fe505bf75ae07db0e75b27c4754061d44b64736f6c634300050c0032"
	tx := types.NewContractCreation(node1.TxPool().Nonce(testBankAddress), big.NewInt(0), gasLimit, nil, common.FromHex(simpleContractByteCode))
	err := node1.AddTx(tx, testBankKey)
	if err != nil {
		t.Fatalf("add tx failed")
	}
	blocks, _ := node1.MineUntilCommitAllTxs(testBankAddress)

	// node 2 import blocks from node 1
	err = node2.ImportBlocks(blocks)
	if err != nil {
		t.Fatalf("import error")
	}

	// get contract address
	var contractAddress common.Address
	empty := common.Address{}
	for _, block := range blocks {
		for _, receipt := range node1.BlockChain().GetReceiptsByHash(block.Hash()) {
			if receipt.ContractAddress != empty {
				contractAddress = receipt.ContractAddress
				break
			}
		}
		if contractAddress != empty {
			break
		}
	}
	if contractAddress == empty {
		t.Fatalf("do not find the created contract")
	}
	state, err := node2.BlockChain().State()
	if err != nil {
		panic(err)
	}
	locZero := common.BigToHash(big.NewInt(0))
	val := state.GetState(contractAddress, locZero)
	if val != common.BigToHash(big.NewInt(1)) {
		t.Fatalf("initailize value failed")
	}

	// check consistency
	state, err = node2.BlockChain().State()
	if err != nil {
		panic(err)
	}
	contractCode := state.GetCode(contractAddress)
	if len(contractCode) == 0 {
		t.Fatalf("node 2 does not have the contract")
	}

	// node 2 set Value to 999 in contract as testUser
	setValueTx := types.NewTransaction(0, contractAddress, big.NewInt(0), gasLimit, nil, common.Hex2Bytes("da358a3c00000000000000000000000000000000000000000000000000000000000003e7"))
	err = node2.AddTx(setValueTx, testUserKey)
	if err != nil {
		t.Fatalf("add tx failed")
	}
	blocks2, _ := node2.MineUntilCommitAllTxs(testUserAddress)

	// node 2 check the execution of last tx
	state, err = node2.BlockChain().State()
	if err != nil {
		panic(err)
	}
	val = state.GetState(contractAddress, locZero)
	if val != common.BigToHash(big.NewInt(999)) {
		t.Fatalf("set value failed")
	}
	if node2.BlockChain().CurrentBlock().NumberU64() != 2 {
		t.Fatalf("node 2 block is not 2")
	}
	fmt.Println("node 2 Td =", node2.BlockChain().GetTdByHash(node2.BlockChain().CurrentBlock().Hash()).Uint64())

	// node 1 set Value to 100 in contract as testBank
	setValueTx = types.NewTransaction(1, contractAddress, big.NewInt(0), gasLimit, nil, common.Hex2Bytes("da358a3c0000000000000000000000000000000000000000000000000000000000000064"))
	err = node1.AddTx(setValueTx, testBankKey)
	if err != nil {
		t.Fatalf("add tx failed")
	}
	_, _ = node1.MineUntilCommitAllTxs(testBankAddress)

	// node 1 check the execution of last tx
	state, err = node1.BlockChain().State()
	if err != nil {
		panic(err)
	}
	val = state.GetState(contractAddress, locZero)
	if val != common.BigToHash(big.NewInt(100)) {
		t.Fatalf("set value failed")
	}
	if node1.BlockChain().CurrentBlock().NumberU64() != 2 {
		t.Fatalf("node 1 block is not 2")
	}
	fmt.Println("node 1 Td =", node1.BlockChain().GetTdByHash(node1.BlockChain().CurrentBlock().Hash()).Uint64())

	// node 1 import blocks from node 2
	if blocks2[0].ParentHash() == node1.BlockChain().CurrentBlock().Hash() {
		t.Fatalf("how can it be?")
	}
	err = node1.ImportBlocks(blocks2)
	if err != nil {
		t.Fatalf("import blocks error")
	}

	// node 1 check state overrided by node 2
	state, err = node1.BlockChain().State()
	if err != nil {
		panic(err)
	}
	val = state.GetState(contractAddress, locZero)
	fmt.Println("123")
	if val != common.BigToHash(big.NewInt(999)) {
		t.Fatalf("override value failed")
	}
}
