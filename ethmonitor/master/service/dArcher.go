package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/log"
	"io/ioutil"
	"net/http"
)

// dArcher is the upper controller to control the dApp testing
type DArcher struct {
	//conn   *rpc.Client
	ip   string
	port int
}

func NewDArcher(ip string, port int) *DArcher {
	//endpoint := fmt.Sprintf("%s:%d", ip, port)
	//c, err := rpc.DialHTTP("tcp", endpoint)
	//if err != nil {
	//	logger.Error("dArcher connection failed", "endpoint", endpoint, "err", err.Error())
	//	return nil
	//}
	dArcher := &DArcher{
		//conn:   c,
		ip:   ip,
		port: port,
	}
	return dArcher
}

func (d *DArcher) jsonRPC(method string, msg interface{}, reply interface{}) {
	jsonRpcMsg := JsonRpcMsg{
		JsonRpc: "2.0",
		Id:      1,
		Method:  method,
		Params:  msg,
	}
	jsonStr, err := json.Marshal(jsonRpcMsg)
	if err != nil {
		log.Error("bad jsonrpc msg", "msg", msg, "err", err)
		return
	}
	resp, err := http.Post(
		fmt.Sprintf("http://%s:%d", d.ip, d.port),
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		log.Error("jsonrpc call error", "err", err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("jsonrpc bad response", "err", err)
		return
	}
	err = json.Unmarshal(b, reply)
	if err != nil {
		log.Error("unmarshal jsonrpc response failed", "err", err)
		return
	}
}

func (d *DArcher) NotifyTxStateChange(txHash string, currentState common.LifecycleState, nextState common.LifecycleState) {
	arg := &TxStateChangeMsg{
		Hash:         txHash,
		CurrentState: currentState,
		NextState:    nextState,
	}
	reply := &Reply{}
	d.jsonRPC("TxStateChangeRPC", arg, reply)
	if reply.Err != nil {
		log.Error("TxStateChangeRPC call failed", "err", reply.Err)
	}
}

func (d *DArcher) NotifyTxReceived(txHash string) {
	reply := &Reply{}
	d.jsonRPC("TxReceivedRPC", &TxReceivedMsg{Hash: txHash}, reply)
	if reply.Err != nil {
		log.Error("TxReceivedRPC call failed", "err", reply.Err)
	}
}

func (d *DArcher) NotifyPivotReached(txHash string, currentState common.LifecycleState, nextState common.LifecycleState) {
	arg := &PivotReachedMsg{
		Hash:         txHash,
		CurrentState: currentState,
		NextState:    nextState,
	}
	reply := &Reply{}
	d.jsonRPC("PivotReachedRPC", arg, reply)
	if reply.Err != nil {
		log.Error("PivotReachedRPC call failed", "err", reply.Err)
	}
}

func (d *DArcher) NotifyTxFinished(txHash string) {
	arg := &TxFinishedMsg{Hash: txHash}
	reply := &Reply{}
	d.jsonRPC("TxFinishedRPC", arg, reply)
	if reply.Err != nil {
		log.Error("TxFinishedRPC call failed", "err", reply.Err)
	}
}

func (d *DArcher) NotifyTxStart(txHash string) {
	arg := &TxStartMsg{Hash: txHash}
	reply := &Reply{}
	d.jsonRPC("TxStartRPC", arg, reply)
	if reply.Err != nil {
		log.Error("TxStartRPC call failed", "err", reply.Err)
	}
}

func (d *DArcher) SelectTxToTraverse(txHashes []string) string {
	arg := &SelectTxToTraverseMsg{Hashes: txHashes}
	reply := &SelectTxToTraverseReply{}
	d.jsonRPC("SelectTxToTraverseRPC", arg, reply)
	if reply.Err != nil {
		log.Error("SelectTxToTraverseRPC call failed", "err", reply.Err)
	}
	return reply.Hash
}

func (d *DArcher) Test() {
	arg := &TestMsg{Msg: "hello json rpc"}
	reply := &Reply{}
	d.jsonRPC("HelloRPC", arg, reply)
	log.Debug("got reply from HelloRPC", "reply", reply)
}
