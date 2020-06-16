package ethMonitor

type Role string

const (
	DOER   Role = "DOER"
	TALKER Role = "TALKER"
)

type Reply struct {
	Err error
}

type NodeStartMsg struct {
	Role        Role
	ServerPort  int
	URL         string
	BlockNumber uint64
	BlockHash   string
	Td          uint64
}

type NewTxMsg struct {
	Role Role
	Hash string
}

type NewChainHeadMsg struct {
	Role   Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

type NewChainSideMsg struct {
	Role   Role
	Hash   string
	Number uint64
	Td     uint64
	Txs    []string
}

type AddPeerMsg struct {
	URL string
}

type PeerAddMsg struct {
	Role   Role
	PeerID string
}

type RemovePeerMsg struct {
	URL string
}

type PeerRemoveMsg struct {
	Role   Role
	PeerID string
}

type MineBlocksMsg struct {
	Count uint64
}

type MineBlocksExceptTxMsg struct {
	Count  uint64
	TxHash string
}

type MineTdMsg struct {
	Td uint64
}

type MineTxMsg struct {
	Hash string
}

type ExecuteTxMsg struct {
	Hash string
}

type RevertTxMsg struct {
	Td uint64
}

type CheckTxInPoolMsg struct {
	Hash string
}

type CheckTxInPoolReply struct {
	Reply
	InPool bool
}

type ScheduleTxMsg struct {
	Hash string
}
