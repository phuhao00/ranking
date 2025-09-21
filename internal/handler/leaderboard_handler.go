// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 排行榜HTTP处理器，适配netcore-go框架

package handler

import (
	"context"
	"strconv"

	"ranking/internal/service"
	"ranking/pkg/logger"
)

// LeaderboardHandler 排行榜处理器
type LeaderboardHandler struct {
	service service.LeaderboardService
	logger  logger.Logger
}

// NewLeaderboardHandler 创建排行榜处理器
func NewLeaderboardHandler(service service.LeaderboardService, logger logger.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{
		service: service,
		logger:  logger,
	}
}

// Create 创建排行榜
func (h *LeaderboardHandler) Create(ctx *HTTPContext, resp *HTTPResponse) {
	var req service.CreateLeaderboardRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.logger.Warn("创建排行榜参数错误", "error", err)
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	leaderboard, err := h.service.CreateLeaderboard(context.Background(), &req)
	if err != nil {
		h.logger.Error("创建排行榜失败", "error", err, "request", req)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "创建排行榜失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 201, map[string]interface{}{
		"code":    201,
		"message": "排行榜创建成功",
		"data":    leaderboard,
	})
}

// Get 获取排行榜配置
func (h *LeaderboardHandler) Get(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	if leaderboardID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID不能为空",
		})
		return
	}

	// 获取查询参数
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	offsetStr := ctx.Query("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000 // 限制最大返回数量
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 获取排行榜配置
	leaderboard, err := h.service.GetLeaderboard(context.Background(), leaderboardID)
	if err != nil {
		h.logger.Error("获取排行榜配置失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 404, map[string]interface{}{
			"code":    404,
			"message": "排行榜不存在: " + err.Error(),
		})
		return
	}

	// 获取排行榜数据
	rankings, err := h.service.GetRankings(context.Background(), leaderboardID, limit, offset)
	if err != nil {
		h.logger.Error("获取排行榜数据失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "获取排行榜数据失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data": map[string]interface{}{
			"leaderboard": leaderboard,
			"rankings":    rankings,
			"limit":       limit,
			"offset":      offset,
		},
	})
}

// UpdateConfig 更新排行榜配置
func (h *LeaderboardHandler) UpdateConfig(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	if leaderboardID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID不能为空",
		})
		return
	}

	// 获取现有配置
	leaderboard, err := h.service.GetLeaderboard(context.Background(), leaderboardID)
	if err != nil {
		h.logger.Error("获取排行榜配置失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 404, map[string]interface{}{
			"code":    404,
			"message": "排行榜不存在: " + err.Error(),
		})
		return
	}

	// 绑定更新请求
	var updateReq service.CreateLeaderboardRequest
	if err := ctx.BindJSON(&updateReq); err != nil {
		h.logger.Warn("更新排行榜参数错误", "error", err)
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 更新配置
	leaderboard.Name = updateReq.Name
	leaderboard.Type = updateReq.Type
	leaderboard.SortOrder = updateReq.SortOrder
	leaderboard.MaxEntries = updateReq.MaxEntries
	leaderboard.Config = updateReq.Config

	err = h.service.UpdateLeaderboard(context.Background(), leaderboard)
	if err != nil {
		h.logger.Error("更新排行榜失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "更新排行榜失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "排行榜更新成功",
		"data":    leaderboard,
	})
}

// Delete 删除排行榜
func (h *LeaderboardHandler) Delete(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	if leaderboardID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID不能为空",
		})
		return
	}

	err := h.service.DeleteLeaderboard(context.Background(), leaderboardID)
	if err != nil {
		h.logger.Error("删除排行榜失败", "error", err, "leaderboard_id", leaderboardID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "删除排行榜失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "排行榜删除成功",
	})
}

// GetUserRank 获取用户排名
func (h *LeaderboardHandler) GetUserRank(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	userID := ctx.Param("userId")

	if leaderboardID == "" || userID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID和用户ID不能为空",
		})
		return
	}

	rank, err := h.service.GetUserRank(context.Background(), leaderboardID, userID)
	if err != nil {
		h.logger.Error("获取用户排名失败", "error", err, "leaderboard_id", leaderboardID, "user_id", userID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "获取用户排名失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data":    rank,
	})
}

// GetRankingsAroundUser 获取用户周围的排名
func (h *LeaderboardHandler) GetRankingsAroundUser(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("id")
	userID := ctx.Param("userId")

	if leaderboardID == "" || userID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID和用户ID不能为空",
		})
		return
	}

	// 获取范围参数
	rangeStr := ctx.Query("range")
	if rangeStr == "" {
		rangeStr = "10"
	}
	rangeSize, err := strconv.ParseInt(rangeStr, 10, 64)
	if err != nil || rangeSize <= 0 {
		rangeSize = 10
	}
	if rangeSize > 50 {
		rangeSize = 50 // 限制最大范围
	}

	rankings, err := h.service.GetRankingsAroundUser(context.Background(), leaderboardID, userID, rangeSize)
	if err != nil {
		h.logger.Error("获取用户周围排名失败", "error", err, "leaderboard_id", leaderboardID, "user_id", userID)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "获取用户周围排名失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data":    rankings,
	})
}

// List 获取排行榜列表
func (h *LeaderboardHandler) List(ctx *HTTPContext, resp *HTTPResponse) {
	// 获取查询参数
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	offsetStr := ctx.Query("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	gameID := ctx.Query("game_id")
	// leaderboardType := ctx.Query("type") // 暂时不使用类型过滤

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 限制最大返回数量
	}

	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil || offset < 0 {
		offset = 0
	}

	leaderboards, total, err := h.service.ListLeaderboards(context.Background(), gameID, limit, offset)
	if err != nil {
		h.logger.Error("获取排行榜列表失败", "error", err)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "获取排行榜列表失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data": map[string]interface{}{
			"leaderboards": leaderboards,
			"total":        total,
			"limit":        limit,
			"offset":       offset,
		},
	})
}