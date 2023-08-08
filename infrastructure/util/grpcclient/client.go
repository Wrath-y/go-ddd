package grpcclient

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewClient(cd ServiceAddress) ClientI {
	return &clientDefinition{
		requestDefinition: requestDefinition{
			address: cd.GetAddress(),
			port:    cd.GetPort(),
		},
	}
}

func (cd *clientDefinition) GetHealthConn() (*grpc.ClientConn, error) {
	if cd.opts == nil {
		cd.opts = make([]grpc.DialOption, 0, 1)
	}
	cd.opts = append(cd.opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", cd.GetAddress(), cd.GetPort()), cd.GetOpts()...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (cd *clientDefinition) WithDialOptions(opts ...grpc.DialOption) ClientI {
	cd.opts = opts
	return cd
}
