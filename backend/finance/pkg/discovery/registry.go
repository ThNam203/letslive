package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	Register(ctx context.Context, hostPort string, serviceHealthCheckURL string, serviceName string, instanceId string, tags []string) error
	Deregister(ctx context.Context, serviceName string, instanceId string) error
	ServiceAddresses(ctx context.Context, serviceName string) ([]string, error)
	ServiceAddress(ctx context.Context, serviceName string) (string, error)
}

var ErrNotFound = errors.New("no service addresses found")

func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
