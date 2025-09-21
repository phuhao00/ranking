// Package service
// Author: HHaou
// Created: 2024-01-20
// Description: 排行榜核心业务逻辑服务层

package service

import (
	"context"
	"fmt"
	"time"

	"ranking/internal/model"
	"ranking/internal/repository/mongodb"
	"ranking/internal/repository/redis"
	"ranking/pkg/logger"

	"github.com/google/uuid"
)

// LeaderboardService 排行榜服务接口
type LeaderboardService interface {
	CreateLeaderboard(ctx context.Context, req *CreateLeaderboardRequest) (*model.Leaderboard, error)
	GetLeaderboard(ctx context.Context, leaderboardID string) (*model.Leaderboard, error)
	UpdateLeaderboard(ctx context.Context, leaderboard *model.Leaderboard) error
	DeleteLeaderboard(ctx context.Context, leaderboardID string) error
	ListLeaderboards(ctx context.Context, gameID string, limit, offset int64) ([]*model.Leaderboard, int64, error)
	SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error)
	BatchSubmitScores(ctx context.Context, req *BatchSubmitScoreRequest) (*BatchSubmitScoreResponse, error)
	GetRankings(ctx context.Context, leaderboardID string, limit, offset int64) ([]*model.RankingEntry, error)
	GetUserRank(ctx context.Context, leaderboardID, userID string) (*UserRankResponse, error)
	GetRankingsAroundUser(ctx context.Context, leaderboardID, userID string, count int64) ([]*model.RankingEntry, error)
	RebuildLeaderboard(ctx context.Context, leaderboardID string) error
}

// leaderboardService 排行榜服务实现
type leaderboardService struct {
	mongoRepo *mongodb.LeaderboardRepository
	redisCache *redis.LeaderboardCache
	logger    logger.Logger
}

// NewLeaderboardService 创建排行榜服务
func NewLeaderboardService(
	mongoRepo *mongodb.LeaderboardRepository,
	redisCache *redis.LeaderboardCache,
	logger logger.Logger,
) LeaderboardService {
	return &leaderboardService{
		mongoRepo:  mongoRepo,
		redisCache: redisCache,
		logger:     logger,
	}
}

// CreateLeaderboardRequest 创建排行榜请求
type CreateLeaderboardRequest struct {
	Name       string                `json:"name" binding:"required"`
	GameID     string                `json:"game_id" binding:"required"`
	Type       model.LeaderboardType `json:"type" binding:"required"`
	SortOrder  model.SortOrder       `json:"sort_order" binding:"required"`
	MaxEntries int64                 `json:"max_entries,omitempty"`
	Config     model.LeaderboardConfig `json:"config,omitempty"`
}

// SubmitScoreRequest 提交分数请求
type SubmitScoreRequest struct {
	LeaderboardID string                 `json:"leaderboard_id" binding:"required"`
	UserID        string                 `json:"user_id" binding:"required"`
	Score         int64                  `json:"score" binding:"required"`
	Source        string                 `json:"source,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SubmitScoreResponse 提交分数响应
type SubmitScoreResponse struct {
	UserID       string `json:"user_id"`
	Score        int64  `json:"score"`
	PreviousScore int64  `json:"previous_score"`
	Rank         int64  `json:"rank"`
	PreviousRank int64  `json:"previous_rank"`
	RankChange   int64  `json:"rank_change"`
}

// BatchSubmitScoreRequest 批量提交分数请求
type BatchSubmitScoreRequest struct {
	LeaderboardID string                   `json:"leaderboard_id" binding:"required"`
	Scores        []SubmitScoreRequest     `json:"scores" binding:"required"`
}

// BatchSubmitScoreResponse 批量提交分数响应
type BatchSubmitScoreResponse struct {
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Results      []SubmitScoreResponse  `json:"results"`
	Errors       []string               `json:"errors,omitempty"`
}

// UserRankResponse 用户排名响应
type UserRankResponse struct {
	UserID    string `json:"user_id"`
	Score     int64  `json:"score"`
	Rank      int64  `json:"rank"`
	TotalUsers int64  `json:"total_users"`
}

// CreateLeaderboard 创建排行榜
func (s *leaderboardService) CreateLeaderboard(ctx context.Context, req *CreateLeaderboardRequest) (*model.Leaderboard, error) {
	// 生成排行榜ID
	leaderboardID := fmt.Sprintf("%s_%s_%s", req.GameID, req.Type, uuid.New().String()[:8])

	// 创建排行榜对象
	leaderboard := &model.Leaderboard{
		LeaderboardID: leaderboardID,
		Name:          req.Name,
		GameID:        req.GameID,
		Type:          req.Type,
		SortOrder:     req.SortOrder,
		MaxEntries:    req.MaxEntries,
		Config:        req.Config,
	}

	// 设置默认值
	if leaderboard.MaxEntries <= 0 {
		leaderboard.MaxEntries = 10000 // 默认最大条目数
	}
	if leaderboard.Config.Timezone == "" {
		leaderboard.Config.Timezone = "UTC"
	}

	// 保存到数据库
	err := s.mongoRepo.CreateLeaderboard(ctx, leaderboard)
	if err != nil {
		return nil, fmt.Errorf("创建排行榜失败: %w", err)
	}

	// 缓存排行榜配置
	if err := s.redisCache.SetLeaderboardConfig(ctx, leaderboard); err != nil {
		s.logger.Warn("缓存排行榜配置失败", "error", err, "leaderboard_id", leaderboardID)
	}

	s.logger.Info("排行榜创建成功",
		"leaderboard_id", leaderboardID,
		"name", req.Name,
		"type", req.Type,
	)

	return leaderboard, nil
}

// GetLeaderboard 获取排行榜配置
func (s *leaderboardService) GetLeaderboard(ctx context.Context, leaderboardID string) (*model.Leaderboard, error) {
	// 先从缓存获取
	leaderboard, err := s.redisCache.GetLeaderboardConfig(ctx, leaderboardID)
	if err == nil && leaderboard != nil {
		return leaderboard, nil
	}

	// 从数据库获取
	leaderboard, err = s.mongoRepo.GetLeaderboard(ctx, leaderboardID)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜失败: %w", err)
	}

	// 缓存配置
	if err := s.redisCache.SetLeaderboardConfig(ctx, leaderboard); err != nil {
		s.logger.Warn("缓存排行榜配置失败", "error", err)
	}

	return leaderboard, nil
}

// UpdateLeaderboard 更新排行榜配置
func (s *leaderboardService) UpdateLeaderboard(ctx context.Context, leaderboard *model.Leaderboard) error {
	// 更新数据库
	err := s.mongoRepo.UpdateLeaderboard(ctx, leaderboard)
	if err != nil {
		return fmt.Errorf("更新排行榜失败: %w", err)
	}

	// 更新缓存
	if err := s.redisCache.SetLeaderboardConfig(ctx, leaderboard); err != nil {
		s.logger.Warn("更新排行榜缓存失败", "error", err)
	}

	s.logger.Info("排行榜更新成功", "leaderboard_id", leaderboard.LeaderboardID)
	return nil
}

// DeleteLeaderboard 删除排行榜
func (s *leaderboardService) DeleteLeaderboard(ctx context.Context, leaderboardID string) error {
	// 软删除数据库记录
	err := s.mongoRepo.DeleteLeaderboard(ctx, leaderboardID)
	if err != nil {
		return fmt.Errorf("删除排行榜失败: %w", err)
	}

	// 清除缓存
	if err := s.redisCache.ClearLeaderboard(ctx, leaderboardID); err != nil {
		s.logger.Warn("清除排行榜缓存失败", "error", err)
	}

	s.logger.Info("排行榜删除成功", "leaderboard_id", leaderboardID)
	return nil
}

// ListLeaderboards 获取排行榜列表
func (s *leaderboardService) ListLeaderboards(ctx context.Context, gameID string, limit, offset int64) ([]*model.Leaderboard, int64, error) {
	return s.mongoRepo.ListLeaderboards(ctx, gameID, limit, offset)
}

// SubmitScore 提交分数
func (s *leaderboardService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
	// 获取排行榜配置
	leaderboard, err := s.GetLeaderboard(ctx, req.LeaderboardID)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %w", err)
	}

	// 获取用户当前分数和排名
	previousScore := int64(0)
	previousRank := int64(0)

	currentRecord, err := s.mongoRepo.GetUserScore(ctx, req.LeaderboardID, req.UserID)
	if err == nil && currentRecord != nil {
		previousScore = currentRecord.Score
		previousRank, _ = s.redisCache.GetRank(ctx, req.LeaderboardID, req.UserID, leaderboard.SortOrder)
	}

	// 创建分数记录
	scoreRecord := &model.ScoreRecord{
		LeaderboardID: req.LeaderboardID,
		UserID:        req.UserID,
		Score:         req.Score,
		PreviousScore: previousScore,
		Source:        req.Source,
		Metadata:      req.Metadata,
		SubmittedAt:   time.Now(),
	}

	if scoreRecord.Source == "" {
		scoreRecord.Source = "game"
	}

	// 保存分数记录到数据库
	err = s.mongoRepo.CreateScoreRecord(ctx, scoreRecord)
	if err != nil {
		return nil, fmt.Errorf("保存分数记录失败: %w", err)
	}

	// 更新缓存中的分数
	err = s.redisCache.SetScore(ctx, req.LeaderboardID, req.UserID, req.Score)
	if err != nil {
		s.logger.Error("更新缓存分数失败", "error", err)
		// 缓存失败不影响主流程
	}

	// 获取新排名
	newRank, err := s.redisCache.GetRank(ctx, req.LeaderboardID, req.UserID, leaderboard.SortOrder)
	if err != nil {
		s.logger.Warn("获取新排名失败", "error", err)
		newRank = 0
	}

	// 计算排名变化
	rankChange := int64(0)
	if previousRank > 0 && newRank > 0 {
		rankChange = previousRank - newRank // 正数表示排名上升
	}

	response := &SubmitScoreResponse{
		UserID:        req.UserID,
		Score:         req.Score,
		PreviousScore: previousScore,
		Rank:          newRank,
		PreviousRank:  previousRank,
		RankChange:    rankChange,
	}

	s.logger.Info("分数提交成功",
		"leaderboard_id", req.LeaderboardID,
		"user_id", req.UserID,
		"score", req.Score,
		"rank", newRank,
		"rank_change", rankChange,
	)

	return response, nil
}

// BatchSubmitScores 批量提交分数
func (s *leaderboardService) BatchSubmitScores(ctx context.Context, req *BatchSubmitScoreRequest) (*BatchSubmitScoreResponse, error) {
	response := &BatchSubmitScoreResponse{
		Results: make([]SubmitScoreResponse, 0, len(req.Scores)),
		Errors:  make([]string, 0),
	}

	// 批量处理分数
	for _, scoreReq := range req.Scores {
		scoreReq.LeaderboardID = req.LeaderboardID // 确保使用统一的排行榜ID

		result, err := s.SubmitScore(ctx, &scoreReq)
		if err != nil {
			response.FailureCount++
			response.Errors = append(response.Errors, fmt.Sprintf("用户%s: %v", scoreReq.UserID, err))
			continue
		}

		response.SuccessCount++
		response.Results = append(response.Results, *result)
	}

	s.logger.Info("批量分数提交完成",
		"leaderboard_id", req.LeaderboardID,
		"total", len(req.Scores),
		"success", response.SuccessCount,
		"failure", response.FailureCount,
	)

	return response, nil
}

// GetRankings 获取排行榜排名
func (s *leaderboardService) GetRankings(ctx context.Context, leaderboardID string, limit, offset int64) ([]*model.RankingEntry, error) {
	// 获取排行榜配置
	leaderboard, err := s.GetLeaderboard(ctx, leaderboardID)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %w", err)
	}

	// 先尝试从缓存获取
	if offset == 0 { // 只有获取前N名时才使用缓存
		rankings, err := s.redisCache.GetTopRankings(ctx, leaderboardID, limit, leaderboard.SortOrder)
		if err == nil && len(rankings) > 0 {
			return rankings, nil
		}
	}

	// 从数据库获取
	records, err := s.mongoRepo.GetTopScores(ctx, leaderboardID, limit+offset, leaderboard.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜数据失败: %w", err)
	}

	// 转换为排名条目
	rankings := make([]*model.RankingEntry, 0, len(records))
	for i, record := range records {
		if int64(i) < offset {
			continue // 跳过偏移量之前的记录
		}

		ranking := &model.RankingEntry{
			UserID:    record.UserID,
			Score:     record.Score,
			Rank:      int64(i + 1),
			UpdatedAt: record.SubmittedAt,
			Metadata:  record.Metadata,
		}
		rankings = append(rankings, ranking)

		if int64(len(rankings)) >= limit {
			break
		}
	}

	// 缓存前N名数据
	if offset == 0 && len(rankings) > 0 {
		scores := make(map[string]int64)
		for _, ranking := range rankings {
			scores[ranking.UserID] = ranking.Score
		}
		if err := s.redisCache.BatchSetScores(ctx, leaderboardID, scores); err != nil {
			s.logger.Warn("缓存排行榜数据失败", "error", err)
		}
	}

	return rankings, nil
}

// GetUserRank 获取用户排名
func (s *leaderboardService) GetUserRank(ctx context.Context, leaderboardID, userID string) (*UserRankResponse, error) {
	// 获取排行榜配置
	leaderboard, err := s.GetLeaderboard(ctx, leaderboardID)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %w", err)
	}

	// 获取用户分数
	score, err := s.redisCache.GetScore(ctx, leaderboardID, userID)
	if err != nil || score == 0 {
		// 从数据库获取
		record, err := s.mongoRepo.GetUserScore(ctx, leaderboardID, userID)
		if err != nil {
			return nil, fmt.Errorf("获取用户分数失败: %w", err)
		}
		if record == nil {
			return nil, fmt.Errorf("用户没有分数记录")
		}
		score = record.Score
	}

	// 获取用户排名
	rank, err := s.redisCache.GetRank(ctx, leaderboardID, userID, leaderboard.SortOrder)
	if err != nil || rank == 0 {
		// 从数据库计算排名
		rank, err = s.mongoRepo.GetUserRank(ctx, leaderboardID, userID, leaderboard.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("获取用户排名失败: %w", err)
		}
	}

	// 获取总用户数
	totalUsers, err := s.redisCache.GetLeaderboardSize(ctx, leaderboardID)
	if err != nil {
		s.logger.Warn("获取排行榜总用户数失败", "error", err)
		totalUsers = 0
	}

	return &UserRankResponse{
		UserID:     userID,
		Score:      score,
		Rank:       rank,
		TotalUsers: totalUsers,
	}, nil
}

// GetRankingsAroundUser 获取用户周围的排名
func (s *leaderboardService) GetRankingsAroundUser(ctx context.Context, leaderboardID, userID string, count int64) ([]*model.RankingEntry, error) {
	// 获取排行榜配置
	leaderboard, err := s.GetLeaderboard(ctx, leaderboardID)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %w", err)
	}

	// 从缓存获取用户周围排名
	rankings, err := s.redisCache.GetRankingsAroundUser(ctx, leaderboardID, userID, count, leaderboard.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("获取用户周围排名失败: %w", err)
	}

	return rankings, nil
}

// RebuildLeaderboard 重建排行榜缓存
func (s *leaderboardService) RebuildLeaderboard(ctx context.Context, leaderboardID string) error {
	s.logger.Info("开始重建排行榜缓存", "leaderboard_id", leaderboardID)

	// 获取排行榜配置
	leaderboard, err := s.GetLeaderboard(ctx, leaderboardID)
	if err != nil {
		return fmt.Errorf("获取排行榜配置失败: %w", err)
	}

	// 获取分布式锁
	lockAcquired, err := s.redisCache.AcquireLock(ctx, leaderboardID, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("获取重建锁失败: %w", err)
	}
	if !lockAcquired {
		return fmt.Errorf("排行榜正在重建中，请稍后再试")
	}
	defer s.redisCache.ReleaseLock(ctx, leaderboardID)

	// 清空现有缓存
	if err := s.redisCache.ClearLeaderboard(ctx, leaderboardID); err != nil {
		return fmt.Errorf("清空排行榜缓存失败: %w", err)
	}

	// 从数据库获取所有分数记录
	records, err := s.mongoRepo.GetTopScores(ctx, leaderboardID, leaderboard.MaxEntries, leaderboard.SortOrder)
	if err != nil {
		return fmt.Errorf("获取分数记录失败: %w", err)
	}

	// 批量设置到缓存
	scores := make(map[string]int64)
	for _, record := range records {
		scores[record.UserID] = record.Score
	}

	if len(scores) > 0 {
		if err := s.redisCache.BatchSetScores(ctx, leaderboardID, scores); err != nil {
			return fmt.Errorf("批量设置缓存失败: %w", err)
		}
	}

	s.logger.Info("排行榜缓存重建完成",
		"leaderboard_id", leaderboardID,
		"records_count", len(records),
	)

	return nil
}