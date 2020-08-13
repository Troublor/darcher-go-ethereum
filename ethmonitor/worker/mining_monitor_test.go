package worker

import (
	"github.com/ethereum/go-ethereum/event"
	"testing"
)

func TestFeedTask(t *testing.T) {
	var feed event.Feed
	ch := make(chan Task, 0)

	// subscribe and immediate unsubscribe to set the Type of Feed
	sub := feed.Subscribe(ch)
	sub.Unsubscribe()

	defer func() {
		err := recover()
		if err != nil {
			t.Fatal(err)
		}
	}()

	feed.Send(&BudgetTask{})
	feed.Send(&TxMonitorTask{})
}
