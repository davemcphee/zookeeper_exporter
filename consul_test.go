package main

import (
	"encoding/json"
	consul "github.com/hashicorp/consul/api"
	"gotest.tools/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type reqStruct struct {
	Name    string
	Tags    []string
	Port    int
	Address string
	Check   struct {
		TTL                            string
		DeregisterCriticalServiceAfter string
	}
}

func TestConsulRegistration(t *testing.T) {
	t.Run("Good registration", func(t *testing.T) {
		// Set up some vars to compare want and got
		cServName := "test-service"
		cTags := []string{"scrapeme", "foo", "bar"}
		cAddress := "127.0.0.2:12345"
		cTTL := 5

		// this will hold PUT json body
		var r reqStruct
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			decoder := json.NewDecoder(req.Body)
			err := decoder.Decode(&r)
			if err != nil {
				t.Errorf("failed to decode json from consul registration request: %s", err)
			}
			assert.Assert(t, req.Method == "PUT", "registration request method should be PUT")
			_, _ = rw.Write([]byte("OK"))

		}))
		defer server.Close()

		consulConfig := consul.DefaultConfig()
		consulConfig.Address = strings.TrimPrefix(server.URL, "http://")

		// Create a consul client and set Address of consul server to our server
		consulClient, err := consul.NewClient(consulConfig)
		if err != nil {
			t.Errorf("failed to create consul client: %s", err)
		}

		// Create a ServiceRegistrar with our consulClient
		sr := ServiceRegistrar{Name: cServName, Addr: cAddress, Tags: cTags, TTLSeconds: cTTL, ConsulAgent: consulClient.Agent()}

		// Attempt to register service, but hitting our test server
		err = sr.RegisterService()
		if err != nil {
			t.Errorf("failed to register service: %s", err)
		}

		// Check if we registered correctly
		wantIP := strings.Split(cAddress, ":")[0]
		wantPortS := strings.Split(cAddress, ":")[1]
		wantPortI, err := strconv.Atoi(wantPortS)
		if err != nil {
			t.Errorf("failed to convert recevied port to integer: %s", err)
		}
		wantTTL := strconv.Itoa(cTTL)
		wantTTL = wantTTL + "s"

		assert.Equal(t, r.Name, cServName)
		assert.DeepEqual(t, r.Tags, cTags)
		assert.Equal(t, r.Address, wantIP)
		assert.Equal(t, r.Port, wantPortI)
		assert.Equal(t, r.Check.TTL, wantTTL)
	})

	t.Run("Bad Registration - bad IP", func(t *testing.T) {
		// Setup new http server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Assert(t, req.Method == "PUT", "registration request method should be PUT")
			http.Error(rw, "no", 501)

		}))
		defer server.Close()

		// Create consul config, replace address with our server's address
		consulConfig := consul.DefaultConfig()
		consulConfig.Address = strings.TrimPrefix(server.URL, "http://")

		consulClient, err := consul.NewClient(consulConfig)
		if err != nil {
			t.Errorf("failed to create consul client: %s", err)
		}

		// Create a ServiceRegistrar with our consulClient
		sr := ServiceRegistrar{Name: "foo", Addr: "127.0.0.0.0.0.1:12345", Tags: []string{"scrapeme"}, TTLSeconds: 5, ConsulAgent: consulClient.Agent()}

		// Attempt to register service, but hitting our test server
		err = sr.RegisterService()
		if err == nil {
			t.Errorf("attempt to register w/ garbage IP address did not fail")
		}
	})

	t.Run("Bad Registration - bad Port", func(t *testing.T) {
		// Setup new http server
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Assert(t, req.Method == "PUT", "registration request method should be PUT")
			http.Error(rw, "no", 501)

		}))
		defer server.Close()

		// Create consul config, replace address with our server's address
		consulConfig := consul.DefaultConfig()
		consulConfig.Address = strings.TrimPrefix(server.URL, "http://")

		consulClient, err := consul.NewClient(consulConfig)
		if err != nil {
			t.Errorf("failed to create consul client: %s", err)
		}

		// Create a ServiceRegistrar with our consulClient
		sr := ServiceRegistrar{Name: "foo", Addr: "127.0.0.1:bananas", Tags: []string{"scrapeme"}, TTLSeconds: 5, ConsulAgent: consulClient.Agent()}

		// Attempt to register service, but hitting our test server
		err = sr.RegisterService()
		if err == nil {
			t.Errorf("attempt to register w/ garbage IP address did not fail")
		}
	})
}
