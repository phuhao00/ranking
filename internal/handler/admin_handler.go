// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 管理员HTTP处理器，适配netcore-go框架

package handler

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/internal/service"
	"ranking/pkg/logger"
)

// AdminHandler 管理员处理器
type AdminHandler struct {
	leaderboardService service.LeaderboardService
	mongoClient        *mongodb.Client
	redisClient        *redis.Client
	logger             logger.Logger
	startTime          time.Time
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(
	leaderboardService service.LeaderboardService,
	mongoClient *mongodb.Client,
	redisClient *redis.Client,
	logger logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		leaderboardService: leaderboardService,
		mongoClient:        mongoClient,
		redisClient:        redisClient,
		logger:             logger,
		startTime:          time.Now(),
	}
}

// GetStats 获取系统统计信息
func (h *AdminHandler) GetStats(ctx *HTTPContext, resp *HTTPResponse) {
	requestCtx := context.Background()

	// 获取系统运行时统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemStats := map[string]interface{}{
		"service_name":       "ranking-service",
		"version":            "1.0.0",
		"uptime_seconds":     time.Since(h.startTime).Seconds(),
		"start_time":         h.startTime.Unix(),
		"current_time":       time.Now().Unix(),
		"goroutines":         runtime.NumGoroutine(),
		"memory_alloc_mb":    float64(m.Alloc) / 1024 / 1024,
		"memory_sys_mb":      float64(m.Sys) / 1024 / 1024,
		"memory_heap_mb":     float64(m.HeapAlloc) / 1024 / 1024,
		"gc_runs":            m.NumGC,
		"cpu_cores":          runtime.NumCPU(),
	}

	// 获取数据库连接状态
	dbStats := map[string]interface{}{
		"mongodb": map[string]interface{}{
			"status": "connected",
		},
		"redis": map[string]interface{}{
			"status": "connected",
		},
	}

	// 检查MongoDB连接
	if err := h.mongoClient.HealthCheck(requestCtx); err != nil {
		dbStats["mongodb"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	// 检查Redis连接
	if err := h.redisClient.HealthCheck(requestCtx); err != nil {
		dbStats["redis"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	}

	// 获取Redis详细统计
	if redisStats, err := h.redisClient.GetStats(requestCtx); err == nil {
		dbStats["redis"].(map[string]interface{})["stats"] = map[string]interface{}{
			"connected_clients":         redisStats["connected_clients"],
			"used_memory_human":         redisStats["used_memory_human"],
			"total_commands_processed":  redisStats["total_commands_processed"],
			"keyspace_hits":             redisStats["keyspace_hits"],
			"keyspace_misses":           redisStats["keyspace_misses"],
			"instantaneous_ops_per_sec": redisStats["instantaneous_ops_per_sec"],
		}
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取统计信息成功",
		"data": map[string]interface{}{
			"system":   systemStats,
			"database": dbStats,
			"timestamp": time.Now().Unix(),
		},
	})
}

// ClearCache 清除缓存
func (h *AdminHandler) ClearCache(ctx *HTTPContext, resp *HTTPResponse) {
	var req struct {
		LeaderboardID string `json:"leaderboard_id,omitempty"`
		Pattern       string `json:"pattern,omitempty"`
	}

	if err := ctx.BindJSON(&req); err != nil {
		h.logger.Warn("清除缓存参数错误", "error", err)
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	requestCtx := context.Background()
	clearedCount := 0

	if req.LeaderboardID != "" {
		// 清除指定排行榜的缓存
		leaderboardKey := fmt.Sprintf("leaderboard:%s", req.LeaderboardID)
		configKey := fmt.Sprintf("leaderboard_config:%s", req.LeaderboardID)
		statsKey := fmt.Sprintf("leaderboard_stats:%s", req.LeaderboardID)

		keys := []string{leaderboardKey, configKey, statsKey}
		if err := h.redisClient.Del(requestCtx, keys...); err != nil {
			h.logger.Error("清除排行榜缓存失败", "error", err, "leaderboard_id", req.LeaderboardID)
			ctx.JSON(resp, 500, map[string]interface{}{
				"code":    500,
				"message": "清除缓存失败: " + err.Error(),
			})
			return
		}
		clearedCount = len(keys)

		h.logger.Info("排行榜缓存已清除", "leaderboard_id", req.LeaderboardID, "keys", keys)
	} else if req.Pattern != "" {
		// 根据模式清除缓存（这里简化实现，实际生产环境需要更安全的方式）
		h.logger.Warn("模式清除缓存功能暂未实现", "pattern", req.Pattern)
		ctx.JSON(resp, 501, map[string]interface{}{
			"code":    501,
			"message": "模式清除缓存功能暂未实现",
		})
		return
	} else {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "必须指定leaderboard_id或pattern",
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "缓存清除成功",
		"data": map[string]interface{}{
			"cleared_count": clearedCount,
			"timestamp":     time.Now().Unix(),
		},
	})
}

// RebuildLeaderboard 重建排行榜
func (h *AdminHandler) RebuildLeaderboard(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	if leaderboardID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID不能为空",
		})
		return
	}

	h.logger.Info("开始重建排行榜", "leaderboard_id", leaderboardID)

	start := time.Now()
	err := h.leaderboardService.RebuildLeaderboard(context.Background(), leaderboardID)
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("重建排行榜失败", "error", err, "leaderboard_id", leaderboardID, "duration", duration)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "重建排行榜失败: " + err.Error(),
		})
		return
	}

	h.logger.Info("排行榜重建完成", "leaderboard_id", leaderboardID, "duration", duration)

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "排行榜重建成功",
		"data": map[string]interface{}{
			"leaderboard_id": leaderboardID,
			"duration_ms":    duration.Milliseconds(),
			"timestamp":      time.Now().Unix(),
		},
	})
}