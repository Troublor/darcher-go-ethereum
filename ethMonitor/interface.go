package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/p2p"
)

type Ethereum interface {
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	StartMining(threads int) error
	StopMining()
	ChainDb() ethdb.Database
}

type Node interface {
	Server() *p2p.Server
}

type Stoppable interface {
	Stop()
}

type Task interface {
	ShouldContinue() bool
	fmt.Stringer
	IsTxAllowed(txHash common.Hash) bool
}

type DaemonTask interface {
	Task
	Stoppable
}

type ProtocolManager interface {
	Synchronise()
}
