package eth

import (
	"errors"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/master/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

// cluster is a cluster of geth nodes, which act as a whole, can do mining, reorg and so on
/**
the action that a cluster can do is:
1. control the geth nodes to mine blocks, execute transactions, and reorg (as RPC client)
2. monitor the status of the geth nodes (as RPC server)
*/
/*
!!! IMPORTANT

For now, cluster can only handle one single transaction at the same time,
this is because the instrumented geth node cannot executed transactions one by one. Whenever mining blocks,
it will execution all transactions in the tx pool.
So we make a restriction on the Cluster here that when cluster traversing lifecycle of transactions,
it can only handle one transaction at one time, if other transactions are submitted during the traverse,
Cluster will ignore them, and these new transactions will be randomly executed, reverted and confirmed.
*/

/**
ClusterConfig provides some settings of Cluster (e.g. the confirmation number)
*/
type ClusterConfig struct {
	ConfirmationNumber uint64
	ServerPort         int
}

type ClusterTask = func()

var (
	TxNotInPoolErr = errors.New("transaction not in tx pool")
	TimeoutErr     = errors.New("timeout")
)

type Cluster struct {
	server *rpc.Server // the rpc server that monitor the status of geth nodes
	doer   *Node       // the rpc client to the doer geth node
	talker *Node       // the rpc client to the talker geth node
	config ClusterConfig

	/** serialize the control task on geth nodes */
	taskQueue *queue.Queue
}

/**
the constructor of Cluster, which will set:
1. the server port of cluster


*/
func NewCluster(config ClusterConfig) *Cluster {
	return &Cluster{
		config:    config,
		taskQueue: &queue.Queue{},
	}
}

/**
Start the cluster will do the following things
1. start the rpc server which monitor the geth nodes
2. start doer and talker and wait for their connection
Upon the start of geth nodes, each node should provide their:
1. control port (the rpc server)
2. rpc port (for ethereum JSON-RPC)

Start() takes the config value and a boolean telling whether cluster should automatically start doer and talker geth node or not
*/
func (c *Cluster) Start(integrated bool) {
	// initiate rpc server and start it
	server := rpc.NewServer(c.config.ServerPort)
	server.Start()
	c.server = server

	if integrated {
		// TODO start doer and talker node
	}

	// wait for doer and talker to connect
	var newNodeCh chan common.NewNodeEvent
	var newNodeSub event.Subscription
	newNodeCh = make(chan common.NewNodeEvent, common.EventCacheSize)
	newNodeSub = c.server.SubscribeNewNodeEvent(newNodeCh)
	defer newNodeSub.Unsubscribe()
	log.Info("Waiting for nodes to connect")
	// TODO not implement receive Ethereum JSON-RPC port from geth node
	for i := 0; i < 2; i++ {
		// wait for node to connect
		ev := <-newNodeCh
		initBlock := &common.Block{Number: ev.BlockNumber, Hash: ev.BlockHash, Td: ev.Td}
		switch ev.Role {
		case common.DOER:
			c.doer = NewNode(ev.Role, ev.URL, "127.0.0.1", ev.ServerPort, initBlock, c.server)
		case common.TALKER:
			c.talker = NewNode(ev.Role, ev.URL, "127.0.0.1", ev.ServerPort, initBlock, c.server)
		}
		log.Info("Node connected", "role", ev.Role)
	}

	log.Info("Both doer and talker connected, start syncing")

	go c.taskLoop()

	c.flushTxs()
	done, err := c.SynchronizeAsyncQueued()
	select {
	case <-done:
	case <-err:
		log.Error("Initial synchronization timeout")
		return
	}
	log.Info("Sync finished, start listening for transactions")
}

func (c *Cluster) taskLoop() {
	for {
		tasks, err := c.taskQueue.Get(1)
		if err != nil {
			log.Error("Fetch cluster task error", "err", err)
			continue
		}
		task := tasks[0].(ClusterTask)
		task()
	}
}

/**
subscribe the new transaction event of geth node
*/
func (c *Cluster) SubscribeNewTxEvent(ch chan<- common.NewTxEvent) event.Subscription {
	return c.server.SubscribeNewTxEvent(ch)
}

func (c *Cluster) SubscribeNewChainHeadEvent(ch chan<- common.NewChainHeadEvent) event.Subscription {
	return c.server.SubscribeNewChainHeadEvent(ch)
}

func (c *Cluster) SubscribeNewChainSideEvent(ch chan<- common.NewChainSideEvent) event.Subscription {
	return c.server.SubscribeNewChainSideEvent(ch)
}

func (c *Cluster) flushTxs() {
	count := common.ConfirmationsCount + 1
	targetNum := count + c.CurrentBlock().Number
	log.Debug("Flush txs", "currentNum", c.CurrentBlock().Number, "targetNum", targetNum)
	doneCh, errCh := c.MineBlocksAsyncQueued(count)
	select {
	case <-doneCh:
	case err := <-errCh:
		log.Error("Flush txs timeout", "err", err)
		return
	}
}

/* getters */

/**
The block before which doer and talker are synchronized
*/
func (c *Cluster) SynchronizedBlock() common.Block {
	return *c.talker.currentBlock
}

/**
The latest block on doer (we hide talker from outside the package)
*/
func (c *Cluster) CurrentBlock() common.Block {
	return *c.doer.currentBlock
}

/* sync methods */
func (c *Cluster) waitUntilSynchronized(timeout time.Duration) (common.Block, error) {
	var doerUpdateCh chan common.NodeUpdateEvent
	var doerUpdateSub event.Subscription
	doerUpdateCh = make(chan common.NodeUpdateEvent, common.EventCacheSize)
	doerUpdateSub = c.doer.SubscribeNodeUpdateEvent(doerUpdateCh)
	defer doerUpdateSub.Unsubscribe()
	var talkerUpdateCh chan common.NodeUpdateEvent
	var talkerUpdateSub event.Subscription
	talkerUpdateCh = make(chan common.NodeUpdateEvent, common.EventCacheSize)
	talkerUpdateSub = c.talker.SubscribeNodeUpdateEvent(talkerUpdateCh)
	defer talkerUpdateSub.Unsubscribe()

	// short circuit if already synchronized
	if c.doer.currentBlock.Td == c.talker.currentBlock.Td {
		log.Info("Peers already Synchronized", "number", c.doer.currentBlock.Number, "td", c.doer.currentBlock.Td)
		return *c.doer.currentBlock, nil
	}
	// wait until doer and talker are synchronized
wait:
	for {
		select {
		case <-time.After(timeout):
			return common.Block{}, TimeoutErr
		case ev := <-doerUpdateCh:
			if ev.Td == c.talker.currentBlock.Td {
				break wait
			}
		case ev := <-talkerUpdateCh:
			if ev.Td == c.doer.currentBlock.Td {
				break wait
			}
		}
	}
	log.Info("Peers synchronized", "number", c.doer.currentBlock.Number, "td", c.doer.currentBlock.Td)
	return *c.doer.currentBlock, nil
}

func (c *Cluster) waitUntilBlock(num uint64, timeout time.Duration) (common.Block, error) {
	var doerUpdateCh chan common.NodeUpdateEvent
	var doerUpdateSub event.Subscription
	doerUpdateCh = make(chan common.NodeUpdateEvent, common.EventCacheSize)
	doerUpdateSub = c.doer.SubscribeNodeUpdateEvent(doerUpdateCh)
	defer doerUpdateSub.Unsubscribe()
wait:
	for {
		select {
		case <-time.After(timeout):
			return common.Block{}, TimeoutErr
		case ev := <-doerUpdateCh:
			if ev.Number >= num {
				break wait
			}
		}
	}
	return *c.doer.currentBlock, nil
}

func (c *Cluster) waitUntilTd(td uint64, timeout time.Duration) (common.Block, error) {
	var talkerUpdateCh chan common.NodeUpdateEvent
	var talkerUpdateSub event.Subscription
	talkerUpdateCh = make(chan common.NodeUpdateEvent, common.EventCacheSize)
	talkerUpdateSub = c.talker.SubscribeNodeUpdateEvent(talkerUpdateCh)
	defer talkerUpdateSub.Unsubscribe()

	// short circuit if talker's Td is already larger than target Td
	if c.talker.currentBlock.Td > td {
		return *c.talker.currentBlock, nil
	}
wait:
	for {
		select {
		case <-time.After(timeout):
			return common.Block{}, TimeoutErr
		case ev := <-talkerUpdateCh:
			if ev.Td > td {
				break wait
			}
		}
	}
	return *c.talker.currentBlock, nil
}

// returns the Td of the block including Tx
func (c *Cluster) waitUntilTxExecuted(hash string, timeout time.Duration) (common.Block, error) {
	var txExecutedCh chan common.TxExecutedEvent
	var txExecutedSub event.Subscription
	txExecutedCh = make(chan common.TxExecutedEvent, common.EventCacheSize)
	txExecutedSub = c.doer.SubscribeTxExecutedEvent(txExecutedCh)
	defer txExecutedSub.Unsubscribe()
	// TODO short circuit if tx is already executed
wait:
	for {
		select {
		case <-time.After(timeout):
			return common.Block{}, TimeoutErr
		case ev := <-txExecutedCh:
			if ev.TxHash == hash {
				break wait
			}
		}
	}
	return *c.doer.currentBlock, nil
}

func (c *Cluster) waitUntilPeerAdded(timeout time.Duration) error {
	var peerAddCh chan common.PeerAddEvent
	var peerAddSub event.Subscription
	peerAddCh = make(chan common.PeerAddEvent, common.EventCacheSize)
	peerAddSub = c.server.SubscribePeerAddEvent(peerAddCh)
	defer peerAddSub.Unsubscribe()
wait1:
	for {
		select {
		case <-time.After(timeout):
			return TimeoutErr
		case <-peerAddCh:
			break wait1
		}
	}
	return nil
}

func (c *Cluster) waitUntilPeerRemoved(timeout time.Duration) error {
	var peerRemoveCh chan common.PeerRemoveEvent
	var peerRemoveSub event.Subscription
	peerRemoveCh = make(chan common.PeerRemoveEvent, common.EventCacheSize)
	peerRemoveSub = c.server.SubscribePeerRemoveEvent(peerRemoveCh)
	defer peerRemoveSub.Unsubscribe()
wait2:
	for {
		select {
		case <-time.After(timeout):
			return TimeoutErr
		case <-peerRemoveCh:
			break wait2
		}
	}
	return nil
}

func (c *Cluster) waitUntilTxInPool(txHash string, timeout time.Duration) error {
	newTxCh := make(chan common.NewTxEvent)
	sub := c.server.SubscribeNewTxEvent(newTxCh)
	defer sub.Unsubscribe()
	if c.doer.CheckTxInPool(txHash) {
		return nil
	}
	for {
		select {
		case <-time.After(timeout):
			return TxNotInPoolErr
		case ev := <-newTxCh:
			if ev.Role != common.DOER {
				continue
			}
			if ev.Hash == txHash {
				return nil
			}
		}
	}
}

/* integrated async methods */
/**
Synchronize the cluster and return the latest block in channel
*/
func (c *Cluster) synchronizeAsync() (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		defer func() {
			doneCh <- *c.doer.currentBlock
		}()
		log.Debug("Start synchronizing")
		// short circuit if already synchronized
		if c.doer.currentBlock.Td == c.talker.currentBlock.Td {
			log.Info("Short circuit when sync", "doer_td", c.doer.currentBlock.Td, "talker_td", c.talker.currentBlock.Td)
			doneCh <- *c.doer.currentBlock
			return
		}

		c.doer.AddPeer(c.talker.url)
		err := c.waitUntilPeerAdded(common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Add peer timeout", "operation", "synchronize")
			errCh <- err
			return
		}

		block, err := c.waitUntilSynchronized(common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Synchronization peer timeout", "operation", "synchronize")
			errCh <- err
			return
		}

		c.doer.RemovePeer(c.talker.url)
		err = c.waitUntilPeerRemoved(common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Remove peer timeout", "operation", "synchronize")
			errCh <- err
			return
		}

		doneCh <- block
	}()
	return doneCh, errCh
}

/**
send synchronize task in the task queue
*/
func (c *Cluster) SynchronizeAsyncQueued() (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.synchronizeAsync()
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) scheduleTxAsync(hash string) (doneCh chan string, errCh chan error) {
	doneCh = make(chan string, 1)
	errCh = make(chan error, 1)
	go func() {
		if !c.doer.CheckTxInPool(hash) {
			errCh <- TxNotInPoolErr
			return
		}
		c.doer.geth.ScheduleTx(hash)
		c.talker.geth.ScheduleTx(hash)
		doneCh <- hash
	}()
	return doneCh, errCh
}

func (c *Cluster) ScheduleTxAsyncQueued(hash string) (doneCh chan string, errCh chan error) {
	doneCh = make(chan string, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.scheduleTxAsync(hash)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) mineTxAsync(hash string) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		if !c.doer.CheckTxInPool(hash) {
			errCh <- TxNotInPoolErr
			return
		}
		c.doer.MineTx(hash)
		block, err := c.waitUntilTxExecuted(hash, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Mine tx timeout", "operation", "mineTx")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) MineTxAsyncQueued(hash string) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.mineTxAsync(hash)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) mineTdAsync(td uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		c.doer.MineTd(td)
		block, err := c.waitUntilTd(td, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Mine td timeout", "operation", "mineTd")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) MineTdAsyncQueued(td uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.mineTdAsync(td)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) mineBlocksAsync(count uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		targetNumber := c.doer.currentBlock.Number + count
		c.doer.MineBlocks(count)
		block, err := c.waitUntilBlock(targetNumber, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Mine blocks timeout", "operation", "mineBlocks")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) MineBlocksAsyncQueued(count uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	err := c.taskQueue.Put(func() {
		d, e := c.mineBlocksAsync(count)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	if err != nil {
		errCh <- err
	}
	return doneCh, errCh
}

func (c *Cluster) mineBlocksWithoutTxAsync(count uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		targetNumber := c.doer.currentBlock.Number + count
		c.doer.MineBlocksWithoutTx(count)
		block, err := c.waitUntilBlock(targetNumber, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Mine blocks without tx timeout", "operation", "mineBlocksWithoutTx")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) MineBlocksWithoutTxAsyncQueued(count uint64) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.mineBlocksWithoutTxAsync(count)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) mineBlocksExceptTxAsync(count uint64, hash string) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		targetNumber := c.doer.currentBlock.Number + count
		c.doer.MineBlocksExceptTx(count, hash)
		block, err := c.waitUntilBlock(targetNumber, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Mine blocks without tx timeout", "operation", "mineBlocksWithoutTx")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) MineBlocksExceptTxAsyncQueued(count uint64, hash string) (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.mineBlocksExceptTxAsync(count, hash)
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) reorgAsync() (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	go func() {
		c.talker.MineTd(c.doer.currentBlock.Td)
		_, err := c.waitUntilTd(c.doer.currentBlock.Td, common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Talker mine td timeout", "operation", "reorg")
			errCh <- err
			return
		}
		dCh, eCh := c.synchronizeAsync()
		var block common.Block
		select {
		case block = <-dCh:
		case err := <-eCh:
			log.Error("Synchronization timeout", "operation", "reorg")
			errCh <- err
			return
		}
		doneCh <- block
	}()
	return doneCh, errCh
}

func (c *Cluster) ReorgAsyncQueued() (doneCh chan common.Block, errCh chan error) {
	doneCh = make(chan common.Block, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.reorgAsync()
		select {
		case data := <-d:
			doneCh <- data
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}
