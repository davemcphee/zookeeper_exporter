package main

import (
	"errors"
	consul "github.com/hashicorp/consul/api"
	"strconv"
	"strings"
	"time"
)

type ServiceRegistrar struct {
	Name        string
	Addr        string
	Tags        []string
	TTLSeconds  int
	ConsulAgent *consul.Agent
}

func NewServiceRegistrar(name, addr, consulTags string, consulTTL int) (*ServiceRegistrar, error) {
	c, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &ServiceRegistrar{
		Name:        name,
		Addr:        addr,
		Tags:        strings.Split(consulTags, ","),
		TTLSeconds:  consulTTL,
		ConsulAgent: c.Agent(),
	}, nil
}

func (sr *ServiceRegistrar) RegisterService() error {
	addrPort := strings.Split(sr.Addr, ":")
	if len(addrPort) != 2 {
		return errors.New(sr.Addr + " does not appear to be a valid address")
	}

	port, err := strconv.Atoi(addrPort[1])
	if err != nil {
		return err
	}

	// Service check definition
	serviceDef := &consul.AgentServiceRegistration{
		Name:    sr.Name,
		Tags:    sr.Tags,
		Address: addrPort[0],
		Port:    port,
		Check: &consul.AgentServiceCheck{
			TTL:                            (time.Duration(sr.TTLSeconds) * time.Second).String(),
			DeregisterCriticalServiceAfter: (time.Duration(sr.TTLSeconds) * time.Second * 10).String(),
		},
	}

	if err := sr.ConsulAgent.ServiceRegister(serviceDef); err != nil {
		return err
	}
	return nil
}

// Sends heartbeats to consul health check
func (sr *ServiceRegistrar) updateCheckTTLForever() {
	// TTL heartbeat every TTL / 2
	ticker := time.NewTicker(time.Duration(sr.TTLSeconds) * time.Second / 2)
	for range ticker.C {
		if err := sr.ConsulAgent.UpdateTTL("service:"+sr.Name, "output", "pass"); err != nil {
			log.Errorf("failed to updateTTL: %v", err)
		}
	}
}

// Deregisters us from consul
func (sr *ServiceRegistrar) deRegister() {
	if err := sr.ConsulAgent.ServiceDeregister("service:" + sr.Name); err != nil {
		log.Errorf("failed to deregister with consul: %v", err)
	}
}

// main registration func, sets up registration structs, registers, runs updateCheckTTLForever
func registerWithConsulAgent(serviceName, serviceTags, consulBindHostPort string, serviceTTL int) error {
	if serviceName != "" {
		log.Infof("attempting to register service %v with consul", serviceName)

		// Creates a new ServiceRegistrar struct
		sr, err := NewServiceRegistrar(serviceName, consulBindHostPort, serviceTags, serviceTTL)
		if err != nil {
			log.Fatalf("failed to create consul service registrar: %v", err)
		}

		// Register the service
		if err := sr.RegisterService(); err != nil {
			log.Fatalf("failed to register service with consul: %v", err)
		}
		go sr.updateCheckTTLForever()
	} else {
		log.Debugf("not registering with consul; consul.service-name flag undefined")
	}
	return nil
}
