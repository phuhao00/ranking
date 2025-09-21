// Package server
// Author: HHaou
// Created: 2024-01-20
// Description: HTTP服务器实现，基于netcore-go v1.0.2框架

package server

import (
	"context"

	"ranking/internal/config"
	"ranking/internal/handler"
	"ranking/internal/middleware"
	"ranking/pkg/logger"

	netcore "github.com/phuhao00/netcore-go"
	netcorehttp "github.com/phuhao00/netcore-go/pkg/http"
)

// Server HTTP服务器
type Server struct {
	config        config.ServerConfig
	logger        logger.Logger
	netcoreServer *netcorehttp.HTTPServer
	handlers      *handler.Handlers
}

// New 创建新的HTTP服务器
func New(cfg config.ServerConfig, logger logger.Logger, handlers *handler.Handlers) (*Server, error) {
	s := &Server{
		config:   cfg,
		logger:   logger,
		handlers: handlers,
	}

	// 创建netcore-go HTTP服务器
	s.netcoreServer = netcore.NewHTTPServer()

	// 设置中间件
	s.setupMiddlewares()

	// 设置路由
	s.setupRoutes()

	return s, nil
}

// setupMiddlewares 设置中间件
func (s *Server) setupMiddlewares() {
	// 添加全局中间件
	s.netcoreServer.Use(middleware.NewLoggerMiddleware(s.logger))
	s.netcoreServer.Use(middleware.NewRecoveryMiddleware(s.logger))
	s.netcoreServer.Use(middleware.NewCORSMiddleware())
	s.netcoreServer.Use(middleware.NewRequestIDMiddleware())
	s.netcoreServer.Use(middleware.NewSecurityMiddleware())
	s.netcoreServer.Use(middleware.NewRateLimitMiddleware(100)) // 每秒100个请求
}





// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.netcoreServer.GET("/health", s.handlers.Health.Check)
	s.netcoreServer.GET("/ready", s.handlers.Health.Ready)

	// 排行榜路由
	s.netcoreServer.POST("/api/v1/leaderboard/create", s.handlers.Leaderboard.Create)
	s.netcoreServer.GET("/api/v1/leaderboard/list", s.handlers.Leaderboard.List)
	s.netcoreServer.GET("/api/v1/leaderboard/:id", s.handlers.Leaderboard.Get)
	s.netcoreServer.DELETE("/api/v1/leaderboard/:id", s.handlers.Leaderboard.Delete)

	// 分数路由
	s.netcoreServer.POST("/api/v1/score/submit", s.handlers.Score.Submit)
	s.netcoreServer.POST("/api/v1/score/batch", s.handlers.Score.BatchSubmit)

	// 监控指标
	s.netcoreServer.GET("/api/v1/metrics", s.handlers.Metrics.GetMetrics)
	s.netcoreServer.GET("/api/v1/metrics/leaderboard/:id/stats", s.handlers.Metrics.GetLeaderboardStats)

	// 管理接口
	s.netcoreServer.GET("/admin/stats", s.handlers.Admin.GetStats)
	s.netcoreServer.POST("/admin/cache/clear", s.handlers.Admin.ClearCache)
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("启动netcore-go HTTP服务器",
		"addr", s.config.GetAddr(),
		"read_timeout", s.config.ReadTimeout,
		"write_timeout", s.config.WriteTimeout,
	)

	// 启动服务器
	go func() {
		if err := s.netcoreServer.Start(); err != nil {
			s.logger.Error("netcore-go HTTP服务器启动失败", "error", err)
		}
	}()

	s.logger.Info("netcore-go HTTP服务器已启动")
	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("开始停止netcore-go HTTP服务器")

	// 停止服务器
	if err := s.netcoreServer.Stop(); err != nil {
		s.logger.Error("netcore-go HTTP服务器停止失败", "error", err)
		return err
	}

	s.logger.Info("netcore-go HTTP服务器已停止")
	return nil
}