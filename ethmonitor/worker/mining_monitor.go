package worker

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"math/big"
	"runtime"
	"strings"
	"sync"
	"time"
)

type MiningMonitor struct {
	role              rpc.Role
	client            *EthMonitorClient
	ethMonitorAddress string
	ctx               context.Context
	cancel            context.CancelFunc

	// a feed for every new Task
	newTaskFeed event.Feed

	// context fields
	eth   Ethereum
	node  *node.Node
	miner Stoppable

	// current mining task
	currentTask        Task
	currentTaskRWMutex sync.RWMutex
	// regularly mining
	regularMineTicker     *time.Ticker
	regularMineTickerDone chan bool

	// tx scheduler to control the SubmitTransaction JSON RPC
	txScheduler TxScheduler
}

func NewMonitor(role rpc.Role, monitorAddress string) *MiningMonitor {
	log.Info("Running as " + strings.ToUpper(fmt.Sprintf("%s", role)) + " node")
	var scheduler TxScheduler
	if monitorAddress == "" {
		scheduler = FakeTxScheduler
		log.Warn("MiningMonitor starts without upstream")
	} else {
		scheduler = newTxScheduler()
	}
	ctx, cancel := context.WithCancel(context.Background())
	m := &MiningMonitor{
		role:              role,
		ethMonitorAddress: monitorAddress,
		txScheduler:       scheduler,
		ctx:               ctx,
		cancel:            cancel,
	}

	// subscribe and immediate unsubscribe to set the Type of newTaskFeed
	ch := make(chan Task, 0)
	sub := m.newTaskFeed.Subscribe(ch)
	sub.Unsubscribe()

	return m
}

/* Public Methods to use ethmonitor start */
func (m *MiningMonitor) AssignMiningTask(task Task) error {
	// should stop persistent tasks (e.g TxMonitorTask, IntervalTask)
	m.StopMiningTask()

	// set new task
	m.currentTaskRWMutex.Lock()
	m.currentTask = task
	m.currentTaskRWMutex.Unlock()
	m.newTaskFeed.Send(task)
	log.Info("New mining task", "task", task.String())
	return m.eth.StartMining(runtime.NumCPU())
}

/**
MineRegularly tells ethmonitor to regularly mine regardless of the MiningTask assigned.
*/
func (m *MiningMonitor) MineRegularly(interval time.Duration) {
	m.StopMiningRegularly()

	if m.regularMineTickerDone == nil {
		m.regularMineTickerDone = make(chan bool, 1)
	}

	m.regularMineTicker = time.NewTicker(interval)
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		for {
			select {
			case <-m.regularMineTickerDone:
				cancel()
				return
			case <-m.regularMineTicker.C:
				currentTask := m.GetCurrentTask()
				if currentTask != nil && currentTask != NilTask {
					log.Info("Regularly mining start", "pause_task", currentTask.String())
				}
				m.StopMiningTask()
				regularTask := NewBudgetWithoutTxTask(ctx, m.eth, 1)
				_ = m.AssignMiningTask(regularTask)
				<-regularTask.TargetAchievedCh()
				if currentTask != nil && currentTask != NilTask {
					log.Info("Regular mining stop", "resume_task", currentTask.String())
				}
			}
		}
	}()
}

/**
MineRegularly tells ethmonitor to stop regularly mining.
*/
func (m *MiningMonitor) StopMiningRegularly() {
	if m.regularMineTickerDone == nil {
		m.regularMineTickerDone = make(chan bool, 1)
	}
	if m.regularMineTicker != nil {
		m.regularMineTicker.Stop()
		m.regularMineTickerDone <- true
	}
}

// stop daemon mining task if currentTask is daemon
// this function is call from outside monitor to stop current running daemon mining task
func (m *MiningMonitor) StopMiningTask() {
	m.currentTaskRWMutex.Lock()
	defer m.currentTaskRWMutex.Unlock()
	if daemonTask, ok := m.currentTask.(DaemonTask); ok {
		daemonTask.Stop()
	}
	// tell eth to stop mining
	m.eth.StopMining()
	// set currentTask to be nil
	if m.currentTask != nil && m.currentTask != NilTask {
		log.Info("Stopped mining task", "task", m.currentTask.String())
	}
	m.currentTask = NilTask
}

/* Public Methods to use ethmonitor end */

/* methods called when initializing geth start */
func (m *MiningMonitor) Start() {
	if m.ethMonitorAddress == "" {
		return
	}
	m.client = NewClient(m.role, m.eth, m.ctx, m.ethMonitorAddress)
	log.Info("ethmonitor client started")

	// start reverse rpcs
	err := m.startReverseRPCs()
	if err != nil {
		log.Error("Serve reverse RPC error", "err", err)
	}

	// start listener loops
	go m.listenChainHeadLoop()
	go m.listenNewTxsLoop()
	go m.listenChainSideLoop()
}

func (m *MiningMonitor) Stop() {
	log.Info("Stopping ethmonitor")
	m.cancel()
}

func (m *MiningMonitor) NotifyNodeStart(node *node.Node) {
	if m.client != nil {
		err := m.client.NotifyNodeStart(node)
		if err != nil {
			log.Error("Notify node start error", "err", err)
		}
	}
}

func (m *MiningMonitor) SetEth(eth Ethereum) {
	m.eth = eth
}

func (m *MiningMonitor) SetNode(node *node.Node) {
	m.node = node
}

func (m *MiningMonitor) SetMiner(miner Stoppable) {
	m.miner = miner
}

/* methods called when initializing geth ends */

/**
start reverse RPC services (via bidirectional grpc)
*/
func (m *MiningMonitor) startReverseRPCs() error {
	err := m.client.ServeGetHeadControl(m.getHeadControlHandler)
	if err != nil {
		log.Error("Serve GetHead reverse RPC error", "err", err)
		return err
	}
	log.Info("GetHead reverse RPC started")
	err = m.client.ServeAddPeerControl(m.addPeerControlHandler)
	if err != nil {
		log.Error("Serve AddPeer reverse RPC error", "err", err)
		return err
	}
	log.Info("AddPeer reverse RPC started")
	err = m.client.ServeRemovePeerControl(m.removePeerControlHandler)
	if err != nil {
		log.Error("Serve RemovePeer reverse RPC error", "err", err)
		return err
	}
	log.Info("RemovePeer reverse RPC started")
	err = m.client.ServeScheduleTxControl(m.scheduleTxControlHandler)
	if err != nil {
		log.Error("Serve ScheduleTx reverse RPC error", "err", err)
		return err
	}
	log.Info("ScheduleTx reverse RPC started")
	err = m.client.ServeMineBlocksControl(m.mineBlocksControlHandler)
	if err != nil {
		log.Error("Serve MineBlocks reverse RPC error", "err", err)
		return err
	}
	log.Info("MineBlocks reverse RPC started")
	err = m.client.ServeMineBlocksExceptTxControl(m.mineBlocksExceptTxControlHandler)
	if err != nil {
		log.Error("Serve MineBlocksExceptTx reverse RPC error", "err", err)
		return err
	}
	log.Info("MineBlocksExceptTx reverse RPC started")
	err = m.client.ServeMineBlocksWithoutTxControl(m.mineBlocksWithoutTxControlHandler)
	if err != nil {
		log.Error("Serve MineBlocksWithoutTx reverse RPC error", "err", err)
		return err
	}
	log.Info("MineBlocksWithoutTx reverse RPC started")
	err = m.client.ServeMineTdControl(m.mineTdControlHandler)
	if err != nil {
		log.Error("Serve MineTd reverse RPC error", "err", err)
		return err
	}
	log.Info("MineTd reverse RPC started")
	err = m.client.ServeMineTxControl(m.mineTxControlHandler)
	if err != nil {
		log.Error("Serve MineTx reverse RPC error", "err", err)
		return err
	}
	log.Info("MineTx reverse RPC started")
	err = m.client.ServeCheckTxInPoolControl(m.checkTxInPoolControlHandler)
	if err != nil {
		log.Error("Serve CheckTxInPool reverse RPC error", "err", err)
		return err
	}
	log.Info("CheckTxInPool reverse RPC started")
	err = m.client.ServeGetReportsByContractControl(m.getReportsByContractHandler)
	if err != nil {
		log.Error("Serve GerReportsByContract reverse RPC error", "err", err)
		return err
	}
	log.Info("GerReportsByContract reverse RPC started")
	err = m.client.ServeGetReportsByTransactionControl(m.getReportsByTransactionHandler)
	if err != nil {
		log.Error("Serve GetReportsByTransaction reverse RPC error", "err", err)
		return err
	}
	log.Info("GetReportsByTransaction reverse RPC started")
	return nil
}

/* reverse RPC handlers start */
func (m *MiningMonitor) _getCurrentChainHead() *rpc.ChainHead {
	currentBlock := m.eth.BlockChain().CurrentBlock()
	txs := make([]string, len(currentBlock.Transactions()))
	for i, tx := range currentBlock.Transactions() {
		txs[i] = tx.Hash().Hex()
	}
	return &rpc.ChainHead{
		Role:   m.role,
		Hash:   currentBlock.Hash().Hex(),
		Number: currentBlock.NumberU64(),
		Td:     m.eth.BlockChain().GetTdByHash(currentBlock.Hash()).Uint64(),
		Txs:    txs,
	}
}

func (m *MiningMonitor) getHeadControlHandler(in *rpc.GetChainHeadControlMsg) (out *rpc.GetChainHeadControlMsg) {
	out = in
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) addPeerControlHandler(in *rpc.AddPeerControlMsg) (out *rpc.AddPeerControlMsg) {
	log.Debug("Receive addPeerControlMsg", "enode", in.GetUrl())
	out = in
	peerCh := make(chan *p2p.PeerEvent)
	peerSub := m.node.Server().SubscribeEvents(peerCh)
	defer peerSub.Unsubscribe()

	// parse enode url to get enode
	eNode, err := enode.Parse(enode.ValidSchemes, in.GetUrl())
	if err != nil {
		log.Error("Invalid enode url", "url", in.GetUrl())
		out.Err = rpc.Error_InternalErr
		return out
	}

	// short circuit if peer is already added
	for _, peerInfo := range m.node.Server().PeersInfo() {
		if peerInfo.ID == eNode.ID().String() {
			out.Err = rpc.Error_NilErr
			return out
		}
	}

	// add peer with enode
	err = m.addPeer(eNode)
	if err != nil {
		log.Error("Add peer error", "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for add success
	for {
		ev := <-peerCh
		if ev.Type == p2p.PeerEventTypeAdd && ev.Peer == eNode.ID() {
			log.Info("Peer added success")
			out.Err = rpc.Error_NilErr
			out.PeerId = eNode.ID().String()
			break
		}
	}
	return out
}

func (m *MiningMonitor) removePeerControlHandler(in *rpc.RemovePeerControlMsg) (out *rpc.RemovePeerControlMsg) {
	out = in
	peerCh := make(chan *p2p.PeerEvent)
	peerSub := m.node.Server().SubscribeEvents(peerCh)
	defer peerSub.Unsubscribe()

	// parse enode url to get enode
	eNode, err := enode.Parse(enode.ValidSchemes, in.GetUrl())
	if err != nil {
		log.Error("Invalid enode url", "url", in.GetUrl())
		out.Err = rpc.Error_InternalErr
		return out
	}

	// short circuit if peer is already removed
	removed := true
	for _, peerInfo := range m.node.Server().PeersInfo() {
		if peerInfo.ID == eNode.ID().String() {
			removed = false
		}
	}
	if removed {
		out.Err = rpc.Error_NilErr
		return out
	}

	// add peer with enode
	log.Info("RemovePeer reverse RPC received")
	err = m.removePeer(eNode)
	if err != nil {
		log.Error("Remove peer error", "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for add success
	for {
		ev := <-peerCh
		if ev.Type == p2p.PeerEventTypeDrop && ev.Peer == eNode.ID() {
			log.Info("Peer removed success")
			out.Err = rpc.Error_NilErr
			out.PeerId = eNode.ID().String()
			break
		}

	}
	return out
}

func (m *MiningMonitor) scheduleTxControlHandler(in *rpc.ScheduleTxControlMsg) (out *rpc.ScheduleTxControlMsg) {
	out = in
	m.GetTxScheduler().ScheduleTx(in.Hash)
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) mineBlocksControlHandler(in *rpc.MineBlocksControlMsg) (out *rpc.MineBlocksControlMsg) {
	out = in

	// assign mining task
	task := NewBudgetTask(m.ctx, m.eth, int64(in.GetCount()))
	err := m.AssignMiningTask(task)
	if err != nil {
		log.Error("Assign mining task error", "task", task.String(), "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for task finish
	<-task.TargetAchievedCh()
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) mineBlocksExceptTxControlHandler(in *rpc.MineBlocksExceptTxControlMsg) (out *rpc.MineBlocksExceptTxControlMsg) {
	out = in

	// assign mining task
	task := NewBudgetExceptTxTask(m.ctx, m.eth, int64(in.GetCount()), in.GetTxHash())
	err := m.AssignMiningTask(task)
	if err != nil {
		log.Error("Assign mining task error", "task", task.String(), "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for task finish
	<-task.TargetAchievedCh()
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) mineBlocksWithoutTxControlHandler(in *rpc.MineBlocksWithoutTxControlMsg) (out *rpc.MineBlocksWithoutTxControlMsg) {
	out = in

	// assign mining task
	task := NewBudgetWithoutTxTask(m.ctx, m.eth, int64(in.GetCount()))
	err := m.AssignMiningTask(task)
	if err != nil {
		log.Error("Assign mining task error", "task", task.String(), "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for task finish
	<-task.TargetAchievedCh()
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) mineTdControlHandler(in *rpc.MineTdControlMsg) (out *rpc.MineTdControlMsg) {
	out = in

	// assign mining task
	task := NewTdTask(m.ctx, m.eth, big.NewInt(int64(in.GetTd())))
	err := m.AssignMiningTask(task)
	if err != nil {
		log.Error("Assign mining task error", "task", task.String(), "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for task finish
	<-task.TargetAchievedCh()
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) mineTxControlHandler(in *rpc.MineTxControlMsg) (out *rpc.MineTxControlMsg) {
	out = in

	// assign mining task
	task := NewTxExecuteTask(m.ctx, m.eth, m.eth.TxPool().Get(common.HexToHash(in.GetHash())))
	err := m.AssignMiningTask(task)
	if err != nil {
		log.Error("Assign mining task error", "task", task.String(), "err", err)
		out.Err = rpc.Error_InternalErr
		return out
	}

	// wait for task finish
	<-task.TargetAchievedCh()
	out.Head = m._getCurrentChainHead()
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) checkTxInPoolControlHandler(in *rpc.CheckTxInPoolControlMsg) (out *rpc.CheckTxInPoolControlMsg) {
	out = in
	out.InPool = m.eth.TxPool().Get(common.HexToHash(in.GetHash())) != nil
	out.Err = rpc.Error_NilErr
	return out
}

func (m *MiningMonitor) getReportsByContractHandler(in *rpc.GetReportsByContractControlMsg) (out *rpc.GetReportsByContractControlMsg) {
	out = in
	reports := vm.GetEVMMonitorProxy().Reports()
	out.Reports = make([]*rpc.ContractVulReport, 0)
	for _, report := range reports {
		if report.GetAddress() == in.GetAddress() {
			out.Reports = append(out.Reports, report)
		}
	}
	return out
}

func (m *MiningMonitor) getReportsByTransactionHandler(in *rpc.GetReportsByTransactionControlMsg) (out *rpc.GetReportsByTransactionControlMsg) {
	out = in
	reports := vm.GetEVMMonitorProxy().Reports()
	out.Reports = make([]*rpc.ContractVulReport, 0)
	for _, report := range reports {
		if report.GetTxHash() == in.GetHash() {
			out.Reports = append(out.Reports, report)
		}
	}
	return out
}

/* reverse RPC handlers end */

func (m *MiningMonitor) addPeer(node *enode.Node) error {
	// Make sure the server is running, fail otherwise
	server := m.node.Server()
	if server == nil {
		return fmt.Errorf("node stopped")
	}
	server.AddPeer(node)
	server.AddTrustedPeer(node)
	return nil
}

func (m *MiningMonitor) removePeer(node *enode.Node) error {
	// Make sure the server is running, fail otherwise
	server := m.node.Server()
	if server == nil {
		return fmt.Errorf("node stopped")
	}
	server.RemoveTrustedPeer(node)
	server.RemovePeer(node)
	return nil
}

/* methods to control tx lifecycle start*/
func (m *MiningMonitor) SubscribeNewTask(ch chan<- Task) event.Subscription {
	return m.newTaskFeed.Subscribe(ch)
}

func (m *MiningMonitor) GetCurrentTask() Task {
	m.currentTaskRWMutex.RLock()
	defer m.currentTaskRWMutex.RUnlock()
	if m.currentTask != nil {
		return m.currentTask
	} else {
		return NilTask
	}
}

func (m *MiningMonitor) GetTxScheduler() TxScheduler {
	return m.txScheduler
}

/**
TxErrorNotifier is used for vm.EVMMonitorProxy to notify tx execution errors
*/
func (m *MiningMonitor) GetTxErrorNotifier() func(msg *rpc.TxErrorMsg) error {
	if m.client != nil {
		return m.client.NotifyTxError
	} else {
		return func(msg *rpc.TxErrorMsg) error {
			log.Warn("Transaction execution failed", "txHash", msg.Hash, "err", msg.Description)
			return nil
		}
	}
}

func (m *MiningMonitor) IsTxAllowed(hash common.Hash) bool {
	m.currentTaskRWMutex.RLock()
	defer m.currentTaskRWMutex.RUnlock()
	if m.currentTask == nil {
		// if there is no mining task, do not allow any transaction
		return false
	}
	if m.role == rpc.Role_TALKER {
		// TALKER should not execute any tx
		return false
	}
	return m.currentTask.IsTxAllowed(hash)
}

/* methods to control tx lifecycle end*/

/*
Transaction listener
*/
func (m *MiningMonitor) listenNewTxsLoop() {
	txsCh := make(chan core.NewTxsEvent, 4096)
	txsSub := m.eth.TxPool().SubscribeNewTxsEvent(txsCh)
	defer txsSub.Unsubscribe()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-txsSub.Err():
			return
		case ev := <-txsCh:
			for _, tx := range ev.Txs {
				// EthMonitor on the remote side should be notified of each tx
				if m.client != nil {
					//m.legacyClient.NotifyNewTx(tx, m.role)
					err := m.client.NotifyNewTx(tx)
					if err != nil {
						log.Error("Notify new tx err", "err", err)
					}
				}
			}
		}

	}
}

/*
ChainHead listener
*/
func (m *MiningMonitor) listenChainHeadLoop() {
	chainHeadCh := make(chan core.ChainHeadEvent, 50)
	chainHeadSub := m.eth.BlockChain().SubscribeChainHeadEvent(chainHeadCh)
	defer chainHeadSub.Unsubscribe()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-chainHeadSub.Err():
			return
		case ev := <-chainHeadCh:
			if m.client != nil {
				//m.legacyClient.NotifyNewChainHead(ev.Block, m.role)
				time.Sleep(50 * time.Millisecond) // delay a little bit to make sure chain side rpc goes first
				err := m.client.NotifyNewChainHead(ev.Block)
				if err != nil {
					log.Error("Notify new chain head err", "err", err)
				}
			}
		}
	}
}

/*
ChainSide listener
*/
func (m *MiningMonitor) listenChainSideLoop() {
	chainSideCh := make(chan core.ChainSideEvent, 50)
	chainSideSub := m.eth.BlockChain().SubscribeChainSideEvent(chainSideCh)
	defer chainSideSub.Unsubscribe()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-chainSideSub.Err():
			return
		case ev := <-chainSideCh:
			if m.client != nil {
				//m.legacyClient.NotifyNewChainSide(ev.Block, m.role)
				err := m.client.NotifyNewChainSide(ev.Block)
				if err != nil {
					log.Error("Notify new chain side err", "err", err)
				}
			}
		}
	}
}
