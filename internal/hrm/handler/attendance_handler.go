package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pb "github.com/lk2023060901/go-next-erp/api/hrm/v1"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/internal/hrm/service"
)

// AttendanceHandler 考勤处理器
type AttendanceHandler struct {
	pb.UnimplementedAttendanceServiceServer
	attendanceService service.AttendanceService
}

// NewAttendanceHandler 创建考勤处理器
func NewAttendanceHandler(attendanceService service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
	}
}

// ClockIn 打卡
func (h *AttendanceHandler) ClockIn(ctx context.Context, req *pb.ClockInRequest) (*pb.ClockInResponse, error) {
	// 1. 参数验证
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, err
	}

	// 2. 解析打卡类型和方法
	clockType := model.AttendanceClockType(req.ClockType)
	method := model.AttendanceMethod(req.CheckInMethod)

	// 3. 构造服务请求
	svcReq := &service.ClockInRequest{
		TenantID:      tenantID,
		EmployeeID:    employeeID,
		ClockTime:     time.Now(),
		ClockType:     clockType,
		CheckInMethod: method,
		SourceType:    model.SourceTypeSystem,
		WiFiSSID:      req.WifiSsid,
		WiFiMAC:       req.WifiMac,
		PhotoURL:      req.PhotoUrl,
		FaceScore:     req.FaceScore,
		Remark:        req.Remark,
	}

	// 设置位置信息
	if req.Location != nil {
		svcReq.Location = &model.LocationInfo{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Accuracy:  req.Location.Accuracy,
		}
		svcReq.Address = req.Location.Address
	}

	// 4. 调用服务层
	record, err := h.attendanceService.ClockIn(ctx, svcReq)
	if err != nil {
		return nil, err
	}

	// 5. 构造响应
	return &pb.ClockInResponse{
		Id:              record.ID.String(),
		EmployeeId:      record.EmployeeID.String(),
		EmployeeName:    record.EmployeeName,
		ClockTime:       record.ClockTime.Format(time.RFC3339),
		ClockType:       string(record.ClockType),
		Status:          string(record.Status),
		ShiftName:       record.ShiftName,
		IsException:     record.IsException,
		ExceptionReason: record.ExceptionReason,
		Message:         "打卡成功",
	}, nil
}

// GetAttendanceRecord 获取考勤记录
func (h *AttendanceHandler) GetAttendanceRecord(ctx context.Context, req *pb.GetAttendanceRecordRequest) (*pb.AttendanceRecordResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	record, err := h.attendanceService.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return convertAttendanceRecordToProto(record), nil
}

// ListEmployeeAttendance 查询员工考勤记录（支持游标分页）
func (h *AttendanceHandler) ListEmployeeAttendance(ctx context.Context, req *pb.ListEmployeeAttendanceRequest) (*pb.ListAttendanceRecordResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeId)
	if err != nil {
		return nil, err
	}

	// 解析日期（使用上海时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startDate, _ := time.ParseInLocation("2006-01-02", req.StartDate, loc)
	endDateParsed, _ := time.ParseInLocation("2006-01-02", req.EndDate, loc)
	endDate := endDateParsed.AddDate(0, 0, 1)

	// 判断是否使用游标分页
	if req.UseCursor {
		return h.listEmployeeAttendanceWithCursor(ctx, tenantID, employeeID, startDate, endDate, req)
	}

	// 使用传统分页（保持兼容性）
	records, err := h.attendanceService.GetEmployeeRecords(ctx, tenantID, employeeID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.AttendanceRecordResponse, 0, len(records))
	for _, record := range records {
		items = append(items, convertAttendanceRecordToProto(record))
	}

	return &pb.ListAttendanceRecordResponse{
		Items: items,
		Total: int32(len(items)),
		Page:  1,
	}, nil
}

// listEmployeeAttendanceWithCursor 游标分页查询（高性能）
func (h *AttendanceHandler) listEmployeeAttendanceWithCursor(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	startDate, endDate time.Time,
	req *pb.ListEmployeeAttendanceRequest,
) (*pb.ListAttendanceRecordResponse, error) {
	// 解析游标
	var cursor *time.Time
	if req.Cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339, req.Cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor format: %w", err)
		}
		cursor = &cursorTime
	}

	// 构建过滤条件
	filter := &repository.AttendanceRecordFilter{
		EmployeeID: &employeeID,
		StartDate:  &startDate,
		EndDate:    &endDate,
	}

	// 页大小（默认 20，最大 100）
	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 调用service层游标分页
	records, nextCursor, hasNext, err := h.attendanceService.ListWithCursor(
		ctx, tenantID, filter, cursor, limit,
	)
	if err != nil {
		return nil, err
	}

	// 转换为proto格式
	items := make([]*pb.AttendanceRecordResponse, 0, len(records))
	for _, record := range records {
		items = append(items, convertAttendanceRecordToProto(record))
	}

	// 构建响应
	response := &pb.ListAttendanceRecordResponse{
		Items:    items,
		Total:    -1, // 游标分页不返回总数
		HasNext:  hasNext,
		HasPrev:  cursor != nil,
		PageSize: int32(limit),
	}

	// 生成下一页游标
	if nextCursor != nil {
		response.NextCursor = nextCursor.Format(time.RFC3339)
	}

	// 生成上一页游标（简化处理）
	if cursor != nil && len(records) > 0 {
		firstRecord := records[0]
		response.PrevCursor = firstRecord.ClockTime.Format(time.RFC3339)
	}

	return response, nil
}

// ListDepartmentAttendance 查询部门考勤记录
func (h *AttendanceHandler) ListDepartmentAttendance(ctx context.Context, req *pb.ListDepartmentAttendanceRequest) (*pb.ListAttendanceRecordResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}
	departmentID, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		return nil, err
	}

	// 解析日期（使用上海时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startDate, _ := time.ParseInLocation("2006-01-02", req.StartDate, loc)
	endDateParsed, _ := time.ParseInLocation("2006-01-02", req.EndDate, loc)
	endDate := endDateParsed.AddDate(0, 0, 1)

	records, err := h.attendanceService.GetDepartmentRecords(ctx, tenantID, departmentID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 转换为proto格式
	items := make([]*pb.AttendanceRecordResponse, 0, len(records))
	for _, record := range records {
		items = append(items, convertAttendanceRecordToProto(record))
	}

	return &pb.ListAttendanceRecordResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// ListExceptionAttendance 查询异常考勤
func (h *AttendanceHandler) ListExceptionAttendance(ctx context.Context, req *pb.ListExceptionAttendanceRequest) (*pb.ListAttendanceRecordResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	// 解析日期（使用上海时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startDate, _ := time.ParseInLocation("2006-01-02", req.StartDate, loc)
	endDateParsed, _ := time.ParseInLocation("2006-01-02", req.EndDate, loc)
	endDate := endDateParsed.AddDate(0, 0, 1)

	records, err := h.attendanceService.GetExceptions(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 转换为proto格式
	items := make([]*pb.AttendanceRecordResponse, 0, len(records))
	for _, record := range records {
		items = append(items, convertAttendanceRecordToProto(record))
	}

	return &pb.ListAttendanceRecordResponse{
		Items: items,
		Total: int32(len(items)),
	}, nil
}

// GetAttendanceStatistics 考勤统计
func (h *AttendanceHandler) GetAttendanceStatistics(ctx context.Context, req *pb.GetAttendanceStatisticsRequest) (*pb.AttendanceStatisticsResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, err
	}

	// 解析日期（使用上海时区）
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startDate, _ := time.ParseInLocation("2006-01-02", req.StartDate, loc)
	endDateParsed, _ := time.ParseInLocation("2006-01-02", req.EndDate, loc)
	endDate := endDateParsed.AddDate(0, 0, 1)

	statusCount, err := h.attendanceService.CountByStatus(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 构造响应
	statCountMap := make(map[string]int32)
	for status, count := range statusCount {
		statCountMap[string(status)] = int32(count)
	}

	return &pb.AttendanceStatisticsResponse{
		NormalDays:  int32(statusCount[model.AttendanceStatusNormal]),
		LateDays:    int32(statusCount[model.AttendanceStatusLate]),
		EarlyDays:   int32(statusCount[model.AttendanceStatusEarly]),
		AbsentDays:  int32(statusCount[model.AttendanceStatusAbsent]),
		LeaveDays:   int32(statusCount[model.AttendanceStatusLeave]),
		StatusCount: statCountMap,
	}, nil
}

// convertAttendanceRecordToProto 转换考勤记录为Proto格式
func convertAttendanceRecordToProto(record *model.AttendanceRecord) *pb.AttendanceRecordResponse {
	resp := &pb.AttendanceRecordResponse{
		Id:              record.ID.String(),
		TenantId:        record.TenantID.String(),
		EmployeeId:      record.EmployeeID.String(),
		EmployeeName:    record.EmployeeName,
		ClockTime:       record.ClockTime.Format(time.RFC3339),
		ClockType:       string(record.ClockType),
		Status:          string(record.Status),
		CheckInMethod:   string(record.CheckInMethod),
		SourceType:      string(record.SourceType),
		Address:         record.Address,
		WifiSsid:        record.WiFiSSID,
		PhotoUrl:        record.PhotoURL,
		FaceScore:       record.FaceScore,
		IsException:     record.IsException,
		ExceptionReason: record.ExceptionReason,
		ExceptionType:   record.ExceptionType,
		Remark:          record.Remark,
		CreatedAt:       record.CreatedAt.Format(time.RFC3339),
	}

	if record.DepartmentID != uuid.Nil {
		resp.DepartmentId = record.DepartmentID.String()
	}
	if record.ShiftID != nil {
		resp.ShiftId = record.ShiftID.String()
	}
	resp.ShiftName = record.ShiftName

	// 设置位置信息
	if record.Location != nil {
		resp.Location = &pb.LocationInfo{
			Latitude:  record.Location.Latitude,
			Longitude: record.Location.Longitude,
			Accuracy:  record.Location.Accuracy,
		}
	}

	return resp
}
