package common

const EventCacheSize = 50

/*
Events used when communicating with geth node
*/
// NewNodeEvent is fired when a new geth node has started
type NewNodeEvent struct {
	Role        Role
	ServerPort  int
	URL         string
	BlockNumber uint64
	BlockHash   string
	Td          uint64
}

// NewTxEvent is fired when a geth node has got a new tx in its txPool
type NewTxEvent struct {
	Role   Role
	Hash   string
	Sender string
	Nonce  uint64
}

// NewChainHeadEvent is fired when a geth node has got a new chain head
type NewChainHeadEvent struct {
	Role   Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

// NewChainSideEvent is fired when a geth node has got a new chain side
type NewChainSideEvent struct {
	Role   Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

// PeerAddEvent is fired when a geth node has added a new peer
type PeerAddEvent struct {
	Role Role
}

// PeerRemoveEvent is fired when a geth node has removed a peer
type PeerRemoveEvent struct {
	Role Role
}

/*
Events used when communicating with upper controller
*/
// TraverseTxEvent is fired when notifier get message from upper controller to start to traverse a certain tx's lifecycle
type TraverseTxEvent struct {
	Hash string
}

type NodeUpdateEvent struct {
	Number uint64
	Td     uint64
}

// TxExecutedEvent is fired when a geth node has executed a certain tx
type TxExecutedEvent struct {
	Role        Role
	TxHash      string
	BlockHash   string
	BlockNumber uint64
	Td          uint64
}
