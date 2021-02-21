package worker

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"time"
)

type baseTask struct {
	targetAchievedCh chan interface{}
	ctx              context.Context
}

func newBaseTask(ctx context.Context, eth Ethereum, targetAchieved func(block *types.Block) (achieved bool)) baseTask {
	task := baseTask{targetAchievedCh: make(chan interface{}), ctx: ctx}
	go func() {
		ch := make(chan core.ChainHeadEvent, 10)
		sub := eth.BlockChain().SubscribeChainHeadEvent(ch)
		defer sub.Unsubscribe()
		for {
			select {
			case <-task.ctx.Done():
				// target achieved
				close(task.targetAchievedCh)
				return
			case ev := <-ch:
				if targetAchieved(ev.Block) {
					// target achieved
					close(task.targetAchievedCh)
					return
				}
			}

		}
	}()
	return task
}

func (t baseTask) TargetAchievedCh() chan interface{} {
	return t.targetAchievedCh
}

func (t baseTask) interrupt() {

}

// BudgetTask is used to control miner to mine according to a budget
type BudgetTask struct {
	baseTask
	eth               Ethereum
	targetBlockNumber *big.Int
}

func NewBudgetTask(ctx context.Context, eth Ethereum, budget int64) *BudgetTask {
	targetNumber := big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget))
	task := &BudgetTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			return block.Number().Cmp(targetNumber) >= 0
		}),
		eth:               eth,
		targetBlockNumber: targetNumber,
	}

	return task
}

func (t *BudgetTask) OnNewMiningWork() {
	return
}

func (t *BudgetTask) String() string {
	return fmt.Sprintf("BudgetTask(%d)", t.targetBlockNumber.Int64())
}

func (t *BudgetTask) IsTxAllowed(txHash common.Hash) bool {
	return true
}

func (t *BudgetTask) OnTxError(txErrors map[common.Hash]error) {
	return
}

// BudgetTask is used to control miner to mine according to a budget
// do not allow any txs
type BudgetWithoutTxTask struct {
	baseTask
	eth               Ethereum
	targetBlockNumber *big.Int
}

func NewBudgetWithoutTxTask(ctx context.Context, eth Ethereum, budget int64) *BudgetWithoutTxTask {
	targetNumber := big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget))
	return &BudgetWithoutTxTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			return block.Number().Cmp(targetNumber) >= 0
		}),
		eth:               eth,
		targetBlockNumber: targetNumber,
	}
}

func (t *BudgetWithoutTxTask) OnNewMiningWork() {
	return
}

func (t *BudgetWithoutTxTask) String() string {
	return fmt.Sprintf("BudgetWithoutTxTask(%d)", t.targetBlockNumber.Int64())
}

func (t *BudgetWithoutTxTask) IsTxAllowed(txHash common.Hash) bool {
	return false
}

func (t *BudgetWithoutTxTask) OnTxError(txErrors map[common.Hash]error) {
	return
}

// BudgetTask is used to control miner to mine according to a budget
// allowing all txs but one specified tx
type BudgetExceptTxTask struct {
	baseTask
	eth               Ethereum
	targetBlockNumber *big.Int
	txHash            string
}

func NewBudgetExceptTxTask(ctx context.Context, eth Ethereum, budget int64, txHash string) *BudgetExceptTxTask {
	targetNumber := big.NewInt(0).Add(eth.BlockChain().CurrentBlock().Number(), big.NewInt(budget))
	return &BudgetExceptTxTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			return block.Number().Cmp(targetNumber) >= 0
		}),
		eth:               eth,
		targetBlockNumber: targetNumber,
		txHash:            txHash,
	}
}

func (t *BudgetExceptTxTask) OnNewMiningWork() {
	return
}

func (t *BudgetExceptTxTask) String() string {
	return fmt.Sprintf("BudgetExceptTxTask(%d, %s)", t.targetBlockNumber.Int64(), PrettifyHash(t.txHash))
}

func (t *BudgetExceptTxTask) IsTxAllowed(txHash common.Hash) bool {
	return t.txHash != txHash.String()
}

func (t *BudgetExceptTxTask) OnTxError(txErrors map[common.Hash]error) {
	return
}

// IntervalTask is used to repeatedly mine blocks with time interval
// TODO concurrency bug: mining process may deadlock when there are many concurrent txs
type IntervalTask struct {
	baseTask
	eth      Ethereum
	interval uint
	clock    <-chan time.Time
	stopCh   chan interface{}

	currentBlockNumber *big.Int
}

func NewIntervalTask(ctx context.Context, eth Ethereum, interval uint) *IntervalTask {
	stopCh := make(chan interface{}, 1)
	return &IntervalTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			// daemon task achieves target only when it is explicitly stopped
			select {
			case <-stopCh:
				return true
			default:
				return false
			}
		}),
		eth:                eth,
		interval:           interval,
		clock:              time.After(time.Duration(interval) * time.Millisecond),
		stopCh:             stopCh,
		currentBlockNumber: eth.BlockChain().CurrentBlock().Number(),
	}

}

func (t *IntervalTask) OnNewMiningWork() {
	bn := t.eth.BlockChain().CurrentBlock().Number()
	if bn.Cmp(t.currentBlockNumber) == 0 {
		return
	}
	select {
	case <-t.clock:
		t.clock = time.After(time.Duration(t.interval) * time.Millisecond)
		t.currentBlockNumber = t.eth.BlockChain().CurrentBlock().Number()
		return
	case <-t.stopCh:
		return
	}
}

func (t *IntervalTask) Stop() {
	close(t.stopCh)
}

func (t *IntervalTask) IsTxAllowed(txHash common.Hash) bool {
	return true
}

func (t *IntervalTask) OnTxError(txErrors map[common.Hash]error) {
	return
}

func (t *IntervalTask) String() string {
	return fmt.Sprintf("IntervalTask(%d)", t.interval)
}

// TxMonitorTask is a forever running task which mines blocks whenever there are tasks in txpool
// this is different from TxExecuteTask which only mines blocks until certain tx is mined
type TxMonitorTask struct {
	baseTask
	txPool *core.TxPool

	txCh   chan core.NewTxsEvent
	txSub  event.Subscription
	stopCh chan interface{}
}

func NewTxMonitorTask(ctx context.Context, eth Ethereum, txPool *core.TxPool) *TxMonitorTask {
	stopCh := make(chan interface{})
	task := &TxMonitorTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			// daemon task achieves target only when it is explicitly stopped
			select {
			case <-stopCh:
				return true
			default:
				return false
			}
		}),
		txPool: txPool,
		txCh:   make(chan core.NewTxsEvent),
		stopCh: stopCh,
	}
	task.txSub = task.txPool.SubscribeNewTxsEvent(task.txCh)
	return task
}

func (t *TxMonitorTask) OnNewMiningWork() {
	select {
	case <-t.stopCh:
		return
	default:
		txs, err := t.txPool.Pending()
		if err != nil {
			t.Stop()
			log.Error("retrieving tx from txPool error, stopping TxMonitor task", "err", err)
			return
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

				return
			} else {
				select {
				case <-t.txCh:
					time.Sleep(200 * time.Millisecond)
					return
				case <-t.stopCh:
					return
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

func (t *TxMonitorTask) OnTxError(txErrors map[common.Hash]error) {
	for txHash, err := range txErrors {
		log.Error("Tx execution error", "tx", txHash, "err", err)
	}
	log.Error("Stopping TxMonitor task")
	t.Stop()
	return
}

func (t *TxMonitorTask) String() string {
	return fmt.Sprintf("TxMonitorTask()")
}

// TxExecuteTask is used to control miner to execute a transaction
type TxExecuteTask struct {
	baseTask
	eth               Ethereum
	targetTransaction *types.Transaction
	err               error
}

func NewTxExecuteTask(ctx context.Context, eth Ethereum, targetTransaction *types.Transaction) *TxExecuteTask {
	task := &TxExecuteTask{
		eth:               eth,
		targetTransaction: targetTransaction,
		err:               nil,
	}
	task.baseTask = newBaseTask(ctx, eth, func(block *types.Block) bool {
		if task.targetTransaction == nil {
			return true
		}
		if task.err != nil {
			return true
		}
		for _, tx := range block.Transactions() {
			if tx.Hash() == targetTransaction.Hash() {
				return true
			}
		}
		return false
	})
	return task
}

func (t *TxExecuteTask) OnNewMiningWork() {
	return
}

func (t *TxExecuteTask) IsTxAllowed(txHash common.Hash) bool {
	if t.targetTransaction == nil {
		return false
	}
	if t.targetTransaction.Hash() == txHash {
		return true
	} else {
		// if there is tx from same sender but lower nonce, also allow those tx to execute
		pendings, _ := t.eth.TxPool().Pending()
		for _, txs := range pendings {
			inQueue := false
			for _, tx := range txs {
				if tx.Hash() == t.targetTransaction.Hash() {
					return inQueue
				} else if tx.Hash() == txHash {
					inQueue = true
				}
			}
		}
		return false
	}
}

func (t *TxExecuteTask) OnTxError(txErrors map[common.Hash]error) {
	if t.targetTransaction == nil {
		return
	}
	if err, ok := txErrors[t.targetTransaction.Hash()]; ok {
		log.Error("Tx execution error, stopping TxExecute task", "tx", t.targetTransaction.Hash(), "err", err)
		t.err = err
	}
}

func (t *TxExecuteTask) String() string {
	if t.targetTransaction == nil {
		return "TxExecuteTask(nil)"
	}
	return fmt.Sprintf("TxExecuteTask(%s)", PrettifyHash(t.targetTransaction.Hash().String()))
}

// a mining task to mine until total difficulty being larger than the given Td
// no tx is allowed
type TdTask struct {
	baseTask
	eth      Ethereum
	targetTd *big.Int
}

func NewTdTask(ctx context.Context, eth Ethereum, targetTd *big.Int) *TdTask {
	return &TdTask{
		baseTask: newBaseTask(ctx, eth, func(block *types.Block) bool {
			return eth.BlockChain().GetTdByHash(block.Hash()).Cmp(targetTd) > 0
		}),
		eth:      eth,
		targetTd: targetTd,
	}
}

func (t *TdTask) OnNewMiningWork() {
	return
}

func (t *TdTask) IsTxAllowed(txHash common.Hash) bool {
	return false
}

func (t *TdTask) OnTxError(txErrors map[common.Hash]error) {
	return
}

func (t *TdTask) String() string {
	return fmt.Sprintf("TdTask(%d)", t.targetTd.Int64())
}

// a placeholder empty task, allowing no tx, always shouldn't continue
// a instance of this should be returned when currentTask is nil
type nilTask struct {
}

var NilTask = &nilTask{}

func (t *nilTask) TargetAchievedCh() chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}

func (t *nilTask) OnNewMiningWork() {
	log.Warn("Nil mining task, please specify")
	return
}

func (t *nilTask) String() string {
	return "NilTask()"
}

func (t *nilTask) IsTxAllowed(txHash common.Hash) bool {
	return false
}

func (t *nilTask) OnTxError(txErrors map[common.Hash]error) {
	return
}
