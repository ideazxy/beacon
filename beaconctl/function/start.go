package function

import (
	"fmt"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
	"github.com/ideazxy/beacon/command"
)

func NewStartCmd() cli.Command {
	return cli.Command{
		Name:  "start",
		Usage: "start a new service instance",
		Flags: []cli.Flag{
			cli.StringSliceFlag{Name: "env, e", Value: &cli.StringSlice{}, Usage: "set container environment variables"},
			cli.StringSliceFlag{Name: "volume, v", Value: &cli.StringSlice{}, Usage: "bind mount a volumn"},
			cli.StringFlag{Name: "service", Usage: "set service name"},
			cli.StringFlag{Name: "proto", Value: "tcp", Usage: "set service protocol, 'tcp' or 'http'"},
			cli.StringFlag{Name: "cluster", Value: "default", Usage: "set cluster name the service belong"},
			cli.StringFlag{Name: "listen", Usage: "set port this container listens"},
			cli.StringSliceFlag{Name: "target", Value: &cli.StringSlice{}, Usage: "set who will receive this command"},
			cli.StringFlag{Name: "username", Usage: "set username for docker registry"},
			cli.StringFlag{Name: "password", Usage: "set password for docker registry"},
			cli.StringFlag{Name: "email", Usage: "set email for docker registry"},
			cli.BoolFlag{Name: "local", Usage: "start a local container, and then register it"},
			cli.BoolFlag{Name: "tls", Usage: "set tls mode for docker daemon"},
			cli.StringFlag{Name: "cert", Usage: "set cert directory for docker daemon"},
			cli.StringFlag{Name: "daemon", Usage: "set docker daemon for local mode"},
			cli.StringFlag{Name: "host", Usage: "set host IP"},
		},
		Action: func(c *cli.Context) {
			handle(c, start)
		},
	}
}

func start(c *cli.Context, client *etcd.Client) {
	if len(c.Args()) < 1 {
		log.Fatalln("image name is required!")
	}
	cmd := &command.Command{
		Id:          time.Now().Format("20060102030405"),
		Type:        "add",
		Image:       c.Args()[0],
		Env:         c.StringSlice("env"),
		Vol:         c.StringSlice("volume"),
		Listen:      c.String("listen"),
		Service:     c.String("service"),
		Cluster:     c.String("cluster"),
		Proto:       c.String("proto"),
		DockerUser:  c.String("username"),
		DockerPswd:  c.String("password"),
		DockerEmail: c.String("email"),
	}
	if len(c.Args()) > 1 {
		cmd.Cmd = c.Args()[1:]
	}
	if len(strings.Split(cmd.Image, ":")) != 2 {
		cmd.Image += ":latest"
	}

	log.Infoln("generate a new command: ", cmd.Marshal())

	if c.Bool("local") {
		log.Infoln("just start container on local host")
		if err := cmd.Process(dockerClient(c), client, c.String("host"), c.GlobalString("prefix")); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Fatalln("execute command failed.")
		}
		return
	}

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
