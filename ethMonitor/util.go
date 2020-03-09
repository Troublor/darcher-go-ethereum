package ethMonitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
)

func inTxPool(txHash common.Hash, txPool *core.TxPool) bool {
	return txPool.Get(txHash) != nil
}
