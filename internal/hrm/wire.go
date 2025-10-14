//go:build wireinject
// +build wireinject

package hrm

import (
	"github.com/google/wire"
	"github.com/lk2023060901/go-next-erp/internal/hrm/handler"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository/postgres"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
	"gorm.io/gorm"
)

// ProviderSet HRM 模块的 Wire Provider Set
var ProviderSet = wire.NewSet(
	// Repository
	postgres.NewAttendanceRecordRepository,
	postgres.NewShiftRepository,
	postgres.NewScheduleRepository,
	postgres.NewAttendanceRuleRepository,
	postgres.NewHRMEmployeeRepository,

	// Service
	service.NewAttendanceService,
	service.NewShiftService,
	service.NewScheduleService,
	service.NewAttendanceRuleService,

	// Handler
	handler.NewAttendanceHandler,
	handler.NewShiftHandler,
	handler.NewScheduleHandler,
	handler.NewAttendanceRuleHandler,
)

// InitHRMModule initializes the HRM module
func InitHRMModule(db *gorm.DB) (*HRMModule, error) {
	panic(wire.Build(ProviderSet, wire.Struct(new(HRMModule), "*")))
}

// HRMModule represents the HRM module
type HRMModule struct {
	AttendanceHandler     *handler.AttendanceHandler
	ShiftHandler          *handler.ShiftHandler
	ScheduleHandler       *handler.ScheduleHandler
	AttendanceRuleHandler *handler.AttendanceRuleHandler
}
