package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	version = getVersion()

	app = kingpin.New("zookeeper_exporter", "A zookeeper metrics exporter for prometheus, with zk_version and leaderServes=no support, with optional consul registration baked in.")

	bindHostPort = app.Flag(
		"web.listen-address",
		"Address on which to expose metrics",
	).Default("127.0.0.1:9898").String()

	zkHostString = app.Flag(
		"zk.hosts",
		"list of ip:port of ZK hosts, comma separated",
	).Required().String()

	pollInterval = app.Flag(
		"zk.poll-interval",
		"How often to poll the ZK servers",
	).Default("30").Int()

	zkTimeout = app.Flag(
		"zk.connect-timeout",
		"Timeout value for opening socket to ZK (s)",
	).Default("4").Int()

	zkRWDeadLine = app.Flag(
		"zk.connect-deadline",
		"Connection deadline for read & write operations (s)",
	).Default("3").Float()

	metricsNamespace = app.Flag(
		"metrics.namespace",
		"string to prepend to all metric names",
	).Default("zookeeper__").String()

	consulName = app.Flag(
		"consul.service-name",
		"If defined, register zookeeper_exporter with local consul agent",
	).Default("").String()

	consulTags = app.Flag(
		"consul.service-tags",
		"Comma separated list of tags for consul service",
	).Default("scrapeme").String()

	consulTTL = app.Flag(
		"consul.service-ttl",
		"consul service TTL - consul will mark service unhealthy if zookeeper_exporter is down for this long (s). Consul will also unregister the service entirely after this service has been unhealthy for this long * 10",
	).Default("60").Int()

	log = logrus.New()
)

func setup() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	app.Version(version)
	app.HelpFlag.Short('h')
	if _, err := app.Parse(os.Args[1:]); err != nil {
		log.Fatal("Couldn't parse command line args")
	}

	// Register w/ consul if *consulName defined on cmd line
	if *consulName != "" {
		if err := registerWithConsulAgent(*consulName, *consulTags, *bindHostPort, *consulTTL); err != nil {
			log.Fatalf("failed to register with consul: %s", err)
		}
	}
}

func getVersion() string {
	b, err := ioutil.ReadFile("VERSION")
	if err != nil {
		log.Errorf("can't read from VERSION file: %s", err)
		return "0.0.0"
	}
	return strings.TrimSpace(string(b))
}

func main() {
	setup()
	zkHosts := strings.Split(*zkHostString, ",")
	if *zkHostString == "" || len(zkHosts) < 1 {
		log.Fatal("Need to define zookeeper servers to monitor")
	}

	log.Printf("Starting zookeeper_exporter v%v", version)
	log.Printf("Listening on http://%v", *bindHostPort)
	log.Printf("Polling %v zookeeper servers every %ds", len(zkHosts), *pollInterval)
	log.Debugf("ZK Servers: %q", zkHosts)

	// convert int to time.Duration
	intervalDuration := time.Duration(*pollInterval) * time.Second

	// Create new metrics interface
	metrics := newMetrics()

	// Start one poller per server
	for _, ipport := range zkHosts {
		if !strings.Contains(ipport, ":") {
			log.Fatalf("zookeeper host \"%s\" is not ip:port format", ipport)
		}
		p := newPoller(intervalDuration, *metrics, *newZKServer(ipport))
		go p.pollForMetrics()
	}

	// Start http handler & server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         *bindHostPort,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
