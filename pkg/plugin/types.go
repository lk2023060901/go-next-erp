package plugin

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PluginStatus 插件状态
type PluginStatus string

const (
	PluginStatusInstalled   PluginStatus = "installed"   // 已安装
	PluginStatusEnabled     PluginStatus = "enabled"     // 已启用
	PluginStatusDisabled    PluginStatus = "disabled"    // 已禁用
	PluginStatusUpdating    PluginStatus = "updating"    // 更新中
	PluginStatusUninstalled PluginStatus = "uninstalled" // 已卸载
)

// PluginMetadata 插件元数据
// 定义插件的基本信息、权限、依赖等
type PluginMetadata struct {
	// 基础信息
	ID          string `json:"id" yaml:"id"`                     // 插件唯一标识（如：com.example.crm）
	Name        string `json:"name" yaml:"name"`                 // 插件名称
	DisplayName string `json:"display_name" yaml:"display_name"` // 显示名称
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`           // 语义化版本号
	Author      string `json:"author" yaml:"author"`             // 作者（个人/公司）
	AuthorEmail string `json:"author_email" yaml:"author_email"`
	Homepage    string `json:"homepage" yaml:"homepage"`
	License     string `json:"license" yaml:"license"`           // MIT, Apache-2.0, Commercial
	Icon        string `json:"icon" yaml:"icon"`                 // 图标 URL

	// 分类和标签（开放式，支持自定义）
	Categories []string `json:"categories" yaml:"categories"` // 多分类支持，如 ["财务", "报表", "第三方集成"]
	Tags       []string `json:"tags" yaml:"tags"`             // 标签，如 ["发票", "支付", "微信"]
	Industry   []string `json:"industry" yaml:"industry"`     // 适用行业，如 ["零售", "制造", "医疗"]

	// 依赖和兼容性
	Dependencies    []Dependency `json:"dependencies" yaml:"dependencies"`
	MinCoreVersion  string       `json:"min_core_version" yaml:"min_core_version"`   // 最低系统版本
	MaxCoreVersion  string       `json:"max_core_version" yaml:"max_core_version"`   // 最高系统版本
	Conflicts       []string     `json:"conflicts" yaml:"conflicts"`                 // 冲突的插件 ID
	RequiredPlugins []string     `json:"required_plugins" yaml:"required_plugins"`   // 必需的其他插件

	// 权限定义（核心）
	Permissions []PermissionDefinition `json:"permissions" yaml:"permissions"`

	// 模块定义
	Modules []ModuleDefinition `json:"modules" yaml:"modules"`

	// 数据库迁移
	Migrations []MigrationDefinition `json:"migrations" yaml:"migrations"`

	// 配置 Schema
	ConfigSchema map[string]interface{} `json:"config_schema" yaml:"config_schema"` // JSON Schema

	// 生命周期钩子
	Hooks PluginHooks `json:"hooks" yaml:"hooks"`

	// API 端点（插件提供的 HTTP API）
	Endpoints []EndpointDefinition `json:"endpoints" yaml:"endpoints"`

	// 事件监听（插件订阅的系统事件）
	EventListeners []EventListener `json:"event_listeners" yaml:"event_listeners"`

	// 费用信息（插件商店）
	Pricing *PricingInfo `json:"pricing,omitempty" yaml:"pricing,omitempty"`

	// 审计和安全
	ChecksumSHA256 string    `json:"checksum_sha256" yaml:"checksum_sha256"` // 插件包校验和
	SignedBy       string    `json:"signed_by" yaml:"signed_by"`             // 签名者公钥
	VerifiedBy     string    `json:"verified_by" yaml:"verified_by"`         // 平台审核者
	SecurityLevel  string    `json:"security_level" yaml:"security_level"`   // safe, trusted, unverified
	PublishedAt    time.Time `json:"published_at" yaml:"published_at"`

	// 商店信息（用于插件市场展示）
	Screenshots []string `json:"screenshots" yaml:"screenshots"`       // 截图 URL
	Video       string   `json:"video" yaml:"video"`                   // 视频介绍 URL
	Changelog   string   `json:"changelog" yaml:"changelog"`           // 更新日志
	Documentation string `json:"documentation" yaml:"documentation"`   // 文档 URL
	SupportEmail  string `json:"support_email" yaml:"support_email"`   // 支持邮箱
	SupportURL    string `json:"support_url" yaml:"support_url"`       // 支持页面
}

// Dependency 依赖定义
type Dependency struct {
	Name    string `json:"name" yaml:"name"`
	Version string `json:"version" yaml:"version"` // 语义化版本约束，如 ">=1.0.0 <2.0.0"
	Type    string `json:"type" yaml:"type"`       // npm, go, python, service, plugin
}

// PermissionDefinition 权限定义
// 插件需要声明需要哪些权限
type PermissionDefinition struct {
	Resource    string   `json:"resource" yaml:"resource"`       // 资源名称（如：customer, order）
	Actions     []string `json:"actions" yaml:"actions"`         // 操作列表
	DisplayName string   `json:"display_name" yaml:"display_name"`
	Description string   `json:"description" yaml:"description"`
	Required    bool     `json:"required" yaml:"required"`       // 是否必需（拒绝则无法安装）
	Scope       string   `json:"scope" yaml:"scope"`             // tenant, global, user
	Sensitive   bool     `json:"sensitive" yaml:"sensitive"`     // 是否为敏感权限（需特别授权）
}

// ModuleDefinition 模块定义
// 插件提供的功能模块（前端菜单/页面）
type ModuleDefinition struct {
	ID          string                 `json:"id" yaml:"id"`                     // 模块 ID
	Name        string                 `json:"name" yaml:"name"`                 // 模块名称
	DisplayName string                 `json:"display_name" yaml:"display_name"` // 显示名称
	Description string                 `json:"description" yaml:"description"`
	Icon        string                 `json:"icon" yaml:"icon"`
	Route       string                 `json:"route" yaml:"route"`               // 前端路由
	Component   string                 `json:"component" yaml:"component"`       // 前端组件路径
	Order       int                    `json:"order" yaml:"order"`               // 排序
	Parent      string                 `json:"parent" yaml:"parent"`             // 父模块 ID
	Enabled     bool                   `json:"enabled" yaml:"enabled"`
	Permissions []string               `json:"permissions" yaml:"permissions"`   // 访问该模块需要的权限
	Config      map[string]interface{} `json:"config" yaml:"config"`
}

// MigrationDefinition 数据库迁移定义
type MigrationDefinition struct {
	Version     string    `json:"version" yaml:"version"`         // 迁移版本号
	Description string    `json:"description" yaml:"description"`
	UpSQL       string    `json:"up_sql" yaml:"up_sql"`           // 升级 SQL
	DownSQL     string    `json:"down_sql" yaml:"down_sql"`       // 回滚 SQL
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
}

// EndpointDefinition API 端点定义
// 插件提供的 HTTP API 端点
type EndpointDefinition struct {
	Path        string   `json:"path" yaml:"path"`               // API 路径（如：/api/v1/crm/customers）
	Method      string   `json:"method" yaml:"method"`           // HTTP 方法
	Handler     string   `json:"handler" yaml:"handler"`         // 处理器名称
	Permissions []string `json:"permissions" yaml:"permissions"` // 需要的权限
	Description string   `json:"description" yaml:"description"`
	Public      bool     `json:"public" yaml:"public"`           // 是否公开（无需认证）
}

// EventListener 事件监听器定义
// 插件订阅的系统事件
type EventListener struct {
	Event       string `json:"event" yaml:"event"`             // 事件名称（如：user.created, order.paid）
	Handler     string `json:"handler" yaml:"handler"`         // 处理器名称
	Priority    int    `json:"priority" yaml:"priority"`       // 优先级（数字越小越优先）
	Async       bool   `json:"async" yaml:"async"`             // 是否异步处理
	Description string `json:"description" yaml:"description"`
}

// PluginHooks 插件生命周期钩子
type PluginHooks struct {
	OnInstall   string `json:"on_install" yaml:"on_install"`     // 安装时执行的脚本/方法
	OnUninstall string `json:"on_uninstall" yaml:"on_uninstall"` // 卸载时执行
	OnEnable    string `json:"on_enable" yaml:"on_enable"`       // 启用时执行
	OnDisable   string `json:"on_disable" yaml:"on_disable"`     // 禁用时执行
	OnUpdate    string `json:"on_update" yaml:"on_update"`       // 更新时执行
	OnConfigure string `json:"on_configure" yaml:"on_configure"` // 配置变更时执行
}

// PricingInfo 价格信息
type PricingInfo struct {
	Type          string  `json:"type"`           // free, one-time, subscription, usage-based
	Price         float64 `json:"price"`          // 价格
	Currency      string  `json:"currency"`       // 货币（USD, CNY, EUR）
	Interval      string  `json:"interval"`       // month, year（订阅模式）
	TrialDays     int     `json:"trial_days"`     // 试用天数
	FreeTier      bool    `json:"free_tier"`      // 是否有免费版
	FreeTierLimit string  `json:"free_tier_limit"` // 免费版限制说明
}

// PluginInstance 插件实例（租户级别）
// 记录某个租户安装的插件及其配置
type PluginInstance struct {
	ID       uuid.UUID              `json:"id"`
	TenantID uuid.UUID              `json:"tenant_id"`
	PluginID string                 `json:"plugin_id"` // 引用 PluginMetadata.ID
	Version  string                 `json:"version"`
	Status   PluginStatus           `json:"status"`
	Config   map[string]interface{} `json:"config"`    // 插件配置（覆盖默认值）

	// 许可证信息
	LicenseKey    string     `json:"license_key,omitempty"`
	LicenseExpiry *time.Time `json:"license_expiry,omitempty"`
	LicensedTo    string     `json:"licensed_to,omitempty"` // 许可持有者

	// 安装信息
	InstalledBy uuid.UUID  `json:"installed_by"` // 安装人
	InstallSource string   `json:"install_source"` // marketplace, manual, git

	// 时间戳
	InstalledAt time.Time  `json:"installed_at"`
	EnabledAt   *time.Time `json:"enabled_at,omitempty"`
	DisabledAt  *time.Time `json:"disabled_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Plugin 插件接口
// 所有插件必须实现此接口
type Plugin interface {
	// GetMetadata 获取插件元数据
	GetMetadata() *PluginMetadata

	// Initialize 初始化插件
	// ctx: 上下文
	// config: 租户配置
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// Validate 验证插件配置
	Validate(config map[string]interface{}) error

	// HandleHook 处理生命周期钩子
	HandleHook(ctx context.Context, hook string, data map[string]interface{}) error

	// HandleEvent 处理系统事件
	HandleEvent(ctx context.Context, event string, data map[string]interface{}) error
}

// PluginRegistry 插件注册表接口
type PluginRegistry interface {
	// Register 注册插件
	Register(plugin Plugin) error

	// Unregister 注销插件
	Unregister(pluginID string) error

	// Get 获取插件
	Get(pluginID string) (Plugin, error)

	// List 列出所有已注册插件
	List() []Plugin

	// ListByCategory 按分类列出插件
	ListByCategory(category string) []Plugin

	// ListByTag 按标签列出插件
	ListByTag(tag string) []Plugin

	// Search 搜索插件
	Search(query string) []Plugin

	// Exists 检查插件是否存在
	Exists(pluginID string) bool
}

// PluginManager 插件管理器接口
type PluginManager interface {
	// Install 安装插件
	Install(ctx context.Context, tenantID uuid.UUID, pluginID, version string, config map[string]interface{}) error

	// Uninstall 卸载插件
	Uninstall(ctx context.Context, tenantID uuid.UUID, pluginID string) error

	// Enable 启用插件
	Enable(ctx context.Context, tenantID uuid.UUID, pluginID string) error

	// Disable 禁用插件
	Disable(ctx context.Context, tenantID uuid.UUID, pluginID string) error

	// Update 更新插件
	Update(ctx context.Context, tenantID uuid.UUID, pluginID, newVersion string) error

	// GetInstance 获取插件实例
	GetInstance(ctx context.Context, tenantID uuid.UUID, pluginID string) (*PluginInstance, error)

	// ListInstances 列出租户的所有插件
	ListInstances(ctx context.Context, tenantID uuid.UUID) ([]*PluginInstance, error)

	// Configure 配置插件
	Configure(ctx context.Context, tenantID uuid.UUID, pluginID string, config map[string]interface{}) error

	// ValidateLicense 验证许可证
	ValidateLicense(ctx context.Context, tenantID uuid.UUID, pluginID string) error

	// GetPermissions 获取插件权限
	GetPermissions(ctx context.Context, pluginID string) ([]PermissionDefinition, error)

	// GrantPermissions 授权插件权限（租户确认）
	GrantPermissions(ctx context.Context, tenantID uuid.UUID, pluginID string, permissions []string) error
}

// PluginStore 插件商店接口
type PluginStore interface {
	// Search 搜索插件
	Search(ctx context.Context, query string, categories []string, tags []string, limit, offset int) ([]*PluginMetadata, error)

	// GetPlugin 获取插件详情
	GetPlugin(ctx context.Context, pluginID string) (*PluginMetadata, error)

	// ListVersions 列出插件的所有版本
	ListVersions(ctx context.Context, pluginID string) ([]string, error)

	// Download 下载插件
	Download(ctx context.Context, pluginID, version string) ([]byte, error)

	// Publish 发布插件（开发者）
	Publish(ctx context.Context, metadata *PluginMetadata, packageData []byte) error

	// Update 更新插件信息
	Update(ctx context.Context, pluginID string, metadata *PluginMetadata) error

	// GetStats 获取插件统计
	GetStats(ctx context.Context, pluginID string) (*PluginStats, error)

	// ListCategories 列出所有分类（动态生成）
	ListCategories(ctx context.Context) ([]string, error)

	// ListTags 列出所有标签（动态生成）
	ListTags(ctx context.Context) ([]string, error)

	// Review 评价插件
	Review(ctx context.Context, pluginID string, userID uuid.UUID, rating int, comment string) error
}

// PluginStats 插件统计信息
type PluginStats struct {
	PluginID        string    `json:"plugin_id"`
	Downloads       int64     `json:"downloads"`
	ActiveInstalls  int64     `json:"active_installs"`
	AverageRating   float64   `json:"average_rating"`
	TotalReviews    int64     `json:"total_reviews"`
	LastUpdated     time.Time `json:"last_updated"`
	PopularityScore float64   `json:"popularity_score"` // 综合评分
}

// PluginEvent 插件事件（用于审计和通知）
type PluginEvent struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	PluginID   string                 `json:"plugin_id"`
	Event      string                 `json:"event"` // installed, uninstalled, enabled, disabled, updated, configured
	Version    string                 `json:"version"`
	UserID     uuid.UUID              `json:"user_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	OccurredAt time.Time              `json:"occurred_at"`
}

// PluginReview 插件评价
type PluginReview struct {
	ID         uuid.UUID `json:"id"`
	PluginID   string    `json:"plugin_id"`
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Rating     int       `json:"rating"`     // 1-5 星
	Comment    string    `json:"comment"`
	Version    string    `json:"version"`    // 评价的版本
	Helpful    int       `json:"helpful"`    // 有帮助的票数
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
