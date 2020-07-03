package worker

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"net/rpc"
)

type Client struct {
	monitor   *Monitor
	netClient *rpc.Client
}

func NewClient(monitor *Monitor, port int) *Client {
	c := &Client{monitor: monitor}
	// connect to rpc server
	var err error
	c.netClient, err = rpc.DialHTTP("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		log.Error("failed to connect to rpc server", "err", err.Error())
		return nil
	}
	return c
}

func (c *Client) NotifyNodeStart(node *node.Node, role Role, serverPort int) {
	currentBlock := c.monitor.eth.BlockChain().CurrentBlock()
	arg := &NodeStartMsg{
		Role:        role,
		ServerPort:  serverPort,
		URL:         node.Server().NodeInfo().Enode,
		BlockNumber: currentBlock.NumberU64(),
		BlockHash:   currentBlock.Hash().Hex(),
		Td:          c.monitor.eth.BlockChain().GetTdByHash(currentBlock.Hash()).Uint64(),
	}
	reply := &Reply{}
	err := c.netClient.Call("Server.OnNodeStartRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call NotifyNodeStart error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("NotifyNodeStart error", "err", reply.Err.Error())
	}
}

func (c *Client) NotifyNewTx(tx *types.Transaction, role Role) {
	signer := types.NewEIP155Signer(c.monitor.eth.BlockChain().Config().ChainID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		log.Error("Get tx sender failed", "tx", tx.Hash())
		return
	}
	arg := &NewTxMsg{
		Role:   role,
		Hash:   tx.Hash().Hex(),
		Sender: sender.Hex(),
		Nonce:  tx.Nonce(),
	}
	reply := &Reply{}
	err = c.netClient.Call("Server.OnNewTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewTx error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("notifyNewTx error", "err", reply.Err.Error())
	}
}

func (c *Client) NotifyNewChainHead(block *types.Block, role Role) {
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	arg := &NewChainHeadMsg{
		Role:   role,
		Hash:   block.Hash().Hex(),
		Number: block.NumberU64(),
		Td:     c.monitor.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(),
		Txs:    txHashes,
	}
	reply := &Reply{}
	err := c.netClient.Call("Server.OnNewChainHeadRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewChainHead error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("notifyNewChainHead error", "err", reply.Err.Error())
	}
}

func (c *Client) NotifyNewChainSide(block *types.Block, role Role) {
	txHashes := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashes = append(txHashes, tx.Hash().Hex())
	}
	arg := &NewChainSideMsg{
		Role:   role,
		Hash:   block.Hash().Hex(),
		Number: block.NumberU64(),
		Td:     c.monitor.eth.BlockChain().GetTdByHash(block.Hash()).Uint64(),
		Txs:    txHashes,
	}
	reply := &Reply{}
	err := c.netClient.Call("Server.OnNewChainSideRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyNewChainSide error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("notifyNewChainSide error", "err", reply.Err.Error())
	}
}

func (c *Client) NotifyPeerDrop(ev *p2p.PeerEvent, role Role) {
	arg := &PeerRemoveMsg{Role: role, PeerID: ev.Peer.String()}
	reply := &Reply{}
	err := c.netClient.Call("Server.OnPeerRemoveRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyPeerDrop error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("notifyPeerDrop error", "err", reply.Err.Error())
	}
}

func (c *Client) NotifyPeerAdd(ev *p2p.PeerEvent, role Role) {
	arg := &PeerAddMsg{Role: role, PeerID: ev.Peer.String()}
	reply := &Reply{}
	err := c.netClient.Call("Server.OnPeerAddRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call notifyPeerAdd error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("notifyPeerAdd error", "err", reply.Err.Error())
	}
}
