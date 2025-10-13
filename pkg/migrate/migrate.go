package migrate

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lk2023060901/go-next-erp/pkg/database"
)

// Migration 迁移定义
type Migration struct {
	Version     string
	Description string
	SQL         string
}

// Migrator 迁移管理器
type Migrator struct {
	db         *database.DB
	migrations []Migration
}

// New 创建迁移管理器
func New(db *database.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

// LoadFromFS 从嵌入的文件系统加载迁移
func (m *Migrator) LoadFromFS(fs embed.FS, dir string) error {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(dir + "/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", entry.Name(), err)
		}

		// 从文件名解析版本号和描述
		// 格式: 001_description.sql
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", entry.Name())
		}

		version := parts[0]
		description := strings.TrimSuffix(parts[1], ".sql")

		m.migrations = append(m.migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// 按版本号排序
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// LoadFromDir 从文件系统目录加载迁移
func (m *Migrator) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", entry.Name(), err)
		}

		// 从文件名解析版本号和描述
		// 格式: 001_description.sql
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename: %s", entry.Name())
		}

		version := parts[0]
		description := strings.TrimSuffix(parts[1], ".sql")

		m.migrations = append(m.migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// 按版本号排序
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up(ctx context.Context) error {
	// 确保迁移表存在
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}

	// 获取已执行的迁移
	executed, err := m.getExecutedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("get executed migrations: %w", err)
	}

	// 执行未执行的迁移
	for _, migration := range m.migrations {
		if _, exists := executed[migration.Version]; exists {
			continue
		}

		fmt.Printf("Running migration %s: %s\n", migration.Version, migration.Description)

		if err := m.executeMigration(ctx, migration); err != nil {
			return fmt.Errorf("execute migration %s: %w", migration.Version, err)
		}

		fmt.Printf("✓ Migration %s completed\n", migration.Version)
	}

	return nil
}

// Status 显示迁移状态
func (m *Migrator) Status(ctx context.Context) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}

	executed, err := m.getExecutedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("get executed migrations: %w", err)
	}

	fmt.Println("Migration Status:")
	fmt.Println("================")

	for _, migration := range m.migrations {
		status := "pending"
		if info, exists := executed[migration.Version]; exists {
			status = fmt.Sprintf("applied at %s", info.AppliedAt.Format(time.RFC3339))
		}

		fmt.Printf("[%s] %s: %s (%s)\n",
			migration.Version,
			migration.Description,
			status,
			"")
	}

	return nil
}

// ensureMigrationTable 确保迁移表存在
func (m *Migrator) ensureMigrationTable(ctx context.Context) error {
	_, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			description TEXT,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getExecutedMigrations 获取已执行的迁移
func (m *Migrator) getExecutedMigrations(ctx context.Context) (map[string]MigrationInfo, error) {
	rows, err := m.db.Query(ctx, `
		SELECT version, description, applied_at
		FROM schema_migrations
		ORDER BY version
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]MigrationInfo)
	for rows.Next() {
		var info MigrationInfo
		if err := rows.Scan(&info.Version, &info.Description, &info.AppliedAt); err != nil {
			return nil, err
		}
		executed[info.Version] = info
	}

	return executed, rows.Err()
}

// executeMigration 执行单个迁移
func (m *Migrator) executeMigration(ctx context.Context, migration Migration) error {
	// 在事务中执行迁移
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 执行迁移SQL
	if _, err := tx.Exec(ctx, migration.SQL); err != nil {
		return err
	}

	// 记录迁移
	if _, err := tx.Exec(ctx, `
		INSERT INTO schema_migrations (version, description)
		VALUES ($1, $2)
	`, migration.Version, migration.Description); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// MigrationInfo 迁移信息
type MigrationInfo struct {
	Version     string
	Description string
	AppliedAt   time.Time
}
