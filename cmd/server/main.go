package main

import (
	"context"
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lk2023060901/go-next-erp/internal/conf"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name 应用名称
	Name string = "go-next-erp"
	// Version 应用版本
	Version string = "v1.0.0"
	// flagconf 配置文件路径
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs/config.yaml", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := conf.Load(flagconf)
	if err != nil {
		panic(err)
	}

	// 初始化日志
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	// 依赖注入
	app, cleanup, err := wireApp(context.Background(), cfg, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 启动应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}
