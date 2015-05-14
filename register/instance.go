package register

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
)

type Instance struct {
	Name    string
	Backend string
	Service string
	Proto   string
	Cluster string
	Ip      string
	Listen  string
	Prefix  string
}

func AddInstance(client *etcd.Client, i *Instance) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/haproxy/%s/%s",
		i.Cluster, i.Proto, i.Service), i.Prefix)

	var key string
	if i.Proto == "http" {
		key = fmt.Sprintf("%s/backends/%s/upstreams/%s/listen", basekey, i.Backend, i.Name)
	} else { // i.Proto == "tcp"
		key = fmt.Sprintf("%s/upstreams/%s/listen", basekey, i.Name)
	}

	value := fmt.Sprintf("%s:%s", i.Ip, strings.TrimPrefix(i.Listen, ":"))
	log.WithFields(log.Fields{
		"key":   key,
		"value": value,
	}).Debugln("update key/value.")
	if _, err := client.Set(key, value, 0); err != nil {
		log.Debugln(err.Error())
		return err
	}
	log.WithFields(log.Fields{
		"id":      i.Name,
		"service": i.Service,
		"backend": i.Backend,
		"value":   value,
	}).Infoln("instance added.")

	return nil
}

func RemoveInstance(client *etcd.Client, i *Instance) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/haproxy/%s/%s",
		i.Cluster, i.Proto, i.Service), i.Prefix)

	var key string
	if i.Proto == "http" {
		key = fmt.Sprintf("%s/backends/%s/upstreams/%s", basekey, i.Backend, i.Name)
	} else { // i.Proto == "tcp"
		key = fmt.Sprintf("%s/upstreams/%s", basekey, i.Name)
	}

	log.WithFields(log.Fields{
		"key": key,
	}).Debugln("delete key.")
	if _, err := client.Delete(key, true); err != nil {
		log.Debugln(err.Error())
		return err
	}
	log.WithFields(log.Fields{
		"id":      i.Name,
		"service": i.Service,
		"backend": i.Backend,
	}).Infoln("instance removed.")

	return nil
}

func FindAndRemoveInstance(client *etcd.Client, cluster, prefix, name string) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/haproxy", cluster), prefix)

	resp, err := client.Get(basekey, false, true)
	if err != nil {
		return err
	}

	if resp.Node == nil || resp.Node.Nodes == nil {
		return errors.New("empty registry.")
	}

	for _, proto := range resp.Node.Nodes {
		if proto.Nodes == nil || len(proto.Nodes) == 0 {
			continue
		}

		for _, service := range proto.Nodes {
			if strings.HasSuffix(proto.Key, "http") {
				if service.Nodes == nil {
					continue
				}
				for _, backend := range service.Nodes {
					if strings.HasSuffix(backend.Key, "backends") {
						if err = findAndRemoveUpstream(client, backend, name); err != nil {
							return err
						}
					}
				}
			} else {
				if err = findAndRemoveUpstream(client, service, name); err != nil {
					return err
				}
			}
		}
	}
	log.Infoln("no more instance needs clean.")
	return nil
}

func findAndRemoveUpstream(client *etcd.Client, node *etcd.Node, name string) error {
	for _, s := range node.Nodes {
		if strings.HasSuffix(s.Key, "upstreams") {
			if s.Nodes == nil {
				continue
			}
			for _, upstream := range s.Nodes {
				log.WithFields(log.Fields{
					"key": upstream.Key,
				}).Debugln("check instance.")

				keyParts := strings.Split(upstream.Key, "/")
				if strings.HasPrefix(keyParts[len(keyParts)-1], name) {
					if _, err := client.Delete(upstream.Key, true); err != nil {
						return err
					}
					log.WithFields(log.Fields{
						"key": upstream.Key,
					}).Infoln("instance removed.")
				}
			}
		}
	}
	return nil
}
