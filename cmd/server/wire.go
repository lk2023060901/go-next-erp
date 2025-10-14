//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
	"github.com/lk2023060901/go-next-erp/internal/approval"
	"github.com/lk2023060901/go-next-erp/internal/auth"
	"github.com/lk2023060901/go-next-erp/internal/conf"
	"github.com/lk2023060901/go-next-erp/internal/file"
	"github.com/lk2023060901/go-next-erp/internal/form"
	"github.com/lk2023060901/go-next-erp/internal/notification"
	"github.com/lk2023060901/go-next-erp/internal/organization"
	"github.com/lk2023060901/go-next-erp/internal/server"
	"github.com/lk2023060901/go-next-erp/pkg"
)

// wireApp 通过 Wire 进行依赖注入
func wireApp(context.Context, *conf.Config, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		pkg.ProviderSet,
		auth.ProviderSet,
		file.ProviderSet,
		form.ProviderSet,
		organization.ProviderSet,
		notification.ProviderSet,
		approval.ProviderSet,
		adapter.ProviderSet,
		server.ProviderSet,
		newApp,
	))
}
