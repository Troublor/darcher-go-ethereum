package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"sync"
)

type txStateChange struct {
	From rpc.TxState
	To   rpc.TxState
}
type Transaction struct {
	hash   string
	sender string
	nonce  uint64

	newHeadCh chan *rpc.ChainHead
	newSideCh chan *rpc.ChainSide
	newTxCh   chan *rpc.Tx

	executedBlock *rpc.ChainHead
	stateRWLock   sync.RWMutex
	state         rpc.TxState

	stateChangeFeed event.Feed

	config EthMonitorConfig
}

func NewTransaction(
	hash string,
	sender string,
	nonce uint64,
	newHeadCh chan *rpc.ChainHead,
	newSideCh chan *rpc.ChainSide,
	newTxCh chan *rpc.Tx,
	config EthMonitorConfig,
) *Transaction {
	t := &Transaction{
		hash:      hash,
		sender:    sender,
		nonce:     nonce,
		state:     rpc.TxState_CREATED,
		newHeadCh: newHeadCh,
		newSideCh: newSideCh,
		newTxCh:   newTxCh,
		config:    config,
	}
	go t.stateLoop()
	return t
}

func (t *Transaction) SubscribeStateChange(ch chan<- txStateChange) event.Subscription {
	return t.stateChangeFeed.Subscribe(ch)
}

func (t *Transaction) WaitForState(state rpc.TxState) {
	stateCh := make(chan txStateChange)
	sub := t.SubscribeStateChange(stateCh)
	defer sub.Unsubscribe()
	if state == t.state {
		return
	}
	for {
		select {
		case change := <-stateCh:
			if change.To == state {
				return
			}
		}
	}
}

func (t *Transaction) stateLoop() {
	for {
		select {
		case ev := <-t.newHeadCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			switch t.State() {
			case rpc.TxState_CREATED:
				fallthrough
			case rpc.TxState_PENDING:
				for _, tx := range ev.Txs {
					if tx == t.hash {
						t.executedBlock = ev
						t.setState(rpc.TxState_EXECUTED)
						break
					}
				}
			case rpc.TxState_EXECUTED:
				if ev.Hash != t.executedBlock.Hash && ev.Number <= t.executedBlock.Number {
					t.setState(rpc.TxState_PENDING)
					return
				} else if ev.Number-t.executedBlock.Number >= t.config.ConfirmationNumber {
					t.setState(rpc.TxState_CONFIRMED)
					return
				}
			case rpc.TxState_CONFIRMED:
				return
			case rpc.TxState_DROPPED:
				return
			}
		case ev := <-t.newSideCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			switch t.State() {
			case rpc.TxState_CREATED:
				continue
			case rpc.TxState_PENDING:
				continue
			case rpc.TxState_EXECUTED:
				if ev.Hash != t.executedBlock.Hash && ev.Number == t.executedBlock.Number {
					t.setState(rpc.TxState_PENDING)
				}
				// TODO what if tx is executed in the incoming side chain?
			case rpc.TxState_CONFIRMED:
				return
			case rpc.TxState_DROPPED:
				return
			}
		case ev := <-t.newTxCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			if t.state == rpc.TxState_PENDING || t.state == rpc.TxState_CREATED {
				if ev.Hash != t.hash && ev.Sender == t.sender && ev.Nonce == t.nonce {
					t.setState(rpc.TxState_DROPPED)
				}
			}
		}
	}
}

func (t *Transaction) Schedule() {
	if t.State() == rpc.TxState_CREATED {
		t.setState(rpc.TxState_PENDING)
	}
}

func (t *Transaction) setState(state rpc.TxState) {
	t.stateRWLock.Lock()
	defer t.stateRWLock.Unlock()
	from := t.state
	t.state = state
	t.stateChangeFeed.Send(txStateChange{
		From: from,
		To:   state,
	})
}

func (t *Transaction) State() rpc.TxState {
	t.stateRWLock.RLock()
	defer t.stateRWLock.RUnlock()
	return t.state
}

func (t *Transaction) Hash() string {
	return t.hash
}

func (t *Transaction) PrettyHash() string {
	return common.PrettifyHash(t.hash)
}

func (t *Transaction) HasFinalized() bool {
	return t.State() == rpc.TxState_CONFIRMED || t.State() == rpc.TxState_DROPPED
}
