package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	log "github.com/inconshreveable/log15"
)

type EthMonitor struct {
	cluster *Cluster
	// the Transaction being traversed lifecycle, represented as a traverser
	traverserMap map[string]*Traverser
	// all non-complete traverser are put in the queue and polled based on the dynamic function
	traverserQueue *common.DynamicPriorityQueue

	txController TxController
}

func NewMonitor(txController TxController, config ClusterConfig) *EthMonitor {
	return &EthMonitor{
		txController: txController,

		cluster: NewCluster(config),

		traverserMap: make(map[string]*Traverser),
		traverserQueue: common.NewDynamicPriorityQueue(func(candidates []interface{}) (selected interface{}) {
			txs := make([]*Transaction, len(candidates))
			m := make(map[*Transaction]*Traverser)
			for i, candidate := range candidates {
				txs[i] = candidate.(*Traverser).tx
				m[txs[i]] = candidate.(*Traverser)
			}
			selectedTx := txController.SelectTxToTraverse(txs)
			return m[selectedTx]
		}),
	}
}

func (m *EthMonitor) Start() {
	m.cluster.Start(false)
	go m.newTxLoop()
	go m.traverseLoop()
	select {}
}

func (m *EthMonitor) newTxLoop() {
	// listen to new transactions
	txCh := make(chan *rpc.Tx, common.EventCacheSize)
	txSub := m.cluster.SubscribeNewTx(txCh)

	subscriptions := make([]event.Subscription, 1)
	subscriptions[0] = txSub
	defer func() {
		for _, sub := range subscriptions {
			sub.Unsubscribe()
		}
	}()

	// listen for transactions
	for {
		ev := <-txCh
		if _, ok := m.traverserMap[ev.Hash]; ok {
			// this scenario happens when reorg
			continue
		}

		if ev.GetRole() == rpc.Role_TALKER {
			log.Warn("Receive tx from talker, ignored", "tx", ev.Hash[:8])
			continue
		}

		newHeadCh := make(chan *rpc.ChainHead, 10)
		newSideCh := make(chan *rpc.ChainSide, 10)
		newTxCh := make(chan *rpc.Tx, 10)
		subscriptions = append(subscriptions, m.cluster.SubscribeNewChainHead(newHeadCh))
		subscriptions = append(subscriptions, m.cluster.SubscribeNewChainSide(newSideCh))
		subscriptions = append(subscriptions, m.cluster.SubscribeNewTx(newTxCh))

		tx := NewTransaction(ev.Hash, ev.Sender, ev.Nonce, newHeadCh, newSideCh, newTxCh)
		traverser := NewTraverser(m.cluster, tx, m.txController)
		m.traverserMap[tx.Hash()] = traverser
		m.traverserQueue.Push(traverser)
	}
}

func (m *EthMonitor) traverseLoop() {
	for {
		selected := m.traverserQueue.Pull().(*Traverser)
		selected.ResumeTraverse()
		if !selected.tx.HasFinalized() {
			// the tx is suspended
			m.traverserQueue.Push(selected)
		}
	}
}
