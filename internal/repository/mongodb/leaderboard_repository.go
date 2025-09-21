// Package mongodb
// Author: HHaou
// Created: 2024-01-20
// Description: 排行榜MongoDB仓储实现

package mongodb

import (
	"context"
	"fmt"
	"time"

	"ranking/internal/model"
	"ranking/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LeaderboardRepository 排行榜仓储接口实现
type LeaderboardRepository struct {
	client *Client
	logger logger.Logger
}

// NewLeaderboardRepository 创建排行榜仓储
func NewLeaderboardRepository(client *Client, logger logger.Logger) *LeaderboardRepository {
	return &LeaderboardRepository{
		client: client,
		logger: logger,
	}
}

// CreateLeaderboard 创建排行榜
func (r *LeaderboardRepository) CreateLeaderboard(ctx context.Context, leaderboard *model.Leaderboard) error {
	collection := r.client.Collection(model.TableLeaderboards)

	// 设置创建时间
	now := time.Now()
	leaderboard.CreatedAt = now
	leaderboard.UpdatedAt = now
	leaderboard.IsActive = true

	// 检查排行榜ID是否已存在
	count, err := collection.CountDocuments(ctx, bson.M{"leaderboard_id": leaderboard.LeaderboardID})
	if err != nil {
		return fmt.Errorf("检查排行榜ID失败: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("排行榜ID已存在: %s", leaderboard.LeaderboardID)
	}

	result, err := collection.InsertOne(ctx, leaderboard)
	if err != nil {
		return fmt.Errorf("创建排行榜失败: %w", err)
	}

	leaderboard.ID = result.InsertedID.(primitive.ObjectID)

	r.logger.Info("排行榜创建成功",
		"leaderboard_id", leaderboard.LeaderboardID,
		"name", leaderboard.Name,
		"type", leaderboard.Type,
	)

	return nil
}

// GetLeaderboard 获取排行榜配置
func (r *LeaderboardRepository) GetLeaderboard(ctx context.Context, leaderboardID string) (*model.Leaderboard, error) {
	collection := r.client.Collection(model.TableLeaderboards)

	var leaderboard model.Leaderboard
	err := collection.FindOne(ctx, bson.M{
		"leaderboard_id": leaderboardID,
		"is_active":      true,
	}).Decode(&leaderboard)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("排行榜不存在: %s", leaderboardID)
		}
		return nil, fmt.Errorf("获取排行榜失败: %w", err)
	}

	return &leaderboard, nil
}

// UpdateLeaderboard 更新排行榜配置
func (r *LeaderboardRepository) UpdateLeaderboard(ctx context.Context, leaderboard *model.Leaderboard) error {
	collection := r.client.Collection(model.TableLeaderboards)

	leaderboard.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":       leaderboard.Name,
			"type":       leaderboard.Type,
			"sort_order": leaderboard.SortOrder,
			"max_entries": leaderboard.MaxEntries,
			"config":     leaderboard.Config,
			"updated_at": leaderboard.UpdatedAt,
			"is_active":  leaderboard.IsActive,
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{
		"leaderboard_id": leaderboard.LeaderboardID,
	}, update)

	if err != nil {
		return fmt.Errorf("更新排行榜失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("排行榜不存在: %s", leaderboard.LeaderboardID)
	}

	r.logger.Info("排行榜更新成功",
		"leaderboard_id", leaderboard.LeaderboardID,
		"modified_count", result.ModifiedCount,
	)

	return nil
}

// DeleteLeaderboard 删除排行榜（软删除）
func (r *LeaderboardRepository) DeleteLeaderboard(ctx context.Context, leaderboardID string) error {
	collection := r.client.Collection(model.TableLeaderboards)

	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{
		"leaderboard_id": leaderboardID,
	}, update)

	if err != nil {
		return fmt.Errorf("删除排行榜失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("排行榜不存在: %s", leaderboardID)
	}

	r.logger.Info("排行榜删除成功", "leaderboard_id", leaderboardID)
	return nil
}

// ListLeaderboards 获取排行榜列表
func (r *LeaderboardRepository) ListLeaderboards(ctx context.Context, gameID string, limit, offset int64) ([]*model.Leaderboard, int64, error) {
	collection := r.client.Collection(model.TableLeaderboards)

	// 构建查询条件
	filter := bson.M{"is_active": true}
	if gameID != "" {
		filter["game_id"] = gameID
	}

	// 获取总数
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("获取排行榜总数失败: %w", err)
	}

	// 查询数据
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if offset > 0 {
		opts.SetSkip(offset)
	}
	opts.SetSort(bson.M{"created_at": -1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("查询排行榜列表失败: %w", err)
	}
	defer cursor.Close(ctx)

	var leaderboards []*model.Leaderboard
	for cursor.Next(ctx) {
		var leaderboard model.Leaderboard
		if err := cursor.Decode(&leaderboard); err != nil {
			return nil, 0, fmt.Errorf("解析排行榜数据失败: %w", err)
		}
		leaderboards = append(leaderboards, &leaderboard)
	}

	if err := cursor.Err(); err != nil {
		return nil, 0, fmt.Errorf("遍历排行榜数据失败: %w", err)
	}

	return leaderboards, total, nil
}

// CreateScoreRecord 创建分数记录
func (r *LeaderboardRepository) CreateScoreRecord(ctx context.Context, record *model.ScoreRecord) error {
	collection := r.client.Collection(model.TableScoreRecords)

	// 设置创建时间
	now := time.Now()
	record.CreatedAt = now
	if record.SubmittedAt.IsZero() {
		record.SubmittedAt = now
	}

	result, err := collection.InsertOne(ctx, record)
	if err != nil {
		return fmt.Errorf("创建分数记录失败: %w", err)
	}

	record.ID = result.InsertedID.(primitive.ObjectID)

	r.logger.Debug("分数记录创建成功",
		"leaderboard_id", record.LeaderboardID,
		"user_id", record.UserID,
		"score", record.Score,
	)

	return nil
}

// GetUserScore 获取用户在指定排行榜的最新分数
func (r *LeaderboardRepository) GetUserScore(ctx context.Context, leaderboardID, userID string) (*model.ScoreRecord, error) {
	collection := r.client.Collection(model.TableScoreRecords)

	opts := options.FindOne().SetSort(bson.M{"submitted_at": -1})
	var record model.ScoreRecord
	err := collection.FindOne(ctx, bson.M{
		"leaderboard_id": leaderboardID,
		"user_id":        userID,
	}, opts).Decode(&record)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 用户没有分数记录
		}
		return nil, fmt.Errorf("获取用户分数失败: %w", err)
	}

	return &record, nil
}

// GetTopScores 获取排行榜前N名分数
func (r *LeaderboardRepository) GetTopScores(ctx context.Context, leaderboardID string, limit int64, sortOrder model.SortOrder) ([]*model.ScoreRecord, error) {
	collection := r.client.Collection(model.TableScoreRecords)

	// 聚合管道：获取每个用户的最新分数，然后排序
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"leaderboard_id": leaderboardID,
			},
		},
		{
			"$sort": bson.M{
				"user_id":      1,
				"submitted_at": -1,
			},
		},
		{
			"$group": bson.M{
				"_id": "$user_id",
				"latest_record": bson.M{"$first": "$$ROOT"},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": "$latest_record",
			},
		},
	}

	// 添加排序
	sortValue := -1
	if sortOrder == model.SortOrderAsc {
		sortValue = 1
	}
	pipeline = append(pipeline, bson.M{
		"$sort": bson.M{
			"score":        sortValue,
			"submitted_at": -1,
		},
	})

	// 添加限制
	if limit > 0 {
		pipeline = append(pipeline, bson.M{
			"$limit": limit,
		})
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("获取排行榜数据失败: %w", err)
	}
	defer cursor.Close(ctx)

	var records []*model.ScoreRecord
	for cursor.Next(ctx) {
		var record model.ScoreRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("解析分数记录失败: %w", err)
		}
		records = append(records, &record)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("遍历分数记录失败: %w", err)
	}

	return records, nil
}

// GetUserRank 获取用户排名
func (r *LeaderboardRepository) GetUserRank(ctx context.Context, leaderboardID, userID string, sortOrder model.SortOrder) (int64, error) {
	collection := r.client.Collection(model.TableScoreRecords)

	// 先获取用户分数
	userRecord, err := r.GetUserScore(ctx, leaderboardID, userID)
	if err != nil {
		return 0, err
	}
	if userRecord == nil {
		return 0, fmt.Errorf("用户没有分数记录")
	}

	// 构建查询条件
	matchCondition := bson.M{
		"leaderboard_id": leaderboardID,
	}

	if sortOrder == model.SortOrderDesc {
		matchCondition["score"] = bson.M{"$gt": userRecord.Score}
	} else {
		matchCondition["score"] = bson.M{"$lt": userRecord.Score}
	}

	// 聚合管道：计算排名
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"leaderboard_id": leaderboardID,
			},
		},
		{
			"$sort": bson.M{
				"user_id":      1,
				"submitted_at": -1,
			},
		},
		{
			"$group": bson.M{
				"_id": "$user_id",
				"latest_score": bson.M{"$first": "$score"},
			},
		},
		{
			"$match": matchCondition,
		},
		{
			"$count": "rank",
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, fmt.Errorf("计算用户排名失败: %w", err)
	}
	defer cursor.Close(ctx)

	var result struct {
		Rank int64 `bson:"rank"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, fmt.Errorf("解析排名结果失败: %w", err)
		}
	}

	// 排名从1开始
	return result.Rank + 1, nil
}