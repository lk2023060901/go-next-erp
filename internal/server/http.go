package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
	"github.com/lk2023060901/go-next-erp/internal/auth/authentication/jwt"
	"github.com/lk2023060901/go-next-erp/internal/conf"
	"github.com/lk2023060901/go-next-erp/internal/notification"
	"github.com/lk2023060901/go-next-erp/internal/notification/service"
	ws "github.com/lk2023060901/go-next-erp/internal/notification/websocket"
	"github.com/lk2023060901/go-next-erp/pkg/middleware"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(
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
	notifService service.NotificationService, // 通知服务
	wsHub *ws.Hub, // WebSocket Hub
	wsHandler *ws.Handler, // WebSocket 处理器
	logger log.Logger,
) *http.Server {
	// 启动 WebSocket Hub
	go wsHub.Run()

	// 初始化 WebSocket 支持
	notification.InitNotificationWebSocket(notifService, wsHandler)
	// 解析超时配置
	timeout := 30 * time.Second
	if cfg.Server.HTTP.Timeout != "" {
		if t, err := time.ParseDuration(cfg.Server.HTTP.Timeout); err == nil {
			timeout = t
		}
	}

	// 不需要认证的接口列表（使用 protobuf operation 格式）
	noAuthPaths := []string{
		"/api.auth.v1.AuthService/Login",
		"/api.auth.v1.AuthService/Register",
		"/api.auth.v1.AuthService/RefreshToken",
	}

	var opts = []http.ServerOption{
		http.Address(cfg.Server.HTTP.Addr),
		http.Network(cfg.Server.HTTP.Network),
		http.Timeout(timeout),
		http.Middleware(
			recovery.Recovery(),
			middleware.Logging(logger),
			selector.Server(
				middleware.Auth(jwtManager),
			).Match(func(ctx context.Context, operation string) bool {
				// 检查是否为不需要认证的接口
				for _, path := range noAuthPaths {
					if operation == path {
						return false // 不需要认证
					}
				}
				return true // 需要认证
			}).Build(),
		),
	}

	srv := http.NewServer(opts...)

	// 注册 Auth 服务
	authv1.RegisterAuthServiceHTTPServer(srv, authAdapter)

	// 注册 User 服务
	authv1.RegisterUserServiceHTTPServer(srv, userAdapter)

	// 注册 Role 服务
	authv1.RegisterRoleServiceHTTPServer(srv, roleAdapter)

	// 注册 Form 服务
	formv1.RegisterFormDefinitionServiceHTTPServer(srv, formAdapter)
	formv1.RegisterFormDataServiceHTTPServer(srv, formAdapter)

	// 注册 Organization 服务
	orgv1.RegisterOrganizationServiceHTTPServer(srv, orgAdapter)
	orgv1.RegisterEmployeeServiceHTTPServer(srv, orgAdapter)

	// 注册 Notification 服务
	notifyv1.RegisterNotificationServiceHTTPServer(srv, notifyAdapter)

	// 注册 Approval 服务
	approvalv1.RegisterProcessDefinitionServiceHTTPServer(srv, approvalAdapter)
	approvalv1.RegisterProcessInstanceServiceHTTPServer(srv, approvalAdapter)
	approvalv1.RegisterApprovalTaskServiceHTTPServer(srv, approvalAdapter)

	// 注分 File 服务
	filev1.RegisterFileServiceHTTPServer(srv, fileAdapter)

	// 注册 WebSocket 通知推送路由
	srv.HandleFunc("/api/v1/notifications/ws", wsHandler.ServeHTTP)

	return srv
}
