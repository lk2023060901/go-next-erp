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
	hrmv1.UnimplementedOvertimeServiceServer
	hrmv1.UnimplementedLeaveTypeServiceServer
	hrmv1.UnimplementedLeaveRequestServiceServer
	hrmv1.UnimplementedLeaveQuotaServiceServer
	hrmv1.UnimplementedBusinessTripServiceServer
	hrmv1.UnimplementedLeaveOfficeServiceServer

	attendanceHandler     *handler.AttendanceHandler
	shiftHandler          *handler.ShiftHandler
	scheduleHandler       *handler.ScheduleHandler
	attendanceRuleHandler *handler.AttendanceRuleHandler
	overtimeHandler       *handler.OvertimeHandler
	leaveHandler          *handler.LeaveHandler
	businessTripHandler   *handler.BusinessTripHandler
	leaveOfficeHandler    *handler.LeaveOfficeHandler
}

// NewHRMAdapter 创建 HRM 适配器
func NewHRMAdapter(
	attendanceHandler *handler.AttendanceHandler,
	shiftHandler *handler.ShiftHandler,
	scheduleHandler *handler.ScheduleHandler,
	attendanceRuleHandler *handler.AttendanceRuleHandler,
	overtimeHandler *handler.OvertimeHandler,
	leaveHandler *handler.LeaveHandler,
	businessTripHandler *handler.BusinessTripHandler,
	leaveOfficeHandler *handler.LeaveOfficeHandler,
) *HRMAdapter {
	return &HRMAdapter{
		attendanceHandler:     attendanceHandler,
		shiftHandler:          shiftHandler,
		scheduleHandler:       scheduleHandler,
		attendanceRuleHandler: attendanceRuleHandler,
		overtimeHandler:       overtimeHandler,
		leaveHandler:          leaveHandler,
		businessTripHandler:   businessTripHandler,
		leaveOfficeHandler:    leaveOfficeHandler,
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

// OvertimeService methods (delegate to handler)
func (a *HRMAdapter) CreateOvertime(ctx context.Context, req *hrmv1.CreateOvertimeRequest) (*hrmv1.OvertimeResponse, error) {
	return a.overtimeHandler.CreateOvertime(ctx, req)
}

func (a *HRMAdapter) UpdateOvertime(ctx context.Context, req *hrmv1.UpdateOvertimeRequest) (*hrmv1.OvertimeResponse, error) {
	return a.overtimeHandler.UpdateOvertime(ctx, req)
}

func (a *HRMAdapter) DeleteOvertime(ctx context.Context, req *hrmv1.DeleteOvertimeRequest) (*hrmv1.DeleteOvertimeResponse, error) {
	return a.overtimeHandler.DeleteOvertime(ctx, req)
}

func (a *HRMAdapter) GetOvertime(ctx context.Context, req *hrmv1.GetOvertimeRequest) (*hrmv1.OvertimeResponse, error) {
	return a.overtimeHandler.GetOvertime(ctx, req)
}

func (a *HRMAdapter) ListOvertimes(ctx context.Context, req *hrmv1.ListOvertimesRequest) (*hrmv1.ListOvertimesResponse, error) {
	return a.overtimeHandler.ListOvertimes(ctx, req)
}

func (a *HRMAdapter) ListEmployeeOvertimes(ctx context.Context, req *hrmv1.ListEmployeeOvertimesRequest) (*hrmv1.ListOvertimesResponse, error) {
	return a.overtimeHandler.ListEmployeeOvertimes(ctx, req)
}

func (a *HRMAdapter) ListPendingOvertimes(ctx context.Context, req *hrmv1.ListPendingOvertimesRequest) (*hrmv1.ListOvertimesResponse, error) {
	return a.overtimeHandler.ListPendingOvertimes(ctx, req)
}

func (a *HRMAdapter) SubmitOvertime(ctx context.Context, req *hrmv1.SubmitOvertimeRequest) (*hrmv1.SubmitOvertimeResponse, error) {
	return a.overtimeHandler.SubmitOvertime(ctx, req)
}

func (a *HRMAdapter) ApproveOvertime(ctx context.Context, req *hrmv1.ApproveOvertimeRequest) (*hrmv1.ApproveOvertimeResponse, error) {
	return a.overtimeHandler.ApproveOvertime(ctx, req)
}

func (a *HRMAdapter) RejectOvertime(ctx context.Context, req *hrmv1.RejectOvertimeRequest) (*hrmv1.RejectOvertimeResponse, error) {
	return a.overtimeHandler.RejectOvertime(ctx, req)
}

func (a *HRMAdapter) SumOvertimeHours(ctx context.Context, req *hrmv1.SumOvertimeHoursRequest) (*hrmv1.SumOvertimeHoursResponse, error) {
	return a.overtimeHandler.SumOvertimeHours(ctx, req)
}

func (a *HRMAdapter) GetCompOffDays(ctx context.Context, req *hrmv1.GetCompOffDaysRequest) (*hrmv1.GetCompOffDaysResponse, error) {
	return a.overtimeHandler.GetCompOffDays(ctx, req)
}

func (a *HRMAdapter) UseCompOffDays(ctx context.Context, req *hrmv1.UseCompOffDaysRequest) (*hrmv1.UseCompOffDaysResponse, error) {
	return a.overtimeHandler.UseCompOffDays(ctx, req)
}

// LeaveTypeService methods (delegate to handler)
func (a *HRMAdapter) CreateLeaveType(ctx context.Context, req *hrmv1.CreateLeaveTypeRequest) (*hrmv1.LeaveTypeResponse, error) {
	return a.leaveHandler.CreateLeaveType(ctx, req)
}

func (a *HRMAdapter) UpdateLeaveType(ctx context.Context, req *hrmv1.UpdateLeaveTypeRequest) (*hrmv1.LeaveTypeResponse, error) {
	return a.leaveHandler.UpdateLeaveType(ctx, req)
}

func (a *HRMAdapter) DeleteLeaveType(ctx context.Context, req *hrmv1.DeleteLeaveTypeRequest) (*hrmv1.DeleteLeaveTypeResponse, error) {
	return a.leaveHandler.DeleteLeaveType(ctx, req)
}

func (a *HRMAdapter) GetLeaveType(ctx context.Context, req *hrmv1.GetLeaveTypeRequest) (*hrmv1.LeaveTypeResponse, error) {
	return a.leaveHandler.GetLeaveType(ctx, req)
}

func (a *HRMAdapter) ListLeaveTypes(ctx context.Context, req *hrmv1.ListLeaveTypesRequest) (*hrmv1.ListLeaveTypesResponse, error) {
	return a.leaveHandler.ListLeaveTypes(ctx, req)
}

func (a *HRMAdapter) ListActiveLeaveTypes(ctx context.Context, req *hrmv1.ListActiveLeaveTypesRequest) (*hrmv1.ListActiveLeaveTypesResponse, error) {
	return a.leaveHandler.ListActiveLeaveTypes(ctx, req)
}

// LeaveRequestService methods (delegate to handler)
func (a *HRMAdapter) CreateLeaveRequest(ctx context.Context, req *hrmv1.CreateLeaveRequestRequest) (*hrmv1.LeaveRequestResponse, error) {
	return a.leaveHandler.CreateLeaveRequest(ctx, req)
}

func (a *HRMAdapter) UpdateLeaveRequest(ctx context.Context, req *hrmv1.UpdateLeaveRequestRequest) (*hrmv1.LeaveRequestResponse, error) {
	return a.leaveHandler.UpdateLeaveRequest(ctx, req)
}

func (a *HRMAdapter) SubmitLeaveRequest(ctx context.Context, req *hrmv1.SubmitLeaveRequestRequest) (*hrmv1.SubmitLeaveRequestResponse, error) {
	return a.leaveHandler.SubmitLeaveRequest(ctx, req)
}

func (a *HRMAdapter) WithdrawLeaveRequest(ctx context.Context, req *hrmv1.WithdrawLeaveRequestRequest) (*hrmv1.WithdrawLeaveRequestResponse, error) {
	return a.leaveHandler.WithdrawLeaveRequest(ctx, req)
}

func (a *HRMAdapter) CancelLeaveRequest(ctx context.Context, req *hrmv1.CancelLeaveRequestRequest) (*hrmv1.CancelLeaveRequestResponse, error) {
	return a.leaveHandler.CancelLeaveRequest(ctx, req)
}

func (a *HRMAdapter) GetLeaveRequest(ctx context.Context, req *hrmv1.GetLeaveRequestRequest) (*hrmv1.LeaveRequestDetailResponse, error) {
	return a.leaveHandler.GetLeaveRequest(ctx, req)
}

func (a *HRMAdapter) ListMyLeaveRequests(ctx context.Context, req *hrmv1.ListMyLeaveRequestsRequest) (*hrmv1.ListLeaveRequestsResponse, error) {
	return a.leaveHandler.ListMyLeaveRequests(ctx, req)
}

func (a *HRMAdapter) ListLeaveRequests(ctx context.Context, req *hrmv1.ListLeaveRequestsRequest) (*hrmv1.ListLeaveRequestsResponse, error) {
	return a.leaveHandler.ListLeaveRequests(ctx, req)
}

func (a *HRMAdapter) ListPendingApprovals(ctx context.Context, req *hrmv1.ListPendingApprovalsRequest) (*hrmv1.ListLeaveRequestsResponse, error) {
	return a.leaveHandler.ListPendingApprovals(ctx, req)
}

func (a *HRMAdapter) ApproveLeaveRequest(ctx context.Context, req *hrmv1.ApproveLeaveRequestRequest) (*hrmv1.ApproveLeaveRequestResponse, error) {
	return a.leaveHandler.ApproveLeaveRequest(ctx, req)
}

func (a *HRMAdapter) RejectLeaveRequest(ctx context.Context, req *hrmv1.RejectLeaveRequestRequest) (*hrmv1.RejectLeaveRequestResponse, error) {
	return a.leaveHandler.RejectLeaveRequest(ctx, req)
}

// LeaveQuotaService methods (delegate to handler)
func (a *HRMAdapter) InitEmployeeQuota(ctx context.Context, req *hrmv1.InitEmployeeQuotaRequest) (*hrmv1.InitEmployeeQuotaResponse, error) {
	return a.leaveHandler.InitEmployeeQuota(ctx, req)
}

func (a *HRMAdapter) UpdateQuota(ctx context.Context, req *hrmv1.UpdateQuotaRequest) (*hrmv1.QuotaResponse, error) {
	return a.leaveHandler.UpdateQuota(ctx, req)
}

func (a *HRMAdapter) GetEmployeeQuotas(ctx context.Context, req *hrmv1.GetEmployeeQuotasRequest) (*hrmv1.GetEmployeeQuotasResponse, error) {
	return a.leaveHandler.GetEmployeeQuotas(ctx, req)
}

// BusinessTripService methods (delegate to handler)
func (a *HRMAdapter) CreateBusinessTrip(ctx context.Context, req *hrmv1.CreateBusinessTripRequest) (*hrmv1.BusinessTripResponse, error) {
	return a.businessTripHandler.CreateBusinessTrip(ctx, req)
}

func (a *HRMAdapter) UpdateBusinessTrip(ctx context.Context, req *hrmv1.UpdateBusinessTripRequest) (*hrmv1.BusinessTripResponse, error) {
	return a.businessTripHandler.UpdateBusinessTrip(ctx, req)
}

func (a *HRMAdapter) DeleteBusinessTrip(ctx context.Context, req *hrmv1.DeleteBusinessTripRequest) (*hrmv1.DeleteBusinessTripResponse, error) {
	return a.businessTripHandler.DeleteBusinessTrip(ctx, req)
}

func (a *HRMAdapter) GetBusinessTrip(ctx context.Context, req *hrmv1.GetBusinessTripRequest) (*hrmv1.BusinessTripResponse, error) {
	return a.businessTripHandler.GetBusinessTrip(ctx, req)
}

func (a *HRMAdapter) ListBusinessTrips(ctx context.Context, req *hrmv1.ListBusinessTripsRequest) (*hrmv1.ListBusinessTripsResponse, error) {
	return a.businessTripHandler.ListBusinessTrips(ctx, req)
}

func (a *HRMAdapter) ListEmployeeBusinessTrips(ctx context.Context, req *hrmv1.ListEmployeeBusinessTripsRequest) (*hrmv1.ListBusinessTripsResponse, error) {
	return a.businessTripHandler.ListEmployeeBusinessTrips(ctx, req)
}

func (a *HRMAdapter) ListPendingBusinessTrips(ctx context.Context, req *hrmv1.ListPendingBusinessTripsRequest) (*hrmv1.ListBusinessTripsResponse, error) {
	return a.businessTripHandler.ListPendingBusinessTrips(ctx, req)
}

func (a *HRMAdapter) SubmitBusinessTrip(ctx context.Context, req *hrmv1.SubmitBusinessTripRequest) (*hrmv1.SubmitBusinessTripResponse, error) {
	return a.businessTripHandler.SubmitBusinessTrip(ctx, req)
}

func (a *HRMAdapter) ApproveBusinessTrip(ctx context.Context, req *hrmv1.ApproveBusinessTripRequest) (*hrmv1.ApproveBusinessTripResponse, error) {
	return a.businessTripHandler.ApproveBusinessTrip(ctx, req)
}

func (a *HRMAdapter) RejectBusinessTrip(ctx context.Context, req *hrmv1.RejectBusinessTripRequest) (*hrmv1.RejectBusinessTripResponse, error) {
	return a.businessTripHandler.RejectBusinessTrip(ctx, req)
}

func (a *HRMAdapter) SubmitTripReport(ctx context.Context, req *hrmv1.SubmitTripReportRequest) (*hrmv1.BusinessTripResponse, error) {
	return a.businessTripHandler.SubmitTripReport(ctx, req)
}

// LeaveOfficeService methods (delegate to handler)
func (a *HRMAdapter) CreateLeaveOffice(ctx context.Context, req *hrmv1.CreateLeaveOfficeRequest) (*hrmv1.LeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.CreateLeaveOffice(ctx, req)
}

func (a *HRMAdapter) UpdateLeaveOffice(ctx context.Context, req *hrmv1.UpdateLeaveOfficeRequest) (*hrmv1.LeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.UpdateLeaveOffice(ctx, req)
}

func (a *HRMAdapter) DeleteLeaveOffice(ctx context.Context, req *hrmv1.DeleteLeaveOfficeRequest) (*hrmv1.DeleteLeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.DeleteLeaveOffice(ctx, req)
}

func (a *HRMAdapter) GetLeaveOffice(ctx context.Context, req *hrmv1.GetLeaveOfficeRequest) (*hrmv1.LeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.GetLeaveOffice(ctx, req)
}

func (a *HRMAdapter) ListLeaveOffices(ctx context.Context, req *hrmv1.ListLeaveOfficesRequest) (*hrmv1.ListLeaveOfficesResponse, error) {
	return a.leaveOfficeHandler.ListLeaveOffices(ctx, req)
}

func (a *HRMAdapter) ListEmployeeLeaveOffices(ctx context.Context, req *hrmv1.ListEmployeeLeaveOfficesRequest) (*hrmv1.ListLeaveOfficesResponse, error) {
	return a.leaveOfficeHandler.ListEmployeeLeaveOffices(ctx, req)
}

func (a *HRMAdapter) ListPendingLeaveOffices(ctx context.Context, req *hrmv1.ListPendingLeaveOfficesRequest) (*hrmv1.ListLeaveOfficesResponse, error) {
	return a.leaveOfficeHandler.ListPendingLeaveOffices(ctx, req)
}

func (a *HRMAdapter) SubmitLeaveOffice(ctx context.Context, req *hrmv1.SubmitLeaveOfficeRequest) (*hrmv1.SubmitLeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.SubmitLeaveOffice(ctx, req)
}

func (a *HRMAdapter) ApproveLeaveOffice(ctx context.Context, req *hrmv1.ApproveLeaveOfficeRequest) (*hrmv1.ApproveLeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.ApproveLeaveOffice(ctx, req)
}

func (a *HRMAdapter) RejectLeaveOffice(ctx context.Context, req *hrmv1.RejectLeaveOfficeRequest) (*hrmv1.RejectLeaveOfficeResponse, error) {
	return a.leaveOfficeHandler.RejectLeaveOffice(ctx, req)
}
