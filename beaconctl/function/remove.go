package function

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/register"
)

func NewRemoveCmd() cli.Command {
	return cli.Command{
		Name:  "remove",
		Usage: "unregister a started service instance.",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "set service instance name"},
			cli.StringFlag{Name: "service", Usage: "set service name"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "set service protocol, 'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name the service belong"},
			cli.StringFlag{Name: "listen", Usage: "set port this container listens"},
			cli.StringFlag{Name: "host", Usage: "set host IP"},
		},
		Action: func(c *cli.Context) {
			handle(c, remove)
		},
	}
}

func remove(c *cli.Context, client *etcd.Client) {
	instance := &register.Instance{
		Name:    c.String("name"),
		Service: c.String("service"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Ip:      c.String("host"),
		Listen:  c.String("listen"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := register.RemoveInstance(client, instance); err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Printf("unregistered a new instance [%s] to service [%s]", instance.Name, instance.Service)
}
