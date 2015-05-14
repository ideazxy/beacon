package function

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	reg "github.com/ideazxy/beacon/register"
)

func NewRegisterCmd() cli.Command {
	return cli.Command{
		Name:  "register",
		Usage: "register a new service to etcd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "service name"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "the target cluster this service register to"},
			cli.StringFlag{Name: "listen", Usage: "the address this service listens"},
			cli.StringFlag{Name: "backend", Usage: "set backend name for http service"},
			cli.StringSliceFlag{Name: "host", Value: &cli.StringSlice{}, Usage: "the host names for http service"},
		},
		Action: func(c *cli.Context) {
			handle(c, doRegisterService)
		},
	}
}

func doRegisterService(c *cli.Context, client *etcd.Client) {
	service := &reg.Service{
		Name:    c.String("name"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Listen:  c.String("listen"),
		Backend: c.String("backend"),
		Hosts:   c.StringSlice("host"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := reg.AddService(client, service); err != nil {
		log.Fatalln(err.Error())
	}
}
