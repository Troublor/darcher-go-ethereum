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
	GASLESS_SEND           = "gasless_send"
	EXCEPTION_DISORDER     = "exception_disorder"
	REENTRANCY             = "reentrancy"
	TIMESTAMP_DEPENDENCY   = "timestamp_dependency"
	BLOCKNUMBER_DEPENDENCY = "blockNumber_dependency"
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
	if op == CALL || op == CALLCODE {
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

type ReentrancyOracle struct {
	reports []Report
	rootTx  *types.Transaction
}

func NewReentrancyOracle() *ReentrancyOracle {
	return &ReentrancyOracle{
		reports: make([]Report, 0),
	}
}

func (o *ReentrancyOracle) Type() VulnerabilityType {
	return REENTRANCY
}

func (o *ReentrancyOracle) BeforeTransaction(tx *types.Transaction) {
	o.rootTx = tx
}

func (o *ReentrancyOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	if call.Type() == TYPE_CREATE {
		// do not track create message call
		return
	}
	var firstOutCall MessageCall
	prevCallDepth := -1
	callStack.Range(func(item interface{}, depth int) (continuous bool) {
		prevCall := item.(MessageCall)
		if prevCall.Function().Equal(call.Function()) {
			// this call is the reentrancy of prevCall
			prevCallDepth = depth
			return false
		}
		return true
	})
	if prevCallDepth == 0 {
		firstOutCall = call
	} else if prevCallDepth > 0 {
		firstOutCall = callStack.Back(prevCallDepth - 1).(MessageCall)
	} else {
		firstOutCall = nil
	}

	if firstOutCall != nil { // the function is reentered
		reenteringCall := callStack.Back(prevCallDepth).(MessageCall)
		if firstOutCall.Value().Cmp(big.NewInt(0)) > 0 && // send non-zero value
			firstOutCall.GasLimit() > 2300 { // have enough gas to do expensive operation
			o.reports = append(o.reports, Report{
				Address:     reenteringCall.Function().Addr(),
				PC:          0,
				Transaction: o.rootTx.Hash(),
				VulType:     REENTRANCY,
				Description: fmt.Sprintf("function reentrancy found at contract %s, function %s", reenteringCall.Function().Addr().Hex(), reenteringCall.Function().Sig()),
			})
		}
	}
}

func (o *ReentrancyOracle) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ReentrancyOracle) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ReentrancyOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	if call.Type() == TYPE_CREATE {
		// do not track create message call
		return
	}
}

func (o *ReentrancyOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.rootTx = nil
}

func (o *ReentrancyOracle) Report() []Report {
	return o.reports
}

type TimestampDependencyOracle struct {
	reports      []Report
	currentTx    *types.Transaction
	currentCall  MessageCall
	useTimestamp bool
}

func NewTimestampDependencyOracle() *TimestampDependencyOracle {
	return &TimestampDependencyOracle{reports: make([]Report, 0)}
}

func (o *TimestampDependencyOracle) Type() VulnerabilityType {
	return TIMESTAMP_DEPENDENCY
}

func (o *TimestampDependencyOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *TimestampDependencyOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = call
}

func (o *TimestampDependencyOracle) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	if op == TIMESTAMP {
		o.useTimestamp = true
	} else if o.useTimestamp && (op == CALL || op == CALLCODE) {
		// use timestamp before CALL in current call
		value := ctx.stack.Back(2)
		if value.Cmp(uint256.NewInt()) > 0 {
			// have transferred non-zero values
			o.reports = append(o.reports, Report{
				Address:     o.currentCall.Callee(),
				PC:          pc,
				Transaction: o.currentTx.Hash(),
				VulType:     TIMESTAMP_DEPENDENCY,
				Description: fmt.Sprintf("timestamp dependency found at contract %s, function %s, CALL opcode at %d", o.currentCall.Callee().Hex(), o.currentCall.Function().Sig(), pc),
			})
		}
	}
}

func (o *TimestampDependencyOracle) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *TimestampDependencyOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = nil
	o.useTimestamp = false
}

func (o *TimestampDependencyOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.currentTx = nil
}

func (o *TimestampDependencyOracle) Report() []Report {
	return o.reports
}

type BlockNumberDependencyOracle struct {
	reports        []Report
	currentTx      *types.Transaction
	currentCall    MessageCall
	useBlockNumber bool
}

func NewBlockNumberDependencyOracle() *BlockNumberDependencyOracle {
	return &BlockNumberDependencyOracle{reports: make([]Report, 0)}
}

func (o *BlockNumberDependencyOracle) Type() VulnerabilityType {
	return BLOCKNUMBER_DEPENDENCY
}

func (o *BlockNumberDependencyOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *BlockNumberDependencyOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = call
}

func (o *BlockNumberDependencyOracle) BeforeOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	if op == NUMBER {
		o.useBlockNumber = true
	} else if o.useBlockNumber && (op == CALL || op == CALLCODE) {
		// use timestamp before CALL in current call
		value := ctx.stack.Back(2)
		if value.Cmp(uint256.NewInt()) > 0 {
			// have transferred non-zero values
			o.reports = append(o.reports, Report{
				Address:     o.currentCall.Callee(),
				PC:          pc,
				Transaction: o.currentTx.Hash(),
				VulType:     BLOCKNUMBER_DEPENDENCY,
				Description: fmt.Sprintf("blockNumber dependency found at contract %s, function %s, CALL opcode at %d", o.currentCall.Callee().Hex(), o.currentCall.Function().Sig(), pc),
			})
		}
	}
}

func (o *BlockNumberDependencyOracle) AfterOperation(op OpCode, operation operation, pc uint64, ctx *callCtx) {
	return
}

func (o *BlockNumberDependencyOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = nil
	o.useBlockNumber = false
}

func (o *BlockNumberDependencyOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.currentTx = nil
}

func (o *BlockNumberDependencyOracle) Report() []Report {
	return o.reports
}
