package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"math/big"
	"testing"
)

func TestArithmeticOpCodes(t *testing.T) {
	// assume we have one value in stack already
	instructions := []OpCode{
		PUSH2,
		ADD,
		PUSH10,
		DUP1,
		DUP3,
		SUB,
		CALLER,
		BALANCE,
		SWAP1,
	}
	jt := newYoloV1InstructionSet()
	tracker := NewTaintTracker(ProgramPoint{
		pc:        0,
		op:        PUSH32,
		operation: jt[PUSH32],
	}, nil)
	var (
		mem         = NewMemory()      // bound memory
		stack       = newstack()       // local stack
		returns     = newReturnStack() // local returns stack
		callContext = &callCtx{
			memory:   mem,
			stack:    stack,
			rstack:   returns,
			contract: nil,
		}
	)
	for _, ins := range instructions {
		tracker.Propagate(ins, jt[ins], callContext)
	}
	shouldNotTaintOps := []OpCode{
		DUP2,
		DUP3,
		DUP5,
	}
	shouldTaintOps := []OpCode{
		DUP1,
		DUP4,
	}
	for _, o := range shouldNotTaintOps {
		if tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
	for _, o := range shouldTaintOps {
		if !tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
}

func TestMemoryOps(t *testing.T) {
	instructions := []OpCode{
		PUSH2,
		MSTORE,
		PUSH10,
		DUP1,
		MLOAD,
		PUSH1,
		PUSH2,
		MSTORE8,
		PUSH1,
		MLOAD,
	}
	jt := newYoloV1InstructionSet()
	tracker := NewTaintTracker(ProgramPoint{
		pc:        0,
		op:        PUSH32,
		operation: jt[PUSH32],
	}, nil)
	var (
		mem         = NewMemory()      // bound memory
		stack       = newstack()       // local stack
		returns     = newReturnStack() // local returns stack
		callContext = &callCtx{
			memory:   mem,
			stack:    stack,
			rstack:   returns,
			contract: nil,
		}
	)
	memOffset := 0
	for _, ins := range instructions {
		if ins == MLOAD {
			p, _ := uint256.FromBig(big.NewInt(int64(memOffset)))
			stack.push(p)
		} else if ins == MSTORE || ins == MSTORE8 {
			p, _ := uint256.FromBig(big.NewInt(int64(memOffset)))
			stack.push(p)
			stack.push(p)
		}
		tracker.Propagate(ins, jt[ins], callContext)
		if ins == MLOAD {
			memOffset++
		}
	}
	shouldNotTaintOps := []OpCode{
		DUP3,
		DUP4,
		DUP5,
	}
	shouldTaintOps := []OpCode{
		DUP1,
		DUP2,
	}
	for _, o := range shouldNotTaintOps {
		if tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
	for _, o := range shouldTaintOps {
		if !tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
}

func TestStorageOps(t *testing.T) {
	instructions := []OpCode{
		PUSH2,
		SSTORE,
		PUSH10,
		SLOAD,
	}
	jt := newYoloV1InstructionSet()
	tracker := NewTaintTracker(ProgramPoint{
		pc:        0,
		op:        PUSH32,
		operation: jt[PUSH32],
	}, nil)
	var (
		mem         = NewMemory()      // bound memory
		stack       = newstack()       // local stack
		returns     = newReturnStack() // local returns stack
		callContext = &callCtx{
			memory: mem,
			stack:  stack,
			rstack: returns,
			contract: NewContract(AccountRef(common.HexToAddress("1337")),
				AccountRef(common.HexToAddress("1338")), new(big.Int), 1000),
		}
	)
	storageOffset := 0
	for _, ins := range instructions {
		if ins == SLOAD {
			p, _ := uint256.FromBig(big.NewInt(int64(storageOffset)))
			stack.push(p)
		} else if ins == SSTORE {
			p, _ := uint256.FromBig(big.NewInt(int64(storageOffset)))
			stack.push(p)
			stack.push(p)
		}
		tracker.Propagate(ins, jt[ins], callContext)
		if ins == SLOAD {
			storageOffset++
		}
	}
	shouldNotTaintOps := []OpCode{
		DUP2,
	}
	shouldTaintOps := []OpCode{
		DUP1,
	}
	for _, o := range shouldNotTaintOps {
		if tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
	for _, o := range shouldTaintOps {
		if !tracker.IsOpTainted(o, jt[o], callContext) {
			t.Fatal()
		}
	}
}
