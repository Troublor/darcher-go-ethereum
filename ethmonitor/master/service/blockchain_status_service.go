package service

import (
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/protobuf/ptypes/empty"
	"io"
	"reflect"
	"sync"
)

/**
Blockchain status service
Provide subscription of new chain head, new chain side, new tx
*/
type BlockchainStatusService struct {
	rpc.UnimplementedBlockchainStatusServiceServer
	newTxFeed        event.Feed
	newChainHeadFeed event.Feed
	newChainSideFeed event.Feed

	/* reverse rpc */
	getHeadReverseRPCs sync.Map

	scope event.SubscriptionScope
}

func NewBlockchainStatusService() *BlockchainStatusService {
	return &BlockchainStatusService{}
}

func (s *BlockchainStatusService) Stop() {
	s.scope.Close()
}

/**
grpc handler: receive new chain head notification from geth
*/
func (s *BlockchainStatusService) NotifyNewChainHead(stream rpc.BlockchainStatusService_NotifyNewChainHeadServer) error {
	for {
		newChainHead, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		log.Debug("New chain head", "role", newChainHead.GetRole().String(), "number", newChainHead.GetNumber(), "hash", common.PrettifyHash(newChainHead.GetHash()))
		s.newChainHeadFeed.Send(newChainHead)
	}
}

/**
grpc handler: receive new chain side notification from geth
*/
func (s *BlockchainStatusService) NotifyNewChainSide(stream rpc.BlockchainStatusService_NotifyNewChainSideServer) error {
	for {
		newChainSide, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		log.Debug("New chain side", "role", newChainSide.GetRole().String(), "number", newChainSide.GetNumber(), "hash", common.PrettifyHash(newChainSide.GetHash()))
		s.newChainSideFeed.Send(newChainSide)
	}
}

/**
grpc handler: receive new tx notification from geth
*/
func (s *BlockchainStatusService) NotifyNewTx(stream rpc.BlockchainStatusService_NotifyNewTxServer) error {
	for {
		newTx, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		log.Debug("New tx", "role", newTx.GetRole().String(), "hash", common.PrettifyHash(newTx.GetHash()))
		s.newTxFeed.Send(newTx)
	}
}

/**
reverse rpc: get current block head
*/
func (s *BlockchainStatusService) GetHeadControl(stream rpc.BlockchainStatusService_GetHeadControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("GetHead control connected")
	rrpc := NewReverseRPC("getHeadControl "+string(in.Role), reflect.TypeOf(in), stream)
	s.getHeadReverseRPCs.Store(in.Role, rrpc)
	rrpc.Serve()
	s.getHeadReverseRPCs.Delete(in.Role)
	return nil
}

/**
Subscribe to the channel of new chain head
*/
func (s *BlockchainStatusService) SubscribeNewChainHead(ch chan<- *rpc.ChainHead) event.Subscription {
	return s.scope.Track(s.newChainHeadFeed.Subscribe(ch))
}

/**
Subscribe to the channel of new chain side
*/
func (s *BlockchainStatusService) SubscribeNewChainSide(ch chan<- *rpc.ChainSide) event.Subscription {
	return s.scope.Track(s.newChainSideFeed.Subscribe(ch))
}

/**
Subscribe to the channel of new tx
*/
func (s *BlockchainStatusService) SubscribeNewTx(ch chan<- *rpc.Tx) event.Subscription {
	return s.scope.Track(s.newTxFeed.Subscribe(ch))
}

func (s *BlockchainStatusService) GetHead(role rpc.Role) (doneCh chan *rpc.GetChainHeadControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.GetChainHeadControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.getHeadReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.GetChainHeadControlMsg{
			Role: role,
			Id:   common.GetUUID(),
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.GetChainHeadControlMsg)
	}()
	return doneCh, errCh
}
