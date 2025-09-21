// Package app
// Author: HHaou
// Created: 2024-01-20
// Description: 应用程序主入口

package app

import (
	"context"
	"fmt"

	"ranking/internal/config"
	"ranking/internal/handler"
	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/internal/server"
	"ranking/internal/service"
	"ranking/pkg/logger"
)

// App 应用程序
type App struct {
	config      *config.Config
	logger      logger.Logger
	mongoClient *mongodb.Client
	redisClient *redis.Client
	server      *server.Server
}

// New 创建新的应用程序实例
func New(cfg *config.Config, log logger.Logger) (*App, error) {
	app := &App{
		config: cfg,
		logger: log,
	}

	// 初始化数据库连接
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化服务和处理器
	if err := app.initServices(); err != nil {
		return nil, fmt.Errorf("初始化服务失败: %w", err)
	}

	// 初始化HTTP服务器
	if err := app.initServer(); err != nil {
		return nil, fmt.Errorf("初始化服务器失败: %w", err)
	}

	return app, nil
}

// initDatabase 初始化数据库连接
func (a *App) initDatabase() error {
	a.logger.Info("初始化数据库连接")

	// 初始化MongoDB
	mongoClient, err := mongodb.New(a.config.MongoDB, a.logger)
	if err != nil {
		return fmt.Errorf("连接MongoDB失败: %w", err)
	}
	a.mongoClient = mongoClient

	// 创建MongoDB索引
	if err := mongoClient.CreateIndexes(context.Background()); err != nil {
		a.logger.Warn("创建MongoDB索引失败", "error", err)
	}

	// 初始化Redis
	redisClient, err := redis.New(a.config.Redis, a.logger)
	if err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}
	a.redisClient = redisClient

	a.logger.Info("数据库连接初始化完成")
	return nil
}

// initServices 初始化服务层
func (a *App) initServices() error {
	a.logger.Info("初始化服务层")

	// 创建仓储层
	leaderboardRepo := mongodb.NewLeaderboardRepository(a.mongoClient, a.logger)
	leaderboardCache := redis.NewLeaderboardCache(a.redisClient, a.logger)

	// 创建服务层
	leaderboardService := service.NewLeaderboardService(
		leaderboardRepo,
		leaderboardCache,
		a.logger,
	)

	// 创建处理器层
	handlers := handler.NewHandlers(
		leaderboardService,
		a.mongoClient,
		a.redisClient,
		a.logger,
	)

	// 创建HTTP服务器
	server, err := server.New(a.config.Server, a.logger, handlers)
	if err != nil {
		return fmt.Errorf("创建HTTP服务器失败: %w", err)
	}
	a.server = server

	a.logger.Info("服务层初始化完成")
	return nil
}

// initServer 初始化服务器
func (a *App) initServer() error {
	a.logger.Info("初始化HTTP服务器")
	return nil
}

// Start 启动应用程序
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("启动应用程序")

	// 启动HTTP服务器
	if err := a.server.Start(ctx); err != nil {
		return fmt.Errorf("启动HTTP服务器失败: %w", err)
	}

	a.logger.Info("应用程序启动成功")
	return nil
}

// Stop 停止应用程序
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("停止应用程序")

	// 停止HTTP服务器
	if a.server != nil {
		if err := a.server.Stop(ctx); err != nil {
			a.logger.Error("停止HTTP服务器失败", "error", err)
		}
	}

	// 关闭数据库连接
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			a.logger.Error("关闭Redis连接失败", "error", err)
		}
	}

	if a.mongoClient != nil {
		if err := a.mongoClient.Close(ctx); err != nil {
			a.logger.Error("关闭MongoDB连接失败", "error", err)
		}
	}

	a.logger.Info("应用程序已停止")
	return nil
}

// HealthCheck 健康检查
func (a *App) HealthCheck(ctx context.Context) error {
	// 检查MongoDB
	if err := a.mongoClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("MongoDB健康检查失败: %w", err)
	}

	// 检查Redis
	if err := a.redisClient.HealthCheck(ctx); err != nil {
		return fmt.Errorf("Redis健康检查失败: %w", err)
	}

	return nil
}