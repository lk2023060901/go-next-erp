package password

import (
	"testing"
)

// BenchmarkArgon2Hash 基准测试：Argon2密码哈希
func BenchmarkArgon2Hash(b *testing.B) {
	hasher := NewArgon2Hasher()
	password := "BenchmarkPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Hash(password)
	}
}

// BenchmarkArgon2Verify 基准测试：Argon2密码验证
func BenchmarkArgon2Verify(b *testing.B) {
	hasher := NewArgon2Hasher()
	password := "BenchmarkPassword123!"
	hash, _ := hasher.Hash(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hasher.Verify(password, hash)
	}
}

// BenchmarkValidatorValidate 基准测试：密码验证
func BenchmarkValidatorValidate(b *testing.B) {
	validator := NewValidator(DefaultPolicy())
	password := "ValidPass123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Validate(password)
	}
}

// BenchmarkValidatorStrength 基准测试：密码强度计算
func BenchmarkValidatorStrength(b *testing.B) {
	validator := NewValidator(DefaultPolicy())
	password := "C0mpl3x!P@ssw0rd!2024"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.Strength(password)
	}
}

// BenchmarkConcurrentHash 基准测试：并发哈希
func BenchmarkConcurrentHash(b *testing.B) {
	hasher := NewArgon2Hasher()
	password := "BenchmarkPassword123!"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = hasher.Hash(password)
		}
	})
}

// BenchmarkConcurrentVerify 基准测试：并发验证
func BenchmarkConcurrentVerify(b *testing.B) {
	hasher := NewArgon2Hasher()
	password := "BenchmarkPassword123!"
	hash, _ := hasher.Hash(password)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = hasher.Verify(password, hash)
		}
	})
}
