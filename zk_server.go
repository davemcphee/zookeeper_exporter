package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	monitorCMD = "mntr"
	okCMD      = "ruok"
	enviCMD    = "envi"
)

// zkServer object
type zkServer struct {
	ipPort string
}

// zkServer constructor
func newZKServer(ipPort string) *zkServer {
	return &zkServer{ipPort: ipPort}
}

// zkServer.getStats() - runs mntr and ruok commands
func (zk *zkServer) getStats() (map[string]string, error) {
	stats, err := zk.getMNTR()
	if err != nil {
		return stats, err
	}

	isOK, err := zk.getOKStatus()
	if err != nil {
		return stats, err
	}

	stats[zkOK] = isOK
	return stats, nil
}

func (zk *zkServer) getMNTR() (map[string]string, error) {
	stats := make(map[string]string)

	byts, err := zk.sendCommand(monitorCMD)
	if err != nil {
		return stats, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(byts))
	for scanner.Scan() {
		splits := strings.Split(scanner.Text(), "\t")
		if splits[0] == "This ZooKeeper instance is not currently serving requests" {
			log.Warnf("[%v] is up but not currently serving requests", zk.ipPort)
			return stats, nil
		}

		if len(splits) != 2 {
			log.Warningf("[%v] expected key:value, got this instead: %v", zk.ipPort, splits)
			continue
		}
		stats[splits[0]] = splits[1]
	}
	return stats, nil
}

func (zk *zkServer) getOKStatus() (string, error) {
	byts, err := zk.sendCommand(okCMD)
	return string(byts), err
}

func (zk *zkServer) sendCommand(cmd string) ([]byte, error) {
	conn, err := net.Dial("tcp", zk.ipPort)
	if err != nil {
		// log.Warnf("[%v] failed to dial zkServer: %v", zk.ipPort, err)
		return []byte{}, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Errorf("[%v] failed to close connection: %v", zk.ipPort, err)
		}
	}()

	// ensure these socket fail fast if ZK having problems
	deadline := 3 * time.Second

	if err := conn.SetReadDeadline(time.Now().Add(deadline)); err != nil {
		log.Errorf("[%v] failed to set Read Deadline on conn: %v", zk.ipPort, err)
	}
	if err := conn.SetWriteDeadline(time.Now().Add(deadline)); err != nil {
		log.Errorf("[%v] failed to set Write Deadline on conn: %v", zk.ipPort, err)
	}

	_, err = fmt.Fprintf(conn, fmt.Sprintf("%s\n", cmd))
	if err != nil {
		log.Errorf("[%v] failed to close connection: %v", zk.ipPort, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, conn)
	if err != nil {
		log.Errorf("[%v] failed to read from zk: %v", zk.ipPort, err)
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
