package consul

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// consulServiceRegistry implements RegistryI
type consulServiceRegistry struct {
	serviceInstances     map[string]map[string]InstanceI
	client               *api.Client
	localServiceInstance InstanceI
}

func newConsulServiceRegistry(address string, token string) (*consulServiceRegistry, error) {
	config := api.DefaultConfig()
	config.Address = address
	config.Token = token
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulServiceRegistry{client: client}, nil
}

func (c *consulServiceRegistry) Register(si InstanceI) error {
	// 创建注册到consul的服务
	registration := new(api.AgentServiceRegistration)
	registration.ID = si.GetInstanceId()
	registration.Name = si.GetServiceName()
	registration.Port = si.GetPort()
	var tags []string
	if si.IsSecure() {
		tags = append(tags, "secure=true")
	} else {
		tags = append(tags, "secure=false")
	}
	registration.Tags = tags
	registration.Meta = si.GetMetadata()
	registration.Address = "127.0.0.1"

	// 增加consul健康检查回调函数
	check := new(api.AgentServiceCheck)

	switch si.GetSchema() {
	case "grpc":
		check.CheckID = si.GetInstanceId()
		check.Name = si.GetServiceName()
		check.GRPC = fmt.Sprintf("%s:%d", si.GetAddress(), si.GetPort())
		if si.IsSecure() {
			check.GRPCUseTLS = true
		}
		check.Interval = "1s"
	}

	check.DeregisterCriticalServiceAfter = "3s" // 故障检查失败30s后 consul自动将注册服务删除
	registration.Check = check

	// 注册服务到consul
	err := c.client.Agent().ServiceRegister(registration)
	if err != nil {
		return err
	}

	if c.serviceInstances == nil {
		c.serviceInstances = map[string]map[string]InstanceI{}
	}

	services := c.serviceInstances[si.GetServiceName()]

	if services == nil {
		services = map[string]InstanceI{}
	}

	services[si.GetInstanceId()] = si

	c.serviceInstances[si.GetServiceName()] = services

	c.localServiceInstance = si

	return nil
}

func (c *consulServiceRegistry) Deregister() error {
	if c.serviceInstances == nil {
		return errors.New("nil serviceInstances")
	}

	services := c.serviceInstances[c.localServiceInstance.GetServiceName()]

	if services == nil {
		return errors.New("nil services")
	}

	delete(services, c.localServiceInstance.GetInstanceId())

	if len(services) == 0 {
		delete(c.serviceInstances, c.localServiceInstance.GetServiceName())
	}

	if err := c.client.Agent().ServiceDeregister(c.localServiceInstance.GetInstanceId()); err != nil {
		return err
	}
	c.localServiceInstance = nil
	return nil
}

func (c consulServiceRegistry) GetInstances(serviceName string, tags ...string) ([]InstanceI, error) {
	catalogService, _, _ := c.client.Catalog().ServiceMultipleTags(serviceName, tags, nil)
	if len(catalogService) > 0 {
		result := make([]InstanceI, len(catalogService))
		for index, sever := range catalogService {
			s := instance{
				InstanceId:  sever.ServiceID,
				ServiceName: sever.ServiceName,
				Address:     sever.ServiceAddress,
				Port:        sever.ServicePort,
				Metadata:    sever.ServiceMeta,
			}
			result[index] = s
		}
		return result, nil
	}
	return nil, nil
}

func (c consulServiceRegistry) GetServices() ([]string, error) {
	services, _, _ := c.client.Catalog().Services(nil)
	result := make([]string, 0, len(services))
	for serviceName := range services {
		result = append(result, serviceName)
	}
	return result, nil
}

func (c consulServiceRegistry) GetGRPCInstanceConn(service InstanceI) (*grpc.ClientConn, error) {
	// 添加grpc.WithBlock()来等连接建立完再返回，否则会默认使用协程异步创建conn。
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", service.GetAddress(), service.GetPort()), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return conn, err
}
