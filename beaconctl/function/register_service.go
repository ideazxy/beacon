package function

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/register"
)

func NewRegisterCmd() cli.Command {
	return cli.Command{
		Name:  "register",
		Usage: "register a new service to etcd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Value: "", Usage: "service name"},
			cli.StringFlag{Name: "proto", Value: "", Usage: "'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "the target cluster this service register to"},
			cli.StringFlag{Name: "listen", Value: "", Usage: "the address this service listens"},
			cli.StringFlag{Name: "host", Value: "", Usage: "the host name this service watch"},
		},
		Action: func(c *cli.Context) {
			handle(c, registerService)
		},
	}
}

func registerService(c *cli.Context, client *etcd.Client) {
	service := &register.Service{
		Name:    c.String("name"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Listen:  c.String("listen"),
		Host:    c.String("host"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := register.AddService(client, service); err != nil {
		log.Fatalln(err.Error())
	}
}
