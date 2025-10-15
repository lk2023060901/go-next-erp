package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/hrm/model"
	"github.com/lk2023060901/go-next-erp/internal/hrm/repository"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

type businessTripRepository struct {
	db *database.DB
}

// NewBusinessTripRepository 创建出差仓储实例
func NewBusinessTripRepository(db *database.DB) repository.BusinessTripRepository {
	return &businessTripRepository{db: db}
}

// Create 创建出差记录
func (r *businessTripRepository) Create(ctx context.Context, trip *model.BusinessTrip) error {
	query := `
		INSERT INTO hrm_business_trips (
			id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_status,
			remark, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	_, err := r.db.Exec(ctx, query,
		trip.ID, trip.TenantID, trip.EmployeeID, trip.EmployeeName, trip.DepartmentID,
		trip.StartTime, trip.EndTime, trip.Duration,
		trip.Destination, trip.Transportation, trip.Accommodation, trip.Companions,
		trip.Purpose, trip.Tasks,
		trip.EstimatedCost, trip.ActualCost,
		trip.ApprovalStatus,
		trip.Remark, trip.CreatedAt, trip.UpdatedAt,
	)

	return err
}

// FindByID 根据ID查找
func (r *businessTripRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.BusinessTrip, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			report, report_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_business_trips
		WHERE id = $1 AND deleted_at IS NULL
	`

	trip := &model.BusinessTrip{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.TenantID, &trip.EmployeeID, &trip.EmployeeName, &trip.DepartmentID,
		&trip.StartTime, &trip.EndTime, &trip.Duration,
		&trip.Destination, &trip.Transportation, &trip.Accommodation, &trip.Companions,
		&trip.Purpose, &trip.Tasks,
		&trip.EstimatedCost, &trip.ActualCost,
		&trip.ApprovalID, &trip.ApprovalStatus, &trip.ApprovedBy, &trip.ApprovedAt, &trip.RejectReason,
		&trip.Report, &trip.ReportAt,
		&trip.Remark, &trip.CreatedAt, &trip.UpdatedAt, &trip.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return trip, nil
}

// Update 更新出差记录
func (r *businessTripRepository) Update(ctx context.Context, trip *model.BusinessTrip) error {
	query := `
		UPDATE hrm_business_trips SET
			employee_name = $2,
			department_id = $3,
			start_time = $4,
			end_time = $5,
			duration = $6,
			destination = $7,
			transportation = $8,
			accommodation = $9,
			companions = $10,
			purpose = $11,
			tasks = $12,
			estimated_cost = $13,
			actual_cost = $14,
			approval_id = $15,
			approval_status = $16,
			approved_by = $17,
			approved_at = $18,
			reject_reason = $19,
			report = $20,
			report_at = $21,
			remark = $22,
			updated_at = $23
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query,
		trip.ID, trip.EmployeeName, trip.DepartmentID,
		trip.StartTime, trip.EndTime, trip.Duration,
		trip.Destination, trip.Transportation, trip.Accommodation, trip.Companions,
		trip.Purpose, trip.Tasks,
		trip.EstimatedCost, trip.ActualCost,
		trip.ApprovalID, trip.ApprovalStatus, trip.ApprovedBy, trip.ApprovedAt, trip.RejectReason,
		trip.Report, trip.ReportAt,
		trip.Remark, trip.UpdatedAt,
	)

	return err
}

// Delete 删除出差记录（软删除）
func (r *businessTripRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE hrm_business_trips SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

// List 列表查询（分页）
func (r *businessTripRepository) List(ctx context.Context, tenantID uuid.UUID, filter *repository.BusinessTripFilter, offset, limit int) ([]*model.BusinessTrip, int, error) {
	conditions := []string{"tenant_id = $1", "deleted_at IS NULL"}
	args := []interface{}{tenantID}
	argIndex := 2

	if filter != nil {
		if filter.EmployeeID != nil {
			conditions = append(conditions, fmt.Sprintf("employee_id = $%d", argIndex))
			args = append(args, *filter.EmployeeID)
			argIndex++
		}
		if filter.DepartmentID != nil {
			conditions = append(conditions, fmt.Sprintf("department_id = $%d", argIndex))
			args = append(args, *filter.DepartmentID)
			argIndex++
		}
		if filter.ApprovalStatus != nil {
			conditions = append(conditions, fmt.Sprintf("approval_status = $%d", argIndex))
			args = append(args, *filter.ApprovalStatus)
			argIndex++
		}
		if filter.StartDate != nil {
			conditions = append(conditions, fmt.Sprintf("start_time >= $%d", argIndex))
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			conditions = append(conditions, fmt.Sprintf("end_time <= $%d", argIndex))
			args = append(args, *filter.EndDate)
			argIndex++
		}
		if filter.Keyword != "" {
			conditions = append(conditions, fmt.Sprintf("(employee_name ILIKE $%d OR destination ILIKE $%d OR purpose ILIKE $%d)", argIndex, argIndex, argIndex))
			args = append(args, "%"+filter.Keyword+"%")
			argIndex++
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM hrm_business_trips WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf(`
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			report, report_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_business_trips
		WHERE %s
		ORDER BY start_time DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	trips := make([]*model.BusinessTrip, 0)
	for rows.Next() {
		trip := &model.BusinessTrip{}
		err := rows.Scan(
			&trip.ID, &trip.TenantID, &trip.EmployeeID, &trip.EmployeeName, &trip.DepartmentID,
			&trip.StartTime, &trip.EndTime, &trip.Duration,
			&trip.Destination, &trip.Transportation, &trip.Accommodation, &trip.Companions,
			&trip.Purpose, &trip.Tasks,
			&trip.EstimatedCost, &trip.ActualCost,
			&trip.ApprovalID, &trip.ApprovalStatus, &trip.ApprovedBy, &trip.ApprovedAt, &trip.RejectReason,
			&trip.Report, &trip.ReportAt,
			&trip.Remark, &trip.CreatedAt, &trip.UpdatedAt, &trip.DeletedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		trips = append(trips, trip)
	}

	return trips, total, rows.Err()
}

// FindByEmployee 查询员工出差记录
func (r *businessTripRepository) FindByEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, year int) ([]*model.BusinessTrip, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			report, report_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_business_trips
		WHERE tenant_id = $1 AND employee_id = $2
			AND EXTRACT(YEAR FROM start_time) = $3
			AND deleted_at IS NULL
		ORDER BY start_time DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID, employeeID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trips := make([]*model.BusinessTrip, 0)
	for rows.Next() {
		trip := &model.BusinessTrip{}
		err := rows.Scan(
			&trip.ID, &trip.TenantID, &trip.EmployeeID, &trip.EmployeeName, &trip.DepartmentID,
			&trip.StartTime, &trip.EndTime, &trip.Duration,
			&trip.Destination, &trip.Transportation, &trip.Accommodation, &trip.Companions,
			&trip.Purpose, &trip.Tasks,
			&trip.EstimatedCost, &trip.ActualCost,
			&trip.ApprovalID, &trip.ApprovalStatus, &trip.ApprovedBy, &trip.ApprovedAt, &trip.RejectReason,
			&trip.Report, &trip.ReportAt,
			&trip.Remark, &trip.CreatedAt, &trip.UpdatedAt, &trip.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	return trips, rows.Err()
}

// FindPending 查询待审批的出差
func (r *businessTripRepository) FindPending(ctx context.Context, tenantID uuid.UUID) ([]*model.BusinessTrip, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			report, report_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_business_trips
		WHERE tenant_id = $1 AND approval_status = 'pending' AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trips := make([]*model.BusinessTrip, 0)
	for rows.Next() {
		trip := &model.BusinessTrip{}
		err := rows.Scan(
			&trip.ID, &trip.TenantID, &trip.EmployeeID, &trip.EmployeeName, &trip.DepartmentID,
			&trip.StartTime, &trip.EndTime, &trip.Duration,
			&trip.Destination, &trip.Transportation, &trip.Accommodation, &trip.Companions,
			&trip.Purpose, &trip.Tasks,
			&trip.EstimatedCost, &trip.ActualCost,
			&trip.ApprovalID, &trip.ApprovalStatus, &trip.ApprovedBy, &trip.ApprovedAt, &trip.RejectReason,
			&trip.Report, &trip.ReportAt,
			&trip.Remark, &trip.CreatedAt, &trip.UpdatedAt, &trip.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	return trips, rows.Err()
}

// FindOverlapping 查询时间重叠的出差记录
func (r *businessTripRepository) FindOverlapping(ctx context.Context, tenantID, employeeID uuid.UUID, startTime, endTime time.Time) ([]*model.BusinessTrip, error) {
	query := `
		SELECT id, tenant_id, employee_id, employee_name, department_id,
			start_time, end_time, duration,
			destination, transportation, accommodation, companions,
			purpose, tasks,
			estimated_cost, actual_cost,
			approval_id, approval_status, approved_by, approved_at, reject_reason,
			report, report_at,
			remark, created_at, updated_at, deleted_at
		FROM hrm_business_trips
		WHERE tenant_id = $1 AND employee_id = $2
			AND approval_status IN ('pending', 'approved')
			AND (
				(start_time <= $3 AND end_time >= $3) OR
				(start_time <= $4 AND end_time >= $4) OR
				(start_time >= $3 AND end_time <= $4)
			)
			AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	rows, err := r.db.Query(ctx, query, tenantID, employeeID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trips := make([]*model.BusinessTrip, 0)
	for rows.Next() {
		trip := &model.BusinessTrip{}
		err := rows.Scan(
			&trip.ID, &trip.TenantID, &trip.EmployeeID, &trip.EmployeeName, &trip.DepartmentID,
			&trip.StartTime, &trip.EndTime, &trip.Duration,
			&trip.Destination, &trip.Transportation, &trip.Accommodation, &trip.Companions,
			&trip.Purpose, &trip.Tasks,
			&trip.EstimatedCost, &trip.ActualCost,
			&trip.ApprovalID, &trip.ApprovalStatus, &trip.ApprovedBy, &trip.ApprovedAt, &trip.RejectReason,
			&trip.Report, &trip.ReportAt,
			&trip.Remark, &trip.CreatedAt, &trip.UpdatedAt, &trip.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	return trips, rows.Err()
}
