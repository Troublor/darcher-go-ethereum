package master

import (
	"errors"
	"github.com/Workiva/go-datastructures/queue"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/master/service"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

// cluster is a cluster of geth nodes, which act as a whole, can do mining, reorg and so on
/**
the action that a cluster can do is:
1. control the geth nodes to mine blocks, execute transactions, and reorg (as RPC client)
2. monitor the status of the geth nodes (as RPC legacyServer)
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

type ClusterTask = func()

var (
	TxNotInPoolErr = errors.New("transaction not in tx pool")
	TimeoutErr     = errors.New("timeout")
	RPCErr         = errors.New("errors happen inside rpc")
)

type Cluster struct {
	//legacyServer *rpc.EthMonitorServer // the rpc legacyServer that monitor the status of geth nodes
	server *service.EthMonitorServer
	config EthMonitorConfig

	doerUrl   string
	talkerUrl string

	/** serialize the control task on geth nodes */
	taskQueue *queue.Queue
}

/**
the constructor of Cluster, which will set:
1. the legacyServer port of cluster


*/
func NewCluster(config EthMonitorConfig) *Cluster {
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
1. control port (the rpc legacyServer)
2. rpc port (for ethereum JSON-RPC)

Start() takes the config value and a boolean telling whether cluster should automatically start doer and talker geth node or not
*/
func (c *Cluster) Start(integrated bool) {
	// initiate rpc server and start it
	c.server = service.NewServer(c.config.ServerPort)
	c.server.Start()

	if integrated {
		// TODO start doer and talker node
	}

	// wait for doer and talker to connect
	var newNodeCh chan *rpc.Node
	var newNodeSub event.Subscription
	newNodeCh = make(chan *rpc.Node, common.EventCacheSize)
	newNodeSub = c.server.P2PNetworkService().SubscribeNewNode(newNodeCh)
	defer newNodeSub.Unsubscribe()
	log.Info("Waiting for nodes to connect")
	// TODO not implement receive Ethereum JSON-RPC port from geth node
	for i := 0; i < 2; i++ {
		// wait for node to connect
		ev := <-newNodeCh
		switch ev.Role {
		case rpc.Role_DOER:
			c.doerUrl = ev.Url
		case rpc.Role_TALKER:
			c.talkerUrl = ev.Url
		}
		log.Info("Node connected", "role", ev.Role)
	}

	log.Info("Both doer and talker connected, start syncing")

	go c.taskLoop()

	c.flushTxs()
	doneCh, errCh := c.SynchronizeAsyncQueued()
	select {
	case <-doneCh:
	case err := <-errCh:
		log.Error("Initial synchronization error", "err", err)
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

func (c *Cluster) GetDoerCurrentHead() *rpc.ChainHead {
	doneCh, errCh := c.server.BlockchainStatusService().GetHead(rpc.Role_DOER)
	select {
	case controlMsg := <-doneCh:
		if controlMsg.Err != rpc.Error_NilErr {
			log.Error("GetDoerCurrentHead RPC error", "err", controlMsg.Err)
			return nil
		}
		return controlMsg.Head
	case err := <-errCh:
		log.Error("GetDoerCurrentHead error", "err", err)
	}
	return nil
}

func (c *Cluster) GetTalkerCurrentHead() *rpc.ChainHead {
	doneCh, errCh := c.server.BlockchainStatusService().GetHead(rpc.Role_TALKER)
	select {
	case controlMsg := <-doneCh:
		if controlMsg.Err != rpc.Error_NilErr {
			log.Error("GetTalkerCurrentHead RPC error", "err", controlMsg.Err)
			return nil
		}
		return controlMsg.Head
	case err := <-errCh:
		log.Error("GetTalkerCurrentHead error", "err", err)
	}
	return nil
}

func (c *Cluster) GetContractVulnerabilityReports(txHash string) []*rpc.ContractVulReport {
	doneCh, errCh := c.server.ContractOracleService().GetReportsByTransaction(rpc.Role_DOER, txHash)
	select {
	case reply := <-doneCh:
		return reply.GetReports()
	case err := <-errCh:
		log.Error("GetReportsByTransaction rpc error", "err", err)
	}
	return nil
}

/**
subscribe the new transaction event of geth node
*/
func (c *Cluster) SubscribeNewTx(ch chan<- *rpc.Tx) event.Subscription {
	return c.server.BlockchainStatusService().SubscribeNewTx(ch)
}

func (c *Cluster) SubscribeNewChainHead(ch chan<- *rpc.ChainHead) event.Subscription {
	return c.server.BlockchainStatusService().SubscribeNewChainHead(ch)
}

func (c *Cluster) SubscribeNewChainSide(ch chan<- *rpc.ChainSide) event.Subscription {
	return c.server.BlockchainStatusService().SubscribeNewChainSide(ch)
}

func (c *Cluster) SubscribeTxError(ch chan<- *rpc.TxErrorMsg) event.Subscription {
	return c.server.ContractOracleService().SubscribeTxError(ch)
}

func (c *Cluster) flushTxs() {
	count := c.config.ConfirmationNumber + 1
	targetNum := count + c.GetDoerCurrentHead().GetNumber()
	log.Debug("Flush txs", "currentNum", c.GetDoerCurrentHead().GetNumber(), "targetNum", targetNum)
	doneCh, errCh := c.MineBlocksAsyncQueued(rpc.Role_DOER, count)
	select {
	case <-doneCh:
	case err := <-errCh:
		log.Error("Flush txs timeout", "err", err)
		return
	}
}

/* sync methods */
func (c *Cluster) waitUntilSynchronized(timeout time.Duration) error {
	var updateCh chan *rpc.ChainHead
	var updateSub event.Subscription
	updateCh = make(chan *rpc.ChainHead, common.EventCacheSize)
	updateSub = c.server.BlockchainStatusService().SubscribeNewChainHead(updateCh)
	defer updateSub.Unsubscribe()

	current := c.GetDoerCurrentHead()
	// short circuit if already synchronized
	if current.GetTd() == c.GetTalkerCurrentHead().GetTd() {
		log.Info("Peers already Synchronized", "number", current.GetNumber(), "td", current.GetTd())
		return nil
	}
	// wait until doer and talker are synchronized
wait:
	for {
		select {
		case <-time.After(timeout):
			return TimeoutErr
		case ev := <-updateCh:
			switch ev.GetRole() {
			case rpc.Role_DOER:
				log.Debug("doer update", "doerNum", ev.GetNumber(), "doerTd", ev.GetTd(), "talkerNum", c.GetTalkerCurrentHead().GetNumber(), "talkerTd", c.GetTalkerCurrentHead().GetTd())
				if c.GetDoerCurrentHead().GetTd() == c.GetTalkerCurrentHead().GetTd() {
					current = ev
					break wait
				}
			case rpc.Role_TALKER:
				log.Debug("talker update", "talkerNum", ev.GetNumber(), "talkerTd", ev.GetTd(), "doerNum", c.GetDoerCurrentHead().GetNumber(), "doerTd", c.GetDoerCurrentHead().GetTd())
				if c.GetDoerCurrentHead().GetTd() == c.GetTalkerCurrentHead().GetTd() {
					current = ev
					break wait
				}
			}
		}
	}
	log.Info("Peers synchronized", "number", c.GetDoerCurrentHead().GetNumber(), "td", c.GetDoerCurrentHead().GetTd())
	return nil
}

func (c *Cluster) waitUntilTxInPool(txHash string, timeout time.Duration) error {
	newTxCh := make(chan *rpc.Tx)
	sub := c.server.BlockchainStatusService().SubscribeNewTx(newTxCh)
	defer sub.Unsubscribe()
	doneCh, errCh := c.server.MiningService().CheckTxInPool(rpc.Role_DOER, txHash)
	select {
	case msg := <-doneCh:
		if msg.GetErr() != rpc.Error_NilErr {
			log.Error("CheckTxInPool RPC error", "err", msg.GetErr())
			return RPCErr
		}
		if msg.GetInPool() {
			// short circuit if tx is already in pool
			return nil
		}
	case err := <-errCh:
		return err
	}
	// wait until tx in pool
	for {
		select {
		case <-time.After(timeout):
			return TimeoutErr
		case ev := <-newTxCh:
			if ev.GetRole() != rpc.Role_DOER {
				continue
			}
			if ev.Hash == txHash {
				return nil
			}
		}
	}
}

/* integrated async methods */
func (c *Cluster) connectPeers() (doneCh chan interface{}, errCh chan error) {
	doneCh = make(chan interface{}, 1)
	errCh = make(chan error, 1)
	go func() {
		adCh, eCh := c.server.P2PNetworkService().AddPeer(rpc.Role_DOER, c.talkerUrl)
		select {
		case msg := <-adCh:
			if msg.GetErr() != rpc.Error_NilErr {
				log.Error("AddPeer reverse RPC error", "err", msg.GetErr().String())
				errCh <- RPCErr
				return
			}
			close(doneCh)
		case err := <-eCh:
			errCh <- err
			return
		}
	}()
	return doneCh, errCh
}

func (c *Cluster) ConnectPeersQueued() (doneCh chan interface{}, errCh chan error) {
	doneCh = make(chan interface{}, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.connectPeers()
		select {
		case <-d:
			close(doneCh)
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) disconnectPeers() (doneCh chan interface{}, errCh chan error) {
	doneCh = make(chan interface{}, 1)
	errCh = make(chan error, 1)
	go func() {
		rdCh, eCh := c.server.P2PNetworkService().RemovePeer(rpc.Role_DOER, c.talkerUrl)
		select {
		case msg := <-rdCh:
			if msg.GetErr() != rpc.Error_NilErr {
				log.Error("RemovePeer reverse RPC error", "err", msg.GetErr().String())
				errCh <- RPCErr
				return
			}
			close(doneCh)
		case err := <-eCh:
			errCh <- err
			return
		}
	}()
	return doneCh, errCh
}

func (c *Cluster) DisconnectPeersQueued() (doneCh chan interface{}, errCh chan error) {
	doneCh = make(chan interface{}, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		d, e := c.disconnectPeers()
		select {
		case <-d:
			close(doneCh)
		case er := <-e:
			errCh <- er
		}
	})
	return doneCh, errCh
}

/**
Synchronize the cluster and return the latest block in channel
*/
func (c *Cluster) synchronizeAsync() (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	go func() {
		log.Debug("Start synchronizing")
		// short circuit if already synchronized
		current := c.GetDoerCurrentHead()
		// short circuit if already synchronized
		if current.GetTd() == c.GetTalkerCurrentHead().GetTd() {
			log.Info("Peers already Synchronized", "number", current.GetNumber(), "td", current.GetTd())
			doneCh <- current
			return
		}

		dCh, eCh := c.connectPeers()
		select {
		case <-dCh:
		case err := <-eCh:
			errCh <- err
			return
		}

		err := c.waitUntilSynchronized(common.ClusterTaskTimeout)
		if err != nil {
			log.Error("Synchronization peer timeout", "operation", "synchronize")
			errCh <- err
			return
		}

		dCh, eCh = c.disconnectPeers()
		select {
		case <-dCh:
		case err := <-eCh:
			errCh <- err
			return
		}

		doneCh <- c.GetDoerCurrentHead()
	}()
	return doneCh, errCh
}

/**
send synchronize task in the task queue
*/
func (c *Cluster) SynchronizeAsyncQueued() (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
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

func (c *Cluster) ScheduleTxAsyncQueued(role rpc.Role, hash string) (doneCh chan string, errCh chan error) {
	doneCh = make(chan string, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		// check tx in pool
		d, e := c.server.MiningService().CheckTxInPool(role, hash)
		select {
		case controlMsg := <-d:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("CheckTxInPool RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			if !controlMsg.GetInPool() {
				errCh <- TxNotInPoolErr
				return
			}
		case err := <-e:
			errCh <- err
			return
		}
		// schedule tx
		dCh, eCh := c.server.MiningService().ScheduleTx(role, hash)
		select {
		case controlMsg := <-dCh:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("ScheduleTx RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			doneCh <- hash
		case er := <-eCh:
			errCh <- er
		}
	})
	return doneCh, errCh
}

func (c *Cluster) MineTxAsyncQueued(role rpc.Role, hash string) (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		// check tx in pool
		d, e := c.server.MiningService().CheckTxInPool(role, hash)
		select {
		case controlMsg := <-d:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("CheckTxInPool RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			if !controlMsg.GetInPool() {
				errCh <- TxNotInPoolErr
				return
			}
		case err := <-e:
			errCh <- err
			return
		}
		// mine tx
		dCh, eCh := c.server.MiningService().MineTx(role, hash)
		select {
		case controlMsg := <-dCh:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("MineTx RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			doneCh <- controlMsg.Head
		case er := <-eCh:
			errCh <- er
			return
		}
	})
	return doneCh, errCh
}

func (c *Cluster) MineBlocksAsyncQueued(role rpc.Role, count uint64) (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	err := c.taskQueue.Put(func() {
		// mine blocks
		dCh, eCh := c.server.MiningService().MineBlocks(role, count)
		select {
		case controlMsg := <-dCh:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("MineBlocks RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			doneCh <- controlMsg.Head
		case er := <-eCh:
			errCh <- er
			return
		}
	})
	if err != nil {
		errCh <- err
	}
	return doneCh, errCh
}

func (c *Cluster) MineBlocksWithoutTxAsyncQueued(role rpc.Role, count uint64) (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		// mine blocks without tx
		dCh, eCh := c.server.MiningService().MineBlocksWithoutTx(role, count)
		select {
		case controlMsg := <-dCh:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("MineBlocksWithoutTx RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			doneCh <- controlMsg.Head
		case er := <-eCh:
			errCh <- er
			return
		}
	})
	return doneCh, errCh
}

func (c *Cluster) MineBlocksExceptTxAsyncQueued(role rpc.Role, count uint64, hash string) (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		// mine blocks without tx
		dCh, eCh := c.server.MiningService().MineBlocksExceptTx(role, count, hash)
		select {
		case controlMsg := <-dCh:
			if controlMsg.Err != rpc.Error_NilErr {
				log.Error("MineBlocksExceptTx RPC error", "err", controlMsg.GetErr().String())
				errCh <- RPCErr
				return
			}
			doneCh <- controlMsg.Head
		case er := <-eCh:
			errCh <- er
			return
		}
	})
	return doneCh, errCh
}

func (c *Cluster) ReorgAsyncQueued() (doneCh chan *rpc.ChainHead, errCh chan error) {
	doneCh = make(chan *rpc.ChainHead, 1)
	errCh = make(chan error, 1)
	_ = c.taskQueue.Put(func() {
		doerCurrent := c.GetDoerCurrentHead()
		dCh, eCh := c.server.MiningService().MineTd(rpc.Role_TALKER, doerCurrent.GetTd())
		select {
		case msg := <-dCh:
			if msg.GetErr() != rpc.Error_NilErr {
				log.Error("MineTd reverse RPC error", "err", msg.GetErr().String())
				errCh <- RPCErr
				return
			}
		case err := <-eCh:
			errCh <- err
			return
		}
		d, e := c.synchronizeAsync()
		select {
		case head := <-d:
			doneCh <- head
		case err := <-e:
			errCh <- err
		}
	})
	return doneCh, errCh
}
