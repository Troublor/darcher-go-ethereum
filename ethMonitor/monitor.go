package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/phayes/freeport"
	"math/big"
	"net/rpc"
	"runtime"
	"strings"
)

type Monitor struct {
	role        Role
	client      *rpc.Client
	serverPort  int
	monitorPort int

	// context fields
	eth   Ethereum
	node  Node
	miner Stoppable
	pm    ProtocolManager

	currentTask Task

	// tx scheduler to control the SubmitTransaction JSON RPC
	txScheduler *TxScheduler
}

func NewMonitor(role Role, monitorPort int) *Monitor {
	log.Info("Running as " + strings.ToUpper(fmt.Sprintf("%s", role)) + " node")
	var port int
	switch role {
	case DOER:
		port, _ = freeport.GetFreePort()
	case TALKER:
		port, _ = freeport.GetFreePort()
	default:
		log.Error("Invalid Role", "role", role)
	}
	return &Monitor{
		role:        role,
		serverPort:  port,
		monitorPort: monitorPort,
		txScheduler: NewTxScheduler(),
	}
}

func (m *Monitor) SetEth(eth Ethereum) {
	m.eth = eth
}

func (m *Monitor) SetNode(node Node) {
	m.node = node
}

func (m *Monitor) SetMiner(miner Stoppable) {
	m.miner = miner
}

func (m *Monitor) SetProtocolManager(pm ProtocolManager) {
	m.pm = pm
}

func (m *Monitor) setCurrent(task Task) {
	// should stop persistent tasks (e.g TxMonitorTask, IntervalTask)
	m.StopDaemonMiningTask()

	// set new task
	m.currentTask = task
	log.Info("new mining task", "task", task.String())
}

func (m *Monitor) GetCurrentTask() Task {
	return m.currentTask
}

func (m *Monitor) GetTxScheduler() *TxScheduler {
	return m.txScheduler
}

func (m *Monitor) isInitialized() bool {
	return m.eth != nil && m.node != nil && m.miner != nil
}

func (m *Monitor) IsTxAllowed(hash common.Hash) bool {
	if m.currentTask == nil {
		// if there is no mining task, do not allow any transaction
		return false
	}
	if m.role == TALKER {
		// TALKER should not execute any tx
		return false
	}
	return m.currentTask.IsTxAllowed(hash)
}

func (m *Monitor) Start() {
	// connect to rpc server
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("127.0.0.1:%d", m.monitorPort))
	if err != nil {
		log.Error("failed to connect to rpc server", "err", err.Error())
		return
	}
	m.client = client
	go m.listenChainHeadLoop()
	go m.listenNewTxsLoop()
	go m.listenChainSideLoop()
	go m.listenPeerEventLoop()
	m.startMonitorRpcServer()
}

func (m *Monitor) NotifyNodeStart(node *node.Node) {
	if m.client == nil {
		return
	}
	currentBlock := m.eth.BlockChain().CurrentBlock()
	arg := &NodeStartMsg{
		Role:        m.role,
		ServerPort:  m.serverPort,
		URL:         node.Server().NodeInfo().Enode,
		BlockNumber: currentBlock.NumberU64(),
		BlockHash:   currentBlock.Hash().Hex(),
		Td:          m.eth.BlockChain().GetTdByHash(currentBlock.Hash()).Uint64(),
	}
	reply := &Reply{}
	err := m.client.Call("Server.OnNodeStartRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call NotifyNodeStart error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("NotifyNodeStart error", "err", reply.Err.Error())
	}
}

// start geth miner
func (m *Monitor) StartMining() error {
	return m.eth.StartMining(runtime.NumCPU())
}

// stop geth miner
func (m *Monitor) StopMining() {
	m.eth.StopMining()
}

/*
Mining Utilities
*/
// mine until certain amount of blocks
func (m *Monitor) MineBlocks(count int64) error {
	task := NewBudgetTask(m.eth, count)
	m.setCurrent(task)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) MineBlocksWithoutTx(count int64) error {
	task := NewBudgetWithoutTxTask(m.eth, count)
	m.setCurrent(task)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) MineBlocksExceptTx(count int64, txHash string) error {
	task := NewBudgetExceptTxTask(m.eth, count, txHash)
	m.setCurrent(task)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

// stop daemon mining task if currentTask is daemon
// this function is call from outside monitor to stop current running daemon mining task
func (m *Monitor) StopDaemonMiningTask() {
	if daemonTask, ok := m.currentTask.(DaemonTask); ok {
		daemonTask.Stop()
	}
}

// mine block with time interval (daemon task)
func (m *Monitor) MineBlockInterval(interval uint) error {
	var intervalTask DaemonTask = NewIntervalTask(m.eth, interval)
	m.setCurrent(intervalTask)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

func (m *Monitor) MineWhenTx() error {
	var txMonitorTask DaemonTask = NewTxMonitorTask(m.eth.TxPool())
	m.setCurrent(txMonitorTask)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

// mine a certain tx, stop if tx does not exist
func (m *Monitor) MineTx(txHash common.Hash) error {
	if tx, _, _, _ := rawdb.ReadTransaction(m.eth.ChainDb(), txHash); tx != nil {
		// the tx has already been executed/mined
		log.Warn("transaction has already been executed", "tx", txHash)
		return nil
	}
	tx := m.eth.TxPool().Get(txHash)
	if tx == nil {
		// the tx is not in the txPool
		return fmt.Errorf("tx does not exist: %s", txHash.Hex())
	}
	task := NewTxExecuteTask(m.eth, tx)
	m.setCurrent(task)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

// mine until Td is larger than the given td
func (m *Monitor) MineTd(td *big.Int) error {
	task := NewTdTask(m.eth, td)
	m.setCurrent(task)
	err := m.StartMining()
	if err != nil {
		return err
	}
	return nil
}

/*
Transaction listener
*/
func (m *Monitor) listenNewTxsLoop() {
	txsCh := make(chan core.NewTxsEvent, 4096)
	txsSub := m.eth.TxPool().SubscribeNewTxsEvent(txsCh)
	defer txsSub.Unsubscribe()

	for {
		select {
		case <-txsSub.Err():
			return
		case ev := <-txsCh:
			for _, tx := range ev.Txs {
				// EthMonitor on the remote side should be notified of each tx
				m.notifyNewTx(tx)
			}
		}

	}
}

func (m *Monitor) notifyNewTx(tx *types.Transaction) {
	if m.client == nil {
		return
	}
	arg := &NewTxMsg{Role: m.role, Hash: tx.Hash().Hex()}
	reply := &Reply{}
	err := m.client.Call("Server.OnNewTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewTx error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("notifyNewTx error", "err", reply.Err.Error())
	}
}

/*
ChainHead listener
*/
func (m *Monitor) listenChainHeadLoop() {
	chainHeadCh := make(chan core.ChainHeadEvent, 50)
	chainHeadSub := m.eth.BlockChain().SubscribeChainHeadEvent(chainHeadCh)
	defer chainHeadSub.Unsubscribe()

	for {
		select {
		case <-chainHeadSub.Err():
			return
		case ev := <-chainHeadCh:
			m.notifyNewChainHead(ev.Block)
		}
	}
}

func (m *Monitor) notifyNewChainHead(block *types.Block) {
	if m.client == nil {
		return
	}
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	arg := &NewChainHeadMsg{Role: m.role, Hash: block.Hash().Hex(), Number: block.NumberU64(), Td: m.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(), Txs: txHashes}
	reply := &Reply{}
	err := m.client.Call("Server.OnNewChainHeadRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewChainHead error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("notifyNewChainHead error", "err", reply.Err.Error())
	}
}

/*
ChainSide listener
*/
func (m *Monitor) listenChainSideLoop() {
	chainSideCh := make(chan core.ChainSideEvent, 50)
	chainSideSub := m.eth.BlockChain().SubscribeChainSideEvent(chainSideCh)
	defer chainSideSub.Unsubscribe()

	for {
		select {
		case <-chainSideSub.Err():
			return
		case ev := <-chainSideCh:
			m.notifyNewChainSide(ev.Block)
		}
	}
}

func (m *Monitor) notifyNewChainSide(block *types.Block) {
	if m.client == nil {
		return
	}
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	arg := &NewChainSideMsg{Role: m.role, Hash: block.Hash().Hex(), Number: block.NumberU64(), Td: m.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(), Txs: txHashes}
	reply := &Reply{}
	err := m.client.Call("Server.OnNewChainSideRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewChainSide error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("notifyNewChainSide error", "err", reply.Err.Error())
	}
}

/*
Peer event listener
*/
func (m *Monitor) listenPeerEventLoop() {
	peerCh := make(chan *p2p.PeerEvent)
	peerSub := m.node.Server().SubscribeEvents(peerCh)
	defer peerSub.Unsubscribe()

	for {
		select {
		case <-peerSub.Err():
			return
		case ev := <-peerCh:
			if ev.Type == p2p.PeerEventTypeAdd {
				log.Info("Peer added success")
				m.notifyPeerAdd(ev)
			} else if ev.Type == p2p.PeerEventTypeDrop {
				log.Info("Peer removed success")
				m.notifyPeerDrop(ev)
			}
		}
	}
}

func (m *Monitor) notifyPeerAdd(ev *p2p.PeerEvent) {
	if m.client == nil {
		return
	}
	arg := &PeerAddMsg{Role: m.role, PeerID: ev.Peer.String()}
	reply := &Reply{}
	err := m.client.Call("Server.OnPeerAddRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyPeerAdd error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("notifyPeerAdd error", "err", reply.Err.Error())
	}
}

func (m *Monitor) notifyPeerDrop(ev *p2p.PeerEvent) {
	if m.client == nil {
		return
	}
	arg := &PeerRemoveMsg{Role: m.role, PeerID: ev.Peer.String()}
	reply := &Reply{}
	err := m.client.Call("Server.OnPeerRemoveRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyPeerDrop error", "err", err.Error())
	}
	if reply != nil && reply.Err != nil {
		log.Error("notifyPeerDrop error", "err", reply.Err.Error())
	}
}
