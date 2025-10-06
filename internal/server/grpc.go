package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(
	authAdapter *adapter.AuthAdapter,
	userAdapter *adapter.UserAdapter,
	roleAdapter *adapter.RoleAdapter,
	logger log.Logger,
) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
		),
	}

	srv := grpc.NewServer(opts...)

	// 注册 Auth 服务
	authv1.RegisterAuthServiceServer(srv, authAdapter)

	// 注册 User 服务
	authv1.RegisterUserServiceServer(srv, userAdapter)

	// 注册 Role 服务
	authv1.RegisterRoleServiceServer(srv, roleAdapter)

	return srv
}
