package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/command"
)

var (
	name     string
	ip       string
	nodes    string
	prefix   string
	interval int
	debug    bool
)

func check(client *etcd.Client) []*command.Command {
	key := fmt.Sprintf("/%s/beacon/commands/single/%s/", strings.Trim(prefix, "/"), name)
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

func init() {
	flag.StringVar(&name, "name", "default", "the host name")
	flag.StringVar(&ip, "ip", "127.0.0.1", "the host ip")
	flag.StringVar(&nodes, "node", "", "ip of etcd")
	flag.StringVar(&prefix, "prefix", "", "prefix of the key that beacon will use")
	flag.IntVar(&interval, "interval", 5, "interval second")
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

	log.Printf("beacond start... host name: %s, host ip: %s", name, ip)
	for {
		commands := check(etcdClient)
		if commands != nil && len(commands) > 0 {
			for _, c := range commands {
				if debug {
					log.Println(c.Marshal())
				}

				err := c.Process()
				if err != nil {
					log.Println(err.Error())
				}
			}
		} else if debug {
			log.Println("No more command to be processd.")
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
