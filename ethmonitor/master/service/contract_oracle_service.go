package service

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"reflect"
	"sync"
)

type ContractOracleService struct {
	rpc.ContractVulnerabilityServiceServer

	/* reverse rpc */
	getReportsByContractRPCs    sync.Map
	getReportsByTransactionRPCs sync.Map
}

func NewContractOracleService() *ContractOracleService {
	return &ContractOracleService{}
}

/**
reverse rpc: get vulnerability reports by contract address
*/
func (s *ContractOracleService) GetReportsByContractControl(stream rpc.ContractVulnerabilityService_GetReportsByContractControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("GetReportsByContract control connected")
	rrpc := NewReverseRPC("getReportsByContract "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.getReportsByContractRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.getReportsByContractRPCs.Delete(in.GetRole())
	return nil
}

/**
reverse rpc: get vulnerability reports by transaction hash
*/
func (s *ContractOracleService) GetReportsByTransactionControl(stream rpc.ContractVulnerabilityService_GetReportsByTransactionControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("GetReportsByTransaction control connected")
	rrpc := NewReverseRPC("getReportsByTransaction "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.getReportsByTransactionRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.getReportsByTransactionRPCs.Delete(in.GetRole())
	return nil
}

func (s *ContractOracleService) GetReportsByContract(role rpc.Role, address string) (doneCh chan *rpc.GetReportsByContractControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.GetReportsByContractControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.getReportsByContractRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.GetReportsByContractControlMsg{
			Role:    role,
			Id:      common.GetUUID(),
			Address: address,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.GetReportsByContractControlMsg)
	}()
	return doneCh, errCh
}
