package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type serverState float64

const (
	// hard codes metric names from mntr command output
	zkAvgLatency              = "zk_avg_latency"
	zkMinLatency              = "zk_min_latency"
	zkMaxLatency              = "zk_max_latency"
	zkPacketsReceived         = "zk_packets_received"
	zkPacketsSent             = "zk_packets_sent"
	zkNumAliveConnections     = "zk_num_alive_connections"
	zkOutstandingRequests     = "zk_outstanding_requests"
	zkZnodeCount              = "zk_znode_count"
	zkWatchCount              = "zk_watch_count"
	zkEphemeralsCount         = "zk_ephemerals_count"
	zkApproximateDataSize     = "zk_approximate_data_size"
	zkOpenFileDescriptorCount = "zk_open_file_descriptor_count"
	zkMaxFileDescriptorCount  = "zk_max_file_descriptor_count"
	zkFollowers               = "zk_followers"
	zkSyncedFollowers         = "zk_synced_followers"
	zkPendingSyncs            = "zk_pending_syncs"
	zkServerState             = "zk_server_state"
	zkFsyncThresholdExceeded  = "zk_fsync_threshold_exceed_count"
	zkVersion                 = "zk_version"

	zkOK = "zk_ok"

	// server states
	unknown    serverState = -1
	follower   serverState = 1
	leader     serverState = 2
	standalone serverState = 3

	// metric namepspace - prepended to all metric names
	namespace = "zookeeper__"
)

func getState(s string) serverState {
	switch s {
	case "follower":
		return follower
	case "leader":
		return leader
	case "standalone":
		return standalone
	default:
		return unknown
	}
}

func prepend_namespace(rawMetricName string) string {
	return namespace + rawMetricName
}

// Creates a map of all known metrics exposed by zookeeper's mntr command
// literal metric name maps to a prometheus Gauge with label zk_instance set to zk's address
func initMetrics() map[string]*prometheus.GaugeVec {

	allMetrics := make(map[string]*prometheus.GaugeVec)

	allMetrics[zkAvgLatency] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkAvgLatency),
		Help: "Average Latency for ZooKeeper network requests",
	}, []string{"zk_instance"})

	allMetrics[zkMinLatency] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkMinLatency),
		Help: "Minimum latency for Zookeeper network requests.",
	}, []string{"zk_instance"})

	allMetrics[zkMaxLatency] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkMaxLatency),
		Help: "Maximum latency for ZooKeeper network requests",
	}, []string{"zk_instance"})

	allMetrics[zkPacketsReceived] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkPacketsReceived),
		Help: "Number of network packets received by the ZooKeeper instance.",
	}, []string{"zk_instance"})

	allMetrics[zkPacketsSent] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkPacketsSent),
		Help: "Number of network packets sent by the ZooKeeper instance.",
	}, []string{"zk_instance"})

	allMetrics[zkNumAliveConnections] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkNumAliveConnections),
		Help: "Number of currently alive connections to the ZooKeeper instance.",
	}, []string{"zk_instance"})

	allMetrics[zkOutstandingRequests] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkOutstandingRequests),
		Help: "Number of requests currently waiting in the queue.",
	}, []string{"zk_instance"})

	allMetrics[zkZnodeCount] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkZnodeCount),
		Help: "Znode count",
	}, []string{"zk_instance"})

	allMetrics[zkWatchCount] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkWatchCount),
		Help: "Watch count",
	}, []string{"zk_instance"})

	allMetrics[zkEphemeralsCount] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkEphemeralsCount),
		Help: "Ephemerals Count",
	}, []string{"zk_instance"})

	allMetrics[zkApproximateDataSize] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkApproximateDataSize),
		Help: "Approximate data size",
	}, []string{"zk_instance"})

	allMetrics[zkOpenFileDescriptorCount] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkOpenFileDescriptorCount),
		Help: "Number of currently open file descriptors",
	}, []string{"zk_instance"})

	allMetrics[zkMaxFileDescriptorCount] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkMaxFileDescriptorCount),
		Help: "Maximum number of open file descriptors",
	}, []string{"zk_instance"})

	allMetrics[zkServerState] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkServerState),
		Help: "Current state of the zk instance: 1 = follower, 2 = leader, 3 = standalone, -1 if unknown",
	}, []string{"zk_instance"})

	allMetrics[zkFollowers] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkFollowers),
		Help: "Leader only: number of followers.",
	}, []string{"zk_instance"})

	allMetrics[zkSyncedFollowers] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkSyncedFollowers),
		Help: "Leader only: number of followers currently in sync",
	}, []string{"zk_instance"})

	allMetrics[zkPendingSyncs] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkPendingSyncs),
		Help: "Current number of pending syncs",
	}, []string{"zk_instance"})

	allMetrics[zkOK] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkOK),
		Help: "Is ZooKeeper currently OK",
	}, []string{"zk_instance"})

	allMetrics[zkFsyncThresholdExceeded] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkFsyncThresholdExceeded),
		Help: "Number of times File sync exceeded fsyncWarningThresholdMS",
	}, []string{"zk_instance"})

	allMetrics[zkVersion] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prepend_namespace(zkVersion),
		Help: "Zookeeper version",
	}, []string{"zk_instance", "zk_version"})

	// Register all gauges with prometheus registry so they're exposed by promhttp Handler
	for _, metric := range allMetrics {
		prometheus.MustRegister(metric)
	}

	return allMetrics
}
