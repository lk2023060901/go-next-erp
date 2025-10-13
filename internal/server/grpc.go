package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	approvalv1 "github.com/lk2023060901/go-next-erp/api/approval/v1"
	authv1 "github.com/lk2023060901/go-next-erp/api/auth/v1"
	filev1 "github.com/lk2023060901/go-next-erp/api/file/v1"
	formv1 "github.com/lk2023060901/go-next-erp/api/form/v1"
	notifyv1 "github.com/lk2023060901/go-next-erp/api/notification/v1"
	orgv1 "github.com/lk2023060901/go-next-erp/api/organization/v1"
	"github.com/lk2023060901/go-next-erp/internal/adapter"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(
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
