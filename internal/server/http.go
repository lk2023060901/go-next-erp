package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(
	authAdapter *adapter.AuthAdapter,
	userAdapter *adapter.UserAdapter,
	roleAdapter *adapter.RoleAdapter,
	formAdapter *adapter.FormAdapter,
	orgAdapter *adapter.OrganizationAdapter,
	notifyAdapter *adapter.NotificationAdapter,
	approvalAdapter *adapter.ApprovalAdapter,
	fileAdapter *adapter.FileAdapter,
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

	// 注册 File 服务
	filev1.RegisterFileServiceHTTPServer(srv, fileAdapter)

	return srv
}
