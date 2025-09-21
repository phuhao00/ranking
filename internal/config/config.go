// Package config
// Author: HHaou
// Created: 2024-01-20
// Description: 配置管理模块

package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server" yaml:"server"`
	MongoDB  MongoDBConfig  `mapstructure:"mongodb" yaml:"mongodb"`
	Redis    RedisConfig    `mapstructure:"redis" yaml:"redis"`
	Log      LogConfig      `mapstructure:"log" yaml:"log"`
	Metrics  MetricsConfig  `mapstructure:"metrics" yaml:"metrics"`
	Consul   ConsulConfig   `mapstructure:"consul" yaml:"consul"`
	Security SecurityConfig `mapstructure:"security" yaml:"security"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Host         string `mapstructure:"host" yaml:"host"`
	Port         int    `mapstructure:"port" yaml:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout" yaml:"idle_timeout"`
	MaxConns     int    `mapstructure:"max_conns" yaml:"max_conns"`
}

// MongoDBConfig MongoDB数据库配置
type MongoDBConfig struct {
	URI            string `mapstructure:"uri" yaml:"uri"`
	Database       string `mapstructure:"database" yaml:"database"`
	MaxPoolSize    int    `mapstructure:"max_pool_size" yaml:"max_pool_size"`
	MinPoolSize    int    `mapstructure:"min_pool_size" yaml:"min_pool_size"`
	MaxIdleTime    int    `mapstructure:"max_idle_time" yaml:"max_idle_time"`
	ConnectTimeout int    `mapstructure:"connect_timeout" yaml:"connect_timeout"`
}

// RedisConfig Redis缓存配置
type RedisConfig struct {
	Addr         string `mapstructure:"addr" yaml:"addr"`
	Password     string `mapstructure:"password" yaml:"password"`
	DB           int    `mapstructure:"db" yaml:"db"`
	PoolSize     int    `mapstructure:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" yaml:"min_idle_conns"`
	IdleTimeout  int    `mapstructure:"idle_timeout" yaml:"idle_timeout"`
	DialTimeout  int    `mapstructure:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout" yaml:"write_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level" yaml:"level"`
	Format     string `mapstructure:"format" yaml:"format"`
	Output     string `mapstructure:"output" yaml:"output"`
	FilePath   string `mapstructure:"file_path" yaml:"file_path"`
	MaxSize    int    `mapstructure:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" yaml:"max_age"`
	Compress   bool   `mapstructure:"compress" yaml:"compress"`
}

// MetricsConfig 监控指标配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Path    string `mapstructure:"path" yaml:"path"`
	Port    int    `mapstructure:"port" yaml:"port"`
}

// ConsulConfig 服务发现配置
type ConsulConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Addr     string `mapstructure:"addr" yaml:"addr"`
	Scheme   string `mapstructure:"scheme" yaml:"scheme"`
	Token    string `mapstructure:"token" yaml:"token"`
	Datacenter string `mapstructure:"datacenter" yaml:"datacenter"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWTSecret     string   `mapstructure:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiration int      `mapstructure:"jwt_expiration" yaml:"jwt_expiration"`
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins"`
	RateLimit     int      `mapstructure:"rate_limit" yaml:"rate_limit"`
}

// Load 加载配置文件
func Load(configPath, env string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径和名称
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 设置环境变量前缀
	v.SetEnvPrefix("RANKING")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v, env)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认配置
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证配置
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper, env string) {
	// 服务器默认配置
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.idle_timeout", 120)
	v.SetDefault("server.max_conns", 10000)

	// MongoDB默认配置
	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "ranking")
	v.SetDefault("mongodb.max_pool_size", 100)
	v.SetDefault("mongodb.min_pool_size", 10)
	v.SetDefault("mongodb.max_idle_time", 300)
	v.SetDefault("mongodb.connect_timeout", 10)

	// Redis默认配置
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 10)
	v.SetDefault("redis.idle_timeout", 300)
	v.SetDefault("redis.dial_timeout", 5)
	v.SetDefault("redis.read_timeout", 3)
	v.SetDefault("redis.write_timeout", 3)

	// 日志默认配置
	logLevel := "info"
	if env == "development" {
		logLevel = "debug"
	}
	v.SetDefault("log.level", logLevel)
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.file_path", "logs/ranking.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	// 监控默认配置
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.port", 9090)

	// Consul默认配置
	v.SetDefault("consul.enabled", false)
	v.SetDefault("consul.addr", "localhost:8500")
	v.SetDefault("consul.scheme", "http")
	v.SetDefault("consul.token", "")
	v.SetDefault("consul.datacenter", "dc1")

	// 安全默认配置
	v.SetDefault("security.jwt_secret", "your-secret-key")
	v.SetDefault("security.jwt_expiration", 3600)
	v.SetDefault("security.allowed_origins", []string{"*"})
	v.SetDefault("security.rate_limit", 1000)
}

// validate 验证配置
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}

	if cfg.MongoDB.URI == "" {
		return fmt.Errorf("MongoDB URI不能为空")
	}

	if cfg.MongoDB.Database == "" {
		return fmt.Errorf("MongoDB数据库名不能为空")
	}

	if cfg.Redis.Addr == "" {
		return fmt.Errorf("Redis地址不能为空")
	}

	if cfg.Security.JWTSecret == "" || cfg.Security.JWTSecret == "your-secret-key" {
		return fmt.Errorf("JWT密钥必须设置且不能使用默认值")
	}

	return nil
}

// GetAddr 获取服务器监听地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetMetricsAddr 获取监控指标服务地址
func (c *MetricsConfig) GetAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}