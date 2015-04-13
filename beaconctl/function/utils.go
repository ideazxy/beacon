package function

import (
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
)

func dockerClient(c *cli.Context) *docker.Client {
	if c.String("daemon") == "" {
		log.Fatalln("daemon is required!")
	}

	var client *docker.Client
	var err error
	if c.Bool("tls") {
		cert := c.String("cert")
		if cert == "" {
			log.Fatalln("cert directory is required!")
		}
		cer := fmt.Sprintf("%s/cert.pem", cert)
		key := fmt.Sprintf("%s/key.pem", cert)
		ca := fmt.Sprintf("%s/ca.pem", cert)
		if client, err = docker.NewTLSClient(c.String("daemon"), cer, key, ca); err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		if client, err = docker.NewClient(c.String("daemon")); err != nil {
			log.Fatalln(err.Error())
		}
	}
	return client
}

func etcdClient(c *cli.Context) *etcd.Client {
	nodes := c.GlobalString("nodes")
	if nodes == "" {
		log.Fatalln("etcd endpoints (nodes) is required!")
	}
	return etcd.NewClient(strings.Split(nodes, ","))
}
