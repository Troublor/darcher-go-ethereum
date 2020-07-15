package vm

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"io/ioutil"
	"math/big"
	"strings"
)

/**
Analyzer will first collect runtime time information from evm and then check each oracles
*/
type Analyzer struct {
	callStack *GeneralStack

	oracles []Oracle
}

func newAnalyzer() *Analyzer {
	return &Analyzer{
		callStack: &GeneralStack{},
		oracles: []Oracle{
			NewGaslessSendOracle(),
			NewExceptionDisorderOracle(),
			NewReentrancyOracle(),
		},
	}
}

func (a *Analyzer) BeforeMessageCall(callType CallType, caller ContractRef, callee ContractRef, input []byte, gas uint64, value *big.Int) {
	var call MessageCall
	switch callType {
	case TYPE_CALL:
		call = &Call{
			caller:   caller.Address(),
			callee:   callee.Address(),
			value:    value,
			input:    input,
			gasLimit: gas,
		}
	case TYPE_CALLCODE:
		call = &CallCode{
			Call{
				caller:   caller.Address(),
				callee:   callee.Address(),
				value:    value,
				input:    input,
				gasLimit: gas,
			},
		}
	case TYPE_STATICCALL:
		call = &StaticCall{
			Call{
				caller:   caller.Address(),
				callee:   callee.Address(),
				value:    value,
				input:    input,
				gasLimit: gas,
			},
		}
	case TYPE_DELEGATECALL:
		call = &DelegateCall{
			Call{
				caller:   caller.Address(),
				callee:   callee.Address(),
				value:    value,
				input:    input,
				gasLimit: gas,
			},
		}
	case TYPE_CREATE:
		call = &Create{
			Call{
				caller:   caller.Address(),
				callee:   callee.Address(),
				value:    value,
				input:    input,
				gasLimit: gas,
			},
		}
	}
	for _, oracle := range a.oracles {
		oracle.BeforeMessageCall(a.callStack, call)
	}
	a.callStack.Push(call)
}

func (a *Analyzer) AfterMessageCall(callType CallType, ret []byte, err error) {
	call := a.callStack.Top().(MessageCall)
	call.setCallReturn(ret, err)
	for _, oracle := range a.oracles {
		oracle.AfterMessageCall(a.callStack, call)
	}
	a.callStack.Pop()
}

func (a *Analyzer) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	for _, oracle := range a.oracles {
		oracle.BeforeOperation(op, operation, pc, ctx)
	}
}

func (a *Analyzer) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	for _, oracle := range a.oracles {
		oracle.AfterOperation(op, operation, pc, ctx)
	}
}

func (a *Analyzer) BeforeTransaction(tx *types.Transaction) {
	for _, oracle := range a.oracles {
		oracle.BeforeTransaction(tx)
	}
}

func (a *Analyzer) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	for _, oracle := range a.oracles {
		oracle.AfterTransaction(tx, receipt)
	}
}

func (a *Analyzer) Report(outputFile string) {
	reports := make([]Report, 0)
	for _, oracle := range a.oracles {
		reports = append(reports, oracle.Report()...)
	}
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