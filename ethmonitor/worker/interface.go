package worker

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
	OnTxError(txErrors map[common.Hash]error)
	// returns a channel which is used to signal whether the task target has been achieved or not
	TargetAchievedCh() chan interface{}
	// a hook function, called before whenever a new mining task is committed
	OnNewMiningWork()
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
