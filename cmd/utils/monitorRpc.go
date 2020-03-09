package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	eth2 "github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/miner"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type TxStatus int

// transaction lifecycle status
const (
	CREATED TxStatus = iota
	PENDING
	EXECUTED
	REVERTED
	CONFIRMED
)

type TxScheduler struct {
	status TxStatus
	hash   common.Hash
	eth    *eth2.Ethereum
}

func NewTxScheduler(txHash common.Hash) *TxScheduler {
	return &TxScheduler{
		status: CREATED,
		hash:   txHash,
	}
}

////execute this transaction in blockchain
//func (s *TxScheduler) Execute() error {
//
//}
//
//// revert this transaction
//func (s *TxScheduler) Revert() error {
//
//}
//
//// re-execute this transaction
//func (s *TxScheduler) ReExecute() error {
//
//}
//
//// confirm this transaction
//func (s *TxScheduler) Confirm(n int) error {
//
//}

type TxHash string

type TxHashAndInt struct {
	Hash TxHash
	N    int
}

type Reply struct {
}

type MonitorRPC struct {
	role       miner.Role
	eth        *eth2.Ethereum
	schedulers []*TxScheduler
}

func NewMonitorRPC(eth *eth2.Ethereum, role miner.Role) *MonitorRPC {
	return &MonitorRPC{
		role:       role,
		eth:        eth,
		schedulers: make([]*TxScheduler, 1),
	}
}

func startMonitorRpcServer(handler *MonitorRPC) error {
	port := 2020
	err := rpc.Register(handler)
	if err != nil {
		return err
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
	return nil
}

//func (m *MonitorRPC) Start() error {
//	err := startMonitorRpcServer(m)
//	if err != nil {
//		return err
//	}
//}
//

// RPC Methods
func (m *MonitorRPC) ExecuteTx(arg *TxHash, reply *Reply) {

}

func (m *MonitorRPC) RevertTx(arg *TxHash, reply *Reply) {

}

func (m *MonitorRPC) ReExecuteTx(arg *TxHash, reply *Reply) {

}

func (m *MonitorRPC) ConfirmTx(arg *TxHashAndInt, reply *Reply) {

}
