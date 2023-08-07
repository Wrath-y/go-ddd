package middleware

import (
	"context"
	"fmt"
	"go-ddd/infrastructure/util/logging"
	"google.golang.org/grpc"
	"os"
	"runtime/debug"
)

// UnaryRecover 捕捉简单代码致命错误
func UnaryRecover() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		defer func() {
			if err := recover(); err != nil {
				var errMsg string
				switch err := err.(type) {
				case error:
					errMsg = string(debug.Stack())
				default:
					errMsg = fmt.Sprintf("%v", err)
				}
				hostname, _ := os.Hostname()
				logging.New().Info(fmt.Sprintf("服务异常: %s", hostname), req, errMsg)
			}
		}()

		return handler(ctx, req)
	}
}
