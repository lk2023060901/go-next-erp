package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	MinIO    MinIOConfig    `yaml:"minio"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"`
	GRPC GRPCConfig `yaml:"grpc"`
}

// HTTPConfig HTTP 服务器配置
type HTTPConfig struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

// GRPCConfig gRPC 服务器配置
type GRPCConfig struct {
	Network string `yaml:"network"`
	Addr    string `yaml:"addr"`
	Timeout string `yaml:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string   `yaml:"driver"`
	Master          string   `yaml:"master"`
	Slaves          []string `yaml:"slaves"`
	MaxOpenConns    int      `yaml:"max_open_conns"`
	MaxIdleConns    int      `yaml:"max_idle_conns"`
	ConnMaxLifetime string   `yaml:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// MinIOConfig MinIO 配置
type MinIOConfig struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	UseSSL          bool   `yaml:"use_ssl"`
	BucketName      string `yaml:"bucket_name"`
	Region          string `yaml:"region"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `yaml:"secret"`
	AccessExpire  int    `yaml:"access_expire"`
	RefreshExpire int    `yaml:"refresh_expire"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
