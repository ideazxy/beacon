package function

import (
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

type handlerFunc func(*cli.Context, *etcd.Client)

func handle(c *cli.Context, fn handlerFunc) {
	fn(c, etcdClient(c))
}
