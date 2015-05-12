package register

import (
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
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/%s/%s",
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
		"key":   key,
		"value": value,
	}).Infoln("added new instance.")

	return nil
}

func RemoveInstance(client *etcd.Client, i *Instance) error {
	basekey := appendPrefix(fmt.Sprintf("/beacon/registry/%s/%s/%s",
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
		"key": key,
	}).Infoln("removed instance.")

	return nil
}
