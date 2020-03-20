package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
)

// RPC Methods

func (m *Monitor) MineBlocksRPC(msg *MineBlocksMsg, reply *Reply) (e error) {
	err := m.MineBlocks(int64(msg.Count))
	if err != nil {
		reply.Err = fmt.Errorf("mine blocks error: %s", err.Error())
	}
	return
}

func (m *Monitor) MineTdRPC(msg *MineTdMsg, reply *Reply) (e error) {
	err := m.MineTd(big.NewInt(int64(msg.Td)))
	if err != nil {
		reply.Err = fmt.Errorf("mine td error: %s", err.Error())
	}
	return
}

func (m *Monitor) MineTxRPC(msg *MineTxMsg, reply *Reply) (e error) {
	err := m.MineTx(common.HexToHash(msg.Hash))
	if err != nil {
		reply.Err = fmt.Errorf("mine transaction error: %s", err.Error())
	}
	return
}

func (m *Monitor) AddPeerRPC(msg *AddPeerMsg, reply *Reply) (e error) {
	log.Info("Start connecting to peer", "url", msg.URL)
	r, err := m.addPeer(msg.URL)
	if !r {
		log.Error("Add peer failed", "url", msg.URL)
	}
	if err != nil {
		reply.Err = fmt.Errorf("add peer error: %s", err.Error())
	}
	return
}

func (m *Monitor) RemovePeerRPC(msg *RemovePeerMsg, reply *Reply) (e error) {
	log.Info("Start disconnecting from peer", "url", msg.URL)
	r, err := m.removePeer(msg.URL)
	if !r {
		log.Error("Add peer failed", "url", msg.URL)
	}
	if err != nil {
		reply.Err = fmt.Errorf("remove peer error: %s", err.Error())
	}
	return
}

func (m *Monitor) ExecuteTxRPC(msg *ExecuteTxMsg, reply *Reply) (e error) {
	hash := common.HexToHash(msg.Hash)
	if m.eth.TxPool().Get(hash) == nil {
		// if the tx to be executed is not in tx pool
		reply.Err = fmt.Errorf("tx %s not found", msg.Hash)
		return
	}
	err := m.MineTx(hash)
	if err != nil {
		reply.Err = fmt.Errorf("error occurs when executing tx: %s", err.Error())
		return
	}
	return
}

func (m *Monitor) RevertTxRPC(msg *RevertTxMsg, reply *Reply) (e error) {
	targetTd := big.NewInt(int64(msg.Td))
	if m.eth.BlockChain().GetTdByHash(m.eth.BlockChain().CurrentBlock().Hash()).Cmp(targetTd) > 0 {
		return
	}
	err := m.MineTd(targetTd)
	if err != nil {
		reply.Err = fmt.Errorf("error occurs when mining td: %s", err.Error())
		return
	}
	return
}

func (m *Monitor) CheckTxInPoolRPC(msg *CheckTxInPoolMsg, reply *CheckTxInPoolReply) (e error) {
	txHash := msg.Hash
	reply.InPool = inTxPool(common.HexToHash(txHash), m.eth.TxPool())
	return nil
}

//func (monitor *Monitor) ReExecuteTx(arg *Hash, reply *Reply) (e error) {
//
//}

//func (monitor *Monitor) ConfirmTx(arg *TxHashAndInt, reply *Reply) (e error) {
//
//}

func (m *Monitor) startMonitorRpcServer() {
	port := m.serverPort
	err := rpc.Register(m)
	if err != nil {
		log.Error("Start RPC Server error", "err", err)
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Error("Start RPC Server error", "err", e)
	}
	go http.Serve(l, nil)
	log.Info("Start RPC Server success", "port", port)
}
