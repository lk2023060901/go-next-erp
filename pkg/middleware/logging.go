package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Logging 日志中间件
func Logging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			startTime := time.Now()

			// 从传输层获取信息
			var operation, path, method string
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
				if header := tr.RequestHeader(); header != nil {
					path = header.Get("path")
					method = header.Get("method")
				}
			}

			// 执行处理器
			reply, err := handler(ctx, req)

			// 计算耗时
			duration := time.Since(startTime)

			// 记录日志
			if err != nil {
				_ = log.WithContext(ctx, logger).Log(
					log.LevelError,
					"operation", operation,
					"path", path,
					"method", method,
					"duration", duration.Milliseconds(),
					"error", err.Error(),
				)
			} else {
				_ = log.WithContext(ctx, logger).Log(
					log.LevelInfo,
					"operation", operation,
					"path", path,
					"method", method,
					"duration", duration.Milliseconds(),
				)
			}

			return reply, err
		}
	}
}
