package main

import (
	"errors"
	"fmt"
	ethmonitor "github.com/ethereum/go-ethereum/ethmonitor/master"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/phayes/freeport"
	cli "gopkg.in/urfave/cli.v1"
	"io"
	"os"
	"os/signal"
	"syscall"
)

var (
	ModeFlag = &cli.StringFlag{
		Name:  "ethmonitor.mode",
		Value: "deploy", // deploy or traverse mode, controllers are only valid for traverse mode
		Usage: "deploy or traverse",
	}
	PortFlag = &cli.IntFlag{
		Name:  "ethmonitor.port",
		Value: 0,
	}
	ControllerFlag = &cli.StringFlag{
		Name:  "ethmonitor.controller",
		Value: "trivial",
		Usage: "trivial, console, darcher or traverse",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}
	AnalyzerAddressFlag = &cli.StringFlag{
		Name:  "analyzer.address",
		Value: "localhost:1234",
	}
	ConfirmationRequirementFlag = &cli.Uint64Flag{
		Name:  "confirmation.requirement",
		Value: 1,
	}
)

var (
	flags = []cli.Flag{
		ModeFlag,
		PortFlag,
		ControllerFlag,
		VerbosityFlag,
		AnalyzerAddressFlag,
		ConfirmationRequirementFlag,
	}
)

func main() {
	app := &cli.App{
		Name:   "ethMonitor",
		Usage:  "geth node cluster monitor!",
		Flags:  flags,
		Action: action,
		Before: before,
		After:  after,
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func action(ctx *cli.Context) error {
	// set log verbosity
	useColor := (isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if useColor {
		output = colorable.NewColorableStderr()
	}
	oStream := log.StreamHandler(output, log.TerminalFormat(useColor))
	glogger := log.NewGlogHandler(oStream)
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(VerbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	var err error
	port := ctx.Int(PortFlag.Name)
	if port == 0 {
		port, err = freeport.GetFreePort()
		if err != nil {
			panic(err)
		}
	}

	var ethMonitor node.Lifecycle

	mode := ctx.String(ModeFlag.Name)
	switch mode {
	case "deploy":
		ethMonitor = ethmonitor.NewDeployMonitor(ethmonitor.EthMonitorConfig{ConfirmationNumber: ctx.Uint64(ConfirmationRequirementFlag.Name), ServerPort: port})

	case "traverse":
		var controller ethmonitor.TxController
		switch ctx.String(ControllerFlag.Name) {
		case "trivial":
			controller = &ethmonitor.TrivialController{}
		case "console":
			controller = ethmonitor.NewConsoleController()
		case "darcher":
			analyzerAddr := ctx.String(AnalyzerAddressFlag.Name)
			controller = ethmonitor.NewDarcherController(analyzerAddr)
		case "robustnessTest":
			controller = ethmonitor.NewRobustnessTestController()
		case "traverse":
			controller = ethmonitor.NewTraverseController()
		default:
			return errors.New("unknown controller " + ctx.String(ControllerFlag.Name))
		}
		log.Info("EthMonitor started", "mode", mode, "port", port, "controller", ctx.String(ControllerFlag.Name))
		ethMonitor = ethmonitor.NewTraverseMonitor(ethmonitor.EthMonitorConfig{
			ConfirmationNumber: ctx.Uint64(ConfirmationRequirementFlag.Name),
			ServerPort:         port,
			Controller:         controller,
		})
	default:
		return errors.New("Unsupported EthMonitor " + mode + " mode")
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Info("Receive SIGINT/SIGTERM signal", "signal", sig.String())
		done <- true
	}()
	err = ethMonitor.Start()
	if err != nil {
		return err
	}
	<-done
	err = ethMonitor.Stop()
	if err != nil {
		return err
	}

	return nil
}

func before(ctx *cli.Context) error {
	return nil
}

func after(ctx *cli.Context) error {
	return nil
}
