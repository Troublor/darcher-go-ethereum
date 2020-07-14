package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
	"math/big"
)

type VulnerabilityType string

const (
	GASLESS_SEND       = "gasless_send"
	EXCEPTION_DISORDER = "exception_disorder"
)

type Report struct {
	Address     common.Address    `json:"address"`
	PC          uint64            `json:"pc"`
	Transaction common.Hash       `json:"transaction"`
	VulType     VulnerabilityType `json:"type"`
	Description string            `json:"description"`
}

type Oracle interface {
	Type() VulnerabilityType
	BeforeTransaction(tx *types.Transaction)
	BeforeMessageCall(callStack *GeneralStack, call MessageCall)
	BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx)
	AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx)
	AfterMessageCall(callStack *GeneralStack, call MessageCall)
	AfterTransaction(tx *types.Transaction, receipt *types.Receipt)

	Report() []Report
}

type GaslessSendOracle struct {
	currentTx *types.Transaction
	reports   []Report

	shouldCheck bool
	pc          int

	sendStack *GeneralStack
	pcStack   *GeneralStack
}

func NewGaslessSendOracle() *GaslessSendOracle {
	return &GaslessSendOracle{
		reports:     make([]Report, 0),
		shouldCheck: false,
		pc:          0,

		sendStack: &GeneralStack{},
		pcStack:   &GeneralStack{},
	}
}

func (o *GaslessSendOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *GaslessSendOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	if o.shouldCheck {
		o.sendStack.Push(call)
		o.shouldCheck = false
		log.Info("send start")
	}
}

func (o *GaslessSendOracle) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	if op == CALL {
		gasLimit := ctx.stack.Back(0)
		value := ctx.stack.Back(2)
		inOffset, inSize := ctx.stack.Back(3), ctx.stack.Back(4)
		args := ctx.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
		log.Info("CALL opcode", "gas", gasLimit, "value", value, "args", len(args))
		tmp, _ := uint256.FromBig(big.NewInt(2300))
		if !value.IsZero() &&
			(gasLimit.Cmp(tmp) == 0 || gasLimit.IsZero()) &&
			len(args) == 0 {
			o.shouldCheck = true
			o.pcStack.Push(pc)
			log.Info("got send")
		}
	}
}

func (o *GaslessSendOracle) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *GaslessSendOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	if o.sendStack.Len() > 0 && o.sendStack.Top().(MessageCall) == call {
		if call.OutOfGas() {
			// gasless send when a send() throws ErrOutOfGas
			o.reports = append(o.reports, Report{
				Address:     call.Caller(),
				VulType:     o.Type(),
				Transaction: o.currentTx.Hash(),
				PC:          o.pcStack.Top().(uint64),
				Description: fmt.Sprintf("gasless send at %d", o.pcStack.Top().(uint64)),
			})
			log.Info("gasless send")
		}
		o.sendStack.Pop()
		o.pcStack.Pop()
		log.Info("send finished")
	}
}

func (o *GaslessSendOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	return
}

func (o *GaslessSendOracle) Report() []Report {
	return o.reports
}

func (o *GaslessSendOracle) Type() VulnerabilityType {
	return GASLESS_SEND
}

type ExceptionDisorderOracle struct {
	reports []Report

	newTx           bool
	rootTx          *types.Transaction
	rootCall        MessageCall
	nestedException bool
}

func NewExceptionDisorderOracle() *ExceptionDisorderOracle {
	return &ExceptionDisorderOracle{reports: make([]Report, 0)}
}

func (o *ExceptionDisorderOracle) Type() VulnerabilityType {
	return EXCEPTION_DISORDER
}

func (o *ExceptionDisorderOracle) BeforeTransaction(tx *types.Transaction) {
	// if a new tx starts to execute
	o.rootTx = tx
	o.newTx = true
}

func (o *ExceptionDisorderOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	if o.newTx {
		// if a new tx starts to execute, store the root message call
		o.rootCall = call
		o.newTx = false
	}
}

func (o *ExceptionDisorderOracle) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ExceptionDisorderOracle) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ExceptionDisorderOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	if call != o.rootCall {
		// this is the end of nested message call
		if call.Exception() != nil {
			// if nested message call throws an exception
			o.nestedException = true
		}
	} else {
		// this is the end of root message call
		if o.nestedException && call.Exception() == nil {
			// if exceptions are thrown in the nested call but root call does not throw any exception
			o.reports = append(o.reports, Report{
				Address:     o.rootCall.Callee(),
				PC:          0,
				Transaction: o.rootTx.Hash(),
				VulType:     o.Type(),
				Description: "exception disorder",
			})
		}
	}
}

func (o *ExceptionDisorderOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.rootTx = nil
	o.rootCall = nil
}

func (o *ExceptionDisorderOracle) Report() []Report {
	return o.reports
}
