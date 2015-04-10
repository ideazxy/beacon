package function

import (
	"log"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

type handlerFunc func(*cli.Context, *etcd.Client)

func handle(c *cli.Context, fn handlerFunc) {
	nodes := c.GlobalString("nodes")
	if nodes == "" {
		log.Fatalln("etcd endpoints (nodes) is required!")
	}
	client := etcd.NewClient(strings.Split(nodes, ","))

	fn(c, client)
}
