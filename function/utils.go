package function

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/ideazxy/beacon/command"
)

func dockerClient(c *cli.Context) *docker.Client {
	if c.String("docker") == "" {
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
		if client, err = docker.NewTLSClient(c.String("docker"), cer, key, ca); err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		if client, err = docker.NewClient(c.String("docker")); err != nil {
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

func appendTag(name string) string {
	n := strings.LastIndex(name, ":")
	if n < 0 {
		return name + ":latest"
	}
	if tag := name[n+1:]; strings.Contains(tag, "/") {
		return name + ":latest"
	}
	return name
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

func dispatchCommand(c *cli.Context, client *etcd.Client, cmd *command.Command) {
	targets := c.StringSlice("target")
	if targets == nil || len(targets) == 0 {
		log.Warningln("no target set! try to send command to all registered host.")
		targets = fetchHosts(c, client)
	}
	if targets == nil {
		log.Fatalln("no target to send command.")
	} else {
		log.Infoln("send command to: ", targets)
	}
	for _, target := range targets {
		key := fmt.Sprintf("/beacon/commands/single/%s/%s/",
			target, cmd.Id)
		if c.GlobalString("prefix") != "" {
			key = fmt.Sprintf("/%s%s", strings.Trim(c.GlobalString("prefix"), "/"), key)
		}

		if _, err := client.Set(key, cmd.Marshal(), 0); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("send command failed.")
		}
	}
}
