package command

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/ideazxy/beacon/register"
)

type Command struct {
	Id          string
	Type        string // add|update|remove
	Image       string
	Name        string
	Cmd         []string
	Env         []string
	Vol         []string
	Listen      string // only one port will be registerd
	Service     string // name of service
	Cluster     string
	Proto       string // tcp|http
	DockerUser  string
	DockerPswd  string
	DockerEmail string
}

func Unmarshal(raw string, cmd *Command) error {
	return json.Unmarshal([]byte(raw), cmd)
}

func (c *Command) Process(dockerClient *docker.Client, etcdClient *etcd.Client, hostIp, prefix string) error {
	// todo: check parameters...
	switch c.Type {
	case "add":
		container, err := c.runContainer(dockerClient)
		if err != nil {
			return err
		}

		if c.Service != "" {
			log.Debugln("prepare to register instance.")
			exportList, ok := container.NetworkSettings.Ports[docker.Port(c.Listen+"/tcp")]
			if !ok || len(exportList) < 1 {
				return errors.New("the port to listen not found")
			}
			instance := &register.Instance{
				Name:    container.ID,
				Service: c.Service,
				Cluster: c.Cluster,
				Proto:   c.Proto,
				Ip:      hostIp,
				Prefix:  prefix,
				Listen:  exportList[0].HostPort,
			}
			if err = register.AddInstance(etcdClient, instance); err != nil {
				return err
			}
		} else {
			log.Infoln("skip over instance register.")
		}

	case "remove":
		containers, err := c.stopContainer(dockerClient)
		if err != nil {
			return err
		}
		if c.Service != "" {
			for _, container := range containers {
				instance := &register.Instance{
					Name:    container.ID,
					Service: c.Service,
					Cluster: c.Cluster,
					Proto:   c.Proto,
					Prefix:  prefix,
				}
				if err = register.RemoveInstance(etcdClient, instance); err != nil {
					return err
				}
			}
			log.WithFields(log.Fields{
				"Num": len(containers),
			}).Infoln("instances unregistered.")
		}

	default:
		log.WithFields(log.Fields{
			"operation": c.Type,
		}).Errorln("unknown operation type.")
		return errors.New("unknown type")
	}
	return nil
}

func (c *Command) runContainer(client *docker.Client) (*docker.Container, error) {
	imageList, err := client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		return nil, err
	}
	var exist bool
outer:
	for _, image := range imageList {
		for _, name := range image.RepoTags {
			if c.Image == name {
				log.Infoln("image already exists.")
				exist = true
				break outer
			}
		}
	}

	if !exist {
		log.Debugln("start to pull image.")
		repository, tag := parseImageName(c.Image)
		indexName, _ := splitReposName(repository)
		err = client.PullImage(docker.PullImageOptions{
			Repository: repository,
			Tag:        tag,
		}, docker.AuthConfiguration{
			Username:      c.DockerUser,
			Password:      c.DockerPswd,
			Email:         c.DockerEmail,
			ServerAddress: indexName,
		})
		if err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"repository": repository,
			"tag":        tag,
		}).Infoln("pulled new image.")
	}

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
	if log.GetLevel() >= log.DebugLevel {
		b, err := json.Marshal(container)
		if err != nil {
			log.Warningln("marshal container info failed: ", err.Error())
		} else {
			log.Debugln("created container: ", string(b))
		}
	}

	err = client.StartContainer(container.ID, &docker.HostConfig{
		Binds:           c.Vol,
		RestartPolicy:   docker.RestartOnFailure(3),
		PublishAllPorts: true,
	})
	if err != nil {
		return nil, err
	}
	log.Debugln("start container: ", container.ID)

	log.Infoln("wait 10 seconds...")
	time.Sleep(10 * time.Second)

	container, err = client.InspectContainer(container.ID)
	if err != nil {
		return nil, err
	}
	if !container.State.Running {
		log.Warnln("container is not running!")
		return nil, errors.New("container not running")
	}
	log.WithFields(log.Fields{
		"Id":     container.ID,
		"Name":   container.Name,
		"Status": "running",
	}).Infoln("run container successfully.")

	return container, nil
}

func (c *Command) stopContainer(client *docker.Client) ([]*docker.APIContainers, error) {
	results, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

	stopped := make([]*docker.APIContainers, 0)
	for _, container := range results {
		if log.GetLevel() >= log.DebugLevel {
			b, err := json.Marshal(&container)
			if err == nil {
				log.Debugln("check container: ", string(b))
			}
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
		log.Debugln("match container: ", id)

		if err = client.StopContainer(id, 10); err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"Id":    container.ID,
			"Name":  container.Names,
			"Image": container.Image,
		}).Infoln("container stopped.")

		stopped = append(stopped, &container)
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
