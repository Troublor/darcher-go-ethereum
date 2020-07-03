package worker

import (
	"reflect"
	"sync"
	"testing"
)

func TestTxScheduler_WaitForTurn(t *testing.T) {
	var mutex sync.Mutex
	result := make([]string, 0)
	scheduler := newTxScheduler()
	var wg1, wg2, wg3 sync.WaitGroup
	wg1.Add(1)
	wg2.Add(1)
	wg3.Add(1)
	go func() {
		hash := "1"
		scheduler.WaitForTurn(hash)
		mutex.Lock()
		result = append(result, hash)
		mutex.Unlock()
		wg1.Done()
	}()
	go func() {
		hash := "3"
		scheduler.WaitForTurn(hash)
		mutex.Lock()
		result = append(result, hash)
		mutex.Unlock()
		wg3.Done()
	}()
	go func() {
		hash := "2"
		scheduler.WaitForTurn(hash)
		mutex.Lock()
		result = append(result, hash)
		mutex.Unlock()
		wg2.Done()
	}()
	scheduler.ScheduleTx("3")
	wg3.Wait()
	if len(result) != 1 {
		t.Fatal("result length is not 1")
	}
	if !reflect.DeepEqual(result, []string{"3"}) {
		t.Fatal("result is incorrect")
	}

	scheduler.ScheduleTx("2")
	wg2.Wait()
	if len(result) != 2 {
		t.Fatal("result length is not 2")
	}
	if !reflect.DeepEqual(result, []string{"3", "2"}) {
		t.Fatal("result is incorrect")
	}

	scheduler.ScheduleTx("1")
	wg1.Wait()
	if len(result) != 3 {
		t.Fatal("result length is not 3")
	}
	if !reflect.DeepEqual(result, []string{"3", "2", "1"}) {
		t.Fatal("result is incorrect")
	}
}
