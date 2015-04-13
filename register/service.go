package register

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type Service struct {
	Name    string
	Proto   string
	Cluster string
	Listen  string
	Host    string
	Prefix  string // prefix for etcd key
}

func AddService(client *etcd.Client, s *Service) error {
	basekey := fmt.Sprintf("/beacon/registry/%s/%s/%s",
		s.Cluster, s.Proto, s.Name)
	if s.Prefix != "" {
		basekey = fmt.Sprintf("/%s%s", strings.Trim(s.Prefix, "/"), basekey)
	}

	k := basekey + "/listen"
	log.Println("Set key: [", k, "], value: [", s.Listen, "]")
	if _, err := client.Set(k, s.Listen, 0); err != nil {
		return err
	}

	if s.Host != "" {
		k = basekey + "/host"
		log.Println("Set key: [", k, "], value: [", s.Host, "]")
		if _, err := client.Set(k, s.Host, 0); err != nil {
			return err
		}
	}

	return nil
}

func RemoveService(client *etcd.Client, s *Service) error {
	key := fmt.Sprintf("/beacon/registry/%s/%s/%s",
		s.Cluster, s.Proto, s.Name)
	if s.Prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(s.Prefix, "/"), key)
	}

	log.Println("Delete key: [", key, "]")
	if _, err := client.Delete(key, true); err != nil {
		return err
	}

	return nil
}
