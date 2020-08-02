package service

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"io"
	"reflect"
	"sync"
	"time"
)

type Identifiable interface {
	GetRole() rpc.Role
	GetId() string
}

/**
Reverse RPC
An application layer on top of bidirectional grpc
To simulate the behaviour that legacyServer send rpc call to clients
*/
type ReverseRPC struct {
	name         string
	replyType    reflect.Type
	mutex        sync.Mutex
	rpcStream    grpc.ServerStream
	pendingCalls sync.Map //map[rpc.Role]map[string]chan Identifiable
}

/**
Constructor takes
1. the name of ReverseRPC service (for log purpose)
2. reflect.Type of reverse rpc reply
3. grpc stream: the bidirectional grpc stream
*/
func NewReverseRPC(name string, replyType reflect.Type, rpcStream grpc.ServerStream) *ReverseRPC {
	for replyType.Kind() == reflect.Ptr {
		replyType = replyType.Elem()
	}
	return &ReverseRPC{
		name:      name,
		replyType: replyType,
		rpcStream: rpcStream,
	}
}

/**
Perform reverse rpc call (call from server to client)
*/
func (rrpc *ReverseRPC) Call(arg Identifiable) (reply Identifiable, err error) {
	replyCh := make(chan Identifiable, 1)
	// save in the pendingCalls map
	var pendingCalls interface{}
	var ok bool
	if pendingCalls, ok = rrpc.pendingCalls.Load(arg.GetRole()); !ok {
		pendingCalls = &sync.Map{}
		rrpc.pendingCalls.Store(arg.GetRole(), pendingCalls)
	}
	pendingCalls.(*sync.Map).Store(arg.GetId(), replyCh)

	// send reverser RPC to client
	timeout := time.After(common.ReverseRPCRetryTime)
	rrpc.mutex.Lock()
	err = rrpc.rpcStream.SendMsg(arg)
	rrpc.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	select {
	case <-timeout:
		log.Warn("Reverse rpc " + rrpc.name + " timeout, retrying")
		return rrpc.Call(arg)
	case reply = <-replyCh:
		return reply, nil
	}
}

/**
Reverse RPC serve loop, be persistent until grpc stream closed
*/
func (rrpc *ReverseRPC) Serve() {
	for {
		reply := reflect.New(rrpc.replyType).Interface().(Identifiable)
		err := rrpc.rpcStream.RecvMsg(reply)
		if err == io.EOF {
			return
		}
		if err != nil && err != context.Canceled {
			log.Error(fmt.Sprintf("ReverseRPC %s receive message error", rrpc.name), "err", err)
			return
		}
		if pendingCalls, ok := rrpc.pendingCalls.Load(reply.GetRole()); ok {
			if call, ok := pendingCalls.(*sync.Map).Load(reply.GetId()); ok {
				call.(chan Identifiable) <- reply
			}
		}
	}
}
