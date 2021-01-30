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
	"sync/atomic"
	"time"
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
	// this method is called when tx execution error
	TxErrorHook(txError *rpc.TxErrorMsg)
	// this method is called at most for each tx if there is contract vulnerability detected
	ContractVulnerabilityHook(vulReport *rpc.ContractVulReport)
}

/**
Trivial Controller does nothing and let traverser go smoothly
*/
type TrivialController struct{}

func (t *TrivialController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	return txs[0]
}

func (t *TrivialController) TxReceivedHook(txHash string) {
	log.Info("Tx received", "hash", txHash)
}

func (t *TrivialController) OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState) {
	log.Info("Tx lifecycle state changed", "hash", txHash)
}

func (t *TrivialController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	return rpc.TxState_CONFIRMED, false
}

func (t *TrivialController) TxFinishedHook(txHash string) {
	log.Info("Tx finished", "hash", txHash)
}

func (t *TrivialController) TxResumeHook(txHash string) {
	return
}

func (t *TrivialController) TxErrorHook(txError *rpc.TxErrorMsg) {
	log.Error("Failed transaction detected", "tx", txError.Hash, "msg", txError.Description)
}

func (t *TrivialController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {
	log.Error("Contract vulnerability detected", "tx", vulReport.TxHash, "msg", vulReport.Description)
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

func (c *ConsoleController) TxErrorHook(txError *rpc.TxErrorMsg) {
	log.Error("Tx execution error detected", "hash", common.PrettifyHash(txError.GetHash()), "type", txError.Type.String(), "err", txError.GetDescription())
}

func (c *ConsoleController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {
	log.Error("Contract vulnerability detected", "contract", vulReport.GetAddress(), "tx", common.PrettifyHash(vulReport.GetTxHash()), "vul", vulReport.GetType().String())
}

// Darcher Controller use upstream Darcher to control the tx lifecycle.
// If connection to upstream is broken, it will try to reconnect and meanwhile fallback to TrivialController
type DarcherController struct {
	txStates map[string]rpc.TxState

	ctx             context.Context
	analyzerAddress string
	client          rpc.EthmonitorControllerServiceClient

	connectionStatus atomic.Value // if not connected, use the fallback controller
	fallback         TxController
}

const (
	DarcherConnected    = "connected"
	DarcherConnecting   = "connecting"
	DarcherDisconnected = "disconnected"
)

func NewDarcherController(analyzerAddress string) *DarcherController {
	// parent context is not used for now, may make use of it in the future
	ctx := context.Background()
	var connectionStatus atomic.Value
	timeoutCtx, _ := context.WithTimeout(ctx, time.Second)
	conn, err := grpc.DialContext(timeoutCtx, analyzerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Error("Connect to Darcher failed", "err", err)
		log.Warn("Fallback to TrivialController")
		connectionStatus.Store(DarcherDisconnected)
	} else {
		log.Info("Connected to Darcher", "address", analyzerAddress)
		connectionStatus.Store(DarcherConnected)
	}
	controller := &DarcherController{
		txStates: make(map[string]rpc.TxState),

		ctx:             ctx,
		analyzerAddress: analyzerAddress,
		client:          rpc.NewEthmonitorControllerServiceClient(conn),

		connectionStatus: connectionStatus,
		fallback:         &TrivialController{},
	}
	// start a goroutine to reconnect
	go controller.connectionLoop()
	return controller
}

func (d *DarcherController) connectionLoop() {
	for {
		select {
		case <-d.ctx.Done():
			break
		default:
			if d.connectionStatus.Load() == DarcherConnected || d.connectionStatus.Load() == DarcherConnecting {
				time.Sleep(time.Second)
			} else if d.connectionStatus.Load() == DarcherDisconnected {
				log.Info("Reconnecting Darcher", "address", d.analyzerAddress)
				timeoutCtx, _ := context.WithTimeout(d.ctx, 10*time.Second)
				conn, err := grpc.DialContext(timeoutCtx, d.analyzerAddress, grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					log.Error("Reconnect to Darcher failed", "address", d.analyzerAddress, "err", err)
					time.Sleep(5 * time.Second)
				} else {
					d.client = rpc.NewEthmonitorControllerServiceClient(conn)
					log.Info("Darcher reconnected")
					d.connectionStatus.Store(DarcherConnecting)
				}
			}
		}
	}
}

func (d *DarcherController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		return d.fallback.SelectTxToTraverse(txs)
	}
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
		log.Error("DarcherController RPC error", "method", "SelectTx", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		return d.fallback.SelectTxToTraverse(txs)
	}
	return m[reply.Selection]

}

func (d *DarcherController) TxReceivedHook(txHash string) {
	if d.connectionStatus.Load() == DarcherDisconnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.TxReceivedHook(txHash)
		return
	} else if d.connectionStatus.Load() == DarcherConnecting {
		// if Darcher is reconnected ready, set the status as connected
		d.connectionStatus.Store(DarcherConnected)
	}
	d.txStates[txHash] = rpc.TxState_CREATED
	_, err := d.client.NotifyTxReceived(d.ctx, &rpc.TxReceivedMsg{Hash: txHash})
	if err != nil {
		log.Error("DarcherController RPC error", "method", "TxReceiveHook", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.TxReceivedHook(txHash)
	}
}

func (d *DarcherController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		return d.fallback.PivotReachedHook(txHash, currentState)
	}
	reply, err := d.client.AskForNextState(d.ctx, &rpc.TxStateControlMsg{
		Hash:         txHash,
		CurrentState: currentState,
	})
	if err != nil {
		log.Error("DarcherController RPC error", "method", "AskForNextState", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		return d.fallback.PivotReachedHook(txHash, currentState)
	}
	return reply.GetNextState(), false

}

func (d *DarcherController) TxFinishedHook(txHash string) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.TxFinishedHook(txHash)
		return
	}
	_, err := d.client.NotifyTxFinished(d.ctx, &rpc.TxFinishedMsg{Hash: txHash})
	if err != nil {
		log.Error("DarcherController RPC error", "method", "NotifyTxFinished", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.TxFinishedHook(txHash)
	}
}

func (d *DarcherController) TxResumeHook(txHash string) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.TxResumeHook(txHash)
		return
	}
	_, err := d.client.NotifyTxTraverseStart(d.ctx, &rpc.TxTraverseStartMsg{Hash: txHash})
	if err != nil {
		log.Error("DarcherController RPC error", "method", "NotifyTxTraverseStart", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.TxFinishedHook(txHash)
	}
}

func (d *DarcherController) OnStateChange(txHash string, currentState rpc.TxState, nextState rpc.TxState) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.OnStateChange(txHash, currentState, nextState)
		return
	}
	d.txStates[txHash] = nextState
	_, err := d.client.NotifyTxStateChangeMsg(d.ctx, &rpc.TxStateChangeMsg{
		Hash: txHash,
		From: currentState,
		To:   nextState,
	})
	if err != nil {
		log.Error("DarcherController RPC error", "method", "NotifyTxStateChangeMsg", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.OnStateChange(txHash, currentState, nextState)
	}
}

func (d *DarcherController) TxErrorHook(txError *rpc.TxErrorMsg) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.TxErrorHook(txError)
		return
	}
	_, err := d.client.NotifyTxError(d.ctx, txError)
	if err != nil {
		log.Error("DarcherController RPC error", "method", "NotifyTxError", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.TxErrorHook(txError)
	}
}

func (d *DarcherController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {
	if d.connectionStatus.Load() != DarcherConnected {
		// if Darcher is disconnected, use fallback controller
		d.fallback.ContractVulnerabilityHook(vulReport)
		return
	}
	_, err := d.client.NotifyContractVulnerability(d.ctx, vulReport)
	if err != nil {
		log.Error("DarcherController RPC error", "method", "NotifyContractVulnerability", "err", err)
		log.Warn("Fallback to TrivialController")
		d.connectionStatus.Store(DarcherDisconnected)
		d.fallback.ContractVulnerabilityHook(vulReport)
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

func (c *DeployController) TxErrorHook(txError *rpc.TxErrorMsg) {
	return
}

func (c *DeployController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {
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

func (c *RobustnessTestController) TxErrorHook(txError *rpc.TxErrorMsg) {
	return
}

func (c *RobustnessTestController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {
	return
}

type TraverseController struct {
	revertedTxs map[string]bool
}

func NewTraverseController() *TraverseController {
	return &TraverseController{revertedTxs: make(map[string]bool)}
}

func (c *TraverseController) SelectTxToTraverse(txs []*Transaction) (tx *Transaction) {
	return txs[0]
}

func (c *TraverseController) TxReceivedHook(txHash string) {
	log.Info("Tx received", "tx", txHash)
}

func (c *TraverseController) OnStateChange(txHash string, fromState rpc.TxState, toState rpc.TxState) {
	log.Info("Tx state change", "tx", txHash, "from", fromState, "to", toState)
	if fromState == rpc.TxState_EXECUTED && toState == rpc.TxState_PENDING {
		c.revertedTxs[txHash] = true
	}
}

func (c *TraverseController) PivotReachedHook(txHash string, currentState rpc.TxState) (nextState rpc.TxState, suspend bool) {
	switch currentState {
	case rpc.TxState_CREATED:
		return rpc.TxState_PENDING, false
	case rpc.TxState_PENDING:
		return rpc.TxState_EXECUTED, false
	case rpc.TxState_EXECUTED:
		if _, ok := c.revertedTxs[txHash]; ok {
			// already reverted
			return rpc.TxState_CONFIRMED, false
		} else {
			return rpc.TxState_PENDING, false
		}
	case rpc.TxState_CONFIRMED:
		return currentState, false
	case rpc.TxState_DROPPED:
		return rpc.TxState_DROPPED, false
	}
	return rpc.TxState_CONFIRMED, false
}

func (c *TraverseController) TxFinishedHook(txHash string) {
	log.Info("Tx finished", "tx", txHash)
}

func (c *TraverseController) TxResumeHook(txHash string) {

}

func (c *TraverseController) TxErrorHook(txError *rpc.TxErrorMsg) {
	log.Error("Tx execution error", "err", txError.GetDescription())
}

func (c *TraverseController) ContractVulnerabilityHook(vulReport *rpc.ContractVulReport) {

}
