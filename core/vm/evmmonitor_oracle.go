package vm

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
	"math/big"
)

type VulnerabilityType string

type Oracle interface {
	Type() rpc.ContractVulType
	BeforeTransaction(tx *types.Transaction)
	BeforeMessageCall(callStack *GeneralStack, call MessageCall)
	BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx)
	AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx)
	AfterMessageCall(callStack *GeneralStack, call MessageCall)
	AfterTransaction(tx *types.Transaction, receipt *types.Receipt)

	Reports() []*rpc.ContractVulReport
	Clear()
}

type GaslessSendOracle struct {
	currentTx *types.Transaction
	reports   []*rpc.ContractVulReport

	shouldCheck bool
	pc          int

	sendStack *GeneralStack
	pcStack   *GeneralStack
}

func NewGaslessSendOracle() *GaslessSendOracle {
	return &GaslessSendOracle{
		reports:     make([]*rpc.ContractVulReport, 0),
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
		log.Debug("send start")
	}
}

func (o *GaslessSendOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if op == CALL || op == CALLCODE {
		gasLimit := ctx.stack.Back(0)
		value := ctx.stack.Back(2)
		inOffset, inSize := ctx.stack.Back(3), ctx.stack.Back(4)
		args := ctx.memory.GetPtr(int64(inOffset.Uint64()), int64(inSize.Uint64()))
		log.Debug("CALL opcode", "gas", gasLimit, "value", value, "args", len(args))
		tmp, _ := uint256.FromBig(big.NewInt(2300))
		if !value.IsZero() &&
			(gasLimit.Cmp(tmp) == 0 || gasLimit.IsZero()) &&
			len(args) == 0 {
			o.shouldCheck = true
			o.pcStack.Push(pc)
			log.Debug("got send")
		}
	}
}

func (o *GaslessSendOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *GaslessSendOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	if o.sendStack.Len() > 0 && o.sendStack.Top().(MessageCall) == call {
		if call.OutOfGas() && callStack.Len() > 1 {
			prevCall := callStack.Back(1).(MessageCall)
			// gasless send when a send() throws ErrOutOfGas
			o.reports = append(o.reports, &rpc.ContractVulReport{
				Address:     call.Caller().Hex(),
				FuncSig:     prevCall.Function().Sig().Hex(),
				Type:        o.Type(),
				TxHash:      o.currentTx.Hash().Hex(),
				Pc:          o.pcStack.Top().(uint64),
				Description: fmt.Sprintf("gasless send at %d", o.pcStack.Top().(uint64)),
			})
			log.Debug("gasless send")
		}
		o.sendStack.Pop()
		o.pcStack.Pop()
		log.Debug("send finished")
	}
}

func (o *GaslessSendOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	return
}

func (o *GaslessSendOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *GaslessSendOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_GASLESS_SEND
}

func (o *GaslessSendOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}

type ExceptionDisorderOracle struct {
	reports []*rpc.ContractVulReport

	newTx           bool
	rootTx          *types.Transaction
	rootCall        MessageCall
	nestedException bool
}

func NewExceptionDisorderOracle() *ExceptionDisorderOracle {
	return &ExceptionDisorderOracle{reports: make([]*rpc.ContractVulReport, 0)}
}

func (o *ExceptionDisorderOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_EXCEPTION_DISORDER
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

func (o *ExceptionDisorderOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ExceptionDisorderOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
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
			o.reports = append(o.reports, &rpc.ContractVulReport{
				Address:     o.rootCall.Callee().Hex(),
				TxHash:      o.rootTx.Hash().Hex(),
				Type:        o.Type(),
				Description: "exception disorder",
			})
		}
		o.nestedException = false
	}
}

func (o *ExceptionDisorderOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.rootTx = nil
	o.rootCall = nil
}

func (o *ExceptionDisorderOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *ExceptionDisorderOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}

type ReentrancyOracle struct {
	reports []*rpc.ContractVulReport
	rootTx  *types.Transaction
}

func NewReentrancyOracle() *ReentrancyOracle {
	return &ReentrancyOracle{
		reports: make([]*rpc.ContractVulReport, 0),
	}
}

func (o *ReentrancyOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_REENTRANCY
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
			o.reports = append(o.reports, &rpc.ContractVulReport{
				Address:     reenteringCall.Function().Addr().Hex(),
				TxHash:      o.rootTx.Hash().Hex(),
				FuncSig:     reenteringCall.Function().Sig().Hex(),
				Type:        o.Type(),
				Description: fmt.Sprintf("function reentrancy found at contract %s, function %s", reenteringCall.Function().Addr().Hex(), reenteringCall.Function().Sig().Hex()),
			})
		}
	}
}

func (o *ReentrancyOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *ReentrancyOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
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

func (o *ReentrancyOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *ReentrancyOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}

type TimestampDependencyOracle struct {
	reports      []*rpc.ContractVulReport
	currentTx    *types.Transaction
	currentCall  MessageCall
	useTimestamp bool
}

func NewTimestampDependencyOracle() *TimestampDependencyOracle {
	return &TimestampDependencyOracle{reports: make([]*rpc.ContractVulReport, 0)}
}

func (o *TimestampDependencyOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_TIMESTAMP_DEPENDENCY
}

func (o *TimestampDependencyOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *TimestampDependencyOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = call
}

func (o *TimestampDependencyOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if op == TIMESTAMP {
		o.useTimestamp = true
	} else if o.useTimestamp && (op == CALL || op == CALLCODE) {
		// use timestamp before CALL in current call
		value := ctx.stack.Back(2)
		if value.Cmp(uint256.NewInt()) > 0 {
			// have transferred non-zero values
			o.reports = append(o.reports, &rpc.ContractVulReport{
				Address:     o.currentCall.Callee().Hex(),
				FuncSig:     o.currentCall.Function().Sig().Hex(),
				Pc:          pc,
				Opcode:      op.Hex(),
				TxHash:      o.currentTx.Hash().Hex(),
				Type:        o.Type(),
				Description: fmt.Sprintf("timestamp dependency found at contract %s, function %s, CALL opcode at %d", o.currentCall.Callee().Hex(), o.currentCall.Function().Sig().Hex(), pc),
			})
		}
	}
}

func (o *TimestampDependencyOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *TimestampDependencyOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall, _ = callStack.Back(1).(MessageCall)
	o.useTimestamp = false
}

func (o *TimestampDependencyOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.currentTx = nil
}

func (o *TimestampDependencyOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *TimestampDependencyOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}

type BlockNumberDependencyOracle struct {
	reports        []*rpc.ContractVulReport
	currentTx      *types.Transaction
	currentCall    MessageCall
	useBlockNumber bool
}

func NewBlockNumberDependencyOracle() *BlockNumberDependencyOracle {
	return &BlockNumberDependencyOracle{reports: make([]*rpc.ContractVulReport, 0)}
}

func (o *BlockNumberDependencyOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_BLOCKNUMBER_DEPENDENCY
}

func (o *BlockNumberDependencyOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *BlockNumberDependencyOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = call
}

func (o *BlockNumberDependencyOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if op == NUMBER {
		o.useBlockNumber = true
	} else if o.useBlockNumber && (op == CALL || op == CALLCODE) {
		// use timestamp before CALL in current call
		value := ctx.stack.Back(2)
		if value.Cmp(uint256.NewInt()) > 0 {
			// have transferred non-zero values
			o.reports = append(o.reports, &rpc.ContractVulReport{
				Address:     o.currentCall.Callee().Hex(),
				FuncSig:     o.currentCall.Function().Sig().Hex(),
				Opcode:      op.Hex(),
				Pc:          pc,
				TxHash:      o.currentTx.Hash().Hex(),
				Type:        o.Type(),
				Description: fmt.Sprintf("blockNumber dependency found at contract %s, function %s, CALL opcode at %d", o.currentCall.Callee().Hex(), o.currentCall.Function().Sig().Hex(), pc),
			})
		}
	}
}

func (o *BlockNumberDependencyOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *BlockNumberDependencyOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall, _ = callStack.Back(1).(MessageCall)
	o.useBlockNumber = false
}

func (o *BlockNumberDependencyOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.currentTx = nil
}

func (o *BlockNumberDependencyOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *BlockNumberDependencyOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}

type DangerousDelegateCallOracle struct {
	reports          []*rpc.ContractVulReport
	currentTx        *types.Transaction
	currentCall      MessageCall
	callDataTrackers []*TaintTracker
}

func NewDangerousDelegateCallOracle() *DangerousDelegateCallOracle {
	return &DangerousDelegateCallOracle{
		reports:          make([]*rpc.ContractVulReport, 0),
		callDataTrackers: make([]*TaintTracker, 0),
	}
}

func (o *DangerousDelegateCallOracle) Type() rpc.ContractVulType {
	return rpc.ContractVulType_DANGEROUS_DELEGATECALL
}

func (o *DangerousDelegateCallOracle) BeforeTransaction(tx *types.Transaction) {
	o.currentTx = tx
}

func (o *DangerousDelegateCallOracle) BeforeMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall = call
}

func (o *DangerousDelegateCallOracle) BeforeOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	if op == CALLDATALOAD || op == CALLDATACOPY {
		// tainted source is call data
		o.callDataTrackers = append(o.callDataTrackers, NewTaintTracker(ProgramPoint{
			pc:        pc,
			op:        op,
			operation: operation,
		}, ctx))
	} else if op == DELEGATECALL {
		// tainted sink is delegate call's input
		inOffset, inSize := ctx.stack.Back(2).Uint64(), ctx.stack.Back(3).Uint64()
		for _, tracker := range o.callDataTrackers {
			// stack depth 1 is the address of delegate call code
			if tracker.IsStackTainted(1) || tracker.IsMemoryTainted(inOffset, inSize) {
				o.reports = append(o.reports, &rpc.ContractVulReport{
					Address:     o.currentCall.Callee().Hex(),
					FuncSig:     o.currentCall.Function().Sig().Hex(),
					Opcode:      op.Hex(),
					Pc:          pc,
					TxHash:      o.currentTx.Hash().Hex(),
					Type:        o.Type(),
					Description: fmt.Sprintf("dangerous delegate call at contract %s, function %s, DELEGATECALL at %d", o.currentCall.Callee().Hex(), o.currentCall.Function().Sig().Hex(), pc),
				})
			}
		}
	} else {
		// taint propagate
		for _, tracker := range o.callDataTrackers {
			tracker.Propagate(op, operation, ctx)
		}
	}
}

func (o *DangerousDelegateCallOracle) AfterOperation(op OpCode, operation *operation, pc uint64, ctx *callCtx) {
	return
}

func (o *DangerousDelegateCallOracle) AfterMessageCall(callStack *GeneralStack, call MessageCall) {
	o.currentCall, _ = callStack.Back(1).(MessageCall)
}

func (o *DangerousDelegateCallOracle) AfterTransaction(tx *types.Transaction, receipt *types.Receipt) {
	o.currentTx = nil
}

func (o *DangerousDelegateCallOracle) Reports() []*rpc.ContractVulReport {
	return o.reports
}

func (o *DangerousDelegateCallOracle) Clear() {
	o.reports = make([]*rpc.ContractVulReport, 0)
}
