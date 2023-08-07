package grpc

import (
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/spf13/viper"
	"go-ddd/infrastructure/util/consul"
	defgrpc "go-ddd/infrastructure/util/def/grpc"
	"go-ddd/infrastructure/util/logging"
	"go-ddd/interfaces/proto"
	"go-ddd/interfaces/proto/facade"
	middleware2 "go-ddd/launch/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func RunGrpc() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		// options
		grpc.InitialWindowSize(defgrpc.InitialWindowSize),
		grpc.InitialConnWindowSize(defgrpc.InitialConnWindowSize),
		grpc.MaxSendMsgSize(defgrpc.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(defgrpc.MaxRecvMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    defgrpc.KeepAliveTime,
			Timeout: defgrpc.KeepAliveTimeout,
		}),
		// middlewares
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			middleware2.UnaryRecover(),
			middleware2.UnaryContext(),
			middleware2.UnaryLogger(),
		)))

	// 在gRPC服务器注册我们的服务
	grpc_health_v1.RegisterHealthServer(grpcServer, facade.NewHealthCheckService())
	proto.RegisterArticleServer(grpcServer, &facade.Article{})

	go func() {
		//用服务器 Serve() 方法以及我们的端口信息区实现阻塞等待，直到进程被杀死或者 Stop() 被调用
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatalf("grpcServer.Serve err: %v", err)
		}
		log.Println("Shutdown grpcServer.Serve")
	}()

	registerService()

	logging.New().Info("Has Start", "", viper.GetString("app.env"))

	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	globalDestroy()

	grpcServer.GracefulStop()
}

func registerService() {
	serviceInstance := consul.NewServiceInstance(strconv.FormatInt(time.Now().Unix(), 10), "article", "grpc", "docker.for.mac.host.internal", 8080, false, map[string]string{})
	if err := consul.Client.Register(serviceInstance); err != nil {
		log.Fatalf("consul.Register err: %v", err)
	}
}

func globalDestroy() {
	if err := consul.Client.Deregister(); err != nil {
		logging.New().ErrorL("consul deregister failed", "", err.Error())
	}
}
