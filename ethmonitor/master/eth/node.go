package eth

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/master/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
)

type Node struct {
	role   common.Role
	url    string
	geth   *rpc.Geth
	server *rpc.Server

	nodeUpdateFeed event.Feed
	txExecutedFeed event.Feed
	scope          event.SubscriptionScope

	currentBlock *common.Block
}

func NewNode(role common.Role, url string, ip string, port int, initBlock *common.Block, server *rpc.Server) *Node {
	cli := rpc.ConnectGeth(ip, port)
	if cli == nil {
		log.Error("Initialize geth connection failed", "role", role, "url", url)
		return nil
	}
	node := &Node{
		role:   role,
		url:    url,
		geth:   cli,
		server: server,

		currentBlock: initBlock,
	}
	// loops
	go node.updateLoop()

	return node
}

func (n *Node) updateLoop() {
	// events
	var newChainHeadCh chan common.NewChainHeadEvent
	var newChainHeadSub event.Subscription
	var newChainSideCh chan common.NewChainSideEvent
	var newChainSideSub event.Subscription
	newChainHeadCh = make(chan common.NewChainHeadEvent, common.EventCacheSize)
	newChainHeadSub = n.server.SubscribeNewChainHeadEvent(newChainHeadCh)
	defer newChainHeadSub.Unsubscribe()
	newChainSideCh = make(chan common.NewChainSideEvent, common.EventCacheSize)
	newChainSideSub = n.server.SubscribeNewChainSideEvent(newChainSideCh)
	defer newChainSideSub.Unsubscribe()
	for {
		select {
		case ev := <-newChainHeadCh:
			if ev.Role != n.role {
				break
			}
			n.currentBlock.Number = ev.Number
			n.currentBlock.Hash = ev.Hash
			n.currentBlock.Td = ev.Td
			log.Debug("Node update", "role", n.role, "Number", ev.Number, "Td", ev.Td)
			n.nodeUpdateFeed.Send(common.NodeUpdateEvent{Number: ev.Number, Td: ev.Td})
			for _, tx := range ev.Txs {
				n.txExecutedFeed.Send(common.TxExecutedEvent{
					Role:        ev.Role,
					TxHash:      tx,
					BlockHash:   ev.Hash,
					BlockNumber: ev.Number,
					Td:          ev.Td,
				})
			}

		case ev := <-newChainSideCh:
			if ev.Role != n.role {
				break
			}
			n.currentBlock.Number = ev.Number
			n.currentBlock.Hash = ev.Hash
			n.currentBlock.Td = ev.Td
			log.Debug("Node update", "role", n.role, "Number", ev.Number, "Td", ev.Td)
			n.nodeUpdateFeed.Send(common.NodeUpdateEvent{Number: ev.Number, Td: ev.Td})
		}
	}
}

func (n *Node) SubscribeNodeUpdateEvent(ch chan<- common.NodeUpdateEvent) event.Subscription {
	return n.scope.Track(n.nodeUpdateFeed.Subscribe(ch))
}

func (n *Node) SubscribeTxExecutedEvent(ch chan<- common.TxExecutedEvent) event.Subscription {
	return n.scope.Track(n.txExecutedFeed.Subscribe(ch))
}

/*
Remote asynchronous methods
*/

func (n *Node) MineBlocks(count uint64) {
	n.geth.MineBlocks(count)
}

func (n *Node) MineBlocksWithoutTx(count uint64) {
	n.geth.MineBlocksWithoutTx(count)
}

func (n *Node) MineBlocksExceptTx(count uint64, txHash string) {
	n.geth.MineBlocksExceptTx(count, txHash)
}

func (n *Node) MineTx(txHash string) {
	n.geth.MineTx(txHash)
}

func (n *Node) MineTd(td uint64) {
	n.geth.MineTd(td)
}

func (n *Node) AddPeer(url string) {
	n.geth.AddPeer(url)
}

func (n *Node) RemovePeer(url string) {
	n.geth.RemovePeer(url)
}

func (n *Node) CheckTxInPool(txHash string) bool {
	return n.geth.CheckTxInPool(txHash)
}
