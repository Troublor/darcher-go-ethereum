package vm

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type CallType string

const (
	TYPE_CALL         = "call"
	TYPE_CALLCODE     = "callcode"
	TYPE_STATICCALL   = "staticcall"
	TYPE_DELEGATECALL = "delegatecall"
	TYPE_CREATE       = "create"
)

type FuncSig [4]byte

var (
	FallbackSig    = [4]byte{0, 0, 0, 0}
	ConstructorSig = [4]byte{0, 0, 0, 1}
)

func (s FuncSig) Hex() string {
	return "0x" + hex.EncodeToString(s[:])
}

func (s FuncSig) Bytes() []byte {
	return s[:]
}

func (op OpCode) Hex() string {
	return "0x" + hex.EncodeToString([]byte{byte(op)})
}

type Function struct {
	addr common.Address
	sig  FuncSig
}

func (f Function) Sig() FuncSig {
	return f.sig
}

func (f Function) Addr() common.Address {
	return f.addr
}

func (f Function) Equal(another interface{}) bool {
	fn, ok := another.(Function)
	if !ok {
		return false
	}
	return f.addr == fn.addr && f.sig == fn.sig
}

type MessageCall interface {
	Type() CallType
	Caller() common.Address
	Callee() common.Address
	Input() []byte
	Function() Function
	GasLimit() uint64
	Value() *big.Int
	OutOfGas() bool
	Exception() error

	// call result
	setCallReturn(ret []byte, err error)
}

type Call struct {
	caller    common.Address
	callee    common.Address
	value     *big.Int
	input     []byte
	gasLimit  uint64
	exception error

	// result fields
	outOfGas bool
}

func (c *Call) Type() CallType {
	return TYPE_CALL
}

func (c *Call) Caller() common.Address {
	return c.caller
}

func (c *Call) Callee() common.Address {
	return c.callee
}

func (c *Call) Input() []byte {
	return c.input
}

func (c *Call) Function() Function {
	var sig FuncSig
	if len(c.Input()) < 4 {
		// special fallback function
		sig = FallbackSig
	} else {
		copy(sig[:], c.Input()[:4])
	}
	return Function{
		addr: c.callee,
		sig:  sig,
	}
}

func (c *Call) GasLimit() uint64 {
	return c.gasLimit
}

func (c *Call) Value() *big.Int {
	return c.value
}

func (c *Call) OutOfGas() bool {
	return c.outOfGas
}

func (c *Call) Exception() error {
	return c.exception
}

func (c *Call) setCallReturn(ret []byte, err error) {
	c.exception = err
	if err == ErrOutOfGas {
		c.outOfGas = true
	}
}

type CallCode struct {
	Call
}

func (c *CallCode) Type() CallType {
	return TYPE_CALLCODE
}

type StaticCall struct {
	Call
}

func (c *StaticCall) Type() CallType {
	return TYPE_STATICCALL
}

type DelegateCall struct {
	Call
}

func (c *DelegateCall) Type() CallType {
	return TYPE_DELEGATECALL
}

type Create struct {
	Call
}

func (c *Create) Type() CallType {
	return TYPE_CREATE
}

func (c *Create) Function() Function {
	// special constructor function
	return Function{
		addr: c.Callee(),
		sig:  ConstructorSig,
	}
}
