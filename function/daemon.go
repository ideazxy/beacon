package function

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/ideazxy/beacon/command"
	reg "github.com/ideazxy/beacon/register"
)

const (
	HOST_TTL            uint64        = 60
	HEARTBEATS_INTERVAL time.Duration = 5 * time.Second
)

var (
	lastIndex uint64
)

func NewDaemonCmd() cli.Command {
	return cli.Command{
		Name:  "daemon",
		Usage: "start beacon as a daemon service.",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "name", Usage: "set host name"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name"},
			cli.StringFlag{Name: "ip", Value: "127.0.0.1", Usage: "set host IP"},
			cli.StringSliceFlag{Name: "add-host", Value: &cli.StringSlice{}, Usage: "add a custom host-to-IP mapping (host:ip)"},
			cli.StringFlag{Name: "docker", Value: "unix:///var/run/docker.sock", Usage: "set docker daemon"},
			cli.BoolFlag{Name: "tls", Usage: "set tls mode for docker daemon"},
			cli.StringFlag{Name: "cert", Usage: "set cert directory for docker daemon if tls flag is set"},
		},
		Action: func(c *cli.Context) {
			handle(c, doDaemon)
		},
	}
}

func check(client *etcd.Client, name string) []*command.Command {
	key := fmt.Sprintf("/beacon/commands/single/%s/", name)
	if prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), key)
	}

	if lastIndex == 0 {
		resp, err := client.Get(key, false, true)
		if err != nil {
			log.Warnln(err.Error())
			return nil
		}
		lastIndex = resp.EtcdIndex
	}

	resp, err := client.Watch(key, lastIndex+1, true, nil, nil)
	if err != nil {
		log.Debugln(err.Error())
		return nil
	}
	lastIndex = resp.Node.ModifiedIndex

	resp, err = client.Get(key, true, true)
	if err != nil {
		log.Debugln(err.Error())
		return nil
	}
	node := resp.Node
	if !node.Dir {
		log.Warningf("dirty data! [%s] should be a dir.\n", key)
		return nil
	}
	cmds := make([]*command.Command, 0)
	for _, n := range node.Nodes {
		if n.Dir {
			log.Warningf("dirty data! [%s] should be command value, dir got.\n", n.Key)
			continue
		}
		log.WithFields(log.Fields{
			"raw": n.Value,
		}).Debugln("find a new command.")
		var cmd command.Command
		err := command.Unmarshal(n.Value, &cmd)
		if err != nil {
			log.Warningln(err.Error())
			continue
		}
		cmds = append(cmds, &cmd)
	}
	return cmds
}

func remove(client *etcd.Client, name, id string) {
	key := fmt.Sprintf("/beacon/commands/single/%s/%s", name, id)
	if prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), key)
	}
	if _, err := client.Delete(key, true); err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Fatalln("remove finished command failed.")
	} else {
		log.WithFields(log.Fields{
			"id": id,
		}).Infoln("command is removed.")
	}
}

func register(client *etcd.Client, cluster, name, ip string) {
	if cluster == "" {
		log.Fatalln("cluster name is required.")
	}
	if name == "" {
		log.Fatalln("host name is required.")
	}
	if ip == "" {
		log.Fatalln("host ip is required.")
	}

	dir := fmt.Sprintf("/beacon/cluster/%s/%s", cluster, name)
	if prefix != "" {
		dir = fmt.Sprintf("/%s%s", strings.Trim(prefix, "/"), dir)
	}
	registerHost(client, dir, ip)

	// refresh ttl every 30 seconds:
	for {
		if _, err := client.UpdateDir(dir, HOST_TTL); err != nil {
			log.WithFields(log.Fields{
				"dir":   dir,
				"error": err.Error(),
			}).Warnln("send heartbeat failed. try to register new ...")

			registerHost(client, dir, ip)
		}
		time.Sleep(HEARTBEATS_INTERVAL)
	}
}

func registerHost(client *etcd.Client, hostKey, ip string) {
	if _, err := client.Get(hostKey, false, false); err != nil {
		if _, err := client.CreateDir(hostKey, HOST_TTL); err != nil {
			log.WithFields(log.Fields{
				"dir":   hostKey,
				"error": err.Error(),
			}).Fatalln("register host failed.")
		}
		log.WithFields(log.Fields{
			"dir": hostKey,
		}).Infoln("register new host.")
	} else {
		log.Infoln("host already exists.")
	}

	key := fmt.Sprintf("%s/ip", hostKey)
	if _, err := client.Set(key, ip, 0); err != nil {
		log.WithFields(log.Fields{
			"key":   key,
			"error": err.Error(),
		}).Fatalln("register host ip failed.")
	}
	log.WithFields(log.Fields{
		"key": key,
		"ip":  ip,
	}).Infoln("set IP for host.")
}

func monitor(dcClient *docker.Client, etClient *etcd.Client, cluster string) {
	listener := make(chan *docker.APIEvents)
	if err := dcClient.AddEventListener(listener); err != nil {
		log.Fatalln(err.Error())
	}
	for {
		select {
		case event := <-listener:
			if event.Status == "stop" || event.Status == "kill" {
				log.WithFields(log.Fields{
					"id":     event.ID,
					"status": event.Status,
					"from":   event.From,
				}).Debugln("capture new event.")

				if err := reg.FindAndRemoveInstance(etClient, cluster, prefix, event.ID); err != nil {
					log.Errorln(err.Error())
				}
			}
		}
	}
}

func doDaemon(c *cli.Context, client *etcd.Client) {
	log.WithFields(log.Fields{
		"hostName": c.String("name"),
		"hostIp":   c.String("ip"),
	}).Infoln("beacond start.")

	go register(client, c.String("cluster"), c.String("name"), c.String("ip"))

	go monitor(dockerClient(c), client, c.String("cluster"))

	lastIndex = 0
	for {
		commands := check(client, c.String("name"))
		if commands != nil {
			for _, command := range commands {
				command.ExtraHosts = append(command.ExtraHosts, c.StringSlice("add-host")...)
				log.WithFields(log.Fields{
					"id":   command.Id,
					"type": command.Type,
				}).Infoln("start to execute a new command.")

				err := command.Process(dockerClient(c), client, c.String("ip"), prefix)
				if err != nil {
					log.Errorln(err.Error())
				}

				remove(client, c.String("name"), command.Id)
			}
		}
	}
}
