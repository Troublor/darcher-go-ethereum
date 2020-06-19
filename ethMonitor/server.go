package ethMonitor

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
)

type Server struct {
	monitor *Monitor
	port    int
}

func NewServer(monitor *Monitor, port int) *Server {
	s := &Server{monitor: monitor, port: port}
	s.startMonitorRpcServer()
	return s
}

func (s *Server) startMonitorRpcServer() {
	port := s.port
	err := rpc.Register(s)
	if err != nil {
		log.Error("Start Monitor Server error", "err", err)
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Error("Start Monitor Server error", "err", e)
	}
	go http.Serve(l, nil)
	log.Info("Start Monitor Server success", "port", port)
}

// RPC Methods

func (s *Server) MineBlocksRPC(msg *MineBlocksMsg, reply *Reply) (e error) {
	err := s.mineBlocks(int64(msg.Count))
	if err != nil {
		reply.Err = fmt.Errorf("mine blocks error: %s", err.Error())
	}
	return
}

func (s *Server) MineBlocksWithoutTxRPC(msg *MineBlocksMsg, reply *Reply) (e error) {
	err := s.mineBlocksWithoutTx(int64(msg.Count))
	if err != nil {
		reply.Err = fmt.Errorf("mine blocks error: %s", err.Error())
	}
	return
}

func (s *Server) MineBlocksExceptTxRPC(msg *MineBlocksExceptTxMsg, reply *Reply) (e error) {
	err := s.mineBlocksExceptTx(int64(msg.Count), msg.TxHash)
	if err != nil {
		reply.Err = fmt.Errorf("mine blocks error: %s", err.Error())
	}
	return
}

func (s *Server) MineTdRPC(msg *MineTdMsg, reply *Reply) (e error) {
	err := s.mineTd(big.NewInt(int64(msg.Td)))
	if err != nil {
		reply.Err = fmt.Errorf("mine td error: %s", err.Error())
	}
	return
}

func (s *Server) MineTxRPC(msg *MineTxMsg, reply *Reply) (e error) {
	err := s.mineTx(common.HexToHash(msg.Hash))
	if err != nil {
		reply.Err = fmt.Errorf("mine transaction error: %s", err.Error())
	}
	return
}

func (s *Server) AddPeerRPC(msg *AddPeerMsg, reply *Reply) (e error) {
	log.Info("Start connecting to peer", "url", msg.URL)
	r, err := s.monitor.addPeer(msg.URL)
	if !r {
		log.Error("Add peer failed", "url", msg.URL)
	}
	if err != nil {
		reply.Err = fmt.Errorf("add peer error: %s", err.Error())
	}
	return
}

func (s *Server) RemovePeerRPC(msg *RemovePeerMsg, reply *Reply) (e error) {
	log.Info("Start disconnecting from peer", "url", msg.URL)
	r, err := s.monitor.removePeer(msg.URL)
	if !r {
		log.Error("Add peer failed", "url", msg.URL)
	}
	if err != nil {
		reply.Err = fmt.Errorf("remove peer error: %s", err.Error())
	}
	return
}

func (s *Server) ExecuteTxRPC(msg *ExecuteTxMsg, reply *Reply) (e error) {
	hash := common.HexToHash(msg.Hash)
	if s.monitor.eth.TxPool().Get(hash) == nil {
		// if the tx to be executed is not in tx pool
		reply.Err = fmt.Errorf("tx %s not found", msg.Hash)
		return
	}
	err := s.mineTx(hash)
	if err != nil {
		reply.Err = fmt.Errorf("error occurs when executing tx: %s", err.Error())
		return
	}
	return
}

func (s *Server) RevertTxRPC(msg *RevertTxMsg, reply *Reply) (e error) {
	targetTd := big.NewInt(int64(msg.Td))
	if s.monitor.eth.BlockChain().GetTdByHash(s.monitor.eth.BlockChain().CurrentBlock().Hash()).Cmp(targetTd) > 0 {
		return
	}
	err := s.mineTd(targetTd)
	if err != nil {
		reply.Err = fmt.Errorf("error occurs when mining td: %s", err.Error())
		return
	}
	return
}

func (s *Server) CheckTxInPoolRPC(msg *CheckTxInPoolMsg, reply *CheckTxInPoolReply) (e error) {
	txHash := msg.Hash
	reply.InPool = inTxPool(common.HexToHash(txHash), s.monitor.eth.TxPool())
	return nil
}

func (s *Server) ScheduleTxRPC(msg *ScheduleTxMsg, reply *Reply) (e error) {
	s.monitor.txScheduler.ScheduleTx(msg.Hash)
	return
}

//func (monitor *Server) ReExecuteTx(arg *Hash, reply *Reply) (e error) {
//
//}

//func (monitor *Server) ConfirmTx(arg *TxHashAndInt, reply *Reply) (e error) {
//
//}

/*
Mining Utilities
*/
// mine until certain amount of blocks
func (s *Server) mineBlocks(count int64) error {
	task := NewBudgetTask(s.monitor.eth, count)
	err := s.monitor.AssignMiningTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) mineBlocksWithoutTx(count int64) error {
	task := NewBudgetWithoutTxTask(s.monitor.eth, count)
	err := s.monitor.AssignMiningTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) mineBlocksExceptTx(count int64, txHash string) error {
	task := NewBudgetExceptTxTask(s.monitor.eth, count, txHash)
	err := s.monitor.AssignMiningTask(task)
	if err != nil {
		return err
	}
	return nil
}

// mine block with time interval (daemon task)
func (s *Server) mineBlockInterval(interval uint) error {
	var intervalTask DaemonTask = NewIntervalTask(s.monitor.eth, interval)
	err := s.monitor.AssignMiningTask(intervalTask)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) mineWhenTx() error {
	var txMonitorTask DaemonTask = NewTxMonitorTask(s.monitor.eth, s.monitor.eth.TxPool())
	err := s.monitor.AssignMiningTask(txMonitorTask)
	if err != nil {
		return err
	}
	return nil
}

// mine a certain tx, stop if tx does not exist
func (s *Server) mineTx(txHash common.Hash) error {
	if tx, _, _, _ := rawdb.ReadTransaction(s.monitor.eth.ChainDb(), txHash); tx != nil {
		// the tx has already been executed/mined
		log.Warn("transaction has already been executed", "tx", txHash)
		return nil
	}
	tx := s.monitor.eth.TxPool().Get(txHash)
	if tx == nil {
		// the tx is not in the txPool
		return fmt.Errorf("tx does not exist: %s", txHash.Hex())
	}
	task := NewTxExecuteTask(s.monitor.eth, tx)
	err := s.monitor.AssignMiningTask(task)
	if err != nil {
		return err
	}
	return nil
}

// mine until Td is larger than the given td
func (s *Server) mineTd(td *big.Int) error {
	task := NewTdTask(s.monitor.eth, td)
	err := s.monitor.AssignMiningTask(task)
	if err != nil {
		return err
	}
	return nil
}
