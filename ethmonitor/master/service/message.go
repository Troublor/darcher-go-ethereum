package service

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
)

type Reply struct {
	Err error
}

/*
Messages used when communicate with upper controller (dArcher)
*/
type JsonRpcMsg struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// TraverseTxMsg is sent from upper controller to start traversing the lifecycle of a certain tx
type TxReceivedMsg struct {
	Hash string `json:"Hash"`
}

// TxStateMsg is sent to upper controller when a certain tx has reached a certain state in the lifecycle
type TxStateChangeMsg struct {
	Hash         string                `json:"Hash"`
	CurrentState common.LifecycleState `json:"CurrentState"`
	NextState    common.LifecycleState `json:"NextState"`
}

// TxTraverseMsg is sent to upper controller to notify that a certain tx's lifecycle has been traversed
type TxFinishedMsg struct {
	Hash string `json:"Hash"`
}

type TxStartMsg struct {
	Hash string `json:"Hash"`
}

type PivotReachedMsg struct {
	Hash         string                `json:"Hash"`
	CurrentState common.LifecycleState `json:"CurrentState"`
	NextState    common.LifecycleState `json:"NextState"`
}

type TestMsg struct {
	Msg string `json:"Msg"`
}

type SelectTxToTraverseMsg struct {
	Hashes []string
}

type SelectTxToTraverseReply struct {
	Reply
	Hash string
}
