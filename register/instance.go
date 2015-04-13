package register

import (
	"fmt"
	"log"
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type Instance struct {
	Name    string
	Service string
	Proto   string
	Cluster string
	Ip      string
	Listen  string
	Prefix  string
}

func AddInstance(client *etcd.Client, i *Instance) error {
	basekey := fmt.Sprintf("/beacon/registry/%s/%s/%s/upstreams/%s",
		i.Cluster, i.Proto, i.Service, i.Name)
	if i.Prefix != "" {
		basekey = fmt.Sprintf("/%s%s", strings.Trim(i.Prefix, "/"), basekey)
	}

	k := basekey + "/listen"
	log.Println("Set key: [", k, "], value: [", i.Listen, "]")
	if _, err := client.Set(k, i.Listen, 0); err != nil {
		return err
	}

	return nil
}

func RemoveInstance(client *etcd.Client, i *Instance) error {
	key := fmt.Sprintf("/beacon/registry/%s/%s/%s/upstreams/%s",
		i.Cluster, i.Proto, i.Service, i.Name)
	if i.Prefix != "" {
		key = fmt.Sprintf("/%s%s", strings.Trim(i.Prefix, "/"), key)
	}

	log.Println("Remove key: [", key, "]")
	if _, err := client.Delete(key, true); err != nil {
		return err
	}

	return nil
}
