package consul

import (
	"github.com/spf13/viper"
	"log"
)

type RegistryI interface {
	Register(serviceInstance InstanceI) error
	Deregister() error

	// GetInstances Gets all ServiceInstances associated with a particular serviceId.
	// param serviceName The serviceId to query.
	// return a List of instance.
	GetInstances(serviceName string, tags ...string) ([]InstanceI, error)

	// GetServices return All known service IDs.
	GetServices() ([]string, error)

	// GetHealthRandomInstance Get a random instance
	GetHealthRandomInstance(serviceName string, tags ...string) (InstanceI, error)
}

type InstanceI interface {
	// GetInstanceId return The unique instance ID as registered.
	GetInstanceId() string

	// GetServiceName return The service name as registered.
	GetServiceName() string

	// GetSchema return The schema of the registered service instance.
	// http/tcp/grpc.
	GetSchema() string

	// GetAddress return The hostname of the registered service instance.
	GetAddress() string

	// GetPort return The port of the registered service instance.
	GetPort() int

	// IsSecure return Whether the port of the registered service instance uses HTTPS.
	IsSecure() bool

	// GetMetadata return The key / value pair metadata associated with the service instance.
	GetMetadata() map[string]string
}

var Client RegistryI

func Setup() {
	Client = newConsulClient("default")
}

func newConsulClient(store string) RegistryI {
	cfg := viper.Sub("consul." + store)
	if cfg == nil {
		log.Fatal("consul config is nil: ", store)
	}
	address := cfg.GetString("address")
	if address == "" {
		log.Fatal("consul address is nil")
	}

	var err error
	Client, err = newConsulServiceRegistry(address, "")
	if err != nil {
		log.Fatal("consul client request failed: ", err)
	}

	return Client
}
