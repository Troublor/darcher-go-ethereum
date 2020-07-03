package rpc

import (
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"net/rpc"
)

type Geth struct {
	conn *rpc.Client
}

func ConnectGeth(ip string, port int) *Geth {
	endpoint := fmt.Sprintf("%s:%d", ip, port)
	c, err := rpc.DialHTTP("tcp", endpoint)
	if err != nil {
		log.Error("Geth connection failed", "endpoint", endpoint, "err", err.Error())
		return nil
	}
	return &Geth{
		conn: c,
	}
}

func (c *Geth) MineBlocks(count uint64) {
	arg := &MineBlocksMsg{Count: count}
	reply := &Reply{}
	err := c.conn.Call("Server.MineBlocksRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call MineBlocksRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("MineBlocksRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) MineBlocksWithoutTx(count uint64) {
	arg := &MineBlocksMsg{Count: count}
	reply := &Reply{}
	err := c.conn.Call("Server.MineBlocksWithoutTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call MineBlocksWithoutTxRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("MineBlocksWithoutTxRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) MineBlocksExceptTx(count uint64, exceptTxHash string) {
	arg := &MineBlocksExceptTxMsg{Count: count, TxHash: exceptTxHash}
	reply := &Reply{}
	err := c.conn.Call("Server.MineBlocksExceptTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call MineBlocksExceptTxRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("MineBlocksExceptTxRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) MineTd(td uint64) {
	arg := &MineTdMsg{Td: td}
	reply := &Reply{}
	err := c.conn.Call("Server.MineTdRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call MineTdRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("MineTdRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) MineTx(txHash string) {
	arg := &MineTxMsg{Hash: txHash}
	reply := &Reply{}
	err := c.conn.Call("Server.MineTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call MineTxRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("MineTxRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) AddPeer(url string) {
	arg := &AddPeerMsg{URL: url}
	reply := &Reply{}
	err := c.conn.Call("Server.AddPeerRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call AddPeerRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("AddPeerRPC invocation error", "err", reply.Err.Error())
	}
}

func (c *Geth) RemovePeer(url string) {
	arg := &RemovePeerMsg{URL: url}
	reply := &Reply{}
	err := c.conn.Call("Server.RemovePeerRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call RemovePeerRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("RemovePeerRPC error", "err", reply.Err.Error())
	}
}

func (c *Geth) CheckTxInPool(txHash string) bool {
	arg := &CheckTxInPoolMsg{Hash: txHash}
	reply := &CheckTxInPoolReply{}
	err := c.conn.Call("Server.CheckTxInPoolRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call CheckTxInPoolRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("CheckTxInPoolRPC error", "err", reply.Err.Error())
	}
	return reply.InPool
}

func (c *Geth) ScheduleTx(txHash string) {
	arg := &ScheduleTxMsg{Hash: txHash}
	reply := &Reply{}
	err := c.conn.Call("Server.ScheduleTxRPC", arg, reply)
	if err != nil {
		log.Error("RPC Call ScheduleTxRPC error", "err", err.Error())
	}
	if reply.Err != nil {
		log.Error("ScheduleTxRPC error", "err", reply.Err.Error())
	}
	return
}
