package service

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"reflect"
	"sync"
)

type MiningService struct {
	rpc.UnimplementedMiningServiceServer
	mineBlocksReverseRPCs          sync.Map
	mineBlocksExceptTxReverseRPCs  sync.Map
	mineBlocksWithoutTxReverseRPCs sync.Map
	mineTdReverseRPCs              sync.Map
	mineTxReverseRPCs              sync.Map
	scheduleTxReverseRPCs          sync.Map
	checkTxInPoolReverseRPCs       sync.Map
}

func NewMiningService() *MiningService {
	return &MiningService{}
}

/* RPC methods start */

func (s *MiningService) ScheduleTxControl(stream rpc.MiningService_ScheduleTxControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("ScheduleTx control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("ScheduleTxControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.scheduleTxReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.scheduleTxReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) MineBlocksControl(stream rpc.MiningService_MineBlocksControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("MineBlocks control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("MineBlocksControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.mineBlocksReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.mineBlocksReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) MineBlocksExceptTxControl(stream rpc.MiningService_MineBlocksExceptTxControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("MineBlocksExceptTx control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("MineBlocksExceptTxControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.mineBlocksExceptTxReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.mineBlocksExceptTxReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) MineBlocksWithoutTxControl(stream rpc.MiningService_MineBlocksWithoutTxControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("MineBlocksWithoutTx control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("MineBlocksWithoutTxControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.mineBlocksWithoutTxReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.mineBlocksWithoutTxReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) MineTdControl(stream rpc.MiningService_MineTdControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("MineTd control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("MineTdControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.mineTdReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.mineTdReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) MineTxControl(stream rpc.MiningService_MineTxControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("MineTx control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("MineTxControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.mineTxReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.mineTxReverseRPCs.Delete(in.GetRole())
	return nil
}

func (s *MiningService) CheckTxInPoolControl(stream rpc.MiningService_CheckTxInPoolControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("CheckTxInPool control connected", "role", in.GetRole())
	rrpc := NewReverseRPC("CheckTxInPoolControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.checkTxInPoolReverseRPCs.Store(in.GetRole(), rrpc)
	rrpc.Serve()
	s.checkTxInPoolReverseRPCs.Delete(in.GetRole())
	return nil
}

/* RPC methods end */

/* Public methods start */

func (s *MiningService) MineBlocks(role rpc.Role, count uint64) (doneCh chan *rpc.MineBlocksControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.MineBlocksControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.mineBlocksReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.MineBlocksControlMsg{
			Role:  role,
			Id:    common.GetUUID(),
			Count: count,
			Err:   rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.MineBlocksControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) MineBlocksExceptTx(role rpc.Role, count uint64, txHash string) (doneCh chan *rpc.MineBlocksExceptTxControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.MineBlocksExceptTxControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.mineBlocksExceptTxReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.MineBlocksExceptTxControlMsg{
			Role:   role,
			Id:     common.GetUUID(),
			Count:  count,
			TxHash: txHash,
			Err:    rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.MineBlocksExceptTxControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) MineBlocksWithoutTx(role rpc.Role, count uint64) (doneCh chan *rpc.MineBlocksWithoutTxControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.MineBlocksWithoutTxControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.mineBlocksWithoutTxReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.MineBlocksWithoutTxControlMsg{
			Role:  role,
			Id:    common.GetUUID(),
			Count: count,
			Err:   rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.MineBlocksWithoutTxControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) MineTd(role rpc.Role, td uint64) (doneCh chan *rpc.MineTdControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.MineTdControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.mineTdReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.MineTdControlMsg{
			Role: role,
			Id:   common.GetUUID(),
			Td:   td,
			Err:  rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.MineTdControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) MineTx(role rpc.Role, txHash string) (doneCh chan *rpc.MineTxControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.MineTxControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.mineTxReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.MineTxControlMsg{
			Role: role,
			Id:   common.GetUUID(),
			Hash: txHash,
			Err:  rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.MineTxControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) ScheduleTx(role rpc.Role, txHash string) (doneCh chan *rpc.ScheduleTxControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.ScheduleTxControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.scheduleTxReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.ScheduleTxControlMsg{
			Role: role,
			Id:   common.GetUUID(),
			Hash: txHash,
			Err:  rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.ScheduleTxControlMsg)
	}()
	return doneCh, errCh
}

func (s *MiningService) CheckTxInPool(role rpc.Role, txHash string) (doneCh chan *rpc.CheckTxInPoolControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.CheckTxInPoolControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.checkTxInPoolReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.CheckTxInPoolControlMsg{
			Role: role,
			Id:   common.GetUUID(),
			Hash: txHash,
			Err:  rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.CheckTxInPoolControlMsg)
	}()
	return doneCh, errCh
}

/* Public methods end */
