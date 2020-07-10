package common

import (
	"sync/atomic"
	"testing"
	"time"
)

type TestStruct struct {
}

func TestDynamicPriorityQueue(t *testing.T) {
	queue := NewDynamicPriorityQueue(func(candidates []interface{}) (selected interface{}) {
		return candidates[0]
	})
	go func() {
		time.Sleep(200 * time.Millisecond)
		queue.Push(&TestStruct{})
	}()
	queue.Pull()

	if queue.Len() > 0 {
		t.Fatal("queue isn't empty")
	}

	queue.Push(1)
	queue.Push(2)
	queue.Push(3)
	if queue.Pull() != 1 {
		t.Fatal("value wrong")
	}
	queue.Pull()
	queue.Pull()
	var count int32 = 0
	go func() {
		queue.Pull()
		atomic.AddInt32(&count, 1)
	}()
	go func() {
		queue.Pull()
		atomic.AddInt32(&count, 1)
	}()
	queue.Push(1)
	time.Sleep(200 * time.Millisecond)
	if count > 1 {
		t.Fatal("shouldn't pull")
	}
}

func TestDynamicPriorityQueue_Prune(t *testing.T) {
	queue := NewDynamicPriorityQueue(func(candidates []interface{}) (selected interface{}) {
		return candidates[0]
	})
	queue.Push(1)
	queue.Push(2)
	queue.Prune(func(i interface{}) bool {
		return i == 1
	})
	n := queue.Pull().(int)
	if n != 2 {
		t.Fatal("prune failed")
	}
}
