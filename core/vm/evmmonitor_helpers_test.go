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

func TestGeneralStack_Range(t *testing.T) {
	defer func() {
		err := recover()
		if err != nil {
			t.Fatal(err)
		}
	}()
	stack := &GeneralStack{}
	count := 0
	stack.Range(func(item interface{}, depth int) (continuous bool) {
		count++
		return true
	})
	if count > 0 {
		t.Fatal()
	}
	stack.Push(1)
	stack.Range(func(item interface{}, depth int) (continuous bool) {
		if item.(int) != 1 {
			t.Fatal()
		}
		if depth != 0 {
			t.Fatal()
		}
		return true
	})
}
