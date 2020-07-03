package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/master/eth"
	"github.com/ethereum/go-ethereum/event"
	log "github.com/inconshreveable/log15"
)

type Monitor struct {
	cluster *eth.Cluster
	// the Transaction being traversed lifecycle, represented as a traverser
	traverserMap map[string]*Traverser

	traverserQueue *common.DynamicPriorityQueue

	txController TxController
}

func NewMonitor(txController TxController, config eth.ClusterConfig) *Monitor {
	return &Monitor{
		txController: txController,

		cluster: eth.NewCluster(config),

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

func (m *Monitor) Start() {
	m.cluster.Start(false)
	go m.newTxLoop()
	go m.traverseLoop()
	select {}
}

func (m *Monitor) newTxLoop() {
	// listen to new transactions
	txCh := make(chan common.NewTxEvent, common.EventCacheSize)
	txSub := m.cluster.SubscribeNewTxEvent(txCh)

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

		if ev.Role == common.TALKER {
			log.Warn("Receive tx from talker, ignored", "tx", ev.Hash[:8])
			continue
		}

		newHeadCh := make(chan common.NewChainHeadEvent, 10)
		newSideCh := make(chan common.NewChainSideEvent, 10)
		newTxCh := make(chan common.NewTxEvent, 10)
		subscriptions = append(subscriptions, m.cluster.SubscribeNewChainHeadEvent(newHeadCh))
		subscriptions = append(subscriptions, m.cluster.SubscribeNewChainSideEvent(newSideCh))
		subscriptions = append(subscriptions, m.cluster.SubscribeNewTxEvent(newTxCh))

		tx := NewTransaction(ev.Hash, ev.Sender, ev.Nonce, newHeadCh, newSideCh, newTxCh)
		traverser := NewTraverser(m.cluster, tx, m.txController)
		m.traverserMap[tx.Hash()] = traverser
		m.traverserQueue.Push(traverser)
	}
}

func (m *Monitor) traverseLoop() {
	for {
		selected := m.traverserQueue.Pull().(*Traverser)
		selected.ResumeTraverse()
		if !selected.tx.HasFinalized() {
			// the tx is suspended
			m.traverserQueue.Push(selected)
		}
	}
}
