package grpcclient

import "google.golang.org/grpc"

type ServiceAddress interface {
	GetAddress() string
	GetPort() int
}

type requestDefinition struct {
	address string
	port    int
	opts    []grpc.DialOption
}

func (r *requestDefinition) GetAddress() string {
	return r.address
}

func (r *requestDefinition) GetPort() int {
	return r.port
}

func (r *requestDefinition) GetOpts() []grpc.DialOption {
	return r.opts
}

type ClientI interface {
	WithDialOptions(opts ...grpc.DialOption) ClientI
	GetHealthConn() (*grpc.ClientConn, error)
}

type clientDefinition struct {
	ClientI
	requestDefinition
}
