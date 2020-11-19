// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.22.0
// 	protoc        v3.11.4
// source: mining_service.proto

package rpc

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type MineBlocksControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role  Role       `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id    string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Count uint64     `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Head  *ChainHead `protobuf:"bytes,4,opt,name=head,proto3" json:"head,omitempty"`
	Err   Error      `protobuf:"varint,5,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *MineBlocksControlMsg) Reset() {
	*x = MineBlocksControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MineBlocksControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MineBlocksControlMsg) ProtoMessage() {}

func (x *MineBlocksControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MineBlocksControlMsg.ProtoReflect.Descriptor instead.
func (*MineBlocksControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{0}
}

func (x *MineBlocksControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *MineBlocksControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MineBlocksControlMsg) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *MineBlocksControlMsg) GetHead() *ChainHead {
	if x != nil {
		return x.Head
	}
	return nil
}

func (x *MineBlocksControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type MineBlocksExceptTxControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role   Role       `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id     string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Count  uint64     `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	TxHash string     `protobuf:"bytes,4,opt,name=tx_hash,json=txHash,proto3" json:"tx_hash,omitempty"`
	Head   *ChainHead `protobuf:"bytes,5,opt,name=head,proto3" json:"head,omitempty"`
	Err    Error      `protobuf:"varint,6,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *MineBlocksExceptTxControlMsg) Reset() {
	*x = MineBlocksExceptTxControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MineBlocksExceptTxControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MineBlocksExceptTxControlMsg) ProtoMessage() {}

func (x *MineBlocksExceptTxControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MineBlocksExceptTxControlMsg.ProtoReflect.Descriptor instead.
func (*MineBlocksExceptTxControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{1}
}

func (x *MineBlocksExceptTxControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *MineBlocksExceptTxControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MineBlocksExceptTxControlMsg) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *MineBlocksExceptTxControlMsg) GetTxHash() string {
	if x != nil {
		return x.TxHash
	}
	return ""
}

func (x *MineBlocksExceptTxControlMsg) GetHead() *ChainHead {
	if x != nil {
		return x.Head
	}
	return nil
}

func (x *MineBlocksExceptTxControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type MineBlocksWithoutTxControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role  Role       `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id    string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Count uint64     `protobuf:"varint,3,opt,name=count,proto3" json:"count,omitempty"`
	Head  *ChainHead `protobuf:"bytes,4,opt,name=head,proto3" json:"head,omitempty"`
	Err   Error      `protobuf:"varint,5,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *MineBlocksWithoutTxControlMsg) Reset() {
	*x = MineBlocksWithoutTxControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MineBlocksWithoutTxControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MineBlocksWithoutTxControlMsg) ProtoMessage() {}

func (x *MineBlocksWithoutTxControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MineBlocksWithoutTxControlMsg.ProtoReflect.Descriptor instead.
func (*MineBlocksWithoutTxControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{2}
}

func (x *MineBlocksWithoutTxControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *MineBlocksWithoutTxControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MineBlocksWithoutTxControlMsg) GetCount() uint64 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *MineBlocksWithoutTxControlMsg) GetHead() *ChainHead {
	if x != nil {
		return x.Head
	}
	return nil
}

func (x *MineBlocksWithoutTxControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type MineTdControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role Role       `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id   string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Td   uint64     `protobuf:"varint,3,opt,name=td,proto3" json:"td,omitempty"`
	Head *ChainHead `protobuf:"bytes,4,opt,name=head,proto3" json:"head,omitempty"`
	Err  Error      `protobuf:"varint,5,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *MineTdControlMsg) Reset() {
	*x = MineTdControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MineTdControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MineTdControlMsg) ProtoMessage() {}

func (x *MineTdControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MineTdControlMsg.ProtoReflect.Descriptor instead.
func (*MineTdControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{3}
}

func (x *MineTdControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *MineTdControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MineTdControlMsg) GetTd() uint64 {
	if x != nil {
		return x.Td
	}
	return 0
}

func (x *MineTdControlMsg) GetHead() *ChainHead {
	if x != nil {
		return x.Head
	}
	return nil
}

func (x *MineTdControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type MineTxControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role Role       `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id   string     `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Hash string     `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	Head *ChainHead `protobuf:"bytes,4,opt,name=head,proto3" json:"head,omitempty"`
	Err  Error      `protobuf:"varint,5,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *MineTxControlMsg) Reset() {
	*x = MineTxControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MineTxControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MineTxControlMsg) ProtoMessage() {}

func (x *MineTxControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MineTxControlMsg.ProtoReflect.Descriptor instead.
func (*MineTxControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{4}
}

func (x *MineTxControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *MineTxControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MineTxControlMsg) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *MineTxControlMsg) GetHead() *ChainHead {
	if x != nil {
		return x.Head
	}
	return nil
}

func (x *MineTxControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type ScheduleTxControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role Role   `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id   string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Hash string `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	Err  Error  `protobuf:"varint,4,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *ScheduleTxControlMsg) Reset() {
	*x = ScheduleTxControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScheduleTxControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScheduleTxControlMsg) ProtoMessage() {}

func (x *ScheduleTxControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScheduleTxControlMsg.ProtoReflect.Descriptor instead.
func (*ScheduleTxControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{5}
}

func (x *ScheduleTxControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *ScheduleTxControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ScheduleTxControlMsg) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *ScheduleTxControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

type CheckTxInPoolControlMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role   Role   `protobuf:"varint,1,opt,name=role,proto3,enum=darcher.Role" json:"role,omitempty"`
	Id     string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Hash   string `protobuf:"bytes,3,opt,name=hash,proto3" json:"hash,omitempty"`
	InPool bool   `protobuf:"varint,4,opt,name=in_pool,json=inPool,proto3" json:"in_pool,omitempty"`
	Err    Error  `protobuf:"varint,5,opt,name=err,proto3,enum=darcher.Error" json:"err,omitempty"`
}

func (x *CheckTxInPoolControlMsg) Reset() {
	*x = CheckTxInPoolControlMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mining_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckTxInPoolControlMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckTxInPoolControlMsg) ProtoMessage() {}

func (x *CheckTxInPoolControlMsg) ProtoReflect() protoreflect.Message {
	mi := &file_mining_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckTxInPoolControlMsg.ProtoReflect.Descriptor instead.
func (*CheckTxInPoolControlMsg) Descriptor() ([]byte, []int) {
	return file_mining_service_proto_rawDescGZIP(), []int{6}
}

func (x *CheckTxInPoolControlMsg) GetRole() Role {
	if x != nil {
		return x.Role
	}
	return Role_DOER
}

func (x *CheckTxInPoolControlMsg) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CheckTxInPoolControlMsg) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *CheckTxInPoolControlMsg) GetInPool() bool {
	if x != nil {
		return x.InPool
	}
	return false
}

func (x *CheckTxInPoolControlMsg) GetErr() Error {
	if x != nil {
		return x.Err
	}
	return Error_NilErr
}

var File_mining_service_proto protoreflect.FileDescriptor

var file_mining_service_proto_rawDesc = []byte{
	0x0a, 0x14, 0x6d, 0x69, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x1a,
	0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa9,
	0x01, 0x0a, 0x14, 0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x12, 0x21, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x12, 0x26, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12,
	0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x48, 0x65,
	0x61, 0x64, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x22, 0xca, 0x01, 0x0a, 0x1c, 0x4d,
	0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x45, 0x78, 0x63, 0x65, 0x70, 0x74, 0x54,
	0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x12, 0x21, 0x0a, 0x04, 0x72,
	0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x74, 0x78, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x12, 0x26, 0x0a,
	0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x61,
	0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x52,
	0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x22, 0xb2, 0x01, 0x0a, 0x1d, 0x4d, 0x69, 0x6e, 0x65,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x57, 0x69, 0x74, 0x68, 0x6f, 0x75, 0x74, 0x54, 0x78, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x12, 0x21, 0x0a, 0x04, 0x72, 0x6f, 0x6c,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65,
	0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x26, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x12, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e,
	0x48, 0x65, 0x61, 0x64, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72,
	0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65,
	0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x22, 0x9f, 0x01, 0x0a,
	0x10, 0x4d, 0x69, 0x6e, 0x65, 0x54, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73,
	0x67, 0x12, 0x21, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x02, 0x74, 0x64, 0x12, 0x26, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x61,
	0x69, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x52, 0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x20, 0x0a, 0x03,
	0x65, 0x72, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x03, 0x65, 0x72, 0x72, 0x22, 0xa3,
	0x01, 0x0a, 0x10, 0x4d, 0x69, 0x6e, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x4d, 0x73, 0x67, 0x12, 0x21, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65,
	0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x26, 0x0a, 0x04, 0x68, 0x65,
	0x61, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68,
	0x65, 0x72, 0x2e, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x52, 0x04, 0x68, 0x65,
	0x61, 0x64, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52,
	0x03, 0x65, 0x72, 0x72, 0x22, 0x7f, 0x0a, 0x14, 0x53, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65,
	0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x12, 0x21, 0x0a, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x64, 0x61, 0x72,
	0x63, 0x68, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68,
	0x61, 0x73, 0x68, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0e, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x52, 0x03, 0x65, 0x72, 0x72, 0x22, 0x9b, 0x01, 0x0a, 0x17, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54,
	0x78, 0x49, 0x6e, 0x50, 0x6f, 0x6f, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73,
	0x67, 0x12, 0x21, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x0d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x52, 0x04,
	0x72, 0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x17, 0x0a, 0x07, 0x69, 0x6e, 0x5f, 0x70,
	0x6f, 0x6f, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x6e, 0x50, 0x6f, 0x6f,
	0x6c, 0x12, 0x20, 0x0a, 0x03, 0x65, 0x72, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e,
	0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x03,
	0x65, 0x72, 0x72, 0x32, 0xa2, 0x05, 0x0a, 0x0d, 0x4d, 0x69, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x57, 0x0a, 0x11, 0x6d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x73, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x1d, 0x2e, 0x64, 0x61, 0x72,
	0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x1a, 0x1d, 0x2e, 0x64, 0x61, 0x72, 0x63,
	0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x43, 0x6f,
	0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x6f,
	0x0a, 0x19, 0x6d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x45, 0x78, 0x63, 0x65,
	0x70, 0x74, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x25, 0x2e, 0x64, 0x61,
	0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73,
	0x45, 0x78, 0x63, 0x65, 0x70, 0x74, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d,
	0x73, 0x67, 0x1a, 0x25, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e,
	0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x45, 0x78, 0x63, 0x65, 0x70, 0x74, 0x54, 0x78, 0x43,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12,
	0x72, 0x0a, 0x1a, 0x6d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x57, 0x69, 0x74,
	0x68, 0x6f, 0x75, 0x74, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x26, 0x2e,
	0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x73, 0x57, 0x69, 0x74, 0x68, 0x6f, 0x75, 0x74, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x1a, 0x26, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x4d, 0x69, 0x6e, 0x65, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x57, 0x69, 0x74, 0x68, 0x6f, 0x75,
	0x74, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x22, 0x00, 0x28,
	0x01, 0x30, 0x01, 0x12, 0x4b, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x65, 0x54, 0x64, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x12, 0x19, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d,
	0x69, 0x6e, 0x65, 0x54, 0x64, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x1a,
	0x19, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x54, 0x64,
	0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01,
	0x12, 0x4b, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x12, 0x19, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65,
	0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x1a, 0x19, 0x2e, 0x64,
	0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x57, 0x0a,
	0x11, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72,
	0x6f, 0x6c, 0x12, 0x1d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x68,
	0x65, 0x64, 0x75, 0x6c, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73,
	0x67, 0x1a, 0x1d, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x53, 0x63, 0x68, 0x65,
	0x64, 0x75, 0x6c, 0x65, 0x54, 0x78, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67,
	0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x60, 0x0a, 0x14, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x54,
	0x78, 0x49, 0x6e, 0x50, 0x6f, 0x6f, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x12, 0x20,
	0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x54, 0x78,
	0x49, 0x6e, 0x50, 0x6f, 0x6f, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d, 0x73, 0x67,
	0x1a, 0x20, 0x2e, 0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x54, 0x78, 0x49, 0x6e, 0x50, 0x6f, 0x6f, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x4d,
	0x73, 0x67, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x72, 0x6f, 0x75, 0x62, 0x6c, 0x6f, 0x72, 0x2f,
	0x64, 0x61, 0x72, 0x63, 0x68, 0x65, 0x72, 0x2d, 0x67, 0x6f, 0x2d, 0x65, 0x74, 0x68, 0x65, 0x72,
	0x65, 0x75, 0x6d, 0x2f, 0x65, 0x74, 0x68, 0x6d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72, 0x2f, 0x72,
	0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mining_service_proto_rawDescOnce sync.Once
	file_mining_service_proto_rawDescData = file_mining_service_proto_rawDesc
)

func file_mining_service_proto_rawDescGZIP() []byte {
	file_mining_service_proto_rawDescOnce.Do(func() {
		file_mining_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_mining_service_proto_rawDescData)
	})
	return file_mining_service_proto_rawDescData
}

var file_mining_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_mining_service_proto_goTypes = []interface{}{
	(*MineBlocksControlMsg)(nil),          // 0: darcher.MineBlocksControlMsg
	(*MineBlocksExceptTxControlMsg)(nil),  // 1: darcher.MineBlocksExceptTxControlMsg
	(*MineBlocksWithoutTxControlMsg)(nil), // 2: darcher.MineBlocksWithoutTxControlMsg
	(*MineTdControlMsg)(nil),              // 3: darcher.MineTdControlMsg
	(*MineTxControlMsg)(nil),              // 4: darcher.MineTxControlMsg
	(*ScheduleTxControlMsg)(nil),          // 5: darcher.ScheduleTxControlMsg
	(*CheckTxInPoolControlMsg)(nil),       // 6: darcher.CheckTxInPoolControlMsg
	(Role)(0),                             // 7: darcher.Role
	(*ChainHead)(nil),                     // 8: darcher.ChainHead
	(Error)(0),                            // 9: darcher.Error
}
var file_mining_service_proto_depIdxs = []int32{
	7,  // 0: darcher.MineBlocksControlMsg.role:type_name -> darcher.Role
	8,  // 1: darcher.MineBlocksControlMsg.head:type_name -> darcher.ChainHead
	9,  // 2: darcher.MineBlocksControlMsg.err:type_name -> darcher.Error
	7,  // 3: darcher.MineBlocksExceptTxControlMsg.role:type_name -> darcher.Role
	8,  // 4: darcher.MineBlocksExceptTxControlMsg.head:type_name -> darcher.ChainHead
	9,  // 5: darcher.MineBlocksExceptTxControlMsg.err:type_name -> darcher.Error
	7,  // 6: darcher.MineBlocksWithoutTxControlMsg.role:type_name -> darcher.Role
	8,  // 7: darcher.MineBlocksWithoutTxControlMsg.head:type_name -> darcher.ChainHead
	9,  // 8: darcher.MineBlocksWithoutTxControlMsg.err:type_name -> darcher.Error
	7,  // 9: darcher.MineTdControlMsg.role:type_name -> darcher.Role
	8,  // 10: darcher.MineTdControlMsg.head:type_name -> darcher.ChainHead
	9,  // 11: darcher.MineTdControlMsg.err:type_name -> darcher.Error
	7,  // 12: darcher.MineTxControlMsg.role:type_name -> darcher.Role
	8,  // 13: darcher.MineTxControlMsg.head:type_name -> darcher.ChainHead
	9,  // 14: darcher.MineTxControlMsg.err:type_name -> darcher.Error
	7,  // 15: darcher.ScheduleTxControlMsg.role:type_name -> darcher.Role
	9,  // 16: darcher.ScheduleTxControlMsg.err:type_name -> darcher.Error
	7,  // 17: darcher.CheckTxInPoolControlMsg.role:type_name -> darcher.Role
	9,  // 18: darcher.CheckTxInPoolControlMsg.err:type_name -> darcher.Error
	0,  // 19: darcher.MiningService.mineBlocksControl:input_type -> darcher.MineBlocksControlMsg
	1,  // 20: darcher.MiningService.mineBlocksExceptTxControl:input_type -> darcher.MineBlocksExceptTxControlMsg
	2,  // 21: darcher.MiningService.mineBlocksWithoutTxControl:input_type -> darcher.MineBlocksWithoutTxControlMsg
	3,  // 22: darcher.MiningService.mineTdControl:input_type -> darcher.MineTdControlMsg
	4,  // 23: darcher.MiningService.mineTxControl:input_type -> darcher.MineTxControlMsg
	5,  // 24: darcher.MiningService.scheduleTxControl:input_type -> darcher.ScheduleTxControlMsg
	6,  // 25: darcher.MiningService.checkTxInPoolControl:input_type -> darcher.CheckTxInPoolControlMsg
	0,  // 26: darcher.MiningService.mineBlocksControl:output_type -> darcher.MineBlocksControlMsg
	1,  // 27: darcher.MiningService.mineBlocksExceptTxControl:output_type -> darcher.MineBlocksExceptTxControlMsg
	2,  // 28: darcher.MiningService.mineBlocksWithoutTxControl:output_type -> darcher.MineBlocksWithoutTxControlMsg
	3,  // 29: darcher.MiningService.mineTdControl:output_type -> darcher.MineTdControlMsg
	4,  // 30: darcher.MiningService.mineTxControl:output_type -> darcher.MineTxControlMsg
	5,  // 31: darcher.MiningService.scheduleTxControl:output_type -> darcher.ScheduleTxControlMsg
	6,  // 32: darcher.MiningService.checkTxInPoolControl:output_type -> darcher.CheckTxInPoolControlMsg
	26, // [26:33] is the sub-list for method output_type
	19, // [19:26] is the sub-list for method input_type
	19, // [19:19] is the sub-list for extension type_name
	19, // [19:19] is the sub-list for extension extendee
	0,  // [0:19] is the sub-list for field type_name
}

func init() { file_mining_service_proto_init() }
func file_mining_service_proto_init() {
	if File_mining_service_proto != nil {
		return
	}
	file_common_proto_init()
	file_blockchain_status_service_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mining_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MineBlocksControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MineBlocksExceptTxControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MineBlocksWithoutTxControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MineTdControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MineTxControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ScheduleTxControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mining_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckTxInPoolControlMsg); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mining_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mining_service_proto_goTypes,
		DependencyIndexes: file_mining_service_proto_depIdxs,
		MessageInfos:      file_mining_service_proto_msgTypes,
	}.Build()
	File_mining_service_proto = out.File
	file_mining_service_proto_rawDesc = nil
	file_mining_service_proto_goTypes = nil
	file_mining_service_proto_depIdxs = nil
}
