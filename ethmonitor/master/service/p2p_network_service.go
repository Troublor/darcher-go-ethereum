package service

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/golang/protobuf/ptypes/empty"
	"io"
	"reflect"
	"sync"
)

var (
	ServiceNotAvailableErr = errors.New("service not available")
)

/**
Geth P2P network service
provide the service of geth node start event, add/remove peer on geth node
*/
type P2PNetworkService struct {
	rpc.UnimplementedP2PNetworkServiceServer
	newNodeFeed event.Feed

	addPeerReverseRPCs    sync.Map
	removePeerReverseRPCs sync.Map

	scope event.SubscriptionScope
}

func NewP2PNetworkService() *P2PNetworkService {
	return &P2PNetworkService{}
}

/**
grpc handler: receive geth node start notification
*/
func (s *P2PNetworkService) NotifyNodeStart(ctx context.Context, msg *rpc.Node) (*empty.Empty, error) {
	s.newNodeFeed.Send(msg)
	return &empty.Empty{}, nil
}

/**
grpc handler: bidirectional rpc, served as reverse rpc to add peer
*/
func (s *P2PNetworkService) AddPeerControl(stream rpc.P2PNetworkService_AddPeerControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("AddPeer control connected")
	rrpc := NewReverseRPC("addPeerControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.addPeerReverseRPCs.Store(in.Role, rrpc)
	rrpc.Serve()
	s.addPeerReverseRPCs.Delete(in.Role)
	return nil
}

/**
grpc handler: bidirectional rpc, served as reverse rpc to remove peer
*/
func (s *P2PNetworkService) RemovePeerControl(stream rpc.P2PNetworkService_RemovePeerControlServer) error {
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	log.Debug("RemovePeer control connected")
	rrpc := NewReverseRPC("removePeerControl "+in.GetRole().String(), reflect.TypeOf(in), stream)
	s.removePeerReverseRPCs.Store(in.Role, rrpc)
	rrpc.Serve()
	s.removePeerReverseRPCs.Delete(in.Role)
	return nil
}

/**
Subscribe to the channel of new node
*/
func (s *P2PNetworkService) SubscribeNewNode(ch chan<- *rpc.Node) event.Subscription {
	return s.scope.Track(s.newNodeFeed.Subscribe(ch))
}

/**
Async add peer method, this will send reverse rpc call to geth to add peer
*/
func (s *P2PNetworkService) AddPeer(role rpc.Role, peerUrl string) (doneCh chan *rpc.AddPeerControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.AddPeerControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.addPeerReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.AddPeerControlMsg{
			Role:   role,
			Id:     common.GetUUID(),
			Url:    peerUrl,
			PeerId: "",
			Err:    rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.AddPeerControlMsg)
	}()
	return doneCh, errCh
}

/**
Async remove peer method, this will send reverse rpc call to geth to remove peer
*/
func (s *P2PNetworkService) RemovePeer(role rpc.Role, peerUrl string) (doneCh chan *rpc.RemovePeerControlMsg, errCh chan error) {
	doneCh = make(chan *rpc.RemovePeerControlMsg, 1)
	errCh = make(chan error, 1)
	go func() {
		rrpc, ok := s.removePeerReverseRPCs.Load(role)
		if !ok {
			errCh <- ServiceNotAvailableErr
			return
		}
		reply, err := rrpc.(*ReverseRPC).Call(&rpc.RemovePeerControlMsg{
			Role:   role,
			Id:     common.GetUUID(),
			Url:    peerUrl,
			PeerId: "",
			Err:    rpc.Error_NilErr,
		})
		if err != nil {
			errCh <- err
			return
		}
		doneCh <- reply.(*rpc.RemovePeerControlMsg)
	}()
	return doneCh, errCh
}
