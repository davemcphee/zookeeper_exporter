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

	app = kingpin.New("zookeeper_exporter", "A zookeeper metrics exporter for prometheus, with zk_version and leaderServes=no support.")

	bindHostPort = app.Flag(
		"web.listen-address",
		"Address on which to expose metrics",
	).Default("0.0.0.0:9898").String()

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
		"Timeout value for connecting to ZK",
	).Default("5").Int()

	zkRWDeadLine = app.Flag(
		"zk.connect-rw-deadline",
		"Socket deadline for read & write operations",
	).Default("5").Int()

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
