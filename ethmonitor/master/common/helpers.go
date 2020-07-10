package common

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
)

func PrettifyHash(hash string) string {
	startIndex := 0
	if strings.Index(hash, "0x") >= 0 {
		startIndex = 2
	}
	return fmt.Sprintf("%s…%s", hash[startIndex:startIndex+6], hash[len(hash)-6:])
}

type DynamicPriorityQueue struct {
	selectFunc func(candidates []interface{}) (selected interface{})

	nonEmptyCh chan interface{}
	rwMutex    sync.RWMutex
	queue      []interface{}
}

func NewDynamicPriorityQueue(selectFunc func(candidates []interface{}) (selected interface{})) *DynamicPriorityQueue {
	ch := make(chan interface{})
	return &DynamicPriorityQueue{
		selectFunc: selectFunc,
		nonEmptyCh: ch,
	}
}

func (q *DynamicPriorityQueue) Pull() interface{} {
	q.rwMutex.Lock()
	if len(q.queue) <= 0 {
		q.rwMutex.Unlock()
		<-q.nonEmptyCh
		q.rwMutex.Lock()
	}
	selected := q.selectFunc(q.queue)
	newQueue := make([]interface{}, len(q.queue)-1)
	index := 0
	for _, item := range q.queue {
		if item == selected {
			continue
		}
		newQueue[index] = item
		index++
	}
	q.queue = newQueue
	if len(q.queue) <= 0 {
		q.nonEmptyCh = make(chan interface{})
	}
	q.rwMutex.Unlock()
	return selected
}

func (q *DynamicPriorityQueue) Push(item interface{}) {
	q.rwMutex.Lock()
	defer q.rwMutex.Unlock()
	if len(q.queue) <= 0 {
		q.queue = append(q.queue, item)
		select {
		case <-q.nonEmptyCh:
		default:
			close(q.nonEmptyCh)
		}
	} else {
		q.queue = append(q.queue, item)
	}
}

func (q *DynamicPriorityQueue) Len() int {
	q.rwMutex.RLock()
	defer q.rwMutex.RUnlock()
	return len(q.queue)
}

/**
Iterate all items in the queue and prune the item if pruneFunc(item) == true
*/
func (q *DynamicPriorityQueue) Prune(pruneFunc func(item interface{}) bool) {
	newQ := make([]interface{}, 0)
	q.rwMutex.Lock()
	q.rwMutex.Unlock()
	for _, item := range q.queue {
		if !pruneFunc(item) {
			newQ = append(newQ, item)
		}
	}
	q.queue = newQ
}

func ExecuteSerially(fns []func() error) error {
	for _, f := range fns {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

var uuidMutex sync.Mutex
var uuid = big.NewInt(0)

func GetUUID() string {
	uuidMutex.Lock()
	defer uuidMutex.Unlock()
	id := uuid.String()
	uuid = uuid.Add(uuid, big.NewInt(1))
	return id
}
