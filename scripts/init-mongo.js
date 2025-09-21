// MongoDB初始化脚本
// Author: HHaou
// Created: 2024-01-20
// Description: 初始化MongoDB数据库和集合

// 切换到ranking数据库
db = db.getSiblingDB('ranking');

// 创建用户
db.createUser({
  user: 'ranking_user',
  pwd: 'ranking_password',
  roles: [
    {
      role: 'readWrite',
      db: 'ranking'
    }
  ]
});

// 创建排行榜集合
db.createCollection('leaderboards');

// 创建分数记录集合
db.createCollection('score_records');

// 创建用户集合
db.createCollection('users');

// 创建统计集合
db.createCollection('leaderboard_stats');

// 创建索引
print('创建排行榜索引...');
db.leaderboards.createIndex({ "leaderboard_id": 1 }, { unique: true });
db.leaderboards.createIndex({ "game_id": 1, "type": 1 });
db.leaderboards.createIndex({ "is_active": 1 });
db.leaderboards.createIndex({ "created_at": 1 });

print('创建分数记录索引...');
db.score_records.createIndex({ "leaderboard_id": 1, "user_id": 1 });
db.score_records.createIndex({ "leaderboard_id": 1, "score": -1 });
db.score_records.createIndex({ "submitted_at": 1 });
db.score_records.createIndex({ "created_at": 1 }, { expireAfterSeconds: 2592000 }); // 30天后过期

print('创建用户索引...');
db.users.createIndex({ "user_id": 1 }, { unique: true });
db.users.createIndex({ "username": 1 });
db.users.createIndex({ "created_at": 1 });

print('创建统计索引...');
db.leaderboard_stats.createIndex({ "leaderboard_id": 1 }, { unique: true });
db.leaderboard_stats.createIndex({ "last_updated": 1 });

// 插入示例数据
print('插入示例排行榜数据...');
db.leaderboards.insertMany([
  {
    leaderboard_id: 'global_score_demo',
    name: '全球积分排行榜',
    game_id: 'demo_game',
    type: 'global',
    sort_order: 'desc',
    max_entries: 10000,
    config: {
      timezone: 'UTC',
      reset_time: '',
      reset_day: 0
    },
    created_at: new Date(),
    updated_at: new Date(),
    is_active: true
  },
  {
    leaderboard_id: 'daily_score_demo',
    name: '每日积分排行榜',
    game_id: 'demo_game',
    type: 'daily',
    sort_order: 'desc',
    max_entries: 1000,
    config: {
      timezone: 'UTC',
      reset_time: '00:00:00',
      reset_day: 0
    },
    created_at: new Date(),
    updated_at: new Date(),
    is_active: true
  },
  {
    leaderboard_id: 'weekly_score_demo',
    name: '每周积分排行榜',
    game_id: 'demo_game',
    type: 'weekly',
    sort_order: 'desc',
    max_entries: 5000,
    config: {
      timezone: 'UTC',
      reset_time: '00:00:00',
      reset_day: 1 // 周一重置
    },
    created_at: new Date(),
    updated_at: new Date(),
    is_active: true
  }
]);

print('插入示例用户数据...');
db.users.insertMany([
  {
    user_id: 'user_001',
    username: 'Player1',
    avatar: '',
    level: 10,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    user_id: 'user_002',
    username: 'Player2',
    avatar: '',
    level: 8,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    user_id: 'user_003',
    username: 'Player3',
    avatar: '',
    level: 12,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

print('插入示例分数记录...');
db.score_records.insertMany([
  {
    leaderboard_id: 'global_score_demo',
    user_id: 'user_001',
    score: 95000,
    previous_score: 90000,
    source: 'game',
    metadata: { level: 10, achievement: 'high_score' },
    submitted_at: new Date(),
    created_at: new Date()
  },
  {
    leaderboard_id: 'global_score_demo',
    user_id: 'user_002',
    score: 87500,
    previous_score: 85000,
    source: 'game',
    metadata: { level: 8, achievement: 'combo_master' },
    submitted_at: new Date(),
    created_at: new Date()
  },
  {
    leaderboard_id: 'global_score_demo',
    user_id: 'user_003',
    score: 102000,
    previous_score: 98000,
    source: 'game',
    metadata: { level: 12, achievement: 'perfect_game' },
    submitted_at: new Date(),
    created_at: new Date()
  }
]);

print('插入排行榜统计数据...');
db.leaderboard_stats.insertMany([
  {
    leaderboard_id: 'global_score_demo',
    total_users: 3,
    total_scores: 3,
    highest_score: 102000,
    lowest_score: 87500,
    average_score: 94833.33,
    last_updated: new Date()
  },
  {
    leaderboard_id: 'daily_score_demo',
    total_users: 0,
    total_scores: 0,
    highest_score: 0,
    lowest_score: 0,
    average_score: 0,
    last_updated: new Date()
  },
  {
    leaderboard_id: 'weekly_score_demo',
    total_users: 0,
    total_scores: 0,
    highest_score: 0,
    lowest_score: 0,
    average_score: 0,
    last_updated: new Date()
  }
]);

print('MongoDB初始化完成！');
print('数据库: ranking');
print('用户: ranking_user');
print('密码: ranking_password');
print('示例排行榜: global_score_demo,