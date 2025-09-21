// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 健康检查HTTP处理器，适配netcore-go框架

package handler

import (
	"context"
	"net/http"
	"time"

	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/pkg/logger"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	mongoClient *mongodb.Client
	redisClient *redis.Client
	logger      logger.Logger
	startTime   time.Time
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(
	mongoClient *mongodb.Client,
	redisClient *redis.Client,
	logger logger.Logger,
) *HealthHandler {
	return &HealthHandler{
		mongoClient: mongoClient,
		redisClient: redisClient,
		logger:      logger,
		startTime:   time.Now(),
	}
}

// Check 健康检查
func (h *HealthHandler) Check(ctx *HTTPContext, resp *HTTPResponse) {
	ctx.JSON(resp, 200, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "ranking-api",
		"version":   "1.0.0",
		"uptime":    time.Since(h.startTime).Seconds(),
	})
}

// Ready 就绪检查
func (h *HealthHandler) Ready(ctx *HTTPContext, resp *HTTPResponse) {
	requestCtx := context.Background()
	status := "ok"
	statusCode := http.StatusOK
	checks := make(map[string]interface{})

	// 检查MongoDB连接
	if err := h.mongoClient.HealthCheck(requestCtx); err != nil {
		status = "error"
		statusCode = http.StatusServiceUnavailable
		checks["mongodb"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
		h.logger.Error("MongoDB健康检查失败", "error", err)
	} else {
		checks["mongodb"] = map[string]interface{}{
			"status": "ok",
		}
	}

	// 检查Redis连接
	if err := h.redisClient.HealthCheck(requestCtx); err != nil {
		status = "error"
		statusCode = http.StatusServiceUnavailable
		checks["redis"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
		h.logger.Error("Redis健康检查失败", "error", err)
	} else {
		checks["redis"] = map[string]interface{}{
			"status": "ok",
		}
	}

	ctx.JSON(resp, statusCode, map[string]interface{}{
		"status":     status,
		"timestamp":  time.Now().Unix(),
		"service":    "ranking-api",
		"version":    "1.0.0",
		"uptime":     time.Since(h.startTime).Seconds(),
		"checks":     checks,
	})
}