package main

import (
	"strconv"
	"strings"
	"time"
)

type zkPoller struct {
	interval time.Duration
	metrics  zkMetrics
	zkServer zkServer
}

func newPoller(interval time.Duration, metrics zkMetrics, zkServer zkServer) *zkPoller {
	return &zkPoller{
		interval: interval,
		metrics:  metrics,
		zkServer: zkServer,
	}
}

func (p *zkPoller) pollForMetrics() {
	// Initialise to counter to 0
	p.metrics.pollingFailureCounter.WithLabelValues(p.zkServer.ipPort).Add(0)
	for {
		expirationTime := time.Now().Add(p.interval)
		m, err := p.zkServer.getStats()
		if err != nil {
			log.Errorf("[%v] failed to get stats: %v", p.zkServer.ipPort, err)
			p.metrics.pollingFailureCounter.WithLabelValues(p.zkServer.ipPort).Inc()
		}

		p.refreshMetrics(m)

		// Instead of sleeping for a further p.interval time, calculate for long we've already spent polling, and sleep
		// the difference
		<-time.After(expirationTime.Sub(time.Now()))
	}
}

func (p *zkPoller) refreshMetrics(updated map[string]string) {
	for name, value := range updated {
		metric, ok := p.metrics.gauges[name]

		if !ok {
			log.Errorf("[%v] stat=%v not defined in metrics.go\n", p.zkServer.ipPort, name)
			continue
		}

		// zkOK is a special case
		if name == zkOK {
			switch value {
			case "imok":
				metric.WithLabelValues(p.zkServer.ipPort).Set(1)
			default:
				metric.WithLabelValues(p.zkServer.ipPort).Set(0)
			}
			continue
		}

		// zk_version is also a special case
		if name == zkVersion {
			versionSplits := strings.Split(value, "-")
			metric.WithLabelValues(p.zkServer.ipPort, versionSplits[0]).Set(1)
			continue
		}

		if name == zkServerState {
			state := getState(value)
			metric.WithLabelValues(p.zkServer.ipPort).Set(float64(state))
			continue
		}

		// all other metrics get converted to float and used as is
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Errorf("[%v] failed to convert string to float, value=%v", p.zkServer.ipPort, value)
		}

		metric.WithLabelValues(p.zkServer.ipPort).Set(f)
	}
}
