package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"strings"
	"time"
)

const version = "0.1"

var (
	bindHostPort = kingpin.Flag(
		"web.listen-address",
		"Address on which to expose metrics",
	).Default("0.0.0.0:9898").String()

	zkHostString = kingpin.Flag(
		"zk.hosts",
		"list of ip:port of ZK hosts, comma separated",
	).Required().String()

	pollInterval = kingpin.Flag(
		"zk.poll-interval",
		"How often to poll the ZK servers",
	).Default("30").Int()

	zkTimeout = kingpin.Flag(
		"zk.connect-timeout",
		"Timeout value for connecting to ZK",
	).Default("5").Int()

	zkRWDeadLine = kingpin.Flag(
		"zk.connect-rw-deadline",
		"Socket deadline for read & write operations",
	).Default("5").Int()

	log = logrus.New()
)

func setup() {
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.DebugLevel)

	kingpin.Version(version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {
	setup()
	zkHosts := strings.Split(*zkHostString, ",")
	if *zkHostString == "" || len(zkHosts) < 0 {
		log.Fatal("Need to define zookeeper servers to monitor")
	}

	log.Printf("Starting zookeeper_exporter v%v", version)
	log.Printf("Listening on http://%v", *bindHostPort)
	log.Printf("Polling %v zookeeper servers every %ds", len(zkHosts), *pollInterval)
	log.Debugf("ZK Servers: %q", zkHosts)

	// convert int to time.Duration
	intervalDuration := time.Duration(*pollInterval) * time.Second

	// Create new metrics interface
	metrics := initMetrics()

	// Start an export thread per server
	for _, ipport := range zkHosts {
		p := newPoller(intervalDuration, metrics, *newZKServer(ipport))
		go p.pollForMetrics()
	}

	// Start http handler & server
	// http.HandleFunc("/", metricsRequestHandler)
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
