// Package redis
// Author: HHaou
// Created: 2024-01-20
// Description: 排行榜Redis缓存仓储实现

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"ranking/internal/model"
	"ranking/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// LeaderboardCache 排行榜缓存仓储
type LeaderboardCache struct {
	client *Client
	logger logger.Logger
}

// NewLeaderboardCache 创建排行榜缓存仓储
func NewLeaderboardCache(client *Client, logger logger.Logger) *LeaderboardCache {
	return &LeaderboardCache{
		client: client,
		logger: logger,
	}
}

// SetScore 设置用户分数到排行榜
func (c *LeaderboardCache) SetScore(ctx context.Context, leaderboardID, userID string, score int64) error {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	// 使用有序集合存储排行榜数据
	err := c.client.ZAdd(ctx, key, redis.Z{
		Score:  float64(score),
		Member: userID,
	})

	if err != nil {
		return fmt.Errorf("设置用户分数失败: %w", err)
	}

	// 设置排行榜过期时间（24小时）
	c.client.Expire(ctx, key, 24*time.Hour)

	// 缓存用户排名信息
	userRankKey := model.GetRedisKey(model.RedisKeyUserRank, leaderboardID, userID)
	rankCache := model.RankingCache{
		LeaderboardID: leaderboardID,
		UserID:        userID,
		Score:         score,
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	rankData, _ := json.Marshal(rankCache)
	c.client.Set(ctx, userRankKey, string(rankData), 24*time.Hour)

	c.logger.Debug("用户分数已缓存",
		"leaderboard_id", leaderboardID,
		"user_id", userID,
		"score", score,
	)

	return nil
}

// GetScore 获取用户分数
func (c *LeaderboardCache) GetScore(ctx context.Context, leaderboardID, userID string) (int64, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	score, err := c.client.ZScore(ctx, key, userID)
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 用户不在排行榜中
		}
		return 0, fmt.Errorf("获取用户分数失败: %w", err)
	}

	return int64(score), nil
}

// GetRank 获取用户排名（从1开始）
func (c *LeaderboardCache) GetRank(ctx context.Context, leaderboardID, userID string, sortOrder model.SortOrder) (int64, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	var rank int64
	var err error

	if sortOrder == model.SortOrderDesc {
		// 降序排列，使用ZRevRank
		rank, err = c.client.ZRevRank(ctx, key, userID)
	} else {
		// 升序排列，使用ZRank
		rank, err = c.client.ZRank(ctx, key, userID)
	}

	if err != nil {
		if err == redis.Nil {
			return 0, nil // 用户不在排行榜中
		}
		return 0, fmt.Errorf("获取用户排名失败: %w", err)
	}

	if rank == -1 {
		return 0, nil // 用户不在排行榜中
	}

	// Redis排名从0开始，转换为从1开始
	return rank + 1, nil
}

// GetTopRankings 获取排行榜前N名
func (c *LeaderboardCache) GetTopRankings(ctx context.Context, leaderboardID string, limit int64, sortOrder model.SortOrder) ([]*model.RankingEntry, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	var members []redis.Z
	var err error

	if sortOrder == model.SortOrderDesc {
		// 降序排列，获取分数最高的
		members, err = c.client.ZRevRangeWithScores(ctx, key, 0, limit-1)
	} else {
		// 升序排列，获取分数最低的
		members, err = c.client.ZRangeWithScores(ctx, key, 0, limit-1)
	}

	if err != nil {
		return nil, fmt.Errorf("获取排行榜数据失败: %w", err)
	}

	rankings := make([]*model.RankingEntry, 0, len(members))
	for i, member := range members {
		userID, ok := member.Member.(string)
		if !ok {
			c.logger.Warn("无效的用户ID类型", "member", member.Member)
			continue
		}

		ranking := &model.RankingEntry{
			UserID:    userID,
			Score:     int64(member.Score),
			Rank:      int64(i + 1), // 排名从1开始
			UpdatedAt: time.Now(),
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// GetRankingsAroundUser 获取用户周围的排名
func (c *LeaderboardCache) GetRankingsAroundUser(ctx context.Context, leaderboardID, userID string, count int64, sortOrder model.SortOrder) ([]*model.RankingEntry, error) {
	// 先获取用户排名
	userRank, err := c.GetRank(ctx, leaderboardID, userID, sortOrder)
	if err != nil {
		return nil, err
	}
	if userRank == 0 {
		return nil, fmt.Errorf("用户不在排行榜中")
	}

	// 计算范围
	half := count / 2
	start := userRank - half - 1 // Redis索引从0开始
	end := userRank + half - 1

	if start < 0 {
		start = 0
	}

	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	var members []redis.Z
	if sortOrder == model.SortOrderDesc {
		members, err = c.client.ZRevRangeWithScores(ctx, key, start, end)
	} else {
		members, err = c.client.ZRangeWithScores(ctx, key, start, end)
	}

	if err != nil {
		return nil, fmt.Errorf("获取用户周围排名失败: %w", err)
	}

	rankings := make([]*model.RankingEntry, 0, len(members))
	for i, member := range members {
		userID, ok := member.Member.(string)
		if !ok {
			continue
		}

		ranking := &model.RankingEntry{
			UserID:    userID,
			Score:     int64(member.Score),
			Rank:      start + int64(i) + 1, // 计算实际排名
			UpdatedAt: time.Now(),
		}
		rankings = append(rankings, ranking)
	}

	return rankings, nil
}

// RemoveUser 从排行榜中移除用户
func (c *LeaderboardCache) RemoveUser(ctx context.Context, leaderboardID, userID string) error {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	err := c.client.ZRem(ctx, key, userID)
	if err != nil {
		return fmt.Errorf("移除用户失败: %w", err)
	}

	// 删除用户排名缓存
	userRankKey := model.GetRedisKey(model.RedisKeyUserRank, leaderboardID, userID)
	c.client.Del(ctx, userRankKey)

	c.logger.Debug("用户已从排行榜移除",
		"leaderboard_id", leaderboardID,
		"user_id", userID,
	)

	return nil
}

// GetLeaderboardSize 获取排行榜用户数量
func (c *LeaderboardCache) GetLeaderboardSize(ctx context.Context, leaderboardID string) (int64, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	size, err := c.client.ZCard(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("获取排行榜大小失败: %w", err)
	}

	return size, nil
}

// ClearLeaderboard 清空排行榜
func (c *LeaderboardCache) ClearLeaderboard(ctx context.Context, leaderboardID string) error {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	err := c.client.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("清空排行榜失败: %w", err)
	}

	c.logger.Info("排行榜已清空", "leaderboard_id", leaderboardID)
	return nil
}

// SetLeaderboardConfig 缓存排行榜配置
func (c *LeaderboardCache) SetLeaderboardConfig(ctx context.Context, leaderboard *model.Leaderboard) error {
	key := model.GetRedisKey(model.RedisKeyLeaderboardConfig, leaderboard.LeaderboardID)

	data, err := json.Marshal(leaderboard)
	if err != nil {
		return fmt.Errorf("序列化排行榜配置失败: %w", err)
	}

	err = c.client.Set(ctx, key, string(data), 24*time.Hour)
	if err != nil {
		return fmt.Errorf("缓存排行榜配置失败: %w", err)
	}

	return nil
}

// GetLeaderboardConfig 获取缓存的排行榜配置
func (c *LeaderboardCache) GetLeaderboardConfig(ctx context.Context, leaderboardID string) (*model.Leaderboard, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboardConfig, leaderboardID)

	data, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜配置失败: %w", err)
	}
	if data == "" {
		return nil, nil // 配置不存在
	}

	var leaderboard model.Leaderboard
	err = json.Unmarshal([]byte(data), &leaderboard)
	if err != nil {
		return nil, fmt.Errorf("反序列化排行榜配置失败: %w", err)
	}

	return &leaderboard, nil
}

// BatchSetScores 批量设置分数
func (c *LeaderboardCache) BatchSetScores(ctx context.Context, leaderboardID string, scores map[string]int64) error {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	// 使用管道批量操作
	pipe := c.client.Pipeline()

	// 添加所有分数
	for userID, score := range scores {
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(score),
			Member: userID,
		})
	}

	// 设置过期时间
	pipe.Expire(ctx, key, 24*time.Hour)

	// 执行管道
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("批量设置分数失败: %w", err)
	}

	c.logger.Debug("批量设置分数完成",
		"leaderboard_id", leaderboardID,
		"count", len(scores),
	)

	return nil
}

// GetScoreRange 获取分数范围内的用户数量
func (c *LeaderboardCache) GetScoreRange(ctx context.Context, leaderboardID string, minScore, maxScore int64) (int64, error) {
	key := model.GetRedisKey(model.RedisKeyLeaderboard, leaderboardID)

	count, err := c.client.ZCount(ctx, key, strconv.FormatInt(minScore, 10), strconv.FormatInt(maxScore, 10))
	if err != nil {
		return 0, fmt.Errorf("获取分数范围用户数量失败: %w", err)
	}

	return count, nil
}

// AcquireLock 获取排行榜计算锁
func (c *LeaderboardCache) AcquireLock(ctx context.Context, leaderboardID string, ttl time.Duration) (bool, error) {
	key := model.GetRedisKey(model.RedisKeyRankingLock, leaderboardID)

	success, err := c.client.SetNX(ctx, key, "locked", ttl)
	if err != nil {
		return false, fmt.Errorf("获取排行榜锁失败: %w", err)
	}

	return success, nil
}

// ReleaseLock 释放排行榜计算锁
func (c *LeaderboardCache) ReleaseLock(ctx context.Context, leaderboardID string) error {
	key := model.GetRedisKey(model.RedisKeyRankingLock, leaderboardID)

	err := c.client.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("释放排行榜锁失败: %w", err)
	}

	return nil
}