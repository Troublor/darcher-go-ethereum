package vm

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
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

func (p *EVMMonitorProxy) Analyzer() *Analyzer {
	return p.analyzer
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
func (p *EVMMonitorProxy) BeforeOperation(op OpCode, pc uint64, ctx *callCtx) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.BeforeCall(op, pc, ctx)
}

/**
Callback function after opcode is executed
*/
func (p *EVMMonitorProxy) AfterOperation(op OpCode, pc uint64, ctx *callCtx) {
	if !p.analyzerEnabled() {
		return
	}
	p.analyzer.AfterCall(op, pc, ctx)
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
