package service

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"net"
)

type EthMonitorServer struct {
	port int

	blockchainStatusService *BlockchainStatusService
	p2pNetworkService       *P2PNetworkService
	miningService           *MiningService
	contractOracleService   *ContractOracleService
}

func NewServer(port int) *EthMonitorServer {
	return &EthMonitorServer{port: port}
}

func (s *EthMonitorServer) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Error("Start master monitor server failed", "err", err)
		return
	}
	grpcServer := grpc.NewServer()
	s.blockchainStatusService = NewBlockchainStatusService()
	s.p2pNetworkService = NewP2PNetworkService()
	s.miningService = NewMiningService()
	s.contractOracleService = NewContractOracleService()
	rpc.RegisterBlockchainStatusServiceServer(grpcServer, s.blockchainStatusService)
	rpc.RegisterP2PNetworkServiceServer(grpcServer, s.p2pNetworkService)
	rpc.RegisterMiningServiceServer(grpcServer, s.miningService)
	rpc.RegisterContractVulnerabilityServiceServer(grpcServer, s.contractOracleService)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			log.Error("Master monitor server serve error", "err", err)
		}
	}()
}

func (s *EthMonitorServer) BlockchainStatusService() *BlockchainStatusService {
	return s.blockchainStatusService
}

func (s *EthMonitorServer) P2PNetworkService() *P2PNetworkService {
	return s.p2pNetworkService
}

func (s *EthMonitorServer) MiningService() *MiningService {
	return s.miningService
}

func (s *EthMonitorServer) ContractOracleService() *ContractOracleService {
	return s.contractOracleService
}
