package consul

import (
	"math/rand"
	"strconv"
	"time"
)

// instance implements InstanceI
type instance struct {
	InstanceId  string // instanceId is the service unique id.
	ServiceName string // serviceName is the service name.
	Schema      string // schema is used to health check request type.
	Address     string // address is the service host.
	Port        int    // port is the service port.
	Secure      bool   // secure is used to health check tls.
	Metadata    map[string]string
}

func NewServiceInstance(instanceId, serviceName string, schema, address string, port int, secure bool, metadata map[string]string) *instance {
	if len(instanceId) == 0 {
		rand.New(rand.NewSource(time.Now().Unix()))
		instanceId = serviceName + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.Itoa(rand.Intn(9000)+1000)
	}

	return &instance{InstanceId: instanceId, ServiceName: serviceName, Schema: schema, Address: address, Port: port, Secure: secure, Metadata: metadata}
}

func (i instance) GetInstanceId() string {
	return i.InstanceId
}

func (i instance) GetServiceName() string {
	return i.ServiceName
}

func (i instance) GetSchema() string {
	return i.Schema
}

func (i instance) GetAddress() string {
	return i.Address
}

func (i instance) GetPort() int {
	return i.Port
}

func (i instance) IsSecure() bool {
	return i.Secure
}

func (i instance) GetMetadata() map[string]string {
	return i.Metadata
}
