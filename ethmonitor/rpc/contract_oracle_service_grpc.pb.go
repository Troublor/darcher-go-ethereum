// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ContractVulnerabilityServiceClient is the client API for ContractVulnerabilityService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContractVulnerabilityServiceClient interface {
	GetReportsByContractControl(ctx context.Context, opts ...grpc.CallOption) (ContractVulnerabilityService_GetReportsByContractControlClient, error)
	GetReportsByTransactionControl(ctx context.Context, opts ...grpc.CallOption) (ContractVulnerabilityService_GetReportsByTransactionControlClient, error)
}

type contractVulnerabilityServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewContractVulnerabilityServiceClient(cc grpc.ClientConnInterface) ContractVulnerabilityServiceClient {
	return &contractVulnerabilityServiceClient{cc}
}

func (c *contractVulnerabilityServiceClient) GetReportsByContractControl(ctx context.Context, opts ...grpc.CallOption) (ContractVulnerabilityService_GetReportsByContractControlClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ContractVulnerabilityService_serviceDesc.Streams[0], "/darcher.ContractVulnerabilityService/getReportsByContractControl", opts...)
	if err != nil {
		return nil, err
	}
	x := &contractVulnerabilityServiceGetReportsByContractControlClient{stream}
	return x, nil
}

type ContractVulnerabilityService_GetReportsByContractControlClient interface {
	Send(*GetReportsByContractControlMsg) error
	Recv() (*GetReportsByContractControlMsg, error)
	grpc.ClientStream
}

type contractVulnerabilityServiceGetReportsByContractControlClient struct {
	grpc.ClientStream
}

func (x *contractVulnerabilityServiceGetReportsByContractControlClient) Send(m *GetReportsByContractControlMsg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *contractVulnerabilityServiceGetReportsByContractControlClient) Recv() (*GetReportsByContractControlMsg, error) {
	m := new(GetReportsByContractControlMsg)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *contractVulnerabilityServiceClient) GetReportsByTransactionControl(ctx context.Context, opts ...grpc.CallOption) (ContractVulnerabilityService_GetReportsByTransactionControlClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ContractVulnerabilityService_serviceDesc.Streams[1], "/darcher.ContractVulnerabilityService/getReportsByTransactionControl", opts...)
	if err != nil {
		return nil, err
	}
	x := &contractVulnerabilityServiceGetReportsByTransactionControlClient{stream}
	return x, nil
}

type ContractVulnerabilityService_GetReportsByTransactionControlClient interface {
	Send(*GetReportsByTransactionControlMsg) error
	Recv() (*GetReportsByTransactionControlMsg, error)
	grpc.ClientStream
}

type contractVulnerabilityServiceGetReportsByTransactionControlClient struct {
	grpc.ClientStream
}

func (x *contractVulnerabilityServiceGetReportsByTransactionControlClient) Send(m *GetReportsByTransactionControlMsg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *contractVulnerabilityServiceGetReportsByTransactionControlClient) Recv() (*GetReportsByTransactionControlMsg, error) {
	m := new(GetReportsByTransactionControlMsg)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ContractVulnerabilityServiceServer is the server API for ContractVulnerabilityService service.
// All implementations must embed UnimplementedContractVulnerabilityServiceServer
// for forward compatibility
type ContractVulnerabilityServiceServer interface {
	GetReportsByContractControl(ContractVulnerabilityService_GetReportsByContractControlServer) error
	GetReportsByTransactionControl(ContractVulnerabilityService_GetReportsByTransactionControlServer) error
	mustEmbedUnimplementedContractVulnerabilityServiceServer()
}

// UnimplementedContractVulnerabilityServiceServer must be embedded to have forward compatible implementations.
type UnimplementedContractVulnerabilityServiceServer struct {
}

func (*UnimplementedContractVulnerabilityServiceServer) GetReportsByContractControl(ContractVulnerabilityService_GetReportsByContractControlServer) error {
	return status.Errorf(codes.Unimplemented, "method GetReportsByContractControl not implemented")
}
func (*UnimplementedContractVulnerabilityServiceServer) GetReportsByTransactionControl(ContractVulnerabilityService_GetReportsByTransactionControlServer) error {
	return status.Errorf(codes.Unimplemented, "method GetReportsByTransactionControl not implemented")
}
func (*UnimplementedContractVulnerabilityServiceServer) mustEmbedUnimplementedContractVulnerabilityServiceServer() {
}

func RegisterContractVulnerabilityServiceServer(s *grpc.Server, srv ContractVulnerabilityServiceServer) {
	s.RegisterService(&_ContractVulnerabilityService_serviceDesc, srv)
}

func _ContractVulnerabilityService_GetReportsByContractControl_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContractVulnerabilityServiceServer).GetReportsByContractControl(&contractVulnerabilityServiceGetReportsByContractControlServer{stream})
}

type ContractVulnerabilityService_GetReportsByContractControlServer interface {
	Send(*GetReportsByContractControlMsg) error
	Recv() (*GetReportsByContractControlMsg, error)
	grpc.ServerStream
}

type contractVulnerabilityServiceGetReportsByContractControlServer struct {
	grpc.ServerStream
}

func (x *contractVulnerabilityServiceGetReportsByContractControlServer) Send(m *GetReportsByContractControlMsg) error {
	return x.ServerStream.SendMsg(m)
}

func (x *contractVulnerabilityServiceGetReportsByContractControlServer) Recv() (*GetReportsByContractControlMsg, error) {
	m := new(GetReportsByContractControlMsg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ContractVulnerabilityService_GetReportsByTransactionControl_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContractVulnerabilityServiceServer).GetReportsByTransactionControl(&contractVulnerabilityServiceGetReportsByTransactionControlServer{stream})
}

type ContractVulnerabilityService_GetReportsByTransactionControlServer interface {
	Send(*GetReportsByTransactionControlMsg) error
	Recv() (*GetReportsByTransactionControlMsg, error)
	grpc.ServerStream
}

type contractVulnerabilityServiceGetReportsByTransactionControlServer struct {
	grpc.ServerStream
}

func (x *contractVulnerabilityServiceGetReportsByTransactionControlServer) Send(m *GetReportsByTransactionControlMsg) error {
	return x.ServerStream.SendMsg(m)
}

func (x *contractVulnerabilityServiceGetReportsByTransactionControlServer) Recv() (*GetReportsByTransactionControlMsg, error) {
	m := new(GetReportsByTransactionControlMsg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _ContractVulnerabilityService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "darcher.ContractVulnerabilityService",
	HandlerType: (*ContractVulnerabilityServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "getReportsByContractControl",
			Handler:       _ContractVulnerabilityService_GetReportsByContractControl_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "getReportsByTransactionControl",
			Handler:       _ContractVulnerabilityService_GetReportsByTransactionControl_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "contract_oracle_service.proto",
}