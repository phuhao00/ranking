// Package handler
// Author: HHaou
// Created: 2024-01-20
// Description: 分数HTTP处理器，适配netcore-go框架

package handler

import (
	"context"
	"fmt"
	"net/http"

	"ranking/internal/service"
	"ranking/pkg/logger"
)

// ScoreHandler 分数处理器
type ScoreHandler struct {
	service service.LeaderboardService
	logger  logger.Logger
}

// NewScoreHandler 创建分数处理器
func NewScoreHandler(service service.LeaderboardService, logger logger.Logger) *ScoreHandler {
	return &ScoreHandler{
		service: service,
		logger:  logger,
	}
}

// Submit 提交分数
func (h *ScoreHandler) Submit(ctx *HTTPContext, resp *HTTPResponse) {
	var req service.SubmitScoreRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.logger.Warn("提交分数参数错误", "error", err)
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证分数范围
	if req.Score < 0 {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "分数不能为负数",
		})
		return
	}

	result, err := h.service.SubmitScore(context.Background(), &req)
	if err != nil {
		h.logger.Error("提交分数失败", "error", err, "request", req)
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "提交分数失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "分数提交成功",
		"data":    result,
	})
}

// BatchSubmit 批量提交分数
func (h *ScoreHandler) BatchSubmit(ctx *HTTPContext, resp *HTTPResponse) {
	var req service.BatchSubmitScoreRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.logger.Warn("批量提交分数参数错误", "error", err)
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 验证批量大小
	if len(req.Scores) == 0 {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "分数列表不能为空",
		})
		return
	}

	if len(req.Scores) > 1000 {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "批量提交分数不能超过1000条",
		})
		return
	}

	// 验证每个分数
	for i, score := range req.Scores {
		if score.Score < 0 {
			ctx.JSON(resp, 400, map[string]interface{}{
				"code":    400,
				"message": fmt.Sprintf("第%d个分数不能为负数", i+1),
			})
			return
		}
		if score.UserID == "" {
			ctx.JSON(resp, 400, map[string]interface{}{
				"code":    400,
				"message": fmt.Sprintf("第%d个用户ID不能为空", i+1),
			})
			return
		}
	}

	result, err := h.service.BatchSubmitScores(context.Background(), &req)
	if err != nil {
		h.logger.Error("批量提交分数失败", "error", err, "leaderboard_id", req.LeaderboardID, "count", len(req.Scores))
		ctx.JSON(resp, 500, map[string]interface{}{
			"code":    500,
			"message": "批量提交分数失败: " + err.Error(),
		})
		return
	}

	// 根据成功率返回不同的状态码
	statusCode := http.StatusOK
	if result.FailureCount > 0 {
		if result.SuccessCount == 0 {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusPartialContent
		}
	}

	ctx.JSON(resp, statusCode, map[string]interface{}{
		"code":    statusCode,
		"message": "批量分数提交完成",
		"data":    result,
	})
}

// GetUserScore 获取用户分数
func (h *ScoreHandler) GetUserScore(ctx *HTTPContext, resp *HTTPResponse) {
	leaderboardID := ctx.Param("leaderboardId")
	userID := ctx.Param("userId")

	if leaderboardID == "" || userID == "" {
		ctx.JSON(resp, 400, map[string]interface{}{
			"code":    400,
			"message": "排行榜ID和用户ID不能为空",
		})
		return
	}

	userRank, err := h.service.GetUserRank(context.Background(), leaderboardID, userID)
	if err != nil {
		h.logger.Error("获取用户分数失败", "error", err, "leaderboard_id", leaderboardID, "user_id", userID)
		ctx.JSON(resp, 404, map[string]interface{}{
			"code":    404,
			"message": "获取用户分数失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(resp, 200, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data":    userRank,
	})
}