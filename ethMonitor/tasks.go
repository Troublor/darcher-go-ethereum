package ethMonitor

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
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
