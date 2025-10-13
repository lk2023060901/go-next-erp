package model

import (
	"time"

	"github.com/google/uuid"
)

// ChangeType 文件变更类型
type ChangeType string

const (
	ChangeTypeCreate ChangeType = "create" // 创建
	ChangeTypeUpdate ChangeType = "update" // 更新
	ChangeTypeRevert ChangeType = "revert" // 回滚
)

// FileVersion 文件版本
type FileVersion struct {
	ID       uuid.UUID `json:"id"`
	FileID   uuid.UUID `json:"file_id"`   // 关联文件 ID
	TenantID uuid.UUID `json:"tenant_id"`

	// Version info
	VersionNumber int    `json:"version_number"` // 版本号
	StorageKey    string `json:"storage_key"`    // 存储键
	Size          int64  `json:"size"`           // 文件大小
	Checksum      string `json:"checksum"`       // SHA-256 哈希

	// Metadata
	Filename string `json:"filename"`  // 文件名
	MimeType string `json:"mime_type"` // MIME 类型
	Comment  string `json:"comment"`   // 版本说明

	// Change tracking
	ChangedBy  uuid.UUID  `json:"changed_by"`  // 创建版本的用户
	ChangeType ChangeType `json:"change_type"` // 变更类型

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
}

// IsCurrent 检查是否为当前版本
func (fv *FileVersion) IsCurrent(file *File) bool {
	return fv.VersionNumber == file.VersionNumber
}
