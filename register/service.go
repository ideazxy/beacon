package register

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/go-etcd/etcd"
)

type Service struct {
	Name    string
	Proto   string
	Cluster string
	Listen  string
	Backend string
	Hosts   []string
	Prefix  string // prefix for etcd key
}

func AddService(client *etcd.Client, s *Service) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/haproxy/%s/%s",
		s.Cluster, s.Proto, s.Name), s.Prefix)

	k := basekey + "/listen"
	log.WithFields(log.Fields{
		"key":   k,
		"value": s.Listen,
	}).Debugln("start to add a new service key.")
	if _, err := client.Set(k, s.Listen, 0); err != nil {
		return err
	}

	if s.Proto == "http" {
		if s.Backend == "" {
			return errors.New("backend name is required.")
		}
		hostDir := fmt.Sprintf("%s/backends/%s/hosts", basekey, s.Backend)
		if s.Hosts != nil {
			for _, host := range s.Hosts {
				log.WithFields(log.Fields{
					"key":  hostDir,
					"host": host,
				}).Debugln("register new host.")
				if _, err := client.Set(fmt.Sprintf("%s/%s", hostDir, host), "", 0); err != nil {
					return err
				}
			}
		}
	}
	log.WithFields(log.Fields{
		"name":    s.Name,
		"backend": s.Backend,
		"proto":   s.Proto,
		"listen":  s.Listen,
		"hosts":   s.Hosts,
	}).Infoln("service registered.")
	return nil
}

func RemoveService(client *etcd.Client, s *Service) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/haproxy/%s",
		s.Cluster, s.Proto), s.Prefix)

	var key string
	if s.Proto == "tcp" {
		key = fmt.Sprintf("%s/%s", basekey, s.Name)
	} else {
		key = fmt.Sprintf("%s/%s/backends/%s", basekey, s.Name, s.Backend)
	}

	log.WithFields(log.Fields{
		"key": key,
	}).Debugln("start to delete a service key.")
	if _, err := client.Delete(key, true); err != nil {
		return err
	}
	log.WithFields(log.Fields{
		"name":    s.Name,
		"backend": s.Backend,
		"proto":   s.Proto,
	}).Infoln("service unregistered.")

	return nil
}
