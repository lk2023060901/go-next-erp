package main

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/lk2023060901/go-next-erp/internal/conf"
)

var (
	// Version 版本号
	Version = "dev"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GitCommit Git 提交哈希
	GitCommit = "unknown"

	// flagconf 配置文件路径
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "configs/config.yaml", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := conf.Load(flagconf)
	if err != nil {
		panic(err)
	}

	// 创建日志
	logger := log.With(
		log.NewStdLogger(os.Stdout),
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit,
	)

	// 依赖注入
	app, cleanup, err := wireApp(cfg, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 启动应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// newApp 创建 Kratos 应用
func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.Name("go-next-erp"),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{
			"build_time": BuildTime,
			"git_commit": GitCommit,
		}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}
