//go:build wireinject
// +build wireinject

package hrm

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/hrm/handler"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository/postgres"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/lk2023060901/go-next-erp/pkg/workflow"
)

// ProviderSet HRM 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewAttendanceRecordRepository,
	postgres.NewShiftRepository,
	postgres.NewScheduleRepository,
	postgres.NewAttendanceRuleRepository,
	postgres.NewHRMEmployeeRepository,
	postgres.NewLeaveTypeRepository,
	postgres.NewLeaveQuotaRepository,
	postgres.NewLeaveRequestRepository,
	postgres.NewLeaveApprovalRepository,
	postgres.NewOvertimeRepository,
	postgres.NewBusinessTripRepository,
	postgres.NewLeaveOfficeRepository,
	postgres.NewPunchCardSupplementRepo,

	// Service
	service.NewAttendanceService,
	service.NewShiftService,
	service.NewScheduleService,
	service.NewAttendanceRuleService,
	service.NewLeaveService,
	service.NewOvertimeService,
	service.NewBusinessTripService,
	service.NewLeaveOfficeService,
	service.NewPunchCardSupplementService,

	// Handler
	handler.NewAttendanceHandler,
	handler.NewShiftHandler,
	handler.NewScheduleHandler,
	handler.NewAttendanceRuleHandler,
	handler.NewLeaveHandler,
	handler.NewOvertimeHandler,
	handler.NewBusinessTripHandler,
	handler.NewLeaveOfficeHandler,
	handler.NewPunchCardSupplementHandler,
)

// InitHRMModule initializes the HRM module
func InitHRMModule(db *database.DB, workflowEngine *workflow.Engine) (*HRMModule, error) {
	panic(wire.Build(ProviderSet, wire.Struct(new(HRMModule), "*")))
}

// HRMModule represents the HRM module
type HRMModule struct {
	AttendanceHandler          *handler.AttendanceHandler
	ShiftHandler               *handler.ShiftHandler
	ScheduleHandler            *handler.ScheduleHandler
	AttendanceRuleHandler      *handler.AttendanceRuleHandler
	LeaveHandler               *handler.LeaveHandler
	OvertimeHandler            *handler.OvertimeHandler
	BusinessTripHandler        *handler.BusinessTripHandler
	LeaveOfficeHandler         *handler.LeaveOfficeHandler
	PunchCardSupplementHandler *handler.PunchCardSupplementHandler
}
