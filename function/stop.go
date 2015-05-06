package function

import (
	"time"

	log "github.com/Sirupsen/logrus"
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
			cli.StringFlag{Name: "backend", Usage: "set backend name (only for http service)"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "set service protocol, 'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name this service belong"},
			cli.StringSliceFlag{Name: "target", Value: &cli.StringSlice{}, Usage: "set who will receive this command"},
			cli.BoolFlag{Name: "local", Usage: "stop a local container, and then register it"},
			cli.BoolFlag{Name: "tls", Usage: "set tls mode for docker daemon"},
			cli.StringFlag{Name: "cert", Usage: "set cert directory for docker daemon"},
			cli.StringFlag{Name: "docker", Usage: "set docker daemon for local mode"},
		},
		Action: func(c *cli.Context) {
			handle(c, doStop)
		},
	}
}

func doStop(c *cli.Context, client *etcd.Client) {
	cmd := &command.Command{
		Id:      time.Now().Format("20060102030405"),
		Type:    "remove",
		Name:    c.String("id"),
		Service: c.String("service"),
		Backend: c.String("backend"),
		Cluster: c.String("cluster"),
		Proto:   c.String("proto"),
	}
	if cmd.Name == "" {
		cmd.Image = c.String("image")
		if cmd.Image == "" {
			log.Fatalln("image name or container id is required!")
		}
	}

	log.Infoln("generate a new command: ", cmd.Marshal())

	if c.Bool("local") {
		log.Infoln("just start container on local host")
		if err := cmd.Process(dockerClient(c), etcdClient(c), c.String("host"), c.GlobalString("prefix")); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("execute command failed.")
		}
		return
	}

	dispatchCommand(c, client, cmd)
}
