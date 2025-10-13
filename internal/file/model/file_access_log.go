package model

import (
"time"

"github.com/google/uuid"
)

// FileAccessLog 文件访问日志模型
type FileAccessLog struct {
	ID       uuid.UUID `json:"id"`
	FileID   uuid.UUID `json:"file_id"`
	TenantID uuid.UUID `json:"tenant_id"`

	// 访问信息
	Action    string     `json:"action"`     // download/preview/delete/update/view
	UserID    *uuid.UUID `json:"user_id"`    // 用户 ID（匿名访问为 NULL）
	IPAddress *string    `json:"ip_address"` // IP 地址
	UserAgent *string    `json:"user_agent"` // User-Agent

	// 结果
	Success      bool    `json:"success"`       // 是否成功
	ErrorMessage *string `json:"error_message"` // 错误信息

	// 元数据
	Metadata map[string]interface{} `json:"metadata"` // 额外上下文信息

	// 时间戳
	CreatedAt time.Time `json:"created_at"`
}

// AccessAction 访问操作类型
type AccessAction string

const (
ActionDownload AccessAction = "download" // 下载
ActionPreview  AccessAction = "preview"  // 预览
ActionDelete   AccessAction = "delete"   // 删除
ActionUpdate   AccessAction = "update"   // 更新
ActionView     AccessAction = "view"     // 查看（获取信息）
ActionCopy     AccessAction = "copy"     // 复制
ActionMove     AccessAction = "move"     // 移动
ActionShare    AccessAction = "share"    // 分享
)
