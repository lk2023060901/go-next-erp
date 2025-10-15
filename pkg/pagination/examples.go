package pagination

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// Example1_SimpleOffsetPagination 示例1：简单的offset分页（适用于小数据量）
func Example1_SimpleOffsetPagination(db *database.DB) error {
	ctx := context.Background()

	// 分页参数
	page := 1
	pageSize := 20
	offset := (page - 1) * pageSize

	// WHERE条件和参数
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{uuid.New()}

	// 1. COUNT查询
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM employees WHERE %s", where)
	var total int64
	err := db.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return err
	}

	// 2. 数据查询
	dataSQL := fmt.Sprintf(`
		SELECT id, name, email, created_at
		FROM employees
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := db.Query(ctx, dataSQL, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 3. 扫描结果
	type Employee struct {
		ID        uuid.UUID
		Name      string
		Email     string
		CreatedAt time.Time
	}

	employees := make([]Employee, 0, pageSize)
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.CreatedAt); err != nil {
			return err
		}
		employees = append(employees, emp)
	}

	fmt.Printf("Total: %d, Page: %d/%d, Items: %d\n",
		total, page, (total+int64(pageSize)-1)/int64(pageSize), len(employees))

	return nil
}

// Example2_OptimizedPagination 示例2：优化的分页（使用分页助手）
func Example2_OptimizedPagination(db *database.DB) error {
	ctx := context.Background()
	paginator := NewPaginator(ctx, db)

	// 分页参数
	limit := 20
	offset := 0

	// WHERE条件
	tenantID := uuid.New()
	where := "tenant_id = $1 AND deleted_at IS NULL"
	args := []interface{}{tenantID}

	// COUNT查询
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM employees WHERE %s", where)

	// 数据查询
	dataSQL := fmt.Sprintf(`
		SELECT id, name, email, created_at
		FROM employees
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	// 扫描函数
	scanFunc := func(rows pgx.Rows) (interface{}, error) {
		type Employee struct {
			ID        uuid.UUID
			Name      string
			Email     string
			CreatedAt time.Time
		}

		var emp Employee
		err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.CreatedAt)
		return emp, err
	}

	// 执行分页查询
	result, err := paginator.Paginate(dataSQL, countSQL, args, limit, offset, scanFunc)
	if err != nil {
		return err
	}

	fmt.Printf("Total: %d, Page: %d/%d, HasNext: %v\n",
		result.Total, result.Page, result.TotalPages, result.HasNext)

	return nil
}

// Example3_CursorPagination 示例3：游标分页（大数据量优化）
func Example3_CursorPagination(db *database.DB) error {
	ctx := context.Background()

	tenantID := uuid.New()
	pageSize := 20
	var lastCreatedAt *time.Time // 游标值（上一页最后一条记录的created_at）

	// 构建SQL
	var sql string
	var args []interface{}

	if lastCreatedAt == nil {
		// 首次查询
		sql = `
			SELECT id, name, email, created_at
			FROM employees
			WHERE tenant_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []interface{}{tenantID, pageSize + 1} // +1用于判断是否有下一页
	} else {
		// 基于游标查询
		sql = `
			SELECT id, name, email, created_at
			FROM employees
			WHERE tenant_id = $1 AND deleted_at IS NULL
			  AND created_at < $2
			ORDER BY created_at DESC, id DESC
			LIMIT $3
		`
		args = []interface{}{tenantID, lastCreatedAt, pageSize + 1}
	}

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Employee struct {
		ID        uuid.UUID
		Name      string
		Email     string
		CreatedAt time.Time
	}

	employees := make([]Employee, 0, pageSize)
	for rows.Next() {
		var emp Employee
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.CreatedAt); err != nil {
			return err
		}
		employees = append(employees, emp)
	}

	// 判断是否有下一页
	hasNext := len(employees) > pageSize
	if hasNext {
		employees = employees[:pageSize]
	}

	// 保存游标（下一次查询使用）
	if len(employees) > 0 {
		lastCreatedAt = &employees[len(employees)-1].CreatedAt
	}

	fmt.Printf("Items: %d, HasNext: %v, LastCursor: %v\n",
		len(employees), hasNext, lastCreatedAt)

	return nil
}

// Example4_EstimatedCount 示例4：估算总数（大表优化）
func Example4_EstimatedCount(db *database.DB) error {
	ctx := context.Background()

	// 对于超大表，精确COUNT代价太高
	// 使用PostgreSQL统计信息估算
	tableName := "employees"

	var estimate int64
	err := db.QueryRow(ctx, `
		SELECT reltuples::bigint AS estimate
		FROM pg_class
		WHERE relname = $1
	`, tableName).Scan(&estimate)

	if err != nil {
		return err
	}

	fmt.Printf("Estimated total: ~%d rows\n", estimate)

	// 配合LIMIT优化的COUNT
	maxCount := int64(10000)
	limitedCountSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT 1 FROM employees 
			WHERE deleted_at IS NULL 
			LIMIT %d
		) limited
	`, maxCount)

	var count int64
	err = db.QueryRow(ctx, limitedCountSQL).Scan(&count)
	if err != nil {
		return err
	}

	if count >= maxCount {
		fmt.Printf("Total: %d+ rows (more than %d)\n", count, maxCount)
	} else {
		fmt.Printf("Total: %d rows\n", count)
	}

	return nil
}

// Example5_OptimizedRepository 示例5：优化的Repository实现
type EmployeeRepository struct {
	db *database.DB
}

func (r *EmployeeRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]Employee, int64, error) {
	offset := (page - 1) * pageSize

	// 优化1：并发执行COUNT和数据查询
	var total int64
	var employees []Employee
	var countErr, dataErr error

	// 使用channel并发执行
	countCh := make(chan struct{})
	dataCh := make(chan struct{})

	// COUNT查询
	go func() {
		defer close(countCh)
		countErr = r.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM employees 
			WHERE tenant_id = $1 AND deleted_at IS NULL
		`, tenantID).Scan(&total)
	}()

	// 数据查询
	go func() {
		defer close(dataCh)
		rows, err := r.db.Query(ctx, `
			SELECT id, name, email, created_at
			FROM employees
			WHERE tenant_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, tenantID, pageSize, offset)

		if err != nil {
			dataErr = err
			return
		}
		defer rows.Close()

		employees = make([]Employee, 0, pageSize)
		for rows.Next() {
			var emp Employee
			if err := rows.Scan(&emp.ID, &emp.Name, &emp.Email, &emp.CreatedAt); err != nil {
				dataErr = err
				return
			}
			employees = append(employees, emp)
		}
		dataErr = rows.Err()
	}()

	// 等待两个查询完成
	<-countCh
	<-dataCh

	if countErr != nil {
		return nil, 0, countErr
	}
	if dataErr != nil {
		return nil, 0, dataErr
	}

	return employees, total, nil
}

type Employee struct {
	ID        uuid.UUID
	Name      string
	Email     string
	CreatedAt time.Time
}
