# zookeeper_exporter

A simple zookeeper exporter for prometheus. Grabs the `mntr` output, and converts each line to a prometheus gauge, including the `zk_version`.

## Usage

~~~
$ zookeeper_exporter --help
usage: zookeeper_exporter --zk.hosts=ZK.HOSTS [<flags>]

Flags:
  -h, --help                  Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address="0.0.0.0:9898"  
                              Address on which to expose metrics
      --zk.hosts=ZK.HOSTS     list of ip:port of ZK hosts, comma separated
      --zk.poll-interval=30   How often to poll the ZK servers
      --zk.connect-timeout=5  Timeout value for connecting to ZK
      --zk.connect-rw-deadline=5  
                              Socket deadline for read & write operations
      --version               Show application version.
~~~

~~~
zookeeper_exporter --zk.hosts=10.0.0.9:2181,10.0.0.10:2181
~~~

## Install
If you want to build from source, you know how. If you prefer, each tagged version release is available as a pre-compiled binary for linux.x86_64 and darwin on github under [releases](https://github.com/davemcphee/zookeeper_exporter/releases).

## ToDo

 - Add consul service registration as default
 - Better test etc