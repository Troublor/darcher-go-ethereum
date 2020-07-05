package master

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	log "github.com/inconshreveable/log15"
)

// Traverser is used to traverse the lifecycle of a Transaction
// Traverser is for a single tx, and should be initialized when tx is on PENDING state
type Traverser struct {
	cluster    *Cluster
	controller TxController

	// tx information
	tx *Transaction
}

func NewTraverser(cluster *Cluster, tx *Transaction, controller TxController) *Traverser {
	traverser := &Traverser{
		controller: controller,
		cluster:    cluster,

		// tx information
		tx: tx,
	}

	traverser.controller.TxReceivedHook(tx.Hash())
	go traverser.notifyTxStateChangeLoop()

	return traverser
}

func (t *Traverser) notifyTxStateChangeLoop() {
	changeCh := make(chan txStateChange, 10)
	sub := t.tx.SubscribeStateChange(changeCh)
	defer sub.Unsubscribe()
	for {
		select {
		case change := <-changeCh:
			t.controller.OnStateChange(t.tx.Hash(), change.From, change.To)
			if t.tx.HasFinalized() {
				return
			}
		}
	}
}

/**
Start doing traverse on the Transaction
*/
func (t *Traverser) ResumeTraverse() {
	// check if this traverser(tx) has already been processed or not
	if t.tx.HasFinalized() {
		log.Error("Transaction is already processed, resume traverse failed", "tx", t.tx.PrettyHash())
		return
	}

	// this may block, controlled by the controller
	t.controller.TxResumeHook(t.tx.Hash())

	// loop until the tx is confirmed or suspend
	for !t.tx.HasFinalized() {
		nextState, suspend := t.controller.PivotReachedHook(t.tx.Hash(), t.tx.State())
		if suspend {
			// suspend the tx life cycle traverse
			return
		}
		err := t.transitStateTo(nextState)
		if err != nil {
			log.Error("Tx transit state error", "tx", t.tx.PrettyHash(), "err", err)
			return
		}
	}

	// this may block, controlled by the controller
	t.controller.TxFinishedHook(t.tx.Hash())
	log.Info("Tx traverse finished", "tx", t.tx.PrettyHash())
}

func (t *Traverser) transitStateTo(state common.LifecycleState) error {
	switch t.tx.State() {
	case common.CREATED:
		return t.transitFromCreatedTo(state)
	case common.PENDING:
		return t.transitFromPendingTo(state)
	case common.EXECUTED:
		return t.transitFromExecutedTo(state)
	case common.CONFIRMED:
		log.Warn("Tx is already confirmed, not transiting state", "tx", t.tx.Hash())
		return nil
	case common.DROPPED:
		log.Warn("Tx is already dropped, not transiting state", "tx", t.tx.Hash())
		return nil
	}
	return nil
}

/**
This assumes tx is currently at created state
*/
func (t *Traverser) transitFromCreatedTo(state common.LifecycleState) error {
	switch state {
	case common.CREATED:
		return nil
	case common.PENDING:
		return t.Schedule()
	case common.EXECUTED:
		return common.ExecuteSerially([]func() error{
			t.Schedule,
			t.Execute,
		})
	case common.CONFIRMED:
		return common.ExecuteSerially([]func() error{
			t.Schedule,
			t.Execute,
			t.Confirm,
			t.Synchronize,
		})
	case common.DROPPED:
		return common.ExecuteSerially([]func() error{
			t.Schedule,
			t.Drop,
		})
	}
	return nil
}

/**
This assumes tx is currently at pending state
*/
func (t *Traverser) transitFromPendingTo(state common.LifecycleState) error {
	switch state {
	case common.CREATED:
		log.Warn("tx cannot transit state from pending to created", "tx", t.tx.Hash())
		return nil
	case common.PENDING:
		return nil
	case common.EXECUTED:
		return t.Execute()
	case common.CONFIRMED:
		return common.ExecuteSerially([]func() error{
			t.Execute,
			t.Confirm,
			t.Synchronize,
		})
	case common.DROPPED:
		return t.Drop()
	}
	return nil
}

func (t *Traverser) transitFromExecutedTo(state common.LifecycleState) error {
	switch state {
	case common.CREATED:
		log.Warn("tx cannot transit state from executed to created", "tx", t.tx.Hash())
		return nil
	case common.PENDING:
		return t.Revert()
	case common.CONFIRMED:
		return common.ExecuteSerially([]func() error{
			t.Confirm,
			t.Synchronize,
		})
	case common.EXECUTED:
		return nil
	case common.DROPPED:
		return common.ExecuteSerially([]func() error{
			t.Revert,
			t.Drop,
		})
	}
	return nil
}

// schedule the tx, tell geth to return tx to dApp
func (t *Traverser) Schedule() error {
	doneCh, errCh := t.cluster.ScheduleTxAsyncQueued(rpc.Role_DOER, t.tx.Hash())
	select {
	case <-doneCh:
	case err := <-errCh:
		log.Error("Schedule tx error", "err", err, "tx", t.tx.PrettyHash())
		return err
	}
	t.tx.Schedule()
	return nil
}

// execute the Transaction
func (t *Traverser) Execute() error {
	doneCh, errCh := t.cluster.MineTxAsyncQueued(rpc.Role_DOER, t.tx.Hash())
	select {
	case block := <-doneCh:
		log.Info("Transaction has been executed on Doer", "tx", t.tx.PrettyHash(), "number", block.Number)
	case err := <-errCh:
		log.Error("Schedule tx error", "err", err, "tx", t.tx.PrettyHash())
		return err
	}
	t.tx.WaitForState(common.EXECUTED)
	return nil
}

// revert the Transaction by blockchain reorganization
func (t *Traverser) Revert() error {
	doneCh, errCh := t.cluster.ReorgAsyncQueued()
	select {
	case <-doneCh:
		t.tx.WaitForState(common.PENDING)
		log.Info("Blockchain reorganization happens")
	case err := <-errCh:
		log.Error("Revert tx failed", "tx", t.tx.PrettyHash(), "err", err)
		return err
	}
	return nil
}

func (t *Traverser) Confirm() error {
	count := (t.tx.executedBlock.Number + common.ConfirmationsCount) - t.cluster.GetDoerCurrentHead().GetNumber()
	doneCh, errCh := t.cluster.MineBlocksWithoutTxAsyncQueued(rpc.Role_DOER, count)
	select {
	case <-doneCh:
		t.tx.WaitForState(common.CONFIRMED)
		log.Info("Transaction confirmed", "tx", t.tx.PrettyHash(), "confirmations", common.ConfirmationsCount)
	case err := <-errCh:
		log.Error("Confirm tx failed", "tx", t.tx.PrettyHash(), "err", err)
		return err
	}
	return nil
}

/**
Assumes tx is at pending state
*/
func (t *Traverser) Drop() error {
	panic("not implemented")
}

func (t *Traverser) Synchronize() error {
	doneCh, errCh := t.cluster.SynchronizeAsyncQueued()
	select {
	case <-doneCh:
	case err := <-errCh:
		return err
	}
	return nil
}
