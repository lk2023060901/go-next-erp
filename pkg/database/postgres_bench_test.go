package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

var benchDB *DB

// setupBenchDB 设置基准测试数据库
func setupBenchDB(b *testing.B) *DB {
	if benchDB != nil {
		return benchDB
	}

	ctx := context.Background()
	cfg := getTestConfig()
	cfg.MaxConns = 50 // 基准测试使用更大的连接池
	cfg.MinConns = 10

	db, err := New(ctx, WithConfig(cfg))
	if err != nil {
		b.Skipf("Cannot connect to database for benchmarking: %v", err)
		return nil
	}

	if err := db.Ping(ctx); err != nil {
		b.Skipf("Database not accessible for benchmarking: %v", err)
		return nil
	}

	// 创建基准测试表
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS bench_test (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			value INT,
			data TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		b.Skipf("Cannot create bench table: %v", err)
		return nil
	}

	// 插入一些初始数据
	for i := 0; i < 1000; i++ {
		db.Exec(ctx, "INSERT INTO bench_test (name, value, data) VALUES ($1, $2, $3)",
			fmt.Sprintf("item_%d", i), i, fmt.Sprintf("data_%d", i))
	}

	benchDB = db
	return db
}

// BenchmarkQueryRow 基准测试单行查询
func BenchmarkQueryRow(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var name string
		var value int
		db.QueryRow(ctx, "SELECT name, value FROM bench_test WHERE id = $1", i%1000+1).Scan(&name, &value)
	}
}

// BenchmarkQuery 基准测试多行查询
func BenchmarkQuery(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query(ctx, "SELECT id, name, value FROM bench_test LIMIT 10")
		if err != nil {
			b.Fatal(err)
		}

		for rows.Next() {
			var id, value int
			var name string
			rows.Scan(&id, &name, &value)
		}
		rows.Close()
	}
}

// BenchmarkExec 基准测试写操作
func BenchmarkExec(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Exec(ctx, "INSERT INTO bench_test (name, value, data) VALUES ($1, $2, $3)",
			fmt.Sprintf("bench_%d", i), i, fmt.Sprintf("benchmark_data_%d", i))
	}
}

// BenchmarkTransaction 基准测试事务
func BenchmarkTransaction(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Transaction(ctx, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "INSERT INTO bench_test (name, value) VALUES ($1, $2)",
				fmt.Sprintf("tx_%d", i), i)
			return err
		})
	}
}

// BenchmarkTransactionMultiOp 基准测试多操作事务
func BenchmarkTransactionMultiOp(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Transaction(ctx, func(tx pgx.Tx) error {
			for j := 0; j < 10; j++ {
				_, err := tx.Exec(ctx, "INSERT INTO bench_test (name, value) VALUES ($1, $2)",
					fmt.Sprintf("multi_%d_%d", i, j), i*10+j)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// BenchmarkPing 基准测试连接检查
func BenchmarkPing(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Ping(ctx)
	}
}

// BenchmarkParallelQueryRow 基准测试并发单行查询
func BenchmarkParallelQueryRow(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			var name string
			var value int
			db.QueryRow(ctx, "SELECT name, value FROM bench_test WHERE id = $1", i%1000+1).Scan(&name, &value)
			i++
		}
	})
}

// BenchmarkParallelQuery 基准测试并发多行查询
func BenchmarkParallelQuery(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rows, err := db.Query(ctx, "SELECT id, name, value FROM bench_test LIMIT 10")
			if err != nil {
				b.Fatal(err)
			}

			for rows.Next() {
				var id, value int
				var name string
				rows.Scan(&id, &name, &value)
			}
			rows.Close()
		}
	})
}

// BenchmarkParallelExec 基准测试并发写操作
func BenchmarkParallelExec(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			db.Exec(ctx, "INSERT INTO bench_test (name, value) VALUES ($1, $2)",
				fmt.Sprintf("parallel_%d", i), i)
			i++
		}
	})
}

// BenchmarkParallelTransaction 基准测试并发事务
func BenchmarkParallelTransaction(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			db.Transaction(ctx, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, "INSERT INTO bench_test (name, value) VALUES ($1, $2)",
					fmt.Sprintf("parallel_tx_%d", i), i)
				i++
				return err
			})
		}
	})
}

// BenchmarkComplexQuery 基准测试复杂查询
func BenchmarkComplexQuery(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := db.Query(ctx, `
			SELECT
				name,
				value,
				COUNT(*) OVER() as total_count,
				AVG(value) OVER() as avg_value
			FROM bench_test
			WHERE value > $1 AND value < $2
			ORDER BY value DESC
			LIMIT 20
		`, i%500, (i%500)+100)

		if err != nil {
			b.Fatal(err)
		}

		for rows.Next() {
			var name string
			var value, total int
			var avg float64
			rows.Scan(&name, &value, &total, &avg)
		}
		rows.Close()
	}
}

// BenchmarkBatchInsert 基准测试批量插入
func BenchmarkBatchInsert(b *testing.B) {
	db := setupBenchDB(b)
	if db == nil {
		return
	}

	ctx := context.Background()
	batchSize := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Transaction(ctx, func(tx pgx.Tx) error {
			for j := 0; j < batchSize; j++ {
				_, err := tx.Exec(ctx, "INSERT INTO bench_test (name, value) VALUES ($1, $2)",
					fmt.Sprintf("batch_%d_%d", i, j), i*batchSize+j)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
}
