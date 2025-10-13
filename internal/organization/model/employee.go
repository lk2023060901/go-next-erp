package model

import (
	"time"

	"github.com/google/uuid"
)

// Employee 员工
// 关联用户表（auth.users），记录员工的组织和职位信息
type Employee struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	UserID   uuid.UUID `json:"user_id"` // 关联用户表

	// 员工信息
	EmployeeNo string `json:"employee_no"` // 工号
	Name       string `json:"name"`        // 姓名
	Gender     string `json:"gender"`      // 性别：male, female
	Mobile     string `json:"mobile"`      // 手机号
	Email      string `json:"email"`       // 邮箱
	Avatar     string `json:"avatar"`      // 头像 URL

	// 组织关系
	OrgID      uuid.UUID  `json:"org_id"`      // 所属组织
	OrgPath    string     `json:"org_path"`    // 组织路径（用于数据权限）
	PositionID *uuid.UUID `json:"position_id"` // 主职位

	// 汇报关系
	DirectLeaderID *uuid.UUID `json:"direct_leader_id"` // 直接上级

	// 入职信息
	JoinDate     *time.Time `json:"join_date"`     // 入职日期
	ProbationEnd *time.Time `json:"probation_end"` // 试用期结束日期
	FormalDate   *time.Time `json:"formal_date"`   // 转正日期
	LeaveDate    *time.Time `json:"leave_date"`    // 离职日期

	// 状态
	Status string `json:"status"` // probation(试用), active(在职), resigned(离职)

	// 审计字段
	CreatedBy uuid.UUID  `json:"created_by"`
	UpdatedBy uuid.UUID  `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

// IsActive 是否在职
func (e *Employee) IsActive() bool {
	return e.Status == "active" || e.Status == "probation"
}

// IsProbation 是否试用期
func (e *Employee) IsProbation() bool {
	return e.Status == "probation"
}

// IsResigned 是否已离职
func (e *Employee) IsResigned() bool {
	return e.Status == "resigned"
}
