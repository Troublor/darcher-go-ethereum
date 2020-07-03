package rpc

import "github.com/ethereum/go-ethereum/ethmonitor/master/common"

type Reply struct {
	Err error
}

/*
Messages used when communicate with geth node
*/
// NodeStartMsg is sent from geth node when the node starts
type NodeStartMsg struct {
	Role        common.Role
	ServerPort  int
	URL         string
	BlockNumber uint64
	BlockHash   string
	Td          uint64
}

// NewTxMsg is sent from geth node when the node gets a new Tx in the txPool
type NewTxMsg struct {
	Role   common.Role
	Hash   string
	Sender string
	Nonce  uint64
}

// NewChainHeadMsg is sent from geth node when the node gets a new chain head
type NewChainHeadMsg struct {
	Role   common.Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

// NewChainSideMsg is sent from geth node when the node gets a new chain side
type NewChainSideMsg struct {
	Role   common.Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

// AddPeerMsg is sent to geth node to make it add peer with url
type AddPeerMsg struct {
	URL string
}

// PeerAddMsg is sent from geth node when the node successfully adds a peer
type PeerAddMsg struct {
	Role   common.Role
	PeerID string
}

// RemovePeerMsg is sent to geth node to make it remove peer with url
type RemovePeerMsg struct {
	URL string
}

// PeerRemoveMsg is sent from geth node when the node successfully removes a peer
type PeerRemoveMsg struct {
	Role   common.Role
	PeerID string
}

// MineBlocksMsg is sent to geth node to make it mine a certain amount of blocks
type MineBlocksMsg struct {
	Count uint64
}

// MineBlocksMsg is sent to geth node to make it mine a certain amount of blocks
type MineBlocksExceptTxMsg struct {
	Count  uint64
	TxHash string
}

// MineTdMsg is sent to geth node to make it mine blocks until a certain total difficulty
type MineTdMsg struct {
	Td uint64
}

// MineTxMsg is sent to geth node make it mine blocks until a certain tx is executed
type MineTxMsg struct {
	Hash string
}

// CheckTxInPoolMsg is sent to geth node whether a certain tx is in its txPool
type CheckTxInPoolMsg struct {
	Hash string
}

// CheckTxInPoolReply is the reply of CheckTxInPoolMsg from geth node
type CheckTxInPoolReply struct {
	Reply
	InPool bool
}

type ScheduleTxMsg struct {
	Hash string
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
