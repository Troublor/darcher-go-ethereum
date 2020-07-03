package master

import (
	"bufio"
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/master/rpc"
	"github.com/ethereum/go-ethereum/log"
	"os"
	"strconv"
	"strings"
)

type TxController interface {
	SelectTxToTraverse(txs []*Transaction) (tx *Transaction)
	// this method is called when traverser for tx is created, returns a bool indicating whether schedule the tx (return hash to dApp)
	TxReceivedHook(txHash string)
	// this method is called when tx state is about to change
	OnStateChange(txHash string, fromState common.LifecycleState, toState common.LifecycleState)
	// this method is called between each tx state transition
	PivotReachedHook(txHash string, currentState common.LifecycleState) (nextState common.LifecycleState, suspend bool)
	// this method is called when traverser finishes traverse
	TxFinishedHook(txHash string)
	// this method is called when traverser starts traverse
	TxResumeHook(txHash string)
}

/**
Trivial Controller does nothing and let traverser go smoothly
*/
type TrivialController struct{}

func (t *TrivialController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	return txs[0]
}

func (t *TrivialController) TxReceivedHook(txHash string) {
	return
}

func (t *TrivialController) OnStateChange(txHash string, fromState common.LifecycleState, toState common.LifecycleState) {
	return
}

func (t *TrivialController) PivotReachedHook(txHash string, currentState common.LifecycleState) (nextState common.LifecycleState, suspend bool) {
	return common.CONFIRMED, false
}

func (t *TrivialController) TxFinishedHook(txHash string) {
	return
}

func (t *TrivialController) TxResumeHook(txHash string) {
	return
}

/**
let console to controller tx lifecycle traverse
*/
type ConsoleController struct {
}

func NewConsoleController() *ConsoleController {
	return &ConsoleController{}
}

func (c *ConsoleController) selectFromOptions(title string, inputTxt string, options []string) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(title)
	for i, opt := range options {
		fmt.Printf("%d: %s\n", i, opt)
	}
	var i int
	for {
		fmt.Print(inputTxt)
		index, err := reader.ReadString('\n')
		if err != nil {
			log.Error("Read from Stdin failed", "err", err)
			continue
		}
		index = strings.TrimSpace(index)
		i, err = strconv.Atoi(index)
		if err != nil {
			log.Error("Parse input failed", "err", err)
			continue
		}
		if i < 0 || i >= len(options) {
			log.Error("Input out of boundary", "err", err)
			continue
		}
		break
	}
	return i
}

func (c *ConsoleController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {

	fmt.Println("================= Console Controller =================")
	options := make([]string, len(txs))

	for i, tx := range txs {
		options[i] = tx.Hash()
	}
	index := c.selectFromOptions("Available transactions: ", "Select a tx to traverse: ", options)
	fmt.Println("================= Console Controller =================")
	return txs[index]
}

func (c *ConsoleController) TxReceivedHook(txHash string) {
	log.Info("Tx received", "tx", txHash)
}

func (c *ConsoleController) PivotReachedHook(txHash string, currentState common.LifecycleState) (nextState common.LifecycleState, suspend bool) {
	fmt.Println("================= Console Controller =================")
	if currentState == common.CONFIRMED || currentState == common.DROPPED {
		return currentState, false
	}
	options := make([]string, 0)
	if currentState != common.PENDING {
		options = append(options, string(common.PENDING))
	}
	if currentState != common.EXECUTED {
		options = append(options, string(common.EXECUTED))
	}
	options = append(options, string(common.CONFIRMED))
	options = append(options, string(common.DROPPED))
	options = append(options, "suspend")

	index := c.selectFromOptions(fmt.Sprintf("Traverse pivot reached, current state: %s", currentState), "Select next state: ", options)
	if index == len(options)-1 {
		// select to suspend
		return "", true
	}
	fmt.Println("================= Console Controller =================")
	return common.LifecycleState(options[index]), false
}

func (c *ConsoleController) TxFinishedHook(txHash string) {
	log.Info("Tx finished", "tx", txHash)
}

func (c *ConsoleController) TxResumeHook(txHash string) {
	log.Debug("Resume traverse tx", "tx", common.PrettifyHash(txHash))
}

func (c *ConsoleController) OnStateChange(txHash string, currentState common.LifecycleState, nextState common.LifecycleState) {
	log.Debug("tx state change", "tx", common.PrettifyHash(txHash), "from", currentState, "to", nextState)
	return
}

type DArcherController struct {
	dArcher  *rpc.DArcher
	txStates map[string]common.LifecycleState
}

func NewDArcherController() *DArcherController {
	return &DArcherController{
		dArcher:  rpc.NewDArcher(common.DArcherIP, common.DArcherPort),
		txStates: make(map[string]common.LifecycleState),
	}
}

func (d *DArcherController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	txHashes := make([]string, len(txs))
	m := make(map[string]*Transaction)
	for i, tx := range txs {
		txHashes[i] = tx.Hash()
		m[txHashes[i]] = tx
	}
	hash := d.dArcher.SelectTxToTraverse(txHashes)
	return m[hash]

}

func (d *DArcherController) TxReceivedHook(txHash string) {
	d.txStates[txHash] = common.CREATED
	d.dArcher.NotifyTxReceived(txHash)
}

func (d *DArcherController) PivotReachedHook(txHash string, currentState common.LifecycleState) (nextState common.LifecycleState, suspend bool) {
	state := d.txStates[txHash]
	switch state {
	case common.CREATED:
		nextState = common.PENDING
		d.dArcher.NotifyPivotReached(txHash, d.txStates[txHash], nextState)
		d.txStates[txHash] = nextState
		return nextState, false
	case common.PENDING:
		nextState = common.EXECUTED
		d.dArcher.NotifyPivotReached(txHash, d.txStates[txHash], nextState)
		d.txStates[txHash] = nextState
		return nextState, false
	case common.EXECUTED:
		nextState = common.PENDING
		d.dArcher.NotifyPivotReached(txHash, d.txStates[txHash], common.REVERTED)
		d.txStates[txHash] = common.REVERTED
		return nextState, false
	case common.REVERTED:
		nextState = common.EXECUTED
		d.dArcher.NotifyPivotReached(txHash, d.txStates[txHash], common.REEXECUTED)
		d.txStates[txHash] = common.REEXECUTED
		return nextState, false
	case common.REEXECUTED:
		nextState = common.CONFIRMED
		d.dArcher.NotifyPivotReached(txHash, d.txStates[txHash], nextState)
		d.txStates[txHash] = nextState
		return nextState, false
	case common.CONFIRMED:
		return common.CONFIRMED, false
	case common.DROPPED:
		return common.DROPPED, false
	}
	return state, false
}

func (d *DArcherController) TxFinishedHook(txHash string) {
	d.dArcher.NotifyTxFinished(txHash)
}

func (d *DArcherController) TxResumeHook(txHash string) {
	d.dArcher.NotifyTxStart(txHash)
}

func (d *DArcherController) OnStateChange(txHash string, currentState common.LifecycleState, nextState common.LifecycleState) {
	d.dArcher.NotifyTxStateChange(txHash, currentState, nextState)
}

type RobustnessTestController struct {
	counter int
}

func NewRobustnessTestController() *RobustnessTestController {
	return &RobustnessTestController{
		counter: 0,
	}
}

func (c *RobustnessTestController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	return txs[0]
}

func (c *RobustnessTestController) TxReceivedHook(txHash string) {
	return
}

func (c *RobustnessTestController) OnStateChange(txHash string, fromState common.LifecycleState, toState common.LifecycleState) {
	log.Debug("tx state change", "tx", common.PrettifyHash(txHash), "from", fromState, "to", toState)
	return
}

func (c *RobustnessTestController) PivotReachedHook(txHash string, currentState common.LifecycleState) (nextState common.LifecycleState, suspend bool) {
	log.Info("Pivot reached", "counter", c.counter)
	c.counter += 1
	if c.counter > 10 {
		return common.CONFIRMED, false
	}
	switch currentState {
	case common.CREATED:
		return common.PENDING, false
	case common.PENDING:
		return common.EXECUTED, false
	case common.EXECUTED:
		return common.PENDING, false
	case common.CONFIRMED:
		return common.CONFIRMED, false
	case common.DROPPED:
		return common.DROPPED, false
	}
	return common.CONFIRMED, false
}

func (c *RobustnessTestController) TxFinishedHook(txHash string) {
	log.Info("Test passed")
	return
}

func (c *RobustnessTestController) TxResumeHook(txHash string) {
	return
}
