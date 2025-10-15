package pagination

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int    `json:"page" form:"page" validate:"min=1"`                    // 页码（从1开始）
	PageSize int    `json:"page_size" form:"page_size" validate:"min=1,max=1000"` // 每页大小
	SortBy   string `json:"sort_by" form:"sort_by"`                               // 排序字段
	SortDesc bool   `json:"sort_desc" form:"sort_desc"`                           // 是否降序

	// 游标分页（可选，性能更好）
	Cursor    string `json:"cursor,omitempty" form:"cursor"`         // 游标（base64编码）
	UseCursor bool   `json:"use_cursor,omitempty" form:"use_cursor"` // 是否使用游标分页
}

// PageResponse 分页响应（泛型）
type PageResponse[T any] struct {
	Items      []T    `json:"items"`                 // 数据列表
	Total      int64  `json:"total"`                 // 总记录数
	Page       int    `json:"page"`                  // 当前页
	PageSize   int    `json:"page_size"`             // 每页大小
	TotalPages int    `json:"total_pages"`           // 总页数
	HasNext    bool   `json:"has_next"`              // 是否有下一页
	HasPrev    bool   `json:"has_prev"`              // 是否有上一页
	NextCursor string `json:"next_cursor,omitempty"` // 下一页游标（游标分页）
	PrevCursor string `json:"prev_cursor,omitempty"` // 上一页游标（游标分页）
}

// CursorInfo 游标信息
type CursorInfo struct {
	LastID    uuid.UUID   `json:"last_id"`
	LastValue interface{} `json:"last_value"` // 排序字段的值
	Direction string      `json:"direction"`  // next/prev
}

// Paginator 分页器
type Paginator struct {
	ctx context.Context
	db  DB // 数据库接口
}

// DB 数据库接口
type DB interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// NewPaginator 创建分页器
func NewPaginator(ctx context.Context, db DB) *Paginator {
	return &Paginator{
		ctx: ctx,
		db:  db,
	}
}

// GetOffset 计算偏移量
func (p *PageRequest) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PageRequest) GetLimit() int {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 1000 {
		p.PageSize = 1000
	}
	return p.PageSize
}

// GetOrderBy 获取排序子句
func (p *PageRequest) GetOrderBy(defaultSort string) string {
	sortField := defaultSort
	if p.SortBy != "" {
		sortField = p.SortBy
	}

	direction := "ASC"
	if p.SortDesc {
		direction = "DESC"
	}

	return fmt.Sprintf("%s %s", sortField, direction)
}

// Paginate 执行分页查询（offset分页 - 适用于小数据量）
func (p *Paginator) Paginate(
	dataSQL string,
	countSQL string,
	args []interface{},
	limit, offset int,
	scanFunc func(rows pgx.Rows) (interface{}, error),
) (*PageResponse[interface{}], error) {
	// 1. 查询总数（可选优化：缓存或估算）
	var total int64
	err := p.db.QueryRow(p.ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count query failed: %w", err)
	}

	// 2. 如果总数为0，直接返回
	if total == 0 {
		return &PageResponse[interface{}]{
			Items:      []interface{}{},
			Total:      0,
			Page:       1,
			PageSize:   limit,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}

	// 3. 查询数据
	rows, err := p.db.Query(p.ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("data query failed: %w", err)
	}
	defer rows.Close()

	items := make([]interface{}, 0, limit)
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 4. 构建响应
	currentPage := (offset / limit) + 1
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &PageResponse[interface{}]{
		Items:      items,
		Total:      total,
		Page:       currentPage,
		PageSize:   limit,
		TotalPages: totalPages,
		HasNext:    currentPage < totalPages,
		HasPrev:    currentPage > 1,
	}, nil
}

// PaginateWithCursor 执行游标分页查询（适用于大数据量）
// 优点：性能稳定，不受偏移量影响
// 缺点：不能跳页，只能上一页/下一页
func (p *Paginator) PaginateWithCursor(
	baseSQL string,
	args []interface{},
	limit int,
	cursorField string, // 游标字段（必须有索引，如 created_at 或 id）
	cursorValue interface{}, // 游标值
	direction string, // next/prev
	scanFunc func(rows pgx.Rows) (interface{}, error),
) (*PageResponse[interface{}], error) {
	// 构建游标查询
	var sql string
	var queryArgs []interface{}

	if cursorValue == nil {
		// 首次查询
		sql = fmt.Sprintf("%s ORDER BY %s DESC LIMIT $%d", baseSQL, cursorField, len(args)+1)
		queryArgs = append(args, limit+1) // +1用于判断是否有下一页
	} else {
		// 基于游标查询
		operator := "<"
		order := "DESC"
		if direction == "prev" {
			operator = ">"
			order = "ASC"
		}

		sql = fmt.Sprintf(
			"%s AND %s %s $%d ORDER BY %s %s LIMIT $%d",
			baseSQL, cursorField, operator, len(args)+1, cursorField, order, len(args)+2,
		)
		queryArgs = append(args, cursorValue, limit+1)
	}

	// 执行查询
	rows, err := p.db.Query(p.ctx, sql, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("cursor query failed: %w", err)
	}
	defer rows.Close()

	items := make([]interface{}, 0, limit)
	for rows.Next() {
		item, err := scanFunc(rows)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 判断是否有更多数据
	hasNext := len(items) > limit
	if hasNext {
		items = items[:limit] // 移除多余的一条
	}

	// 游标分页不返回total（避免COUNT性能问题）
	return &PageResponse[interface{}]{
		Items:      items,
		Total:      -1, // 表示未知
		PageSize:   limit,
		HasNext:    hasNext,
		HasPrev:    cursorValue != nil,
		NextCursor: "", // TODO: 编码游标
		PrevCursor: "",
	}, nil
}

// EstimateTotal 估算总数（适用于大表，避免精确COUNT）
// 使用PostgreSQL的统计信息快速估算
func (p *Paginator) EstimateTotal(tableName string, whereClause string) (int64, error) {
	// 如果有WHERE条件，使用EXPLAIN估算
	if whereClause != "" {
		sql := fmt.Sprintf("EXPLAIN (FORMAT JSON) SELECT * FROM %s WHERE %s", tableName, whereClause)
		var jsonPlan string
		err := p.db.QueryRow(p.ctx, sql).Scan(&jsonPlan)
		if err != nil {
			return 0, err
		}
		// TODO: 解析EXPLAIN结果获取估算行数
		return 0, fmt.Errorf("not implemented")
	}

	// 无WHERE条件，直接从统计信息获取
	sql := `
		SELECT reltuples::bigint AS estimate
		FROM pg_class
		WHERE relname = $1
	`
	var estimate int64
	err := p.db.QueryRow(p.ctx, sql, tableName).Scan(&estimate)
	return estimate, err
}

// OptimizedCount 优化的COUNT查询
// 对于大表，如果offset+limit < 10000，只COUNT到需要的部分
func (p *Paginator) OptimizedCount(
	countSQL string,
	args []interface{},
	maxCount int64, // 最大计数限制，超过此值不精确计数
) (int64, bool, error) {
	// 方案1：如果只需要知道"是否超过N条"，使用LIMIT优化
	limitedSQL := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT 1 FROM (%s) t LIMIT %d) limited", countSQL, maxCount)

	var count int64
	err := p.db.QueryRow(p.ctx, limitedSQL, args...).Scan(&count)
	if err != nil {
		return 0, false, err
	}

	hasMore := count >= maxCount
	return count, hasMore, nil
}

// CacheKey 生成分页缓存键
func CacheKey(prefix, sql string, args []interface{}, page, pageSize int) string {
	// TODO: 实现缓存键生成逻辑
	return fmt.Sprintf("%s:%d:%d:%v", prefix, page, pageSize, args)
}

// SmartPaginate 智能分页（自动选择最优策略）
func (p *Paginator) SmartPaginate(
	req *PageRequest,
	baseSQL string,
	countSQL string,
	args []interface{},
	scanFunc func(rows pgx.Rows) (interface{}, error),
) (*PageResponse[interface{}], error) {
	limit := req.GetLimit()
	offset := req.GetOffset()

	// 策略1: 使用游标分页（性能最好）
	if req.UseCursor {
		// TODO: 实现游标分页
		return nil, fmt.Errorf("cursor pagination not implemented yet")
	}

	// 策略2: 小偏移量 - 使用标准offset分页
	if offset < 1000 {
		return p.Paginate(baseSQL, countSQL, args, limit, offset, scanFunc)
	}

	// 策略3: 大偏移量 - 建议使用游标或限制最大页数
	return nil, fmt.Errorf("offset %d too large, please use cursor pagination", offset)
}
