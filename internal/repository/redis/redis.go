// Package redis
// Author: HHaou
// Created: 2024-01-20
// Description: Redis缓存层实现

package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ranking/internal/config"
	"ranking/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Client Redis客户端封装
type Client struct {
	client *redis.Client
	logger logger.Logger
	config config.RedisConfig
}

// New 创建新的Redis客户端
func New(cfg config.RedisConfig, log logger.Logger) (*Client, error) {
	// 创建Redis客户端选项
	opts := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	// 创建客户端
	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接测试失败: %w", err)
	}

	log.Info("Redis连接成功",
		"addr", cfg.Addr,
		"db", cfg.DB,
		"pool_size", cfg.PoolSize,
	)

	return &Client{
		client: client,
		logger: log,
		config: cfg,
	}, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			c.logger.Error("关闭Redis连接失败", "error", err)
			return err
		}
		c.logger.Info("Redis连接已关闭")
	}
	return nil
}

// Client 获取Redis客户端实例
func (c *Client) Client() *redis.Client {
	return c.client
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Set 设置键值
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	result := c.client.Get(ctx, key)
	if result.Err() == redis.Nil {
		return "", nil
	}
	return result.Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// ZAdd 添加有序集合成员
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return c.client.ZAdd(ctx, key, members...).Err()
}

// ZRem 删除有序集合成员
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.ZRem(ctx, key, members...).Err()
}

// ZRevRangeWithScores 按分数倒序获取有序集合成员（带分数）
func (c *Client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRevRangeWithScores(ctx, key, start, stop).Result()
}

// ZRangeWithScores 按分数正序获取有序集合成员（带分数）
func (c *Client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	return c.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRevRank 获取成员在有序集合中的倒序排名
func (c *Client) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	result := c.client.ZRevRank(ctx, key, member)
	if result.Err() == redis.Nil {
		return -1, nil
	}
	return result.Result()
}

// ZRank 获取成员在有序集合中的正序排名
func (c *Client) ZRank(ctx context.Context, key, member string) (int64, error) {
	result := c.client.ZRank(ctx, key, member)
	if result.Err() == redis.Nil {
		return -1, nil
	}
	return result.Result()
}

// ZScore 获取成员分数
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	result := c.client.ZScore(ctx, key, member)
	if result.Err() == redis.Nil {
		return 0, nil
	}
	return result.Result()
}

// ZCard 获取有序集合成员数量
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.client.ZCard(ctx, key).Result()
}

// ZCount 统计分数区间内的成员数量
func (c *Client) ZCount(ctx context.Context, key, min, max string) (int64, error) {
	return c.client.ZCount(ctx, key, min, max).Result()
}

// HSet 设置哈希字段
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段值
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	result := c.client.HGet(ctx, key, field)
	if result.Err() == redis.Nil {
		return "", nil
	}
	return result.Result()
}

// HGetAll 获取哈希所有字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel 删除哈希字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HExists 检查哈希字段是否存在
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// Incr 递增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr 递减
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// SetNX 仅当键不存在时设置
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Eval 执行Lua脚本
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return c.client.Eval(ctx, script, keys, args...).Result()
}

// Pipeline 创建管道
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// TxPipeline 创建事务管道
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("Redis健康检查失败: %w", err)
	}

	return nil
}

// GetStats 获取Redis统计信息
func (c *Client) GetStats(ctx context.Context) (map[string]string, error) {
	info := c.client.Info(ctx, "stats")
	if info.Err() != nil {
		return nil, fmt.Errorf("获取Redis统计信息失败: %w", info.Err())
	}

	// 解析INFO命令返回的统计信息
	stats := make(map[string]string)
	lines := strings.Split(info.Val(), "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				stats[parts[0]] = parts[1]
			}
		}
	}

	return stats, nil
}