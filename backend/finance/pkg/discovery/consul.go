package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"

	capi "github.com/hashicorp/consul/api"
)

type ConsulRegistry struct {
	client *capi.Client
}

func NewConsulRegistry(addr string) (Registry, error) {
	config := capi.DefaultConfig()
	config.Address = addr
	client, err := capi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ConsulRegistry{client: client}, nil
}

func (r *ConsulRegistry) Register(ctx context.Context, serviceHostPort string, serviceHealthCheckURL string, serviceName string, instanceID string, tags []string) error {
	host, portStr, err := net.SplitHostPort(serviceHostPort)
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}
	return r.client.Agent().ServiceRegister(&capi.AgentServiceRegistration{
		Address: host,
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Check: &capi.AgentServiceCheck{
			HTTP:     serviceHealthCheckURL,
			Interval: "15s",
			Timeout:  "2s",
		},
		Tags: tags,
	})
}

func (r *ConsulRegistry) Deregister(ctx context.Context, _ string, instanceID string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}

func (r *ConsulRegistry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, ErrNotFound
	}
	var res []string
	for _, e := range entries {
		res = append(res, fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port))
	}
	return res, nil
}

func (r *ConsulRegistry) ServiceAddress(ctx context.Context, serviceName string) (string, error) {
	addrs, err := r.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return "", err
	}
	return addrs[rand.Intn(len(addrs))], nil
}
