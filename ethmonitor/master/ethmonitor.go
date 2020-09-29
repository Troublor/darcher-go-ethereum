package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"sync"
)

type EthMonitor struct {
	cluster *Cluster
	// the Transaction being traversed lifecycle, represented as a traverser
	traverserMap map[string]*Traverser
	// all non-complete traverser are put in the queue and polled based on the dynamic function
	traverserQueue         *common.DynamicPriorityQueue
	traverserSubscriptions sync.Map //map[string]*event.SubscriptionScope

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
			log.Debug("Transaction is selected from pool", "tx", selectedTx.PrettyHash())
			return m[selectedTx]
		}),
	}
}

func (m *EthMonitor) Start() {
	m.cluster.Start(false)
	go m.newTxLoop()
	go m.traverseLoop()
	go m.txErrorLoop()
	select {}
}

func (m *EthMonitor) newTxLoop() {
	// listen to new transactions
	txCh := make(chan *rpc.Tx, common.EventCacheSize)
	txSub := m.cluster.SubscribeNewTx(txCh)
	defer txSub.Unsubscribe()

	defer func() {
		m.traverserSubscriptions.Range(func(key, value interface{}) bool {
			value.(*event.SubscriptionScope).Close()
			return true
		})
	}()

	// listen for transactions
	for {
		ev := <-txCh

		if ev.GetRole() == rpc.Role_TALKER {
			log.Warn("Receive tx from talker, ignored", "tx", common.PrettifyHash(ev.GetHash()))
			continue
		}

		if _, ok := m.traverserMap[ev.Hash]; ok {
			// this scenario happens when reorg
			continue
		}

		newHeadCh := make(chan *rpc.ChainHead, 10)
		newSideCh := make(chan *rpc.ChainSide, 10)
		newTxCh := make(chan *rpc.Tx, 10)
		scope := &event.SubscriptionScope{}
		scope.Track(m.cluster.SubscribeNewChainHead(newHeadCh))
		scope.Track(m.cluster.SubscribeNewChainSide(newSideCh))
		scope.Track(m.cluster.SubscribeNewTx(newTxCh))
		m.traverserSubscriptions.Store(ev.GetHash(), scope)

		tx := NewTransaction(ev.Hash, ev.Sender, ev.Nonce, newHeadCh, newSideCh, newTxCh)
		traverser := NewTraverser(m.cluster, tx, m.txController)
		m.traverserMap[tx.Hash()] = traverser
		m.traverserQueue.Push(traverser)
	}
}

func (m *EthMonitor) traverseLoop() {
	for {
		// fist prune the items in the queue
		m.traverserQueue.Prune(func(item interface{}) bool {
			tr := item.(*Traverser)
			return tr.Tx().HasFinalized()
		})
		// select next transaction to play with
		selected := m.traverserQueue.Pull().(*Traverser)
		selected.ResumeTraverse()
		if !selected.tx.HasFinalized() {
			// the tx is suspended
			m.traverserQueue.Push(selected)
		} else {
			// unsubscribe subscriptions
			scope, _ := m.traverserSubscriptions.Load(selected.Tx().Hash())
			scope.(*event.SubscriptionScope).Close()
			m.traverserSubscriptions.Delete(selected.Tx().Hash())
		}
	}
}

/**
txErrorLoop listen for txError from contractOracleService.txErrorFeed and call controller.TxErrorHook
*/
func (m *EthMonitor) txErrorLoop() {
	txErrorCh := make(chan *rpc.TxErrorMsg, 10)
	txErrorSub := m.cluster.SubscribeTxError(txErrorCh)
	defer txErrorSub.Unsubscribe()
	for {
		txError := <-txErrorCh
		if txError != nil {
			m.txController.TxErrorHook(txError)
			log.Debug("Transaction error found", "tx", txError.GetHash(), "err", txError.GetDescription())
		}
	}
}
