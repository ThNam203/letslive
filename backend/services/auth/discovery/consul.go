package discovery

import (
	"context"
	"fmt"
	"net"
	"sen1or/lets-live/auth/config"
	"sen1or/lets-live/auth/logger"
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

	return &ConsulRegistry{
		client: client,
	}, nil
}

func (r *ConsulRegistry) Register(ctx context.Context, hostPort string, serviceName string, instanceID string) error {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	logger.Infof("http://%s/v1/health", config.MyConfig.Service.Address)

	return r.client.Agent().ServiceRegister(&capi.AgentServiceRegistration{
		Address: host,
		ID:      instanceID,
		Name:    serviceName,
		Port:    port,
		Check: &capi.AgentServiceCheck{
			//HTTP:     fmt.Sprintf("http://%s:%s/v1/health", config.MyConfig.Service.Host, config.MyConfig.Service.Port),
			HTTP:     fmt.Sprintf("http://%s/v1/health", config.MyConfig.Service.Address),
			Interval: "10s",
			Timeout:  "1s",
		},
	})
}

func (r *ConsulRegistry) Deregister(ctx context.Context, _ string, instanceID string) error {
	return r.client.Agent().ServiceDeregister(instanceID)
}

func (r *ConsulRegistry) ServiceAddresses(ctx context.Context, serviceName string) ([]string, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	} else if len(entries) == 0 {
		return nil, ErrNotFound
	}

	var res []string
	for _, e := range entries {
		res = append(res, fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port))
	}

	return res, nil
}
