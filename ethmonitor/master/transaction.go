package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"sync"
)

type txStateChange struct {
	From common.LifecycleState
	To   common.LifecycleState
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
	state         common.LifecycleState

	stateChangeFeed event.Feed
}

func NewTransaction(
	hash string,
	sender string,
	nonce uint64,
	newHeadCh chan *rpc.ChainHead,
	newSideCh chan *rpc.ChainSide,
	newTxCh chan *rpc.Tx,
) *Transaction {
	t := &Transaction{
		hash:      hash,
		sender:    sender,
		nonce:     nonce,
		state:     common.CREATED,
		newHeadCh: newHeadCh,
		newSideCh: newSideCh,
		newTxCh:   newTxCh,
	}
	go t.stateLoop()
	return t
}

func (t *Transaction) SubscribeStateChange(ch chan<- txStateChange) event.Subscription {
	return t.stateChangeFeed.Subscribe(ch)
}

func (t *Transaction) WaitForState(state common.LifecycleState) {
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
			case common.CREATED:
				fallthrough
			case common.PENDING:
				for _, tx := range ev.Txs {
					if tx == t.hash {
						t.executedBlock = ev
						t.setState(common.EXECUTED)
						break
					}
				}
			case common.EXECUTED:
				if ev.Hash != t.executedBlock.Hash && ev.Number <= t.executedBlock.Number {
					t.setState(common.PENDING)
				} else if ev.Number-t.executedBlock.Number >= common.ConfirmationsCount {
					t.setState(common.CONFIRMED)
					return
				}
			case common.CONFIRMED:
				return
			case common.DROPPED:
				return
			}
		case ev := <-t.newSideCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			switch t.State() {
			case common.CREATED:
				continue
			case common.PENDING:
				continue
			case common.EXECUTED:
				if ev.Hash != t.executedBlock.Hash && ev.Number == t.executedBlock.Number {
					t.setState(common.PENDING)
				}
				// TODO what if tx is executed in the incoming side chain?
			case common.CONFIRMED:
				return
			case common.DROPPED:
				return
			}
		case ev := <-t.newTxCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			if t.state == common.PENDING || t.state == common.CREATED {
				if ev.Hash != t.hash && ev.Sender == t.sender && ev.Nonce == t.nonce {
					t.setState(common.DROPPED)
				}
			}
		}
	}
}

func (t *Transaction) Schedule() {
	if t.State() == common.CREATED {
		t.setState(common.PENDING)
	}
}

func (t *Transaction) setState(state common.LifecycleState) {
	t.stateRWLock.Lock()
	defer t.stateRWLock.Unlock()
	from := t.state
	t.state = state
	t.stateChangeFeed.Send(txStateChange{
		From: from,
		To:   state,
	})
}

func (t *Transaction) State() common.LifecycleState {
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
	return t.State() == common.CONFIRMED || t.State() == common.DROPPED
}
