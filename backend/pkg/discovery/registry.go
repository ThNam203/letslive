package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type Registry interface {
	// Register creates a service instance record in the registry.
	// hostPort: consul host and port (ex: consul:8500)
	Register(ctx context.Context, hostPort string, serviceHealthCheckURL string, serviceName string, instanceId string, tags []string) error
	// Deregister removes a service insttance record from the registry.
	Deregister(ctx context.Context, serviceName string, instanceId string) error
	// ServiceAddresses returns the list of addresses of active instances of the given service.
	ServiceAddresses(ctx context.Context, serviceName string) ([]string, error)
	// ServiceAddress choose one adddress from service addresses and return
	ServiceAddress(ctx context.Context, serviceName string) (string, error)
}

// ErrNotFound is returned when no service addresses are found.
var ErrNotFound = errors.New("no service addresses found")

// GenerateInstanceID generates a pseudo-unique service instance identifier, using a service name
// suffixed by dash and a random number.
func GenerateInstanceID(serviceName string) string {
	return fmt.Sprintf("%s-%d", serviceName, rand.New(rand.NewSource(time.Now().UnixNano())).Int())
}
