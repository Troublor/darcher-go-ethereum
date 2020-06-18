package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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

func (t *BudgetTask) ShouldContinue() bool {
	return t.eth.BlockChain().CurrentBlock().Number().Cmp(t.targetBlockNumber) < 0
}

func (t *BudgetTask) String() string {
	return fmt.Sprintf("BudgetTask(%d)", t.targetBlockNumber.Int64())
}

func (t *BudgetTask) IsTxAllowed(txHash common.Hash) bool {
	return true
}

// BudgetTask is used to control miner to mine according to a budget
// do not allow any txs
type BudgetWithoutTxTask struct {
	eth               Ethereum
	targetBlockNumber *big.Int
}

func NewBudgetWithoutTxTask(eth Ethereum, budget int64) *BudgetWithoutTxTask {
	return &BudgetWithoutTxTask{
		eth:               eth,
		targetBlockNumber: big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget)),
	}
}

func (t *BudgetWithoutTxTask) ShouldContinue() bool {
	return t.eth.BlockChain().CurrentBlock().Number().Cmp(t.targetBlockNumber) < 0
}

func (t *BudgetWithoutTxTask) String() string {
	return fmt.Sprintf("BudgetWithoutTxTask(%d)", t.targetBlockNumber.Int64())
}

func (t *BudgetWithoutTxTask) IsTxAllowed(txHash common.Hash) bool {
	return false
}

// BudgetTask is used to control miner to mine according to a budget
// allowing all txs but one specified tx
type BudgetExceptTxTask struct {
	eth               Ethereum
	targetBlockNumber *big.Int
	txHash            string
}

func NewBudgetExceptTxTask(eth Ethereum, budget int64, txHash string) *BudgetExceptTxTask {
	return &BudgetExceptTxTask{
		eth:               eth,
		targetBlockNumber: big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget)),
		txHash:            txHash,
	}
}

func (t *BudgetExceptTxTask) ShouldContinue() bool {
	return t.eth.BlockChain().CurrentBlock().Number().Cmp(t.targetBlockNumber) < 0
}

func (t *BudgetExceptTxTask) String() string {
	return fmt.Sprintf("BudgetExceptTxTask(%d, %s)", t.targetBlockNumber.Int64(), PrettifyHash(t.txHash))
}

func (t *BudgetExceptTxTask) IsTxAllowed(txHash common.Hash) bool {
	return t.txHash != txHash.String()
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

func (t *IntervalTask) ShouldContinue() bool {
	bn := t.eth.BlockChain().CurrentBlock().Number()
	if bn.Cmp(t.currentBlockNumber) == 0 {
		return true
	}
	select {
	case <-t.clock:
		t.clock = time.After(time.Duration(t.interval) * time.Millisecond)
		t.currentBlockNumber = t.eth.BlockChain().CurrentBlock().Number()
		return true
	case <-t.stopCh:
		return false
	}
}

func (t *IntervalTask) Stop() {
	close(t.stopCh)
}

func (t *IntervalTask) IsTxAllowed(txHash common.Hash) bool {
	return true
}

func (t *IntervalTask) String() string {
	return fmt.Sprintf("IntervalTask(%d)", t.interval)
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

func (t *TxMonitorTask) ShouldContinue() bool {
	select {
	case <-t.stopCh:
		return false
	default:
		txs, err := t.txPool.Pending()
		if err != nil {
			t.Stop()
			log.Error("retrieving tx from txPool error", "err", err)
			return false
		} else {
			if len(txs) > 0 {
			out:
				for {
					select {
					case <-t.txCh:
					default:
						break out
					}
				}

				return true
			} else {
				select {
				case <-t.txCh:
					time.Sleep(500 * time.Millisecond)
					return true
				case <-t.stopCh:
					return false
				}
			}
		}
	}
}

func (t *TxMonitorTask) Stop() {
	close(t.stopCh)
	t.txSub.Unsubscribe()
}

func (t *TxMonitorTask) IsTxAllowed(txHash common.Hash) bool {
	return true
}

func (t *TxMonitorTask) String() string {
	return fmt.Sprintf("TxMonitorTask()")
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

func (t *TxExecuteTask) ShouldContinue() bool {
	return t.eth.TxPool().Get(t.targetTransaction.Hash()) != nil
}

func (t *TxExecuteTask) IsTxAllowed(txHash common.Hash) bool {
	return t.targetTransaction.Hash() == txHash
}

func (t *TxExecuteTask) String() string {
	return fmt.Sprintf("TxExecuteTask(%s)", PrettifyHash(t.targetTransaction.Hash().String()))
}

// a mining task to mine until total difficulty being larger than the given Td
// no tx is allowed
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

func (t *TdTask) ShouldContinue() bool {
	bc := t.eth.BlockChain()
	td := bc.GetTdByHash(bc.CurrentBlock().Hash())
	log.Info("TdTask", "td", td, "target", t.targetTd)
	return td.Cmp(t.targetTd) <= 0
}

func (t *TdTask) IsTxAllowed(txHash common.Hash) bool {
	return false
}

func (t *TdTask) String() string {
	return fmt.Sprintf("TdTask(%d)", t.targetTd.Int64())
}
