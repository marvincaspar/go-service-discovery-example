package main

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

//Client provides an interface for getting data out of Consul
type Client interface {
	// Get a Service from consul
	Service(string, string) ([]string, error)
	// Register a service with local agent
	Register(string, int) error
	// Deregister a service with local agent
	DeRegister(string) error
}

type client struct {
	consul *consul.Client
}

//NewConsul returns a Client interface for given consul address
func NewConsulClient(addr string) (*client, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	c, err := consul.NewClient(config)
	if err != nil {
		return &client{}, err
	}
	return &client{consul: c}, nil
}

// Register a service with consul local agent - note the tags to define path-prefix is to be used.
func (c *client) Register(id, name, host string, port int, path, health string) error {

	reg := &consul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Port:    port,
		Address: host,
		Check: &consul.AgentServiceCheck{
			CheckID:       id,
			Name:          "HTTP API health",
			HTTP:          health,
			TLSSkipVerify: true,
			Method:        "GET",
			Interval:      "10s",
			Timeout:       "1s",
		},
		Tags: []string{
			"traefik.enable=true",
			"traefik.tags=api",
			"traefik.tags=external",
			"traefik.backend=" + name,
		},
	}
	return c.consul.Agent().ServiceRegister(reg)
}

// DeRegister a service with consul local agent
func (c *client) DeRegister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service return a service
func (c *client) Service(serviceName, tag string) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	passingOnly := true
	addrs, meta, err := c.consul.Health().Service(serviceName, tag, passingOnly, nil)
	if len(addrs) == 0 && err == nil {
		return nil, nil, fmt.Errorf("service ( %s ) was not found", serviceName)
	}
	if err != nil {
		return nil, nil, err
	}
	return addrs, meta, nil
}

func (c *client) ServiceAddress(serviceName string) (string, error) {
	srvc, _, err := c.Service(serviceName, "")
	if err != nil {
		return "", err
	}

	address := srvc[0].Service.Address
	port := srvc[0].Service.Port
	return fmt.Sprintf("http://%s:%v", address, port), nil
}
