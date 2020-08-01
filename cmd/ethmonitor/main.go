package main

import (
	"fmt"
	ethmonitor "github.com/ethereum/go-ethereum/ethmonitor/master"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/phayes/freeport"
	cli "gopkg.in/urfave/cli.v1"
	"io"
	"os"
)

var (
	Port = &cli.IntFlag{
		Name:  "ethmonitor.port",
		Value: 0,
	}
	Controller = &cli.StringFlag{
		Name:  "ethmonitor.controller",
		Value: "trivial",
	}
	VerbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Usage: "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail",
		Value: 3,
	}
	AnalyzerAddress = &cli.StringFlag{
		Name:  "analyzer.address",
		Value: "localhost:1234",
	}
)

var (
	flags = []cli.Flag{
		Port,
		Controller,
		VerbosityFlag,
		AnalyzerAddress,
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

	port := ctx.Int(Port.Name)
	var controller ethmonitor.TxController
	switch ctx.String(Controller.Name) {
	case "trivial":
		controller = &ethmonitor.TrivialController{}
	case "console":
		controller = ethmonitor.NewConsoleController()
	case "darcher":
		analyzerAddr := ctx.String(AnalyzerAddress.Name)
		controller = ethmonitor.NewDarcherController(analyzerAddr)
	case "deploy":
		controller = ethmonitor.NewDeployController()
	case "robustnessTest":
		controller = ethmonitor.NewRobustnessTestController()
	}
	var err error
	if port == 0 {
		port, err = freeport.GetFreePort()
		if err != nil {
			panic(err)
		}
	}
	log.Info("EthMonitor started", "port", port, "controller", ctx.String(Controller.Name))
	cluster := ethmonitor.NewMonitor(controller, ethmonitor.ClusterConfig{ConfirmationNumber: 1, ServerPort: port})
	cluster.Start()
	return nil
}

func before(ctx *cli.Context) error {
	return nil
}

func after(ctx *cli.Context) error {
	return nil
}
