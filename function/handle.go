package function

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

var (
	prefix string
)

type handlerFunc func(*cli.Context, *etcd.Client)

func handle(c *cli.Context, fn handlerFunc) {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	prefix = c.GlobalString("prefix")

	fn(c, etcdClient(c))
}
