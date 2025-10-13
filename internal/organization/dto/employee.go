package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/organization/model"
)

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
	UserID         string     `json:"user_id" binding:"required,uuid"`
	EmployeeNo     string     `json:"employee_no" binding:"required,max=50"`
	Name           string     `json:"name" binding:"required,max=100"`
	Gender         string     `json:"gender" binding:"omitempty,oneof=male female"`
	Mobile         string     `json:"mobile" binding:"max=20"`
	Email          string     `json:"email" binding:"omitempty,email,max=100"`
	Avatar         string     `json:"avatar" binding:"max=500"`
	OrgID          string     `json:"org_id" binding:"required,uuid"`
	PositionID     *string    `json:"position_id" binding:"omitempty,uuid"`
	DirectLeaderID *string    `json:"direct_leader_id" binding:"omitempty,uuid"`
	JoinDate       *time.Time `json:"join_date"`
	ProbationEnd   *time.Time `json:"probation_end"`
	Status         string     `json:"status" binding:"required,oneof=probation active"`
}

// UpdateEmployeeRequest 更新员工请求
type UpdateEmployeeRequest struct {
	Name           string  `json:"name" binding:"required,max=100"`
	Gender         string  `json:"gender" binding:"omitempty,oneof=male female"`
	Mobile         string  `json:"mobile" binding:"max=20"`
	Email          string  `json:"email" binding:"omitempty,email,max=100"`
	Avatar         string  `json:"avatar" binding:"max=500"`
	DirectLeaderID *string `json:"direct_leader_id" binding:"omitempty,uuid"`
}

// TransferEmployeeRequest 调岗请求
type TransferEmployeeRequest struct {
	NewOrgID      string  `json:"new_org_id" binding:"required,uuid"`
	NewPositionID *string `json:"new_position_id" binding:"omitempty,uuid"`
}

// ChangePositionRequest 更换职位请求
type ChangePositionRequest struct {
	NewPositionID string `json:"new_position_id" binding:"required,uuid"`
}

// ChangeLeaderRequest 更换上级请求
type ChangeLeaderRequest struct {
	NewLeaderID string `json:"new_leader_id" binding:"required,uuid"`
}

// RegularizeRequest 转正请求
type RegularizeRequest struct {
	FormalDate time.Time `json:"formal_date" binding:"required"`
}

// ResignRequest 离职请求
type ResignRequest struct {
	LeaveDate time.Time `json:"leave_date" binding:"required"`
}

// EmployeeResponse 员工响应
type EmployeeResponse struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	UserID         uuid.UUID  `json:"user_id"`
	EmployeeNo     string     `json:"employee_no"`
	Name           string     `json:"name"`
	Gender         string     `json:"gender,omitempty"`
	Mobile         string     `json:"mobile,omitempty"`
	Email          string     `json:"email,omitempty"`
	Avatar         string     `json:"avatar,omitempty"`
	OrgID          uuid.UUID  `json:"org_id"`
	OrgPath        string     `json:"org_path"`
	PositionID     *uuid.UUID `json:"position_id,omitempty"`
	DirectLeaderID *uuid.UUID `json:"direct_leader_id,omitempty"`
	JoinDate       *time.Time `json:"join_date,omitempty"`
	ProbationEnd   *time.Time `json:"probation_end,omitempty"`
	FormalDate     *time.Time `json:"formal_date,omitempty"`
	LeaveDate      *time.Time `json:"leave_date,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ToEmployeeResponse 转换为响应对象
func ToEmployeeResponse(emp *model.Employee) *EmployeeResponse {
	return &EmployeeResponse{
		ID:             emp.ID,
		TenantID:       emp.TenantID,
		UserID:         emp.UserID,
		EmployeeNo:     emp.EmployeeNo,
		Name:           emp.Name,
		Gender:         emp.Gender,
		Mobile:         emp.Mobile,
		Email:          emp.Email,
		Avatar:         emp.Avatar,
		OrgID:          emp.OrgID,
		OrgPath:        emp.OrgPath,
		PositionID:     emp.PositionID,
		DirectLeaderID: emp.DirectLeaderID,
		JoinDate:       emp.JoinDate,
		ProbationEnd:   emp.ProbationEnd,
		FormalDate:     emp.FormalDate,
		LeaveDate:      emp.LeaveDate,
		Status:         emp.Status,
		CreatedAt:      emp.CreatedAt,
		UpdatedAt:      emp.UpdatedAt,
	}
}

// ToEmployeeResponseList 批量转换
func ToEmployeeResponseList(emps []*model.Employee) []*EmployeeResponse {
	result := make([]*EmployeeResponse, len(emps))
	for i, emp := range emps {
		result[i] = ToEmployeeResponse(emp)
	}
	return result
}
