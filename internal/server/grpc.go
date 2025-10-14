package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/conf"
	"github.com/lk2023060901/go-next-erp/pkg/middleware"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(
	cfg *conf.Config,
	jwtManager *jwt.Manager,
	authAdapter *adapter.AuthAdapter,
	userAdapter *adapter.UserAdapter,
	roleAdapter *adapter.RoleAdapter,
	formAdapter *adapter.FormAdapter,
	orgAdapter *adapter.OrganizationAdapter,
	notifyAdapter *adapter.NotificationAdapter,
	approvalAdapter *adapter.ApprovalAdapter,
	fileAdapter *adapter.FileAdapter,
	logger log.Logger,
) *grpc.Server {
	// 解析超时配置
	timeout := 30 * time.Second
	if cfg.Server.GRPC.Timeout != "" {
		if t, err := time.ParseDuration(cfg.Server.GRPC.Timeout); err == nil {
			timeout = t
		}
	}

	// 不需要认证的gRPC方法列表
	noAuthMethods := []string{
		"/api.auth.v1.AuthService/Login",
		"/api.auth.v1.AuthService/Register",
		"/api.auth.v1.AuthService/RefreshToken",
	}

	var opts = []grpc.ServerOption{
		grpc.Address(cfg.Server.GRPC.Addr),
		grpc.Network(cfg.Server.GRPC.Network),
		grpc.Timeout(timeout),
		grpc.Middleware(
			recovery.Recovery(),
			middleware.Logging(logger),
			selector.Server(
				middleware.Auth(jwtManager),
			).Match(func(ctx context.Context, operation string) bool {
				// 检查是否为不需要认证的方法
				for _, method := range noAuthMethods {
					if operation == method {
						return false // 不需要认证
					}
				}
				return true // 需要认证
			}).Build(),
		),
	}

	srv := grpc.NewServer(opts...)

	// 注册 Auth 服务
	authv1.RegisterAuthServiceServer(srv, authAdapter)

	// 注册 User 服务
	authv1.RegisterUserServiceServer(srv, userAdapter)

	// 注册 Role 服务
	authv1.RegisterRoleServiceServer(srv, roleAdapter)

	// 注册 Form 服务
	formv1.RegisterFormDefinitionServiceServer(srv, formAdapter)
	formv1.RegisterFormDataServiceServer(srv, formAdapter)

	// 注册 Organization 服务
	orgv1.RegisterOrganizationServiceServer(srv, orgAdapter)
	orgv1.RegisterEmployeeServiceServer(srv, orgAdapter)

	// 注册 Notification 服务
	notifyv1.RegisterNotificationServiceServer(srv, notifyAdapter)

	// 注册 Approval 服务
	approvalv1.RegisterProcessDefinitionServiceServer(srv, approvalAdapter)
	approvalv1.RegisterProcessInstanceServiceServer(srv, approvalAdapter)
	approvalv1.RegisterApprovalTaskServiceServer(srv, approvalAdapter)

	// 注册 File 服务
	filev1.RegisterFileServiceServer(srv, fileAdapter)

	return srv
}
