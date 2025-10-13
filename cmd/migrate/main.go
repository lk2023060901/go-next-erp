package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/lk2023060901/go-next-erp/pkg/database"
	"github.com/lk2023060901/go-next-erp/pkg/migrate"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	ctx := context.Background()

	// 初始化数据库连接
	db, err := initDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 创建迁移管理器
	migrator := migrate.New(db)

	// 获取 migrations 目录路径
	migrationsDir := getMigrationsDir()

	// 从文件系统加载迁移文件
	if err := migrator.LoadFromDir(migrationsDir); err != nil {
		log.Fatalf("Failed to load migrations from %s: %v", migrationsDir, err)
	}

	// 执行命令
	switch command {
	case "up":
		if err := migrator.Up(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("All migrations completed successfully!")

	case "status":
		if err := migrator.Status(ctx); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up      - Run all pending migrations")
	fmt.Println("  status  - Show migration status")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  DB_HOST     - Database host (default: localhost)")
	fmt.Println("  DB_PORT     - Database port (default: 5432)")
	fmt.Println("  DB_NAME     - Database name (default: erp)")
	fmt.Println("  DB_USER     - Database user (default: postgres)")
	fmt.Println("  DB_PASSWORD - Database password (default: postgres)")
}

func initDatabase(ctx context.Context) (*database.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnvAsInt("DB_PORT", 5432)
	dbName := getEnv("DB_NAME", "erp")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")

	db, err := database.New(ctx,
		database.WithHost(host),
		database.WithPort(port),
		database.WithDatabase(dbName),
		database.WithUsername(user),
		database.WithPassword(password),
	)

	if err != nil {
		return nil, err
	}

	fmt.Printf("Connected to database: %s@%s:%d/%s\n", user, host, port, dbName)
	return db, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}

// getMigrationsDir 获取 migrations 目录路径
func getMigrationsDir() string {
	// 优先使用环境变量
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}

	// 默认使用根目录下的 migrations
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Warning: failed to get executable path: %v", err)
		return "./migrations"
	}

	// 获取项目根目录（假设 bin/migrate 结构）
	rootDir := filepath.Dir(filepath.Dir(execPath))
	migrationsPath := filepath.Join(rootDir, "migrations")

	// 检查目录是否存在
	if _, err := os.Stat(migrationsPath); err == nil {
		return migrationsPath
	}

	// 如果上面都失败，尝试当前目录
	if _, err := os.Stat("./migrations"); err == nil {
		return "./migrations"
	}

	// 最后尝试相对路径（开发环境）
	if _, err := os.Stat("../../migrations"); err == nil {
		return "../../migrations"
	}

	log.Fatalf("migrations directory not found. Please set MIGRATIONS_DIR environment variable or ensure migrations/ exists")
	return ""
}
