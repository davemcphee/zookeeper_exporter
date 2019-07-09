[![Build Status](https://travis-ci.org/davemcphee/zookeeper_exporter.svg?branch=master)](https://travis-ci.org/davemcphee/zookeeper_exporter)
# zookeeper_exporter
A simple zookeeper exporter for prometheus. Grabs the `mntr` output, and converts each line to a prometheus gauge,  
including the `zk_version`.  
  
Unlike other zookeeper exporter, this one understands that the `leaderServes=no` option without barfing, and is able to  
export the `zk_version` metric as a label, instead of skipping it.   
  
This exporter manages sockets to zookeeper servers very carefully, setting timeouts and read / write deadlines to  
prevent hanging the stats poller when servers are slow to respond or unresponsive.   
  
`zk.poll-interval` takes into account how long a poll took, so if you set your interval to 30s, it will poll every 30s  
*exactly*, and not every `(30s + the time it took to poll the server)`.

`-metrics.namespace="zookeeper__"` prepends this string to all exported metric names.

## consul registration
If the flag `--consul.service-name` is set, this exporter will attempt to register itself with the local consul agent.

`--consul.service-name="zookeeper_exporter` defines the service name in consul.

`--consul.service-tags="list,of,tags"` is an optional, comma separated list of tags to register the service with.

`--consul.service-ttl=60` is the check TTL for the service health check. This exporter will update the health check while
it's alive, but if it dies, consul will mark the service as unhealthy after this many seconds, and will unregister it
entirely after `consul.service-ttl * 10` seconds.
  
## Usage  
  
~~~  
$ zookeeper_exporter --help  
usage: zookeeper_exporter --zk.hosts=ZK.HOSTS [<flags>]

A zookeeper metrics exporter for prometheus, with zk_version and leaderServes=no support, with optional consul registration baked in.

Flags:
  -h, --help                    Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address="127.0.0.1:9898"  
                                Address on which to expose metrics
      --zk.hosts=ZK.HOSTS       list of ip:port of ZK hosts, comma separated
      --zk.poll-interval=30     How often to poll the ZK servers
      --zk.connect-timeout=4    Timeout value for opening socket to ZK (s)
      --zk.connect-deadline=3   Connection deadline for read & write operations (s)
      --metrics.namespace="zookeeper__"  
                                string to prepend to all metric names
      --consul.service-name=""  If defined, register zookeeper_exporter with local consul agent
      --consul.service-tags="scrapeme"  
                                Comma separated list of tags for consul service
      --consul.service-ttl=60   consul service TTL - consul will mark service unhealthy if zookeeper_exporter is down for this long (s).
                                Consul will also unregister the service entirely after this service has been unhealthy for this long * 10
      --version                 Show application version.

$ zookeeper_exporter --zk.hosts=10.0.0.9:2181,10.0.0.10:2181  
~~~  
  
## Install  
If you want to build from source, you know how. If you prefer, each tagged version release is available as a  
pre-compiled binary for linux.x86_64 and darwin on github   
under [releases](https://github.com/davemcphee/zookeeper_exporter/releases).  
  
## ToDo  
 
 - Better tests - need to mock a ZK server; argh.
 