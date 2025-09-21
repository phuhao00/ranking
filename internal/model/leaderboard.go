// Package model
// Author: HHaou
// Created: 2024-01-20
// Description: 排行榜数据模型定义

package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LeaderboardType 排行榜类型
type LeaderboardType string

const (
	// LeaderboardTypeGlobal 全局排行榜
	LeaderboardTypeGlobal LeaderboardType = "global"
	// LeaderboardTypeDaily 日排行榜
	LeaderboardTypeDaily LeaderboardType = "daily"
	// LeaderboardTypeWeekly 周排行榜
	LeaderboardTypeWeekly LeaderboardType = "weekly"
	// LeaderboardTypeMonthly 月排行榜
	LeaderboardTypeMonthly LeaderboardType = "monthly"
	// LeaderboardTypeSeason 赛季排行榜
	LeaderboardTypeSeason LeaderboardType = "season"
)

// SortOrder 排序方式
type SortOrder string

const (
	// SortOrderDesc 降序（分数高的在前）
	SortOrderDesc SortOrder = "desc"
	// SortOrderAsc 升序（分数低的在前）
	SortOrderAsc SortOrder = "asc"
)

// Leaderboard 排行榜配置
type Leaderboard struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LeaderboardID string           `bson:"leaderboard_id" json:"leaderboard_id"`
	Name        string             `bson:"name" json:"name"`
	GameID      string             `bson:"game_id" json:"game_id"`
	Type        LeaderboardType    `bson:"type" json:"type"`
	SortOrder   SortOrder          `bson:"sort_order" json:"sort_order"`
	MaxEntries  int64              `bson:"max_entries" json:"max_entries"`
	Config      LeaderboardConfig  `bson:"config" json:"config"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
}

// LeaderboardConfig 排行榜配置详情
type LeaderboardConfig struct {
	// ResetTime 重置时间（对于周期性排行榜）
	ResetTime string `bson:"reset_time,omitempty" json:"reset_time,omitempty"`
	// ResetDay 重置日期（1-7表示周一到周日，1-31表示每月的日期）
	ResetDay int `bson:"reset_day,omitempty" json:"reset_day,omitempty"`
	// Timezone 时区
	Timezone string `bson:"timezone,omitempty" json:"timezone,omitempty"`
	// ScoreDecay 分数衰减配置
	ScoreDecay *ScoreDecayConfig `bson:"score_decay,omitempty" json:"score_decay,omitempty"`
	// Rewards 奖励配置
	Rewards []RewardConfig `bson:"rewards,omitempty" json:"rewards,omitempty"`
}

// ScoreDecayConfig 分数衰减配置
type ScoreDecayConfig struct {
	Enabled    bool    `bson:"enabled" json:"enabled"`
	DecayRate  float64 `bson:"decay_rate" json:"decay_rate"`     // 衰减率（每天）
	MinScore   int64   `bson:"min_score" json:"min_score"`       // 最小分数
	DecayStart int     `bson:"decay_start" json:"decay_start"`   // 开始衰减的天数
}

// RewardConfig 奖励配置
type RewardConfig struct {
	RankStart int                    `bson:"rank_start" json:"rank_start"`
	RankEnd   int                    `bson:"rank_end" json:"rank_end"`
	Rewards   map[string]interface{} `bson:"rewards" json:"rewards"`
}

// ScoreRecord 分数记录
type ScoreRecord struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	LeaderboardID string             `bson:"leaderboard_id" json:"leaderboard_id"`
	UserID        string             `bson:"user_id" json:"user_id"`
	Score         int64              `bson:"score" json:"score"`
	PreviousScore int64              `bson:"previous_score" json:"previous_score"`
	Source        string             `bson:"source" json:"source"`
	Metadata      map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	SubmittedAt   time.Time          `bson:"submitted_at" json:"submitted_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// RankingEntry 排名条目
type RankingEntry struct {
	UserID    string                 `bson:"user_id" json:"user_id"`
	Score     int64                  `bson:"score" json:"score"`
	Rank      int64                  `bson:"rank" json:"rank"`
	PrevRank  int64                  `bson:"prev_rank,omitempty" json:"prev_rank,omitempty"`
	UpdatedAt time.Time              `bson:"updated_at" json:"updated_at"`
	Metadata  map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// RankingCache Redis中的排名缓存
type RankingCache struct {
	LeaderboardID string    `json:"leaderboard_id"`
	UserID        string    `json:"user_id"`
	Score         int64     `json:"score"`
	Rank          int64     `json:"rank"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// User 用户信息（简化版）
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Username  string             `bson:"username" json:"username"`
	Avatar    string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Level     int                `bson:"level,omitempty" json:"level,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// LeaderboardStats 排行榜统计信息
type LeaderboardStats struct {
	LeaderboardID string    `bson:"leaderboard_id" json:"leaderboard_id"`
	TotalUsers    int64     `bson:"total_users" json:"total_users"`
	TotalScores   int64     `bson:"total_scores" json:"total_scores"`
	HighestScore  int64     `bson:"highest_score" json:"highest_score"`
	LowestScore   int64     `bson:"lowest_score" json:"lowest_score"`
	AverageScore  float64   `bson:"average_score" json:"average_score"`
	LastUpdated   time.Time `bson:"last_updated" json:"last_updated"`
}

// TableNames 数据库表名常量
const (
	TableLeaderboards = "leaderboards"
	TableScoreRecords = "score_records"
	TableUsers        = "users"
	TableStats        = "leaderboard_stats"
)

// Redis键名模板
const (
	// RedisKeyLeaderboard 排行榜数据键
	RedisKeyLeaderboard = "leaderboard:%s"
	// RedisKeyUserRank 用户排名键
	RedisKeyUserRank = "user_rank:%s:%s"
	// RedisKeyLeaderboardConfig 排行榜配置键
	RedisKeyLeaderboardConfig = "leaderboard_config:%s"
	// RedisKeyLeaderboardStats 排行榜统计键
	RedisKeyLeaderboardStats = "leaderboard_stats:%s"
	// RedisKeyRankingLock 排名计算锁
	RedisKeyRankingLock = "ranking_lock:%s"
)

// GetRedisKey 获取Redis键名
func GetRedisKey(template string, args ...interface{}) string {
	switch len(args) {
	case 0:
		return template
	case 1:
		return fmt.Sprintf(template, args[0])
	case 2:
		return fmt.Sprintf(template, args[0], args[1])
	default:
		return fmt.Sprintf(template, args...)
	}
}