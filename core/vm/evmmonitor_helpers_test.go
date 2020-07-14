package vm

import "testing"

func TestStack_Push(t *testing.T) {
	stack := &GeneralStack{}
	stack.Push(1, 2)
	top := stack.Top().(int)
	if top != 2 {
		t.Fatal()
	}
	top = stack.Pop().(int)
	if top != 2 {
		t.Fatal()
	}
	top = stack.Pop().(int)
	if top != 1 {
		t.Fatal()
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal()
		}
		if err != StackUnderflowErr {
			t.Fatal()
		}
	}()
	stack.Pop()
}
