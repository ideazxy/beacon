# beacon
A lightweight service register/discovery framework.

Beacon mainly does three things:
* run or stop docker container on cluster;
* register or unregister container as a instance of the service;
* reload configuration of haproxy

Beacon contains three modules:
* Beacond: a daemon used to run or stop docker container according to the commands received from etcd, and then to send register or unregister info back to etcd.
* Beaconctl: a command line tool to send beacon commands to etcd. It can also start or stop docker container directly under local mode (--local).
* [Confd](https://github.com/kelseyhightower/confd): a daemon used to reload configuration of haproxy according to etcd.

Yes, beacon rely on etcd.