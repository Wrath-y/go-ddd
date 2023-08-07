package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	grpcCtx "go-ddd/infrastructure/common/context"
	"go-ddd/infrastructure/common/errcode"
	"go-ddd/infrastructure/util/logging"
	"go-ddd/infrastructure/util/util/random"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

const bodyLimitKB = 5000

func UnaryLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		start := time.Now()

		c := grpcCtx.GetContext(ctx)

		xRequestID := random.UUID()

		logger := logging.New()
		logger.SetRequestID(xRequestID)
		logger.Setv1(info.FullMethod)
		c.Logger = logger

		raw, _ := json.Marshal(req)

		rawKB := len(raw) / 1024 // => to KB
		if rawKB > bodyLimitKB {
			c.Info("接口请求与响应", string(raw[:1024]), nil)
			return nil, errcode.BlogBodyTooLarge.WithDetail(fmt.Sprintf("消息限制%dKB, 本消息%dKB", bodyLimitKB, rawKB))
		}

		defer func() {
			out, _ := json.Marshal(resp)
			md, _ := metadata.FromIncomingContext(c)

			request := map[string]any{
				"path":     info.FullMethod,
				"metadata": md,
				"body":     string(raw),
			}
			c.Info("接口请求与响应", request, string(out), logging.AttrOption{StartTime: &start})
		}()

		return handler(c, req)
	}
}
