package vm

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type ProgramPoint struct {
	pc        uint64
	op        OpCode
	operation *operation
}

type TaintedPosition interface {
	// returns the propagated positions after op, if current position is not consumed, it should also be returned
	Propagate(op OpCode, operation *operation, ctx *callCtx) []TaintedPosition
	// whether this tainted position is used in op
	IsUsed(op OpCode, operation *operation, ctx *callCtx) bool
	// equality of two position, to remove duplicate
	Equal(another TaintedPosition) bool
	// Invalid, the position has been consumed
	Invalid() bool
}

type StackPosition struct {
	invalid bool
	// the position in stack is represented as the depth from the top (start from 0)
	depth int
}

func (p *StackPosition) Propagate(op OpCode, operation *operation, ctx *callCtx) []TaintedPosition {
	if p.Invalid() {
		return make([]TaintedPosition, 0)
	}
	numPop := operation.minStack
	numPush := int(params.StackLimit + uint64(numPop) - uint64(operation.maxStack))
	newPositions := make([]TaintedPosition, 0)
	// some special cases
	if IsSwap(op) {
		swapDepth := numPop - 1
		if p.depth == 0 {
			p.depth = swapDepth
		} else if p.depth == swapDepth {
			p.depth = 0
		}
		newPositions = append(newPositions, p)
	} else if IsDup(op) {
		dupDepth := numPop - 1
		if p.depth == dupDepth {
			newPositions = append(newPositions, &StackPosition{depth: 0})
		}
		p.depth++
		newPositions = append(newPositions, p)
	} else if op == MSTORE && p.IsUsed(op, operation, ctx) {
		p.invalid = true
		offset := ctx.stack.Back(0).Uint64()
		newPositions = append(newPositions, &MemoryPosition{
			offset: offset,
			size:   32,
		})
	} else if op == MSTORE8 && p.IsUsed(op, operation, ctx) {
		p.invalid = true
		offset := ctx.stack.Back(0).Uint64()
		newPositions = append(newPositions, &MemoryPosition{
			offset: offset,
			size:   8,
		})
	} else if op == SSTORE && p.IsUsed(op, operation, ctx) {
		p.invalid = true
		key := ctx.stack.Back(0).Bytes32()
		newPositions = append(newPositions, &StoragePosition{
			account: ctx.contract.Address(),
			key:     key,
		})
	} else if p.IsUsed(op, operation, ctx) {
		// mark this position as invalid, cuz it has been consumed
		p.invalid = true
		// if this tainted position is used, then the new pushed values are also tainted
		for i := 0; i < numPush; i++ {
			newPositions = append(newPositions, &StackPosition{depth: i})
		}
	} else {
		// update the position after op
		p.depth = p.depth - numPop + numPush
		newPositions = append(newPositions, p)
	}
	return newPositions
}

func (p *StackPosition) IsUsed(op OpCode, operation *operation, ctx *callCtx) bool {
	if p.Invalid() {
		return false
	}
	// special cases
	if IsSwap(op) {
		return p.depth == operation.minStack-1 || p.depth == 0
	} else if IsDup(op) {
		return p.depth == operation.minStack-1
	}
	numPop := operation.minStack
	return numPop > p.depth
}

func (p *StackPosition) Equal(another TaintedPosition) bool {
	if p.Invalid() || another.Invalid() {
		return false
	}
	sp, ok := another.(*StackPosition)
	if !ok {
		return false
	}
	return p.depth == sp.depth
}

func (p *StackPosition) Invalid() bool {
	return p.invalid
}

type MemoryPosition struct {
	invalid bool
	offset  uint64
	size    uint64
}

func (p *MemoryPosition) Propagate(op OpCode, operation *operation, ctx *callCtx) []TaintedPosition {
	if p.Invalid() {
		return make([]TaintedPosition, 0)
	}
	newPositions := make([]TaintedPosition, 0)
	if op == MLOAD && p.IsUsed(op, operation, ctx) {
		// if mload touches the position, the position is still tainted and the data pushed on stack is tainted
		newPositions = append(newPositions, p)
		newPositions = append(newPositions, &StackPosition{depth: 0})
	} else if op == MSTORE || op == MSTORE8 {
		// if the mstore target is the tainted position, then the position will be invalid (override)
		start := ctx.stack.Back(0)
		var length uint64
		switch op {
		case MSTORE8:
			length = 8
		case MSTORE:
			length = 32
		}
		if start.Uint64()+length < p.offset {
			// no overlap
			newPositions = append(newPositions, p)
		} else if start.Uint64() >= p.offset+p.size {
			// no overlap
			newPositions = append(newPositions, p)
		} else if start.Uint64() < p.offset && start.Uint64()+length >= p.offset+p.size {
			// completely override
			p.invalid = true
		} else if start.Uint64() < p.offset {
			// left overlap
			p.offset = start.Uint64() + length
			newPositions = append(newPositions, p)
		} else if start.Uint64() >= p.offset && start.Uint64()+length < p.offset+p.size {
			// middle overlap
			newPositions = append(newPositions, &MemoryPosition{
				offset: start.Uint64() + length,
				size:   p.offset + p.size - (start.Uint64() + length),
			})
			p.size = start.Uint64() - p.offset
			newPositions = append(newPositions, p)
		} else if start.Uint64() >= p.offset {
			// right overlap
			p.size = start.Uint64() - p.offset
			newPositions = append(newPositions, p)
		} else {
			// should not happen
			panic(errors.New(fmt.Sprintf("memory overlap unusual case, offset=%d, size=%d, start=%d, len=%d", p.offset, p.size, start, length)))
		}
	} else {
		newPositions = append(newPositions, p)
	}
	return newPositions
}

func (p *MemoryPosition) IsUsed(op OpCode, operation *operation, ctx *callCtx) bool {
	if p.Invalid() {
		return false
	}
	if op == MLOAD {
		start := ctx.stack.Back(0)
		if start.Uint64()+32 < uint64(p.offset) {
			// no overlap
			return false
		} else if start.Uint64() >= uint64(p.offset+p.size) {
			// no overlap
			return false
		} else {
			return true
		}
	}
	return false
}

func (p *MemoryPosition) Equal(another TaintedPosition) bool {
	if p.Invalid() || another.Invalid() {
		return false
	}
	mp, ok := another.(*MemoryPosition)
	if !ok {
		return false
	}
	return p.offset == mp.offset && p.size == mp.size
}

func (p *MemoryPosition) Invalid() bool {
	return p.invalid
}

type StoragePosition struct {
	invalid bool
	account common.Address
	key     common.Hash
}

func (p *StoragePosition) Propagate(op OpCode, operation *operation, ctx *callCtx) []TaintedPosition {
	if p.Invalid() {
		return make([]TaintedPosition, 0)
	}
	newPositions := make([]TaintedPosition, 0)
	if p.IsUsed(op, operation, ctx) {
		newPositions = append(newPositions, &StackPosition{depth: 0})
		newPositions = append(newPositions, p)
	} else if op == SSTORE {
		key := ctx.stack.Back(0).Bytes32()
		if p.account != ctx.contract.Address() || p.key != key {
			// the storage is override
			newPositions = append(newPositions, p)
		} else {
			p.invalid = true
		}
	} else {
		newPositions = append(newPositions, p)
	}
	return newPositions
}

func (p *StoragePosition) IsUsed(op OpCode, operation *operation, ctx *callCtx) bool {
	if p.Invalid() {
		return false
	}
	if op == SLOAD {
		key := ctx.stack.Back(0).Bytes32()
		if p.account == ctx.contract.Address() && p.key == key {
			return true
		}
	}
	return false
}

func (p *StoragePosition) Equal(another TaintedPosition) bool {
	if p.Invalid() {
		return false
	}
	sp, ok := another.(*StoragePosition)
	if !ok {
		return false
	}
	return p.account == sp.account && p.key == sp.key
}

func (p *StoragePosition) Invalid() bool {
	return p.invalid
}

type TaintTracker struct {
	taintSource      ProgramPoint
	taintedPositions []TaintedPosition
}

func NewTaintTracker(source ProgramPoint, ctx *callCtx) *TaintTracker {
	positions := make([]TaintedPosition, 0)
	// special cases
	if source.op == CALLDATACOPY {
		memStart, length := ctx.stack.Back(0), ctx.stack.Back(2)
		positions = append(positions, &MemoryPosition{
			offset: memStart.Uint64(),
			size:   length.Uint64(),
		})
	} else if source.op == MSTORE {
		start := ctx.stack.Back(0)
		positions = append(positions, &MemoryPosition{
			offset: start.Uint64(),
			size:   32,
		})

	} else if source.op == MSTORE8 {
		start := ctx.stack.Back(0)
		positions = append(positions, &MemoryPosition{
			offset: start.Uint64(),
			size:   8,
		})
	} else if source.op == SSTORE {
		start := ctx.stack.Back(0)
		positions = append(positions, &StoragePosition{
			account: ctx.contract.Address(),
			key:     start.Bytes32(),
		})
	} else {
		numPop := source.operation.minStack
		numPush := int(params.StackLimit + uint64(numPop) - uint64(source.operation.maxStack))
		for i := 0; i < numPush; i++ {
			positions = append(positions, &StackPosition{depth: i})
		}
	}
	return &TaintTracker{
		taintSource:      source,
		taintedPositions: positions,
	}
}

func (t *TaintTracker) Propagate(op OpCode, operation *operation, ctx *callCtx) {
	old := t.taintedPositions
	t.taintedPositions = make([]TaintedPosition, 0)
	for _, pos := range old {
		newPositions := pos.Propagate(op, operation, ctx)
		for _, nP := range newPositions {
			t.addTaintedPosition(nP)
		}
	}
}

func (t *TaintTracker) addTaintedPosition(position TaintedPosition) {
	for _, p := range t.taintedPositions {
		if p.Equal(position) {
			return
		}
	}
	t.taintedPositions = append(t.taintedPositions, position)
}

func (t *TaintTracker) IsOpTainted(op OpCode, operation *operation, ctx *callCtx) bool {
	for _, p := range t.taintedPositions {
		if p.IsUsed(op, operation, ctx) {
			return true
		}
	}
	return false
}

func (t *TaintTracker) IsStackTainted(depths ...int) bool {
	for _, pos := range t.taintedPositions {
		if sp, ok := pos.(*StackPosition); ok {
			for _, d := range depths {
				if sp.depth == d {
					return true
				}
			}
		}
	}
	return false
}

func (t *TaintTracker) IsMemoryTainted(offset uint64, size uint64) bool {
	for _, pos := range t.taintedPositions {
		if mp, ok := pos.(*MemoryPosition); ok {
			if !(offset+size <= mp.offset || offset >= mp.offset+mp.offset) {
				return true
			}
		}
	}
	return false
}

func (t *TaintTracker) IsStorageTainted(account common.Address, key common.Hash) bool {
	for _, pos := range t.taintedPositions {
		if sp, ok := pos.(*StoragePosition); ok {
			if account == sp.account && sp.key == key {
				return true
			}
		}
	}
	return false
}
