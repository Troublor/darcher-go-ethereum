package worker

import (
	"sync"
)

type TxScheduler interface {
	WaitForTurn(hash string)
	ScheduleTx(hash string)
}

/**
This struct is used to schedule new transactions, when a new transaction is submitted, the SubmitTransaction RPC will
not return immediately, it will return until it is scheduled by this scheduler.

This scheduler is a slave worker of the dArcher transaction scheduler, it will scheduler tx only when the tx is
scheduled in the dArcher

By doing so, the dApp will get the transaction hash only when the transaction is scheduled to traverse lifecycle
*/
type txScheduler struct {
	// this is queue is the set of transactions waiting for being scheduled, txHash as string is stored as key in order to
	// compare with those in dArcher, the value of the map is a channel to signal the corresponding tx that it is scheduled
	// the channel should be closed when tx is scheduled instead of send value, in case multiple goroutine is waiting for it.
	queue map[string]chan interface{}
	mutex sync.Mutex
}

func newTxScheduler() *txScheduler {
	return &txScheduler{
		queue: make(map[string]chan interface{}),
	}
}

/**
This method will block, until the given transaction is scheduled to traverse lifecycle
*/
func (s *txScheduler) WaitForTurn(hash string) {
	var ch chan interface{}
	s.mutex.Lock()
	var ok bool
	if ch, ok = s.queue[hash]; !ok {
		ch = make(chan interface{})
		s.queue[hash] = ch
	}
	// in case the transaction is already in the queue, just wait for the existing channel
	s.mutex.Unlock()
	<-ch
}

/**
This method should be called by dArcher through RPC to schedule the given transaction
*/
func (s *txScheduler) ScheduleTx(hash string) {
	s.mutex.Lock()
	if ch, ok := s.queue[hash]; ok {
		close(ch)
	} else {
		// this case may happen when ScheduleTx is called before WaitForTurn.
		// In this case, we just prepare a already closed channel so that WaitForTurn will not block any more
		ch = make(chan interface{})
		s.queue[hash] = ch
		close(ch)
	}
	s.mutex.Unlock()
}

type fakeTxScheduler struct {
}

var FakeTxScheduler = fakeTxScheduler{}

func (s fakeTxScheduler) WaitForTurn(hash string) {
	return
}

func (s fakeTxScheduler) ScheduleTx(hash string) {
	return
}
