// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 处理器集合

package handler

import (
	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/internal/service"
	"ranking/pkg/logger"

	netcorehttp "github.com/phuhao00/netcore-go/pkg/http"
)

// HTTPContext 使用 netcore-go 的 HTTPContext
type HTTPContext = netcorehttp.HTTPContext

// HTTPRequest 使用 netcore-go 的 HTTPRequest
type HTTPRequest = netcorehttp.HTTPRequest

// HTTPResponse 使用 netcore-go 的 HTTPResponse
type HTTPResponse = netcorehttp.HTTPResponse

// Handlers 处理器集合
type Handlers struct {
	Leaderboard *LeaderboardHandler
	Score       *ScoreHandler
	Health      *HealthHandler
	Metrics     *MetricsHandler
	Admin       *AdminHandler
}

// NewHandlers 创建处理器集合
func NewHandlers(
	leaderboardService service.LeaderboardService,
	mongoClient *mongodb.Client,
	redisClient *redis.Client,
	logger logger.Logger,
) *Handlers {
	return &Handlers{
		Leaderboard: NewLeaderboardHandler(leaderboardService, logger),
		Score:       NewScoreHandler(leaderboardService, logger),
		Health:      NewHealthHandler(mongoClient, redisClient, logger),
		Metrics:     NewMetricsHandler(mongoClient, redisClient, logger),
		Admin:       NewAdminHandler(leaderboardService, mongoClient, redisClient, logger),
	}
}