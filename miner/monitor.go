package miner

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Role int

const (
	Doer Role = iota
	Talker
)

type MiningMode int

const (
	BudgetMode MiningMode = iota
	TxExecuteMode
	TxRevertMode
	TxConfirmMode
)

type Monitor struct {
	role  Role
	miner *Miner
	mode  MiningMode

	// used for BudgetMode
	targetBlockNumber *big.Int

	// used for TxExecuteMode, TxRevertMode, TxConfirmMode
	targetTransactionHash *common.Hash
}

func NewMonitor(miner *Miner) *Monitor {
	role := Doer
	return &Monitor{
		role:  role,
		miner: miner,
		mode:  BudgetMode,

		targetBlockNumber: big.NewInt(0),
	}
}

func (m *Monitor) Start() {
}

func (m *Monitor) Stop() {
	m.miner.Stop()
}

func (m *Monitor) SetMiningMode(mode MiningMode) {
	m.mode = mode
}

func (m *Monitor) ShouldContinue() bool {
	switch m.mode {
	case BudgetMode:
		return m.miner.eth.BlockChain().CurrentBlock().Number().Cmp(m.targetBlockNumber) < 0
	case TxExecuteMode:

		if m.targetTransactionHash == nil {
			return false
		}
		return m.miner.eth.TxPool().Get(*m.targetTransactionHash) != nil
	default:
		return false
	}
}

func (m *Monitor) GetBudget(currentBlockNumber *big.Int) uint64 {
	if currentBlockNumber.Cmp(m.targetBlockNumber) >= 0 {
		return 0
	} else {
		return big.NewInt(0).Sub(m.targetBlockNumber, currentBlockNumber).Uint64()
	}
}

func (m *Monitor) SetBudget(currentBlockNumber *big.Int, n uint64) uint64 {
	m.targetBlockNumber.Add(currentBlockNumber, big.NewInt(int64(n)))
	return m.targetBlockNumber.Uint64()
}

func (m *Monitor) checkBudget(currentBlockNumber *big.Int) bool {
	return currentBlockNumber.Cmp(m.targetBlockNumber) < 0
}

func (m *Monitor) SetTargetTransactionHash(txHash *common.Hash) {
	m.targetTransactionHash = txHash
}

func (m *Monitor) checkTransaction() bool {
	if m.targetTransactionHash == nil {
		return false
	}
	return m.miner.eth.TxPool().Get(*m.targetTransactionHash) != nil
}
