package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"strings"
)

func inTxPool(txHash common.Hash, txPool *core.TxPool) bool {
	return txPool.Get(txHash) != nil
}

func PrettifyHash(hash string) string {
	offset := 0
	if strings.Index(hash, "0x") == 0 {
		offset = 2
	}
	return fmt.Sprintf("%s...%s", hash[offset:offset+6], hash[len(hash)-6:])
}
