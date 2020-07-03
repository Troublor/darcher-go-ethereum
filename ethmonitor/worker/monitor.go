package worker

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/phayes/freeport"
	"runtime"
	"strings"
)

type Monitor struct {
	role           Role
	client         *Client // used to communicate with ethMonitor
	server         *Server // used to receive control message from ethMonitor
	serverPort     int
	ethMonitorPort int

	// a feed for every new Task
	newTaskFeed event.Feed

	// context fields
	eth   Ethereum
	node  Node
	miner Stoppable
	pm    ProtocolManager

	currentTask Task

	// tx scheduler to control the SubmitTransaction JSON RPC
	txScheduler TxScheduler
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
	var scheduler TxScheduler
	if monitorPort == 0 {
		scheduler = FakeTxScheduler
		log.Warn("Monitor starts without upstream")
	} else {
		scheduler = newTxScheduler()
	}
	m := &Monitor{
		role:           role,
		ethMonitorPort: monitorPort,
		txScheduler:    scheduler,
		serverPort:     port,
	}
	return m
}

func (m *Monitor) StartMonitoring() {
	if m.ethMonitorPort == 0 {
		return
	}
	// start Monitor server to handle control message from ethMonitor
	m.server = NewServer(m, m.serverPort)
	// connect to ethMonitor as a client
	m.client = NewClient(m, m.ethMonitorPort)

	go m.listenChainHeadLoop()
	go m.listenNewTxsLoop()
	go m.listenChainSideLoop()
	go m.listenPeerEventLoop()
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

func (m *Monitor) setCurrentTask(task Task) {
	// should stop persistent tasks (e.g TxMonitorTask, IntervalTask)
	m.StopMiningTask()

	// set new task
	m.currentTask = task
	m.newTaskFeed.Send(task)
	log.Info("new mining task", "task", task.String())
}

func (m *Monitor) SubscribeNewTask(ch chan<- Task) event.Subscription {
	return m.newTaskFeed.Subscribe(ch)
}

func (m *Monitor) GetCurrentTask() Task {
	if m.currentTask != nil {
		return m.currentTask
	} else {
		return NilTask
	}
}

func (m *Monitor) GetTxScheduler() TxScheduler {
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

func (m *Monitor) NotifyNodeStart(node *node.Node) {
	if m.client != nil {
		m.client.NotifyNodeStart(node, m.role, m.serverPort)
	}
}

func (m *Monitor) AssignMiningTask(task Task) error {
	// should stop persistent tasks (e.g TxMonitorTask, IntervalTask)
	m.StopMiningTask()

	// set new task
	m.currentTask = task
	m.newTaskFeed.Send(task)
	log.Info("New mining task", "task", task.String())
	return m.eth.StartMining(runtime.NumCPU())
}

// stop daemon mining task if currentTask is daemon
// this function is call from outside monitor to stop current running daemon mining task
func (m *Monitor) StopMiningTask() {
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
				if m.client != nil {
					m.client.NotifyNewTx(tx, m.role)
				}
			}
		}

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
			if m.client != nil {
				m.client.NotifyNewChainHead(ev.Block, m.role)
			}
		}
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
			if m.client != nil {
				m.client.NotifyNewChainSide(ev.Block, m.role)
			}
		}
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
				if m.client != nil {
					m.client.NotifyPeerAdd(ev, m.role)
				}
			} else if ev.Type == p2p.PeerEventTypeDrop {
				log.Info("Peer removed success")
				if m.client != nil {
					m.client.NotifyPeerDrop(ev, m.role)
				}
			}
		}
	}
}
