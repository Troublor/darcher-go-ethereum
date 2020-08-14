package worker

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"sync"
)

type EthMonitorClient struct {
	role       rpc.Role
	eth        Ethereum
	serverAddr string // ethmonitor master server address
	conn       *grpc.ClientConn
	ctx        context.Context

	/* blockchain status service */
	blockchainStatusService rpc.BlockchainStatusServiceClient
	// client-side streams
	notifyNewChainHeadStreamMutex sync.RWMutex
	notifyNewChainHeadStream      rpc.BlockchainStatusService_NotifyNewChainHeadClient
	notifyNewChainSideStreamMutex sync.RWMutex
	notifyNewChainSideStream      rpc.BlockchainStatusService_NotifyNewChainSideClient
	notifyNewTxStreamMutex        sync.RWMutex
	notifyNewTxStream             rpc.BlockchainStatusService_NotifyNewTxClient
	// reverse RPCs
	getHeadControl      rpc.BlockchainStatusService_GetHeadControlClient
	getHeadControlMutex sync.Mutex

	/* p2p network service */
	p2pNetworkService rpc.P2PNetworkServiceClient
	// simple rpc
	notifyNodeStartMutex sync.RWMutex
	// reverse RPCs
	addPeerControl         rpc.P2PNetworkService_AddPeerControlClient
	addPeerControlMutex    sync.Mutex
	removePeerControl      rpc.P2PNetworkService_RemovePeerControlClient
	removePeerControlMutex sync.Mutex

	/* mining service */
	miningService rpc.MiningServiceClient
	// reverse RPCs
	scheduleTxControl               rpc.MiningService_ScheduleTxControlClient
	scheduleTxControlMutex          sync.Mutex
	mineBlocksControl               rpc.MiningService_MineBlocksControlClient
	mineBlocksControlMutex          sync.Mutex
	mineBlocksExceptTxControl       rpc.MiningService_MineBlocksExceptTxControlClient
	mineBlocksExceptTxControlMutex  sync.Mutex
	mineBlocksWithoutTxControl      rpc.MiningService_MineBlocksWithoutTxControlClient
	mineBlocksWithoutTxControlMutex sync.Mutex
	mineTdControl                   rpc.MiningService_MineTdControlClient
	mineTdControlMutex              sync.Mutex
	mineTxControl                   rpc.MiningService_MineTxControlClient
	mineTxControlMutex              sync.Mutex
	checkTxInPoolControl            rpc.MiningService_CheckTxInPoolControlClient
	checkTxInPoolControlMutex       sync.Mutex

	/* contract oracle service */
	contractOracleService rpc.ContractVulnerabilityServiceClient
	// simple rpc
	notifyTxErrorMutex sync.RWMutex
	// reverse RPCs
	getReportsByContractControl         rpc.ContractVulnerabilityService_GetReportsByContractControlClient
	getReportsByContractControlMutex    sync.Mutex
	getReportsByTransactionControl      rpc.ContractVulnerabilityService_GetReportsByTransactionControlClient
	getReportsByTransactionControlMutex sync.Mutex
}

func NewClient(role rpc.Role, eth Ethereum, ctx context.Context, masterAddr string) *EthMonitorClient {
	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Error("Connect to monitor master error", "err", err)
		return nil
	}

	blockchainStatusService := rpc.NewBlockchainStatusServiceClient(conn)
	p2pNetworkService := rpc.NewP2PNetworkServiceClient(conn)
	miningService := rpc.NewMiningServiceClient(conn)
	contractOracleService := rpc.NewContractVulnerabilityServiceClient(conn)

	c := &EthMonitorClient{
		role:       role,
		eth:        eth,
		serverAddr: masterAddr,
		conn:       conn,
		ctx:        ctx,

		blockchainStatusService: blockchainStatusService,
		p2pNetworkService:       p2pNetworkService,
		miningService:           miningService,
		contractOracleService:   contractOracleService,
	}
	return c
}

func (c *EthMonitorClient) Close() {
	err := c.conn.Close()
	if err != nil {
		log.Error("Close to monitor master connect error", "err", err)
	}
}

/**
Thread safe
*/
func (c *EthMonitorClient) NotifyNewChainHead(block *types.Block) error {
	c.notifyNewChainHeadStreamMutex.Lock()
	defer c.notifyNewChainHeadStreamMutex.Unlock()
	return c._notifyNewChainHead(block)
}

func (c *EthMonitorClient) _notifyNewChainHead(block *types.Block) error {
	var err error
	if c.notifyNewChainHeadStream == nil {
		c.notifyNewChainHeadStream, err = c.blockchainStatusService.NotifyNewChainHead(c.ctx)
		if err != nil {
			return err
		}
	}
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	err = c.notifyNewChainHeadStream.Send(&rpc.ChainHead{
		Role:   c.role,
		Hash:   block.Hash().Hex(),
		Number: block.Number().Uint64(),
		Td:     c.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(),
		Txs:    txHashes,
	})
	if err == io.EOF {
		_, e := c.notifyNewChainHeadStream.CloseAndRecv()
		if e != nil {
			return e
		}
		c.notifyNewChainHeadStream = nil
		// notify failed because of EOF, try to notify again
		log.Warn("NotifyNewChainHead stream closed, retrying")
		e = c._notifyNewChainHead(block)
		if e != nil {
			return e
		}
	} else if err != nil {
		return err
	}
	return nil
}

/**
Thread safe
*/
func (c *EthMonitorClient) NotifyNewChainSide(block *types.Block) error {
	c.notifyNewChainSideStreamMutex.Lock()
	defer c.notifyNewChainSideStreamMutex.Unlock()
	return c._notifyNewChainSide(block)
}

func (c *EthMonitorClient) _notifyNewChainSide(block *types.Block) error {
	var err error
	if c.notifyNewChainSideStream == nil {
		c.notifyNewChainSideStream, err = c.blockchainStatusService.NotifyNewChainSide(c.ctx)
		if err != nil {
			return err
		}
	}
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	err = c.notifyNewChainSideStream.Send(&rpc.ChainSide{
		Role:   c.role,
		Hash:   block.Hash().Hex(),
		Number: block.Number().Uint64(),
		Td:     c.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(),
		Txs:    txHashes,
	})
	if err == io.EOF {
		_, e := c.notifyNewChainSideStream.CloseAndRecv()
		if e != nil {
			return e
		}
		c.notifyNewChainSideStream = nil
		// notify failed because of EOF, try to notify again
		log.Warn("NotifyNewChainSide stream closed, retrying")
		e = c._notifyNewChainSide(block)
		if e != nil {
			return e
		}
	} else if err != nil {
		return err
	}
	return nil
}

/**
Thread safe
*/
func (c *EthMonitorClient) NotifyNewTx(tx *types.Transaction) error {
	c.notifyNewTxStreamMutex.Lock()
	defer c.notifyNewTxStreamMutex.Unlock()
	return c._notifyNewTx(tx)
}

func (c *EthMonitorClient) _notifyNewTx(tx *types.Transaction) error {
	var err error
	if c.notifyNewTxStream == nil {
		c.notifyNewTxStream, err = c.blockchainStatusService.NotifyNewTx(c.ctx)
		if err != nil {
			return err
		}
	}
	signer := types.NewEIP155Signer(c.eth.BlockChain().Config().ChainID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		log.Error("Get tx sender failed", "tx", tx.Hash())
		return err
	}
	err = c.notifyNewTxStream.Send(&rpc.Tx{
		Role:   c.role,
		Hash:   tx.Hash().Hex(),
		Sender: sender.Hex(),
		Nonce:  tx.Nonce(),
	})
	if err == io.EOF {
		_, e := c.notifyNewTxStream.CloseAndRecv()
		if e != nil {
			return e
		}
		c.notifyNewTxStream = nil
		// notify failed because of EOF, try to notify again
		log.Warn("NotifyNewTx stream closed, retrying")
		e = c._notifyNewTx(tx)
		if e != nil {
			return e
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (c *EthMonitorClient) NotifyNodeStart(node *node.Node) error {
	c.notifyNodeStartMutex.Lock()
	defer c.notifyNodeStartMutex.Unlock()
	return c._notifyNodeStart(node)
}

func (c *EthMonitorClient) _notifyNodeStart(node *node.Node) error {
	currentBlock := c.eth.BlockChain().CurrentBlock()
	txs := make([]string, len(currentBlock.Transactions()))
	for i, tx := range currentBlock.Transactions() {
		txs[i] = tx.Hash().Hex()
	}
	_, err := c.p2pNetworkService.NotifyNodeStart(c.ctx, &rpc.Node{
		Role: c.role,
		Url:  node.Server().NodeInfo().Enode,
		Head: &rpc.ChainHead{
			Role:   c.role,
			Hash:   currentBlock.Hash().Hex(),
			Number: currentBlock.NumberU64(),
			Td:     c.eth.BlockChain().GetTdByHash(currentBlock.Hash()).Uint64(),
			Txs:    txs,
		},
	})
	return err
}

/**
Notify TxError (rpc.TxErrorMsg) to ethmonitor.master, this should be called when tx execution throws an error
*/
func (c *EthMonitorClient) NotifyTxError(txErrorMsg *rpc.TxErrorMsg) error {
	log.Warn("Transaction execution failed", "txHash", txErrorMsg.Hash, "err", txErrorMsg.Description)
	c.notifyTxErrorMutex.Lock()
	defer c.notifyTxErrorMutex.Unlock()
	return c._notifyTxError(txErrorMsg)
}

func (c *EthMonitorClient) _notifyTxError(txErrorMsg *rpc.TxErrorMsg) error {
	_, err := c.contractOracleService.NotifyTxError(c.ctx, txErrorMsg)
	return err
}

/**
Serve AddPeer reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeAddPeerControl(handler func(in *rpc.AddPeerControlMsg) (out *rpc.AddPeerControlMsg)) error {
	var err error
	c.addPeerControl, err = c.p2pNetworkService.AddPeerControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.addPeerControl.Send(&rpc.AddPeerControlMsg{
		Role:   c.role,
		Id:     "",
		Url:    "",
		PeerId: "",
		Err:    0,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.addPeerControlMutex.Lock()
				_ = c.addPeerControl.CloseSend()
				c.addPeerControlMutex.Unlock()
				return
			default:
			}
			in, err := c.addPeerControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("AddPeer reverse rpc closed, retrying")
				c.addPeerControlMutex.Lock()
				c.addPeerControl, err = c.p2pNetworkService.AddPeerControl(c.ctx)
				c.addPeerControlMutex.Unlock()
				if err != nil {
					log.Error("AddPeer reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.addPeerControl.Recv()
			}
			if err == context.Canceled {
				log.Info("AddPeer reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("AddPeer reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.addPeerControlMutex.Lock()
				err = c.addPeerControl.Send(out)
				c.addPeerControlMutex.Unlock()
				if err != nil {
					log.Error("AddPeer reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve RemovePeer reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeRemovePeerControl(handler func(in *rpc.RemovePeerControlMsg) (out *rpc.RemovePeerControlMsg)) error {
	var err error
	c.removePeerControl, err = c.p2pNetworkService.RemovePeerControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.removePeerControl.Send(&rpc.RemovePeerControlMsg{
		Role:   c.role,
		Id:     "",
		Url:    "",
		PeerId: "",
		Err:    0,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.removePeerControlMutex.Lock()
				_ = c.removePeerControl.CloseSend()
				c.removePeerControlMutex.Unlock()
				return
			default:
			}
			in, err := c.removePeerControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("RemovePeer reverse rpc closed, retrying")
				c.removePeerControlMutex.Lock()
				c.removePeerControl, err = c.p2pNetworkService.RemovePeerControl(c.ctx)
				c.removePeerControlMutex.Unlock()
				if err != nil {
					log.Error("RemovePeer reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.removePeerControl.Recv()
			}
			if err == context.Canceled {
				log.Info("RemovePeer reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("RemovePeer reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.removePeerControlMutex.Lock()
				err = c.removePeerControl.Send(out)
				c.removePeerControlMutex.Unlock()
				if err != nil {
					log.Error("RemovePeer reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve GetHead reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeGetHeadControl(handler func(in *rpc.GetChainHeadControlMsg) (out *rpc.GetChainHeadControlMsg)) error {
	var err error
	c.getHeadControl, err = c.blockchainStatusService.GetHeadControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.getHeadControl.Send(&rpc.GetChainHeadControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.getHeadControlMutex.Lock()
				_ = c.getHeadControl.CloseSend()
				c.getHeadControlMutex.Unlock()
				return
			default:
			}
			in, err := c.getHeadControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("GetHead reverse rpc closed, retrying")
				c.getHeadControlMutex.Lock()
				c.getHeadControl, err = c.blockchainStatusService.GetHeadControl(c.ctx)
				c.getHeadControlMutex.Unlock()
				if err != nil {
					log.Error("GetHead reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.getHeadControl.Recv()
			}
			if err == context.Canceled {
				log.Info("GetHead reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("GetHead reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.getHeadControlMutex.Lock()
				err = c.getHeadControl.Send(out)
				c.getHeadControlMutex.Unlock()
				if err != nil {
					log.Error("GetHead reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve ScheduleTx reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeScheduleTxControl(handler func(in *rpc.ScheduleTxControlMsg) (out *rpc.ScheduleTxControlMsg)) error {
	var err error
	c.scheduleTxControl, err = c.miningService.ScheduleTxControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.scheduleTxControl.Send(&rpc.ScheduleTxControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.scheduleTxControlMutex.Lock()
				_ = c.scheduleTxControl.CloseSend()
				c.scheduleTxControlMutex.Unlock()
				return
			default:
			}
			in, err := c.scheduleTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("ScheduleTx reverse rpc closed, retrying")
				c.scheduleTxControlMutex.Lock()
				c.scheduleTxControl, err = c.miningService.ScheduleTxControl(c.ctx)
				c.scheduleTxControlMutex.Unlock()
				if err != nil {
					log.Error("ScheduleTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.scheduleTxControl.Recv()
			}
			if err == context.Canceled {
				log.Info("ScheduleTx reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("ScheduleTx reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.scheduleTxControlMutex.Lock()
				err = c.scheduleTxControl.Send(out)
				c.scheduleTxControlMutex.Unlock()
				if err != nil {
					log.Error("ScheduleTx reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve MineBlocks reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeMineBlocksControl(handler func(in *rpc.MineBlocksControlMsg) (out *rpc.MineBlocksControlMsg)) error {
	var err error
	c.mineBlocksControl, err = c.miningService.MineBlocksControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.mineBlocksControl.Send(&rpc.MineBlocksControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.mineBlocksControlMutex.Lock()
				_ = c.mineBlocksControl.CloseSend()
				c.mineBlocksControlMutex.Unlock()
				return
			default:
			}
			in, err := c.mineBlocksControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocks reverse rpc closed, retrying")
				c.mineBlocksControlMutex.Lock()
				c.mineBlocksControl, err = c.miningService.MineBlocksControl(c.ctx)
				c.mineBlocksControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocks reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksControl.Recv()
			}
			if err == context.Canceled {
				log.Info("MineBlocks reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("MineBlocks reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.mineBlocksControlMutex.Lock()
				err = c.mineBlocksControl.Send(out)
				c.mineBlocksControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocks reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve MineBlocksExceptTx reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeMineBlocksExceptTxControl(handler func(in *rpc.MineBlocksExceptTxControlMsg) (out *rpc.MineBlocksExceptTxControlMsg)) error {
	var err error
	c.mineBlocksExceptTxControl, err = c.miningService.MineBlocksExceptTxControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.mineBlocksExceptTxControl.Send(&rpc.MineBlocksExceptTxControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.mineBlocksExceptTxControlMutex.Lock()
				_ = c.mineBlocksExceptTxControl.CloseSend()
				c.mineBlocksExceptTxControlMutex.Unlock()
				return
			default:
			}
			in, err := c.mineBlocksExceptTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocksExceptTx reverse rpc closed, retrying")
				c.mineBlocksExceptTxControlMutex.Lock()
				c.mineBlocksExceptTxControl, err = c.miningService.MineBlocksExceptTxControl(c.ctx)
				c.mineBlocksExceptTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocksExceptTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksExceptTxControl.Recv()
			}
			if err == context.Canceled {
				log.Info("MineBlocksExceptTx reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("MineBlocksExceptTx reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.mineBlocksExceptTxControlMutex.Lock()
				err = c.mineBlocksExceptTxControl.Send(out)
				c.mineBlocksExceptTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocksExceptTx reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve MineBlocksWithoutTx reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeMineBlocksWithoutTxControl(handler func(in *rpc.MineBlocksWithoutTxControlMsg) (out *rpc.MineBlocksWithoutTxControlMsg)) error {
	var err error
	c.mineBlocksWithoutTxControl, err = c.miningService.MineBlocksWithoutTxControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.mineBlocksWithoutTxControl.Send(&rpc.MineBlocksWithoutTxControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.mineBlocksWithoutTxControlMutex.Lock()
				_ = c.mineBlocksWithoutTxControl.CloseSend()
				c.mineBlocksWithoutTxControlMutex.Unlock()
				return
			default:
			}
			in, err := c.mineBlocksWithoutTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocksWithoutTx reverse rpc closed, retrying")
				c.mineBlocksWithoutTxControlMutex.Lock()
				c.mineBlocksWithoutTxControl, err = c.miningService.MineBlocksWithoutTxControl(c.ctx)
				c.mineBlocksWithoutTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocksWithoutTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksWithoutTxControl.Recv()
			}
			if err == context.Canceled {
				log.Info("MineBlocksWithoutTx reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("MineBlocksWithoutTx reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.mineBlocksWithoutTxControlMutex.Lock()
				err = c.mineBlocksWithoutTxControl.Send(out)
				c.mineBlocksWithoutTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineBlocksWithoutTx reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve MineTd reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeMineTdControl(handler func(in *rpc.MineTdControlMsg) (out *rpc.MineTdControlMsg)) error {
	var err error
	c.mineTdControl, err = c.miningService.MineTdControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.mineTdControl.Send(&rpc.MineTdControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.mineTdControlMutex.Lock()
				_ = c.mineTdControl.CloseSend()
				c.mineTdControlMutex.Unlock()
				return
			default:
			}
			in, err := c.mineTdControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineTd reverse rpc closed, retrying")
				c.mineTdControlMutex.Lock()
				c.mineTdControl, err = c.miningService.MineTdControl(c.ctx)
				c.mineTdControlMutex.Unlock()
				if err != nil {
					log.Error("MineTd reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineTdControl.Recv()
			}
			if err == context.Canceled {
				log.Info("MineTd reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("MineTd reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.mineTdControlMutex.Lock()
				err = c.mineTdControl.Send(out)
				c.mineTdControlMutex.Unlock()
				if err != nil {
					log.Error("MineTd reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve MineTx reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeMineTxControl(handler func(in *rpc.MineTxControlMsg) (out *rpc.MineTxControlMsg)) error {
	var err error
	c.mineTxControl, err = c.miningService.MineTxControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.mineTxControl.Send(&rpc.MineTxControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.mineTxControlMutex.Lock()
				_ = c.mineTxControl.CloseSend()
				c.mineTxControlMutex.Unlock()
				return
			default:
			}
			in, err := c.mineTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineTx reverse rpc closed, retrying")
				c.mineTxControlMutex.Lock()
				c.mineTxControl, err = c.miningService.MineTxControl(c.ctx)
				c.mineTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineTxControl.Recv()
			}
			if err == context.Canceled {
				log.Info("MineTx reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("MineTx reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.mineTxControlMutex.Lock()
				err = c.mineTxControl.Send(out)
				c.mineTxControlMutex.Unlock()
				if err != nil {
					log.Error("MineTx reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve CheckTxInPool reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeCheckTxInPoolControl(handler func(in *rpc.CheckTxInPoolControlMsg) (out *rpc.CheckTxInPoolControlMsg)) error {
	var err error
	c.checkTxInPoolControl, err = c.miningService.CheckTxInPoolControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.checkTxInPoolControl.Send(&rpc.CheckTxInPoolControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.checkTxInPoolControlMutex.Lock()
				_ = c.checkTxInPoolControl.CloseSend()
				c.checkTxInPoolControlMutex.Unlock()
				return
			default:
			}
			in, err := c.checkTxInPoolControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("CheckTxInPool reverse rpc closed, retrying")
				c.checkTxInPoolControlMutex.Lock()
				c.checkTxInPoolControl, err = c.miningService.CheckTxInPoolControl(c.ctx)
				c.checkTxInPoolControlMutex.Unlock()
				if err != nil {
					log.Error("CheckTxInPool reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.checkTxInPoolControl.Recv()
			}
			if err == context.Canceled {
				log.Info("CheckTxInPool reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("CheckTxInPool reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.checkTxInPoolControlMutex.Lock()
				err = c.checkTxInPoolControl.Send(out)
				c.checkTxInPoolControlMutex.Unlock()
				if err != nil {
					log.Error("CheckTxInPool reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve GetReportsByContract reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeGetReportsByContractControl(handler func(in *rpc.GetReportsByContractControlMsg) (out *rpc.GetReportsByContractControlMsg)) error {
	var err error
	c.getReportsByContractControl, err = c.contractOracleService.GetReportsByContractControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.getReportsByContractControl.Send(&rpc.GetReportsByContractControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.getReportsByContractControlMutex.Lock()
				_ = c.getReportsByContractControl.CloseSend()
				c.getReportsByContractControlMutex.Unlock()
				return
			default:
			}
			in, err := c.getReportsByContractControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("GetReportsByContract reverse rpc closed, retrying")
				c.getReportsByContractControlMutex.Lock()
				c.getReportsByContractControl, err = c.contractOracleService.GetReportsByContractControl(c.ctx)
				c.getReportsByContractControlMutex.Unlock()
				if err != nil {
					log.Error("GetReportsByContract reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.getReportsByContractControl.Recv()
			}
			if err == context.Canceled {
				log.Info("GetReportsByContract reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("GetReportsByContract reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.getReportsByContractControlMutex.Lock()
				err = c.getReportsByContractControl.Send(out)
				c.getReportsByContractControlMutex.Unlock()
				if err != nil {
					log.Error("GetReportsByContract reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}

/**
Serve GetReportsByTransaction reverse rpc
This is a synchronized method (will block) and should only be called once
*/
func (c *EthMonitorClient) ServeGetReportsByTransactionControl(handler func(in *rpc.GetReportsByTransactionControlMsg) (out *rpc.GetReportsByTransactionControlMsg)) error {
	var err error
	c.getReportsByTransactionControl, err = c.contractOracleService.GetReportsByTransactionControl(c.ctx)
	if err != nil {
		return err
	}
	// send init message to register reverse rpc
	_ = c.getReportsByTransactionControl.Send(&rpc.GetReportsByTransactionControlMsg{
		Role: c.role,
	})
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				c.getReportsByTransactionControlMutex.Lock()
				_ = c.getReportsByTransactionControl.CloseSend()
				c.getReportsByTransactionControlMutex.Unlock()
				return
			default:
			}
			in, err := c.getReportsByTransactionControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("GetReportsByTransaction reverse rpc closed, retrying")
				c.getReportsByTransactionControlMutex.Lock()
				c.getReportsByTransactionControl, err = c.contractOracleService.GetReportsByTransactionControl(c.ctx)
				c.getReportsByTransactionControlMutex.Unlock()
				if err != nil {
					log.Error("GetReportsByTransaction reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.getReportsByTransactionControl.Recv()
			}
			if err == context.Canceled {
				log.Info("GetReportsByTransaction reverse rpc closed", "err", err)
				return
			}
			if err != nil {
				log.Error("GetReportsByTransaction reverse rpc receive data error", "err", err)
				return
			}
			go func() {
				out := handler(in)
				c.getReportsByTransactionControlMutex.Lock()
				err = c.getReportsByTransactionControl.Send(out)
				c.getReportsByTransactionControlMutex.Unlock()
				if err != nil {
					log.Error("GetReportsByTransaction reverse rpc failed to reply", "err", err)
					return
				}
			}()
		}
	}()
	return nil
}
