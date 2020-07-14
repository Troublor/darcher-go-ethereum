package vm

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type CallType string

const (
	TYPE_CALL         = "call"
	TYPE_CALLCODE     = "callcode"
	TYPE_STATICCALL   = "staticcall"
	TYPE_DELEGATECALL = "delegatecall"
)

type MessageCall interface {
	Type() CallType
	Caller() common.Address
	Callee() common.Address
	Input() []byte
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
