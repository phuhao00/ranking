// Package mongodb
// Author: HHaou
// Created: 2024-01-20
// Description: MongoDB数据访问层

package mongodb

import (
	"context"
	"fmt"
	"time"

	"ranking/internal/config"
	"ranking/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client MongoDB客户端封装
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	logger   logger.Logger
	config   config.MongoDBConfig
}

// New 创建新的MongoDB客户端
func New(cfg config.MongoDBConfig, log logger.Logger) (*Client, error) {
	// 设置客户端选项
	clientOptions := options.Client().ApplyURI(cfg.URI)

	// 连接池配置
	if cfg.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(cfg.MaxPoolSize))
	}
	if cfg.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(uint64(cfg.MinPoolSize))
	}
	if cfg.MaxIdleTime > 0 {
		clientOptions.SetMaxConnIdleTime(time.Duration(cfg.MaxIdleTime) * time.Second)
	}
	if cfg.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(time.Duration(cfg.ConnectTimeout) * time.Second)
	}

	// 创建客户端
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("MongoDB连接测试失败: %w", err)
	}

	database := client.Database(cfg.Database)

	log.Info("MongoDB连接成功",
		"database", cfg.Database,
		"max_pool_size", cfg.MaxPoolSize,
		"min_pool_size", cfg.MinPoolSize,
	)

	return &Client{
		client:   client,
		database: database,
		logger:   log,
		config:   cfg,
	}, nil
}

// Close 关闭MongoDB连接
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		if err := c.client.Disconnect(ctx); err != nil {
			c.logger.Error("关闭MongoDB连接失败", "error", err)
			return err
		}
		c.logger.Info("MongoDB连接已关闭")
	}
	return nil
}

// Database 获取数据库实例
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection 获取集合实例
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// StartSession 开始会话
func (c *Client) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return c.client.StartSession(opts...)
}

// WithTransaction 执行事务
func (c *Client) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := c.StartSession()
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}
	defer session.EndSession(ctx)

	return session.WithTransaction(ctx, fn)
}

// CreateIndexes 创建索引
func (c *Client) CreateIndexes(ctx context.Context) error {
	c.logger.Info("开始创建MongoDB索引")

	// 排行榜集合索引
	leaderboardCollection := c.Collection("leaderboards")
	leaderboardIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"leaderboard_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{
				"game_id": 1,
				"type":    1,
			},
		},
		{
			Keys: map[string]interface{}{
				"is_active": 1,
			},
		},
	}

	if _, err := leaderboardCollection.Indexes().CreateMany(ctx, leaderboardIndexes); err != nil {
		return fmt.Errorf("创建排行榜索引失败: %w", err)
	}

	// 分数记录集合索引
	scoreCollection := c.Collection("score_records")
	scoreIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"leaderboard_id": 1,
				"user_id":        1,
			},
		},
		{
			Keys: map[string]interface{}{
				"leaderboard_id": 1,
				"score":          -1,
			},
		},
		{
			Keys: map[string]interface{}{
				"submitted_at": 1,
			},
		},
		{
			Keys: map[string]interface{}{
				"created_at": 1,
			},
			Options: options.Index().SetExpireAfterSeconds(30 * 24 * 3600), // 30天后过期
		},
	}

	if _, err := scoreCollection.Indexes().CreateMany(ctx, scoreIndexes); err != nil {
		return fmt.Errorf("创建分数记录索引失败: %w", err)
	}

	// 用户集合索引
	userCollection := c.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"user_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{
				"username": 1,
			},
		},
	}

	if _, err := userCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("创建用户索引失败: %w", err)
	}

	// 统计集合索引
	statsCollection := c.Collection("leaderboard_stats")
	statsIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"leaderboard_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := statsCollection.Indexes().CreateMany(ctx, statsIndexes); err != nil {
		return fmt.Errorf("创建统计索引失败: %w", err)
	}

	c.logger.Info("MongoDB索引创建完成")
	return nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("MongoDB健康检查失败: %w", err)
	}

	return nil
}