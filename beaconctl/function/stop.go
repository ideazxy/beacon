package function

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/command"
)

func NewStopCmd() cli.Command {
	return cli.Command{
		Name:  "stop",
		Usage: "stop service instance",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "id", Usage: "set container id"},
			cli.StringFlag{Name: "image", Usage: "set image name"},
			cli.StringFlag{Name: "service", Usage: "set service name"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "set service protocol, 'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name this service belong"},
			cli.StringSliceFlag{Name: "target", Value: &cli.StringSlice{}, Usage: "set who will receive this command"},
			cli.BoolFlag{Name: "local", Usage: "stop a local container, and then register it"},
			cli.BoolFlag{Name: "tls", Usage: "set tls mode for docker daemon"},
			cli.StringFlag{Name: "cert", Usage: "set cert directory for docker daemon"},
			cli.StringFlag{Name: "daemon", Usage: "set docker daemon for local mode"},
			cli.StringFlag{Name: "host", Usage: "set host IP"},
		},
		Action: func(c *cli.Context) {
			handle(c, stop)
		},
	}
}

func stop(c *cli.Context, client *etcd.Client) {
	cmd := &command.Command{
		Id:      time.Now().Format("20060102030405"),
		Type:    "remove",
		Name:    c.String("id"),
		Service: c.String("service"),
		Cluster: c.String("cluster"),
		Proto:   c.String("proto"),
	}
	if cmd.Name == "" {
		cmd.Image = c.String("image")
		if cmd.Image == "" {
			log.Fatalln("image name or container id is required!")
		}
	}

	if c.GlobalBool("debug") {
		log.Println("generate a new command: ", cmd.Marshal())
	}

	if c.Bool("local") {
		if c.GlobalBool("debug") {
			log.Println("just stop container on local host")
		}
		if err := cmd.Process(dockerClient(c), etcdClient(c), c.String("host"), c.GlobalString("prefix")); err != nil {
			log.Fatalln(err.Error())
		}
		return
	}

	targets := c.StringSlice("target")
	if targets == nil || len(targets) == 0 {
		log.Fatalln("at least one target should be set!")
	}
	for _, target := range targets {
		key := fmt.Sprintf("/beacon/commands/single/%s/%s/",
			target, cmd.Id)
		if c.GlobalString("prefix") != "" {
			key = fmt.Sprintf("/%s%s", strings.Trim(c.GlobalString("prefix"), "/"), key)
		}

		if _, err := client.Set(key, cmd.Marshal(), 0); err != nil {
			log.Fatalln(err.Error())
		}
	}
}
