// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 监控指标HTTP处理器，适配netcore-go框架

package handler

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/pkg/logger"
)

// MetricsHandler 监控指标处理器
type MetricsHandler struct {
	mongoClient *mongodb.Client
	redisClient *redis.Client
	logger      logger.Logger
	startTime   time.Time
}

// NewMetricsHandler 创建监控指标处理器
func NewMetricsHandler(
	mongoClient *mongodb.Client,
	redisClient *redis.Client,
	logger logger.Logger,
) *MetricsHandler {
	return &MetricsHandler{
		mongoClient: mongoClient,
		redisClient: redisClient,
		logger:      logger,
		startTime:   time.Now(),
	}
}

// GetMetrics 获取系统指标
func (h *MetricsHandler) GetMetrics(ctx *HTTPContext, resp *HTTPResponse) {
	requestCtx := context.Background()

	// 获取系统指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemMetrics := map[string]interface{}{
		"uptime_seconds":     time.Since(h.startTime).Seconds(),
		"goroutines":         runtime.NumGoroutine(),
		"memory_alloc_bytes": m.Alloc,
		"memory_sys_bytes":   m.Sys,
		"memory_heap_bytes":  m.HeapAlloc,
		"gc_runs":            m.NumGC,
		"cpu_cores":          runtime.NumCPU(),
	}

	// 获取Redis指标
	redisMetrics := make(map[string]interface{})
	if stats, err := h.redisClient.GetStats(requestCtx); err == nil {
		redisMetrics = map[string]interface{}{
			"connected_clients":    stats["connected_clients"],
			"used_memory":          stats["used_memory"],
			"used_memory_human":    stats["used_memory_human"],
			"total_commands_processed": stats["total_commands_processed"],
			"keyspace_hits":        stats["keyspace_hits"],
			"keyspace_misses":      stats["keyspace_misses"],
		}
	} else {
		h.logger.Warn("获取Redis指标失败", "error", err)
		redisMetrics["error"] = err.Error()
	}

	// MongoDB指标（简化版）
	mongoMetrics := map[string]interface{}{
		"status": "connected",
	}
	if err := h.mongoClient.HealthCheck(requestCtx); err != nil {
		mongoMetrics["status"] = "error"
		mongoMetrics["error"] = err.Error()
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"timestamp": time.Now().Unix(),
		"system":    systemMetrics,
		"redis":     redisMetrics,
		"mongodb":   mongoMetrics,
	})
}

// GetLeaderboardStats 获取排行榜统计信息
func (h *MetricsHandler) GetLeaderboardStats(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	if leaderboardID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID不能为空",
		})
		return
	}

	requestCtx := context.Background()

	// 获取排行榜大小
	key := fmt.Sprintf("leaderboard:%s", leaderboardID)
	size, err := h.redisClient.ZCard(requestCtx, key)
	if err != nil {
		h.logger.Error("获取排行榜大小失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "获取排行榜统计失败: " + err.Error(),
		})
		return
	}

	stats := map[string]interface{}{
		"leaderboard_id": leaderboardID,
		"total_users":    size,
		"timestamp":      time.Now().Unix(),
	}

	// 获取分数范围统计（可选）
	if size > 0 {
		// 获取前10名作为样本
		topRankings, err := h.redisClient.ZRevRangeWithScores(requestCtx, key, 0, 9)
		if err == nil && len(topRankings) > 0 {
			stats["highest_score"] = int64(topRankings[0].Score)
			if len(topRankings) > 1 {
				stats["top_10_lowest_score"] = int64(topRankings[len(topRankings)-1].Score)
			}
		}
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}