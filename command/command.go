package command

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/ideazxy/beacon/register"
)

type Command struct {
	Id      string
	Type    string // add|update|remove
	Image   string
	Name    string
	Cmd     []string
	Env     []string
	Vol     []string
	Listen  string // only one port will be registerd
	Service string // name of service
	Cluster string
	Proto   string // tcp|http
}

func Unmarshal(raw string, cmd *Command) error {
	return json.Unmarshal([]byte(raw), cmd)
}

func (c *Command) Process(dockerClient *docker.Client, etcdClient *etcd.Client, hostIp, prefix string) error {
	// todo: check parameters...
	switch c.Type {
	case "add":
		instance, err := c.runContainer(dockerClient)
		if err != nil {
			return err
		}

		if c.Service != "" {
			instance.Ip = hostIp
			instance.Prefix = prefix
			if err = register.AddInstance(etcdClient, instance); err != nil {
				return err
			}
		}

	case "remove":
		instances, err := c.stopContainer(dockerClient)
		if err != nil {
			return err
		}
		if c.Service != "" {
			for _, instance := range instances {
				instance.Prefix = prefix
				if err = register.RemoveInstance(etcdClient, instance); err != nil {
					return err
				}
			}
			log.Printf("unregister %d instances.\n", len(instances))
		}

	default:
		return errors.New("unknown type")
	}
	return nil
}

func (c *Command) runContainer(client *docker.Client) (*register.Instance, error) {
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: c.Name,
		Config: &docker.Config{
			Env:   c.Env,
			Cmd:   c.Cmd,
			Image: c.Image,
		},
	})
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(container)
	if err != nil {
		log.Println("Warn: ", err.Error())
	}
	log.Println("created container: ", string(b))

	err = client.StartContainer(container.ID, &docker.HostConfig{
		Binds:           c.Vol,
		RestartPolicy:   docker.RestartOnFailure(3),
		PublishAllPorts: true,
	})
	if err != nil {
		return nil, err
	}
	b, err = json.Marshal(container)
	if err != nil {
		log.Println("Warn: ", err.Error())
	}
	log.Println("started container: ", string(b))

	log.Println("wait 10 seconds...")
	time.Sleep(10 * time.Second)

	container, err = client.InspectContainer(container.ID)
	if err != nil {
		return nil, err
	}
	if !container.State.Running {
		log.Println("container not running!")
		return nil, errors.New("container not running")
	}
	log.Println("container is running.")

	instance := &register.Instance{
		Name:    container.ID,
		Service: c.Service,
		Cluster: c.Cluster,
		Proto:   c.Proto,
	}
	exportList, ok := container.NetworkSettings.Ports[docker.Port(c.Listen+"/tcp")]
	if !ok || len(exportList) < 1 {
		return nil, errors.New("the port to listen not found")
	}
	instance.Listen = exportList[0].HostPort
	return instance, nil
}

func (c *Command) stopContainer(client *docker.Client) ([]*register.Instance, error) {
	results, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

	stopped := make([]*register.Instance, 0)
	for _, container := range results {
		b, err := json.Marshal(&container)
		if err == nil {
			log.Println("check container: ", string(b))
		}

		var id string
		if c.Name != "" && strings.HasPrefix(container.ID, c.Name) {
			id = container.ID
		} else if container.Image == c.Image {
			id = container.ID
		}

		if id == "" {
			continue
		}
		log.Println("match container: ", id)

		if err = client.StopContainer(id, 10); err != nil {
			return nil, err
		}
		log.Println("stopped container: ", id)

		instance := &register.Instance{
			Name:    id,
			Service: c.Service,
			Cluster: c.Cluster,
			Proto:   c.Proto,
		}
		stopped = append(stopped, instance)
	}
	return stopped, nil
}

func (c *Command) Marshal() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(b)
}
