package master

import (
	"bufio"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethmonitor/master/common"
	"github.com/ethereum/go-ethereum/ethmonitor/rpc"
	"github.com/ethereum/go-ethereum/log"
	"google.golang.org/grpc"
	"os"
	"strconv"
	"strings"
)

type TxController interface {
	SelectTxToTraverse(txs []*Transaction) (tx *Transaction)
	// this method is called when traverser for tx is created, returns a bool indicating whether schedule the tx (return hash to dApp)
	TxReceivedHook(txHash string)
	// this method is called when tx state is about to change
	OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState)
	// this method is called between each tx state transition
	PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool)
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

func (t *TrivialController) OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState) {
	return
}

func (t *TrivialController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	return rpc.TxState_CONFIRMED, false
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

func (c *ConsoleController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	fmt.Println("================= Console Controller =================")
	if currentState == rpc.TxState_CONFIRMED || currentState == rpc.TxState_DROPPED {
		return currentState, false
	}
	options := make([]string, 0)
	if currentState != rpc.TxState_PENDING {
		options = append(options, rpc.TxState_PENDING.String())
	}
	if currentState != rpc.TxState_EXECUTED {
		options = append(options, rpc.TxState_EXECUTED.String())
	}
	options = append(options, rpc.TxState_CONFIRMED.String())
	options = append(options, rpc.TxState_DROPPED.String())
	options = append(options, "suspend")

	index := c.selectFromOptions(fmt.Sprintf("Traverse pivot reached, current state: %s", currentState), "Select next state: ", options)
	if index == len(options)-1 {
		// select to suspend
		return 0, true
	}
	fmt.Println("================= Console Controller =================")
	return rpc.TxState(rpc.TxState_value[options[index]]), false
}

func (c *ConsoleController) TxFinishedHook(txHash string) {
	log.Info("Tx finished", "tx", txHash)
}

func (c *ConsoleController) TxResumeHook(txHash string) {
	log.Debug("Resume traverse tx", "tx", common.PrettifyHash(txHash))
}

func (c *ConsoleController) OnStateChange(txHash string, currentState rpc.TxState, nextState rpc.TxState) {
	log.Debug("tx state change", "tx", common.PrettifyHash(txHash), "from", currentState, "to", nextState)
	return
}

type DarcherController struct {
	txStates map[string]rpc.TxState

	ctx    context.Context
	client rpc.EthmonitorControllerServiceClient
}

func NewDarcherController(darcherPort int) *DarcherController {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "localhost", darcherPort), grpc.WithInsecure())
	if err != nil {
		log.Error("Connect to Darcher failed", "err", err)
		return nil
	}
	log.Info("Connected to Darcher")
	return &DarcherController{
		txStates: make(map[string]rpc.TxState),

		ctx:    context.Background(),
		client: rpc.NewEthmonitorControllerServiceClient(conn),
	}
}

func (d *DarcherController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	txHashes := make([]string, len(txs))
	m := make(map[string]*Transaction)
	for i, tx := range txs {
		txHashes[i] = tx.Hash()
		m[txHashes[i]] = tx
	}
	reply, err := d.client.SelectTx(d.ctx, &rpc.SelectTxControlMsg{
		CandidateHashes: txHashes,
	})
	if err != nil {
		log.Error("SelectTx RPC error", "err", err)
		return txs[0]
	}
	return m[reply.Selection]

}

func (d *DarcherController) TxReceivedHook(txHash string) {
	d.txStates[txHash] = rpc.TxState_CREATED
	_, err := d.client.NotifyTxReceived(d.ctx, &rpc.TxReceivedMsg{Hash: txHash})
	if err != nil {
		log.Error("NotifyTxReceived RPC error", "err", err)
	}
}

func (d *DarcherController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	reply, err := d.client.AskForNextState(d.ctx, &rpc.TxStateControlMsg{
		Hash:         txHash,
		CurrentState: currentState,
	})
	if err != nil {
		log.Error("AskForNextState RPC error", "err", err)
		return rpc.TxState_CONFIRMED, false
	}
	return reply.GetNextState(), false

}

func (d *DarcherController) TxFinishedHook(txHash string) {
	d.txStates[txHash] = rpc.TxState_CREATED
	_, err := d.client.NotifyTxFinished(d.ctx, &rpc.TxFinishedMsg{Hash: txHash})
	if err != nil {
		log.Error("NotifyTxFinished RPC error", "err", err)
	}
}

func (d *DarcherController) TxResumeHook(txHash string) {
	return
}

func (d *DarcherController) OnStateChange(txHash string, currentState rpc.TxState, nextState rpc.TxState) {
	d.txStates[txHash] = rpc.TxState_CREATED
	_, err := d.client.NotifyTxStateChangeMsg(d.ctx, &rpc.TxStateChangeMsg{
		Hash: txHash,
		From: currentState,
		To:   nextState,
	})
	if err != nil {
		log.Error("NotifyTxStateChangeMsg RPC error", "err", err)
	}
}

/**
Deploy controller is useful when deploy contracts, it is much faster because
it doesn't require
*/
type DeployController struct {
}

func NewDeployController() *DeployController {
	return &DeployController{}
}

func (c *DeployController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	for _, tx := range txs {
		if tx.state == rpc.TxState_CREATED || tx.state == rpc.TxState_PENDING {
			return tx
		}
	}
	return txs[0]
}

func (c *DeployController) TxReceivedHook(txHash string) {
	log.Info("Tx received", "hash", txHash)
}

func (c *DeployController) OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState) {
	return
}

func (c *DeployController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	switch currentState {
	case rpc.TxState_CREATED:
		fallthrough
	case rpc.TxState_PENDING:
		log.Debug("Execute tx", "hash", txHash)
		return rpc.TxState_EXECUTED, false
	default:
		// suspend the transaction
		return rpc.TxState_CONFIRMED, false
	}
}

func (c *DeployController) TxFinishedHook(txHash string) {
	log.Info("Tx finished", "hash", txHash)
}

func (c *DeployController) TxResumeHook(txHash string) {
	return
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

func (c *RobustnessTestController) OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState) {
	log.Debug("tx state change", "tx", common.PrettifyHash(txHash), "from", fromState, "to", toState)
	return
}

func (c *RobustnessTestController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	log.Info("Pivot reached", "counter", c.counter)
	c.counter += 1
	if c.counter > 10 {
		return rpc.TxState_CONFIRMED, false
	}
	switch currentState {
	case rpc.TxState_CREATED:
		return rpc.TxState_PENDING, false
	case rpc.TxState_PENDING:
		return rpc.TxState_EXECUTED, false
	case rpc.TxState_EXECUTED:
		return rpc.TxState_PENDING, false
	case rpc.TxState_CONFIRMED:
		return rpc.TxState_CONFIRMED, false
	case rpc.TxState_DROPPED:
		return rpc.TxState_DROPPED, false
	}
	return rpc.TxState_CONFIRMED, false
}

func (c *RobustnessTestController) TxFinishedHook(txHash string) {
	log.Info("Test passed")
	return
}

func (c *RobustnessTestController) TxResumeHook(txHash string) {
	return
}
