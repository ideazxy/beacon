package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/ideazxy/beacon/function"
)

func main() {
	app := cli.NewApp()
	app.Name = "beacon"
	app.Usage = "A light weight tool to manage docker containers."
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "nodes", Value: "http://0.0.0.0:2379", Usage: "etcd endpoints"},
		cli.StringFlag{Name: "prefix", Value: "", Usage: "key prefix for /beacon"},
		cli.BoolFlag{Name: "debug", Usage: "print debug logs"},
	}
	app.Commands = []cli.Command{
		function.NewDaemonCmd(),
		function.NewRegisterCmd(),
		function.NewUnregisterCmd(),
		function.NewAddCmd(),
		function.NewRemoveCmd(),
		function.NewStartCmd(),
		function.NewStopCmd(),
	}

	app.Run(os.Args)
}
