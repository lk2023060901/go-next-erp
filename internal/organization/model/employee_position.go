package model

import (
	"time"

	"github.com/google/uuid"
)

// EmployeePosition 员工职位关联（支持一人多职）
type EmployeePosition struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null"`
	EmployeeID uuid.UUID `gorm:"type:uuid;not null;index:idx_emp_position_employee"`
	PositionID uuid.UUID `gorm:"type:uuid;not null;index:idx_emp_position_position"`
	OrgID      uuid.UUID `gorm:"type:uuid;not null;index:idx_emp_position_org"` // 该职位所在组织

	// 是否主职位
	IsPrimary bool `gorm:"default:false"`

	// 生效时间
	StartDate *time.Time
	EndDate   *time.Time

	// 审计字段
	CreatedBy uuid.UUID  `gorm:"type:uuid"`
	CreatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `gorm:"index"`
}

// TableName 指定表名
func (EmployeePosition) TableName() string {
	return "employee_positions"
}

// IsActive 是否有效
func (ep *EmployeePosition) IsActive() bool {
	now := time.Now()
	if ep.StartDate != nil && now.Before(*ep.StartDate) {
		return false
	}
	if ep.EndDate != nil && now.After(*ep.EndDate) {
		return false
	}
	return true
}
