package facade

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthCheckService implements grpc_health_v1.HealthServer
type HealthCheckService struct{}

func NewHealthCheckService() *HealthCheckService {
	return &HealthCheckService{}
}

// Check 实现健康检查接口，这里直接返回健康状态，这里也可以有更复杂的健康检查策略，比如根据服务器负载来返回
func (h *HealthCheckService) Check(_ context.Context, _ *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthCheckService) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return nil
}
