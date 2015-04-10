package function

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/register"
)

func NewUnregisterCmd() cli.Command {
	return cli.Command{
		Name:  "unregister",
		Usage: "unregister a new service from etcd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Value: "", Usage: "service name"},
			cli.StringFlag{Name: "proto", Value: "", Usage: "'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "the target cluster this service register to"},
		},
		Action: func(c *cli.Context) {
			handle(c, unregisterService)
		},
	}
}

func unregisterService(c *cli.Context, client *etcd.Client) {
	service := &register.Service{
		Name:    c.String("name"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := register.RemoveService(client, service); err != nil {
		log.Fatalln(err.Error())
	}
}
