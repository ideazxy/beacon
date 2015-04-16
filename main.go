package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/ideazxy/beacon/command"
)

var (
	name     string
	ip       string
	nodes    string
	prefix   string
	interval int
	cert     string
	daemon   string
	tls      bool
	debug    bool
)

func check(client *etcd.Client) []*command.Command {
	key := fmt.Sprintf("/beacon/commands/single/%s/", name)
	if prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), key)
	}
	resp, err := client.Get(key, true, true)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	node := resp.Node
	if !node.Dir {
		log.Println("dirty data! should be dir.")
		return nil
	}

	cmds := make([]*command.Command, 0)
	for _, n := range node.Nodes {
		if n.Dir {
			log.Println("dirty data! should be command value, dir got. key: ", n.Key)
			continue
		}
		if debug {
			log.Println("raw command: ", n.Value)
		}
		var cmd command.Command
		err := command.Unmarshal(n.Value, &cmd)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		cmds = append(cmds, &cmd)
	}
	return cmds
}

func remove(client *etcd.Client, id string) {
	key := fmt.Sprintf("/beacon/commands/single/%s/%s", name, id)
	if prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), key)
	}
	if _, err := client.Delete(key, true); err != nil {
		log.Fatalln("remove finished command failed: ", err.Error())
	}
}

func dockerClient() (*docker.Client, error) {
	var client *docker.Client
	var err error
	if tls {
		cer := fmt.Sprintf("%s/cert.pem", cert)
		key := fmt.Sprintf("%s/key.pem", cert)
		ca := fmt.Sprintf("%s/ca.pem", cert)
		if client, err = docker.NewTLSClient(daemon, cer, key, ca); err != nil {
			return nil, err
		}
	} else {
		if client, err = docker.NewClient(daemon); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func init() {
	flag.StringVar(&name, "name", "default", "the host name")
	flag.StringVar(&ip, "ip", "127.0.0.1", "the host ip")
	flag.StringVar(&nodes, "nodes", "", "ip of etcd")
	flag.StringVar(&prefix, "prefix", "", "prefix of the key that beacon will use")
	flag.IntVar(&interval, "interval", 5, "interval second")
	flag.StringVar(&cert, "cert", "", "set cert directory for docker daemon")
	flag.StringVar(&daemon, "daemon", "unix:///var/run/docker.sock", "set docker daemon")
	flag.BoolVar(&tls, "tls", false, "set tls mode for docker daemon")
	flag.BoolVar(&debug, "debug", false, "turn on debug log")
}

func main() {
	flag.Parse()

	if nodes == "" {
		log.Fatalln("etcd node is required.")
	}

	etcdClient := etcd.NewClient(strings.Split(nodes, ","))
	if debug {
		log.Println(etcdClient.GetCluster())
	}
	client, err := dockerClient()
	if err != nil {
		log.Fatalln("failed to connect to docker daemon.")
	}

	log.Printf("beacond start... host name: %s, host ip: %s", name, ip)
	for {
		commands := check(etcdClient)
		if commands != nil && len(commands) > 0 {
			for _, c := range commands {
				if debug {
					log.Println(c.Marshal())
				}

				err := c.Process(client, etcdClient, ip, prefix)
				if err != nil {
					log.Println(err.Error())
				}

				remove(etcdClient, c.Id)
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
