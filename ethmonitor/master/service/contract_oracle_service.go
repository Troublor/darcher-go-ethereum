package service

import (
	"context"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/protobuf/ptypes/empty"
	"io"
	"reflect"
	"sync"
)

type ContractOracleService struct {
	rpc.ContractVulnerabilityServiceServer

	// feed *rpc.TxErrorMsg, which comes from NotifyTxError rpc
	txErrorFeed event.Feed

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

// NotifyTxError rpc, called by ethmonitor.worker when tx execution error occurs
func (s *ContractOracleService) NotifyTxError(ctx context.Context, msg *rpc.TxErrorMsg) (*empty.Empty, error) {
	s.txErrorFeed.Send(msg)
	return &empty.Empty{}, nil
}

/**
Subscribe to *rpc.TxError
*/
func (s *ContractOracleService) SubscribeTxError(ch chan<- *rpc.TxErrorMsg) event.Subscription {
	return s.txErrorFeed.Subscribe(ch)
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

func (s *ContractOracleService) GetReportsByTransaction(role rpc.Role, txHash string) (doneCh chan *rpc.GetReportsByTransactionControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.GetReportsByTransactionControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.getReportsByTransactionRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.GetReportsByTransactionControlMsg{
			Role: role,
			Id:   common.GetUUID(),
			Hash: txHash,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.GetReportsByTransactionControlMsg)
	}()
	return doneCh, errCh
}
