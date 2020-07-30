package vm

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"io/ioutil"
	"math/big"
	"strings"
)

var proxy = &EVMMonitorProxy{}

/**
Proxy takes callback from evm and dispatch to analyzer when necessary
*/
func GetEVMMonitorProxy() *EVMMonitorProxy {
	return proxy
}

type EVMMonitorProxy struct {
	analyzer *Analyzer

	// tx error notifier, we use this notifier, which is set by ethmonitor.worker, to notify tx error for ethmonitor.master when any tx execution error happens
	txErrorNotifier func(msg *rpc.TxErrorMsg) error
}

/**
Called from outside evmmonitor to enable evm dynamic analyzer
*/
func EnableEVMAnalyzer() {
	proxy.analyzer = newAnalyzer()
	log.Info("EVM Analyzer enabled")
}

func (p *EVMMonitorProxy) analyzerEnabled() bool {
	return p.analyzer != nil
}

/**
Callback function before message call is performed
*/
func (p *EVMMonitorProxy) BeforeMessageCall(callType CallType, caller ContractRef, callee ContractRef, input []byte, gas uint64, value *big.Int) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.BeforeMessageCall(callType, caller, callee, input, gas, value)
}

/**
Callback function after message call is performed
*/
func (p *EVMMonitorProxy) AfterMessageCall(callType CallType, ret []byte, err error) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.AfterMessageCall(callType, ret, err)
}

/**
Callback function before opcode is executed
*/
func (p *EVMMonitorProxy) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.BeforeOperation(op, operation, pc, ctx)
}

/**
Callback function after opcode is executed
*/
func (p *EVMMonitorProxy) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.AfterOperation(op, operation, pc, ctx)
}

func (p *EVMMonitorProxy) BeforeTransaction(tx *types.Transaction) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.BeforeTransaction(tx)
}

func (p *EVMMonitorProxy) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.AfterTransaction(tx, receipt)
}

func (p *EVMMonitorProxy) Reports() []*rpc.ContractVulReport {
	if !p.analyzerEnabled() {
		return make([]*rpc.ContractVulReport, 0)
	}
	return p.analyzer.Reports()
}

func (p *EVMMonitorProxy) ReportsToFile(outputFile string) {
	reports := p.Reports()
	data, _ := json.MarshalIndent(reports, "", "  ")
	if strings.ToLower(outputFile) == "stdout" {
		fmt.Println(string(data))
	} else {
		err := ioutil.WriteFile(outputFile, data, 0644)
		if err != nil {
			log.Error("Output contract oracle report failed", "err", err)
		}
	}
}

func (p *EVMMonitorProxy) SetTxErrorNotifier(notifier func(msg *rpc.TxErrorMsg) error) {
	p.txErrorNotifier = notifier
}

func (p *EVMMonitorProxy) NotifyTxError(msg *rpc.TxErrorMsg) {
	if p.txErrorNotifier != nil {
		err := p.txErrorNotifier(msg)
		if err != nil {
			log.Error("Notify TxError failed", "err", err)
		}
	}
}
