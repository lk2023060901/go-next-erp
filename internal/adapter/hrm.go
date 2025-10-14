package adapter

import (
	"context"

	hrmv1 "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/handler"
)

// HRMAdapter HRM 适配器
type HRMAdapter struct {
	hrmv1.UnimplementedAttendanceServiceServer
	hrmv1.UnimplementedShiftServiceServer
	hrmv1.UnimplementedScheduleServiceServer
	hrmv1.UnimplementedAttendanceRuleServiceServer

	attendanceHandler     *handler.AttendanceHandler
	shiftHandler          *handler.ShiftHandler
	scheduleHandler       *handler.ScheduleHandler
	attendanceRuleHandler *handler.AttendanceRuleHandler
}

// NewHRMAdapter 创建 HRM 适配器
func NewHRMAdapter(
	attendanceHandler *handler.AttendanceHandler,
	shiftHandler *handler.ShiftHandler,
	scheduleHandler *handler.ScheduleHandler,
	attendanceRuleHandler *handler.AttendanceRuleHandler,
) *HRMAdapter {
	return &HRMAdapter{
		attendanceHandler:     attendanceHandler,
		shiftHandler:          shiftHandler,
		scheduleHandler:       scheduleHandler,
		attendanceRuleHandler: attendanceRuleHandler,
	}
}

// AttendanceService methods (delegate to handler)
func (a *HRMAdapter) ClockIn(ctx context.Context, req *hrmv1.ClockInRequest) (*hrmv1.ClockInResponse, error) {
	return a.attendanceHandler.ClockIn(ctx, req)
}

func (a *HRMAdapter) GetAttendanceRecord(ctx context.Context, req *hrmv1.GetAttendanceRecordRequest) (*hrmv1.AttendanceRecordResponse, error) {
	return a.attendanceHandler.GetAttendanceRecord(ctx, req)
}

func (a *HRMAdapter) ListEmployeeAttendance(ctx context.Context, req *hrmv1.ListEmployeeAttendanceRequest) (*hrmv1.ListAttendanceRecordResponse, error) {
	return a.attendanceHandler.ListEmployeeAttendance(ctx, req)
}

func (a *HRMAdapter) ListDepartmentAttendance(ctx context.Context, req *hrmv1.ListDepartmentAttendanceRequest) (*hrmv1.ListAttendanceRecordResponse, error) {
	return a.attendanceHandler.ListDepartmentAttendance(ctx, req)
}

func (a *HRMAdapter) ListExceptionAttendance(ctx context.Context, req *hrmv1.ListExceptionAttendanceRequest) (*hrmv1.ListAttendanceRecordResponse, error) {
	return a.attendanceHandler.ListExceptionAttendance(ctx, req)
}

func (a *HRMAdapter) GetAttendanceStatistics(ctx context.Context, req *hrmv1.GetAttendanceStatisticsRequest) (*hrmv1.AttendanceStatisticsResponse, error) {
	return a.attendanceHandler.GetAttendanceStatistics(ctx, req)
}

// ShiftService methods (delegate to handler)
func (a *HRMAdapter) CreateShift(ctx context.Context, req *hrmv1.CreateShiftRequest) (*hrmv1.ShiftResponse, error) {
	return a.shiftHandler.CreateShift(ctx, req)
}

func (a *HRMAdapter) GetShift(ctx context.Context, req *hrmv1.GetShiftRequest) (*hrmv1.ShiftResponse, error) {
	return a.shiftHandler.GetShift(ctx, req)
}

func (a *HRMAdapter) UpdateShift(ctx context.Context, req *hrmv1.UpdateShiftRequest) (*hrmv1.ShiftResponse, error) {
	return a.shiftHandler.UpdateShift(ctx, req)
}

func (a *HRMAdapter) DeleteShift(ctx context.Context, req *hrmv1.DeleteShiftRequest) (*hrmv1.DeleteShiftResponse, error) {
	return a.shiftHandler.DeleteShift(ctx, req)
}

func (a *HRMAdapter) ListShifts(ctx context.Context, req *hrmv1.ListShiftsRequest) (*hrmv1.ListShiftsResponse, error) {
	return a.shiftHandler.ListShifts(ctx, req)
}

func (a *HRMAdapter) ListActiveShifts(ctx context.Context, req *hrmv1.ListActiveShiftsRequest) (*hrmv1.ListShiftsResponse, error) {
	return a.shiftHandler.ListActiveShifts(ctx, req)
}

// ScheduleService methods (delegate to handler)
func (a *HRMAdapter) CreateSchedule(ctx context.Context, req *hrmv1.CreateScheduleRequest) (*hrmv1.ScheduleResponse, error) {
	return a.scheduleHandler.CreateSchedule(ctx, req)
}

func (a *HRMAdapter) BatchCreateSchedules(ctx context.Context, req *hrmv1.BatchCreateSchedulesRequest) (*hrmv1.BatchCreateSchedulesResponse, error) {
	return a.scheduleHandler.BatchCreateSchedules(ctx, req)
}

func (a *HRMAdapter) GetSchedule(ctx context.Context, req *hrmv1.GetScheduleRequest) (*hrmv1.ScheduleResponse, error) {
	return a.scheduleHandler.GetSchedule(ctx, req)
}

func (a *HRMAdapter) UpdateSchedule(ctx context.Context, req *hrmv1.UpdateScheduleRequest) (*hrmv1.ScheduleResponse, error) {
	return a.scheduleHandler.UpdateSchedule(ctx, req)
}

func (a *HRMAdapter) DeleteSchedule(ctx context.Context, req *hrmv1.DeleteScheduleRequest) (*hrmv1.DeleteScheduleResponse, error) {
	return a.scheduleHandler.DeleteSchedule(ctx, req)
}

func (a *HRMAdapter) ListEmployeeSchedules(ctx context.Context, req *hrmv1.ListEmployeeSchedulesRequest) (*hrmv1.ListSchedulesResponse, error) {
	return a.scheduleHandler.ListEmployeeSchedules(ctx, req)
}

func (a *HRMAdapter) ListDepartmentSchedules(ctx context.Context, req *hrmv1.ListDepartmentSchedulesRequest) (*hrmv1.ListSchedulesResponse, error) {
	return a.scheduleHandler.ListDepartmentSchedules(ctx, req)
}

// AttendanceRuleService methods (delegate to handler)
func (a *HRMAdapter) CreateAttendanceRule(ctx context.Context, req *hrmv1.CreateAttendanceRuleRequest) (*hrmv1.AttendanceRuleResponse, error) {
	return a.attendanceRuleHandler.CreateAttendanceRule(ctx, req)
}

func (a *HRMAdapter) GetAttendanceRule(ctx context.Context, req *hrmv1.GetAttendanceRuleRequest) (*hrmv1.AttendanceRuleResponse, error) {
	return a.attendanceRuleHandler.GetAttendanceRule(ctx, req)
}

func (a *HRMAdapter) UpdateAttendanceRule(ctx context.Context, req *hrmv1.UpdateAttendanceRuleRequest) (*hrmv1.AttendanceRuleResponse, error) {
	return a.attendanceRuleHandler.UpdateAttendanceRule(ctx, req)
}

func (a *HRMAdapter) DeleteAttendanceRule(ctx context.Context, req *hrmv1.DeleteAttendanceRuleRequest) (*hrmv1.DeleteAttendanceRuleResponse, error) {
	return a.attendanceRuleHandler.DeleteAttendanceRule(ctx, req)
}

func (a *HRMAdapter) ListAttendanceRules(ctx context.Context, req *hrmv1.ListAttendanceRulesRequest) (*hrmv1.ListAttendanceRulesResponse, error) {
	return a.attendanceRuleHandler.ListAttendanceRules(ctx, req)
}
