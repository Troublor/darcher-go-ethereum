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
	getHeadControl rpc.BlockchainStatusService_GetHeadControlClient

	/* p2p network service */
	p2pNetworkService rpc.P2PNetworkServiceClient
	// simple rpc
	notifyNodeStartMutex sync.RWMutex
	// reverse RPCs
	addPeerControl    rpc.P2PNetworkService_AddPeerControlClient
	removePeerControl rpc.P2PNetworkService_RemovePeerControlClient

	/* mining service */
	miningService rpc.MiningServiceClient
	// reverse RPCs
	scheduleTxControl          rpc.MiningService_ScheduleTxControlClient
	mineBlocksControl          rpc.MiningService_MineBlocksControlClient
	mineBlocksExceptTxControl  rpc.MiningService_MineBlocksExceptTxControlClient
	mineBlocksWithoutTxControl rpc.MiningService_MineBlocksWithoutTxControlClient
	mineTdControl              rpc.MiningService_MineTdControlClient
	mineTxControl              rpc.MiningService_MineTxControlClient
	checkTxInPoolControl       rpc.MiningService_CheckTxInPoolControlClient
}

func NewClient(role rpc.Role, eth Ethereum, ctx context.Context, masterAddr string) *EthMonitorClient {
	conn, err := grpc.Dial(masterAddr, grpc.WithInsecure())
	if err != nil {
		log.Error("Connect to monitor master error", "err", err)
		return nil
	}

	blockchainStatusService := rpc.NewBlockchainStatusServiceClient(conn)
	p2pNetworkService := rpc.NewP2PNetworkServiceClient(conn)
	miningService := rpc.NewMiningServiceClient(conn)

	c := &EthMonitorClient{
		role:       role,
		eth:        eth,
		serverAddr: masterAddr,
		conn:       conn,
		ctx:        ctx,

		blockchainStatusService: blockchainStatusService,
		p2pNetworkService:       p2pNetworkService,
		miningService:           miningService,
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
				_ = c.addPeerControl.CloseSend()
				return
			default:
			}
			in, err := c.addPeerControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("AddPeer reverse rpc closed, retrying")
				c.addPeerControl, err = c.p2pNetworkService.AddPeerControl(c.ctx)
				if err != nil {
					log.Error("AddPeer reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.addPeerControl.Recv()
			}
			if err != nil {
				log.Error("AddPeer reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.addPeerControl.Send(out)
			if err != nil {
				log.Error("AddPeer reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.removePeerControl.CloseSend()
				return
			default:
			}
			in, err := c.removePeerControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("RemovePeer reverse rpc closed, retrying")
				c.removePeerControl, err = c.p2pNetworkService.RemovePeerControl(c.ctx)
				if err != nil {
					log.Error("RemovePeer reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.removePeerControl.Recv()
			}
			if err != nil {
				log.Error("RemovePeer reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.removePeerControl.Send(out)
			if err != nil {
				log.Error("RemovePeer reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.getHeadControl.CloseSend()
				return
			default:
			}
			in, err := c.getHeadControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("GetHead reverse rpc closed, retrying")
				c.getHeadControl, err = c.blockchainStatusService.GetHeadControl(c.ctx)
				if err != nil {
					log.Error("GetHead reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.getHeadControl.Recv()
			}
			if err != nil {
				log.Error("GetHead reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.getHeadControl.Send(out)
			if err != nil {
				log.Error("GetHead reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.scheduleTxControl.CloseSend()
				return
			default:
			}
			in, err := c.scheduleTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("ScheduleTx reverse rpc closed, retrying")
				c.scheduleTxControl, err = c.miningService.ScheduleTxControl(c.ctx)
				if err != nil {
					log.Error("ScheduleTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.scheduleTxControl.Recv()
			}
			if err != nil {
				log.Error("ScheduleTx reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.scheduleTxControl.Send(out)
			if err != nil {
				log.Error("ScheduleTx reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.mineBlocksControl.CloseSend()
				return
			default:
			}
			in, err := c.mineBlocksControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocks reverse rpc closed, retrying")
				c.mineBlocksControl, err = c.miningService.MineBlocksControl(c.ctx)
				if err != nil {
					log.Error("MineBlocks reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksControl.Recv()
			}
			if err != nil {
				log.Error("MineBlocks reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.mineBlocksControl.Send(out)
			if err != nil {
				log.Error("MineBlocks reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.mineBlocksExceptTxControl.CloseSend()
				return
			default:
			}
			in, err := c.mineBlocksExceptTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocksExceptTx reverse rpc closed, retrying")
				c.mineBlocksExceptTxControl, err = c.miningService.MineBlocksExceptTxControl(c.ctx)
				if err != nil {
					log.Error("MineBlocksExceptTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksExceptTxControl.Recv()
			}
			if err != nil {
				log.Error("MineBlocksExceptTx reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.mineBlocksExceptTxControl.Send(out)
			if err != nil {
				log.Error("MineBlocksExceptTx reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.mineBlocksWithoutTxControl.CloseSend()
				return
			default:
			}
			in, err := c.mineBlocksWithoutTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineBlocksWithoutTx reverse rpc closed, retrying")
				c.mineBlocksWithoutTxControl, err = c.miningService.MineBlocksWithoutTxControl(c.ctx)
				if err != nil {
					log.Error("MineBlocksWithoutTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineBlocksWithoutTxControl.Recv()
			}
			if err != nil {
				log.Error("MineBlocksWithoutTx reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.mineBlocksWithoutTxControl.Send(out)
			if err != nil {
				log.Error("MineBlocksWithoutTx reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.mineTdControl.CloseSend()
				return
			default:
			}
			in, err := c.mineTdControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineTd reverse rpc closed, retrying")
				c.mineTdControl, err = c.miningService.MineTdControl(c.ctx)
				if err != nil {
					log.Error("MineTd reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineTdControl.Recv()
			}
			if err != nil {
				log.Error("MineTd reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.mineTdControl.Send(out)
			if err != nil {
				log.Error("MineTd reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.mineTxControl.CloseSend()
				return
			default:
			}
			in, err := c.mineTxControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("MineTx reverse rpc closed, retrying")
				c.mineTxControl, err = c.miningService.MineTxControl(c.ctx)
				if err != nil {
					log.Error("MineTx reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.mineTxControl.Recv()
			}
			if err != nil {
				log.Error("MineTx reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.mineTxControl.Send(out)
			if err != nil {
				log.Error("MineTx reverse rpc failed to reply", "err", err)
				return
			}
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
				_ = c.checkTxInPoolControl.CloseSend()
				return
			default:
			}
			in, err := c.checkTxInPoolControl.Recv()
			for err == io.EOF {
				// reverse rpc stream closed, try to reconnect
				log.Warn("CheckTxInPool reverse rpc closed, retrying")
				c.checkTxInPoolControl, err = c.miningService.CheckTxInPoolControl(c.ctx)
				if err != nil {
					log.Error("CheckTxInPool reverse rpc reconnect error", "err", err)
					return
				}
				in, err = c.checkTxInPoolControl.Recv()
			}
			if err != nil {
				log.Error("CheckTxInPool reverse rpc receive data error", "err", err)
				return
			}
			out := handler(in)
			err = c.checkTxInPoolControl.Send(out)
			if err != nil {
				log.Error("CheckTxInPool reverse rpc failed to reply", "err", err)
				return
			}
		}
	}()
	return nil
}
