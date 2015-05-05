package function

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	reg "github.com/ideazxy/beacon/register"
)

func NewUnregisterCmd() cli.Command {
	return cli.Command{
		Name:  "unregister",
		Usage: "unregister a new service from etcd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "service name"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "'tcp' or 'http'"},
			cli.StringFlag{Name: "backend", Usage: "set backend name for http service"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "the target cluster this service register to"},
		},
		Action: func(c *cli.Context) {
			handle(c, doUnregisterService)
		},
	}
}

func doUnregisterService(c *cli.Context, client *etcd.Client) {
	service := &reg.Service{
		Name:    c.String("name"),
		Backend: c.String("backend"),
		Proto:   c.String("proto"),
		Cluster: c.String("cluster"),
		Prefix:  c.GlobalString("prefix"),
	}
	if err := reg.RemoveService(client, service); err != nil {
		log.Fatalln(err.Error())
	}
	log.WithFields(log.Fields{
		"name":     service.Name,
		"backend":  service.Backend,
		"protocol": service.Proto,
	}).Infoln("unregister a service.")
}
