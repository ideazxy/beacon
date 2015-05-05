package function

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	reg "github.com/ideazxy/beacon/register"
)

func NewAddCmd() cli.Command {
	return cli.Command{
		Name:  "add",
		Usage: "register a started service instance.",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "set service instance name"},
			cli.StringFlag{Name: "service", Usage: "set service name"},
			cli.StringFlag{Name: "backend", Usage: "set backend name (only for http service)"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "set service protocol, 'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name the service belong"},
			cli.StringFlag{Name: "port", Usage: "set port this container listens"},
			cli.StringFlag{Name: "host", Usage: "set host IP"},
		},
		Action: func(c *cli.Context) {
			handle(c, add)
		},
	}
}

func add(c *cli.Context, client *etcd.Client) {
	instance := &reg.Instance{
		Name:    c.String("name"),
		Service: c.String("service"),
		Backend: c.String("backend"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Ip:      c.String("host"),
		Listen:  c.String("port"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := reg.AddInstance(client, instance); err != nil {
		log.Fatalln(err.Error())
	}
	log.WithFields(log.Fields{
		"name":    instance.Name,
		"service": instance.Service,
		"backend": instance.Backend,
	}).Infoln("register a new instance.")
}
