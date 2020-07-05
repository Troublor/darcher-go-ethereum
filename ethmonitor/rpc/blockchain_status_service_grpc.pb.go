// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpc

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// BlockchainStatusServiceClient is the client API for BlockchainStatusService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlockchainStatusServiceClient interface {
	NotifyNewChainHead(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewChainHeadClient, error)
	NotifyNewChainSide(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewChainSideClient, error)
	NotifyNewTx(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewTxClient, error)
	GetHeadControl(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_GetHeadControlClient, error)
}

type blockchainStatusServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockchainStatusServiceClient(cc grpc.ClientConnInterface) BlockchainStatusServiceClient {
	return &blockchainStatusServiceClient{cc}
}

func (c *blockchainStatusServiceClient) NotifyNewChainHead(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewChainHeadClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockchainStatusService_serviceDesc.Streams[0], "/ethmonitor.BlockchainStatusService/notifyNewChainHead", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockchainStatusServiceNotifyNewChainHeadClient{stream}
	return x, nil
}

type BlockchainStatusService_NotifyNewChainHeadClient interface {
	Send(*ChainHead) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type blockchainStatusServiceNotifyNewChainHeadClient struct {
	grpc.ClientStream
}

func (x *blockchainStatusServiceNotifyNewChainHeadClient) Send(m *ChainHead) error {
	return x.ClientStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewChainHeadClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blockchainStatusServiceClient) NotifyNewChainSide(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewChainSideClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockchainStatusService_serviceDesc.Streams[1], "/ethmonitor.BlockchainStatusService/notifyNewChainSide", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockchainStatusServiceNotifyNewChainSideClient{stream}
	return x, nil
}

type BlockchainStatusService_NotifyNewChainSideClient interface {
	Send(*ChainSide) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type blockchainStatusServiceNotifyNewChainSideClient struct {
	grpc.ClientStream
}

func (x *blockchainStatusServiceNotifyNewChainSideClient) Send(m *ChainSide) error {
	return x.ClientStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewChainSideClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blockchainStatusServiceClient) NotifyNewTx(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_NotifyNewTxClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockchainStatusService_serviceDesc.Streams[2], "/ethmonitor.BlockchainStatusService/notifyNewTx", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockchainStatusServiceNotifyNewTxClient{stream}
	return x, nil
}

type BlockchainStatusService_NotifyNewTxClient interface {
	Send(*Tx) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type blockchainStatusServiceNotifyNewTxClient struct {
	grpc.ClientStream
}

func (x *blockchainStatusServiceNotifyNewTxClient) Send(m *Tx) error {
	return x.ClientStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewTxClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blockchainStatusServiceClient) GetHeadControl(ctx context.Context, opts ...grpc.CallOption) (BlockchainStatusService_GetHeadControlClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockchainStatusService_serviceDesc.Streams[3], "/ethmonitor.BlockchainStatusService/getHeadControl", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockchainStatusServiceGetHeadControlClient{stream}
	return x, nil
}

type BlockchainStatusService_GetHeadControlClient interface {
	Send(*GetChainHeadControlMsg) error
	Recv() (*GetChainHeadControlMsg, error)
	grpc.ClientStream
}

type blockchainStatusServiceGetHeadControlClient struct {
	grpc.ClientStream
}

func (x *blockchainStatusServiceGetHeadControlClient) Send(m *GetChainHeadControlMsg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *blockchainStatusServiceGetHeadControlClient) Recv() (*GetChainHeadControlMsg, error) {
	m := new(GetChainHeadControlMsg)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BlockchainStatusServiceServer is the server API for BlockchainStatusService service.
// All implementations must embed UnimplementedBlockchainStatusServiceServer
// for forward compatibility
type BlockchainStatusServiceServer interface {
	NotifyNewChainHead(BlockchainStatusService_NotifyNewChainHeadServer) error
	NotifyNewChainSide(BlockchainStatusService_NotifyNewChainSideServer) error
	NotifyNewTx(BlockchainStatusService_NotifyNewTxServer) error
	GetHeadControl(BlockchainStatusService_GetHeadControlServer) error
	mustEmbedUnimplementedBlockchainStatusServiceServer()
}

// UnimplementedBlockchainStatusServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBlockchainStatusServiceServer struct {
}

func (*UnimplementedBlockchainStatusServiceServer) NotifyNewChainHead(BlockchainStatusService_NotifyNewChainHeadServer) error {
	return status.Errorf(codes.Unimplemented, "method NotifyNewChainHead not implemented")
}
func (*UnimplementedBlockchainStatusServiceServer) NotifyNewChainSide(BlockchainStatusService_NotifyNewChainSideServer) error {
	return status.Errorf(codes.Unimplemented, "method NotifyNewChainSide not implemented")
}
func (*UnimplementedBlockchainStatusServiceServer) NotifyNewTx(BlockchainStatusService_NotifyNewTxServer) error {
	return status.Errorf(codes.Unimplemented, "method NotifyNewTx not implemented")
}
func (*UnimplementedBlockchainStatusServiceServer) GetHeadControl(BlockchainStatusService_GetHeadControlServer) error {
	return status.Errorf(codes.Unimplemented, "method GetHeadControl not implemented")
}
func (*UnimplementedBlockchainStatusServiceServer) mustEmbedUnimplementedBlockchainStatusServiceServer() {
}

func RegisterBlockchainStatusServiceServer(s *grpc.Server, srv BlockchainStatusServiceServer) {
	s.RegisterService(&_BlockchainStatusService_serviceDesc, srv)
}

func _BlockchainStatusService_NotifyNewChainHead_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockchainStatusServiceServer).NotifyNewChainHead(&blockchainStatusServiceNotifyNewChainHeadServer{stream})
}

type BlockchainStatusService_NotifyNewChainHeadServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*ChainHead, error)
	grpc.ServerStream
}

type blockchainStatusServiceNotifyNewChainHeadServer struct {
	grpc.ServerStream
}

func (x *blockchainStatusServiceNotifyNewChainHeadServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewChainHeadServer) Recv() (*ChainHead, error) {
	m := new(ChainHead)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BlockchainStatusService_NotifyNewChainSide_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockchainStatusServiceServer).NotifyNewChainSide(&blockchainStatusServiceNotifyNewChainSideServer{stream})
}

type BlockchainStatusService_NotifyNewChainSideServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*ChainSide, error)
	grpc.ServerStream
}

type blockchainStatusServiceNotifyNewChainSideServer struct {
	grpc.ServerStream
}

func (x *blockchainStatusServiceNotifyNewChainSideServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewChainSideServer) Recv() (*ChainSide, error) {
	m := new(ChainSide)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BlockchainStatusService_NotifyNewTx_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockchainStatusServiceServer).NotifyNewTx(&blockchainStatusServiceNotifyNewTxServer{stream})
}

type BlockchainStatusService_NotifyNewTxServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*Tx, error)
	grpc.ServerStream
}

type blockchainStatusServiceNotifyNewTxServer struct {
	grpc.ServerStream
}

func (x *blockchainStatusServiceNotifyNewTxServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockchainStatusServiceNotifyNewTxServer) Recv() (*Tx, error) {
	m := new(Tx)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BlockchainStatusService_GetHeadControl_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockchainStatusServiceServer).GetHeadControl(&blockchainStatusServiceGetHeadControlServer{stream})
}

type BlockchainStatusService_GetHeadControlServer interface {
	Send(*GetChainHeadControlMsg) error
	Recv() (*GetChainHeadControlMsg, error)
	grpc.ServerStream
}

type blockchainStatusServiceGetHeadControlServer struct {
	grpc.ServerStream
}

func (x *blockchainStatusServiceGetHeadControlServer) Send(m *GetChainHeadControlMsg) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockchainStatusServiceGetHeadControlServer) Recv() (*GetChainHeadControlMsg, error) {
	m := new(GetChainHeadControlMsg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _BlockchainStatusService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ethmonitor.BlockchainStatusService",
	HandlerType: (*BlockchainStatusServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "notifyNewChainHead",
			Handler:       _BlockchainStatusService_NotifyNewChainHead_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "notifyNewChainSide",
			Handler:       _BlockchainStatusService_NotifyNewChainSide_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "notifyNewTx",
			Handler:       _BlockchainStatusService_NotifyNewTx_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "getHeadControl",
			Handler:       _BlockchainStatusService_GetHeadControl_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "blockchain_status_service.proto",
}
