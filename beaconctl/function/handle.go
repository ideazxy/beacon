package function

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

type handlerFunc func(*cli.Context, *etcd.Client)

func handle(c *cli.Context, fn handlerFunc) {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	fn(c, etcdClient(c))
}
