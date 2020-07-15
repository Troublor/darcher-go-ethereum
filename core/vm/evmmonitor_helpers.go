package vm

import (
	"errors"
	"sync"
)

var (
	StackUnderflowErr = errors.New("stack under flow")
)

type GeneralStack struct {
	mutex sync.RWMutex
	arr   []interface{}
}

func (s *GeneralStack) prepare() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.arr == nil {
		s.arr = make([]interface{}, 0)
	}
}

func (s *GeneralStack) Push(item ...interface{}) {
	s.prepare()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.arr = append(s.arr, item...)
}

func (s *GeneralStack) Pop() interface{} {
	s.prepare()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.arr) == 0 {
		panic(StackUnderflowErr)
	}
	top := s.arr[len(s.arr)-1]
	s.arr = s.arr[:len(s.arr)-1]
	return top
}

func (s *GeneralStack) Len() int {
	s.prepare()
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.arr)
}

func (s *GeneralStack) Range(rangeFunc func(item interface{}, depth int) (continuous bool)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for i := len(s.arr) - 1; i >= 0; i-- {
		if !rangeFunc(s.arr[i], len(s.arr)-1-i) {
			break
		}
	}
}

func (s *GeneralStack) Top() interface{} {
	s.prepare()
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if len(s.arr) == 0 {
		panic(StackUnderflowErr)
	}
	return s.arr[len(s.arr)-1]
}

func (s *GeneralStack) Back(n int) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if n > len(s.arr)-1 {
		return nil
	}
	return s.arr[len(s.arr)-1-n]
}

func IsSwap(op OpCode) bool {
	return op == SWAP || op >= SWAP1 && op <= SWAP16
}

func IsDup(op OpCode) bool {
	return op == DUP || op >= DUP1 && op <= DUP16
}
