package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(
	authAdapter *adapter.AuthAdapter,
	userAdapter *adapter.UserAdapter,
	roleAdapter *adapter.RoleAdapter,
	logger log.Logger,
) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}

	srv := http.NewServer(opts...)

	// 注册 Auth 服务
	authv1.RegisterAuthServiceHTTPServer(srv, authAdapter)

	// 注册 User 服务
	authv1.RegisterUserServiceHTTPServer(srv, userAdapter)

	// 注册 Role 服务
	authv1.RegisterRoleServiceHTTPServer(srv, roleAdapter)

	return srv
}
