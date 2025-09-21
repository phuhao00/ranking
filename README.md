# 分布式游戏排行榜系统

一个高性能、高并发的分布式游戏排行榜微服务系统，支持动态伸缩和多节点部署，确保数据一致性。

## 🚀 特性

- **高性能**: 基于Go语言开发，支持高并发处理
- **分布式**: 支持多节点部署和动态伸缩
- **数据一致性**: 使用MongoDB + Redis确保数据一致性
- **实时排名**: Redis缓存提供毫秒级排名查询
- **多种排行榜**: 支持全局、日榜、周榜、月榜等多种类型
- **RESTful API**: 提供完整的REST API接口
- **监控告警**: 集成Prometheus + Grafana监控
- **容器化**: 支持Docker容器化部署
- **健康检查**: 完善的健康检查和故障转移机制

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   游戏客户端    │    │   游戏服务器    │    │   管理后台      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   负载均衡器    │
                    └─────────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ 排行榜服务节点1  │    │ 排行榜服务节点2  │    │ 排行榜服务节点N  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                    ┌─────────────────┐
                    │   Redis集群     │
                    └─────────────────┘
                                │
                    ┌─────────────────┐
                    │  MongoDB集群    │
                    └─────────────────┘
```

## 🛠️ 技术栈

- **后端**: Go 1.21 + Gin + netcore-go
- **数据库**: MongoDB 7.0 (持久化存储)
- **缓存**: Redis 7.2 (实时排名缓存)
- **日志**: spoor日志库
- **监控**: Prometheus + Grafana
- **容器化**: Docker + Docker Compose
- **负载均衡**: Nginx (可选)
- **服务发现**: Consul (可选)

## 📦 快速开始

### 环境要求

- Go 1.21+
- Docker & Docker Compose
- MongoDB 7.0+
- Redis 7.2+

### 1. 克隆项目

```bash
git clone <repository-url>
cd ranking
```

### 2. 使用Docker Compose启动

```bash
# 启动完整环境
make dev-full

# 或者手动启动
docker-compose --profile tools --profile monitoring up -d
```

### 3. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 获取排行榜列表
curl http://localhost:8080/api/v1/leaderboard/list
```

### 4. 访问管理界面

- **应用服务**: http://localhost:8080
- **MongoDB管理**: http://localhost:8081 (admin/admin)
- **Redis管理**: http://localhost:8082 (admin/admin)
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000 (admin/admin)

## 🔧 本地开发

### 1. 安装依赖

```bash
make deps
```

### 2. 启动开发环境

```bash
# 启动数据库服务
make dev

# 运行应用
make run
```

### 3. 代码检查

```bash
# 格式化代码
make fmt

# 运行所有检查
make check
```

### 4. 运行测试

```bash
make test
```

## 📚 API 文档

### 排行榜管理

#### 创建排行榜

```http
POST /api/v1/leaderboard/create
Content-Type: application/json

{
  "name": "全球积分排行榜",
  "game_id": "my_game",
  "type": "global",
  "sort_order": "desc",
  "max_entries": 10000,
  "config": {
    "timezone": "UTC"
  }
}
```

#### 获取排行榜

```http
GET /api/v1/leaderboard/{id}?limit=100&offset=0
```

#### 获取用户排名

```http
GET /api/v1/leaderboard/{id}/rank/{userId}
```

### 分数管理

#### 提交分数

```http
POST /api/v1/score/submit
Content-Type: application/json

{
  "leaderboard_id": "global_score_demo",
  "user_id": "user_123",
  "score": 95000,
  "source": "game",
  "metadata": {
    "level": 10,
    "achievement": "high_score"
  }
}
```

#### 批量提交分数

```http
POST /api/v1/score/batch
Content-Type: application/json

{
  "leaderboard_id": "global_score_demo",
  "scores": [
    {
      "user_id": "user_123",
      "score": 95000
    },
    {
      "user_id": "user_456",
      "score": 87500
    }
  ]
}
```

### 监控接口

#### 健康检查

```http
GET /health
GET /ready
```

#### 系统指标

```http
GET /api/v1/metrics/
GET /api/v1/metrics/leaderboard/{id}
```

## 🔧 配置说明

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `RANKING_SERVER_HOST` | 服务器监听地址 | `0.0.0.0` |
| `RANKING_SERVER_PORT` | 服务器端口 | `8080` |
| `RANKING_MONGODB_URI` | MongoDB连接字符串 | `mongodb://localhost:27017` |
| `RANKING_MONGODB_DATABASE` | MongoDB数据库名 | `ranking` |
| `RANKING_REDIS_ADDR` | Redis地址 | `localhost:6379` |
| `RANKING_LOG_LEVEL` | 日志级别 | `info` |
| `RANKING_LOG_OUTPUT` | 日志输出 | `stdout` |

### 配置文件

配置文件位于 `configs/config.yaml`，支持以下配置：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30
  write_timeout: 30

mongodb:
  uri: "mongodb://localhost:27017"
  database: "ranking"
  max_pool_size: 100

redis:
  addr: "localhost:6379"
  db: 0
  pool_size: 100

log:
  level: "info"
  format: "json"
  output: "stdout"
```

## 🚀 部署指南

### Docker部署

```bash
# 构建镜像
make docker-build

# 启动服务
make docker-run
```

### Kubernetes部署

```bash
# 应用Kubernetes配置
kubectl apply -f k8s/
```

### 生产环境配置

1. **数据库优化**
   - MongoDB副本集配置
   - Redis集群配置
   - 数据备份策略

2. **性能调优**
   - 连接池大小调整
   - 缓存策略优化
   - 索引优化

3. **监控告警**
   - Prometheus指标收集
   - Grafana仪表板配置
   - 告警规则设置

## 📊 性能指标

- **QPS**: 支持10,000+ QPS
- **延迟**: P99 < 100ms
- **并发**: 支持10,000+并发连接
- **可用性**: 99.9%+

## 🧪 测试

### 单元测试

```bash
make test
```

### 性能测试

```bash
make load-test
```

### 集成测试

```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行集成测试
go test -tags=integration ./test/...
```

## 🔍 故障排查

### 常见问题

1. **连接数据库失败**
   ```bash
   # 检查数据库状态
   make status
   
   # 查看日志
   make logs
   ```

2. **排名计算错误**
   ```bash
   # 重建排行榜缓存
   curl -X POST http://localhost:8080/admin/leaderboard/{id}/rebuild
   ```

3. **性能问题**
   ```bash
   # 查看系统指标
   curl http://localhost:8080/api/v1/metrics/
   ```

### 日志分析

```bash
# 查看应用日志
make logs

# 查看错误日志
docker-compose logs ranking-service | grep ERROR
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范

- 遵循Go代码规范
- 添加必要的注释
- 编写单元测试
- 更新相关文档

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 👥 作者

- **HHaou** - *初始开发* - [GitHub](https://github.com/HHaou)

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - HTTP Web框架
- [MongoDB](https://www.mongodb.com/) - 文档数据库
- [Redis](https://redis.io/) - 内存数据库
- [Docker](https://www.docker.com/) - 容器化平台
- [Prometheus](https://prometheus.io/) - 监控系统

## 📞 支持

如果您有任何问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/HHaou/ranking/issues)
- 发送邮件到 [your-email@example.com]
- 加入讨论群 [QQ群号或微信群]

---

**⭐ 如果这个项目对您有帮助，请给它一个星标！**