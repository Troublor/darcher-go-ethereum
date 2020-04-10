package ethMonitor

import (
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"
)

// BudgetTask is used to control miner to mine according to a budget
type BudgetTask struct {
	eth               Ethereum
	targetBlockNumber *big.Int
}

func NewBudgetTask(eth Ethereum, budget int64) *BudgetTask {
	return &BudgetTask{
		eth:               eth,
		targetBlockNumber: big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget)),
	}
}

func (m *BudgetTask) ShouldContinue() bool {
	return m.eth.BlockChain().CurrentBlock().Number().Cmp(m.targetBlockNumber) < 0
}

// IntervalTask is used to repeatedly mine blocks with time interval
// TODO concurrency bug: mining process may deadlock when there are many concurrent txs
type IntervalTask struct {
	eth      Ethereum
	interval uint
	clock    <-chan time.Time
	stopCh   chan interface{}

	currentBlockNumber *big.Int
}

func NewIntervalTask(eth Ethereum, interval uint) *IntervalTask {
	return &IntervalTask{
		eth:                eth,
		interval:           interval,
		clock:              time.After(time.Duration(interval) * time.Millisecond),
		stopCh:             make(chan interface{}, 1),
		currentBlockNumber: eth.BlockChain().CurrentBlock().Number(),
	}

}

func (m *IntervalTask) ShouldContinue() bool {
	bn := m.eth.BlockChain().CurrentBlock().Number()
	if bn.Cmp(m.currentBlockNumber) == 0 {
		return true
	}
	select {
	case <-m.clock:
		m.clock = time.After(time.Duration(m.interval) * time.Millisecond)
		m.currentBlockNumber = m.eth.BlockChain().CurrentBlock().Number()
		return true
	case <-m.stopCh:
		return false
	}
}

func (m *IntervalTask) Stop() {
	close(m.stopCh)
}

// TxMonitorTask is a forever running task which mines blocks whenever there are tasks in txpool
// this is different from TxExecuteTask which only mines blocks until certain tx is mined
type TxMonitorTask struct {
	txPool *core.TxPool

	txCh   chan core.NewTxsEvent
	txSub  event.Subscription
	stopCh chan interface{}
}

func NewTxMonitorTask(txPool *core.TxPool) *TxMonitorTask {
	task := &TxMonitorTask{
		txPool: txPool,
		txCh:   make(chan core.NewTxsEvent),
		stopCh: make(chan interface{}),
	}
	task.txSub = task.txPool.SubscribeNewTxsEvent(task.txCh)
	return task
}

func (m *TxMonitorTask) ShouldContinue() bool {
	select {
	case <-m.stopCh:
		return false
	default:
		txs, err := m.txPool.Pending()
		if err != nil {
			m.Stop()
			log.Error("retrieving tx from txPool error", "err", err)
			return false
		} else {
			if len(txs) > 0 {
			out:
				for {
					select {
					case <-m.txCh:
					default:
						break out
					}
				}

				return true
			} else {
				select {
				case <-m.txCh:
					time.Sleep(500 * time.Millisecond)
					return true
				case <-m.stopCh:
					return false
				}
			}
		}
	}
}

func (m *TxMonitorTask) Stop() {
	close(m.stopCh)
	m.txSub.Unsubscribe()
}

// TxExecuteTask is used to control miner to execute a transaction
type TxExecuteTask struct {
	eth               Ethereum
	targetTransaction *types.Transaction
}

func NewTxExecuteTask(eth Ethereum, targetTransaction *types.Transaction) *TxExecuteTask {
	return &TxExecuteTask{
		eth:               eth,
		targetTransaction: targetTransaction,
	}
}

func (m *TxExecuteTask) ShouldContinue() bool {
	return m.eth.TxPool().Get(m.targetTransaction.Hash()) != nil
}

type TdTask struct {
	eth      Ethereum
	targetTd *big.Int
}

func NewTdTask(eth Ethereum, targetTd *big.Int) *TdTask {
	return &TdTask{
		eth:      eth,
		targetTd: targetTd,
	}
}

func (m *TdTask) ShouldContinue() bool {
	bc := m.eth.BlockChain()
	td := bc.GetTdByHash(bc.CurrentBlock().Hash())
	log.Info("TdTask", "td", td, "target", m.targetTd)
	return td.Cmp(m.targetTd) <= 0
}
