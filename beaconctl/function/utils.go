package function

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
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

func fetchHosts(c *cli.Context, client *etcd.Client) []string {
	key := fmt.Sprintf("/beacon/cluster/%s", c.String("cluster"))
	if c.GlobalString("prefix") != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(c.GlobalString("prefix"), "/"), key)
	}
	resp, err := client.Get(key, false, false)
	if err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Warningln("failed to fetch cluster info.")
		return nil
	}
	hosts := make([]string, 0, len(resp.Node.Nodes))
	for _, node := range resp.Node.Nodes {
		log.Debugln("find host:", node.Key)
		parts := strings.Split(node.Key, "/")
		hosts = append(hosts, parts[len(parts)-1])
	}
	return hosts
}
