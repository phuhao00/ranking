# 游戏排行榜系统

基于 Go 语言和 netcore-go 网络框架开发的高性能游戏排行榜系统，支持排行榜管理、分数提交和实时查询功能。

## ✨ 核心特性

- **netcore-go 网络框架**: 基于高性能的 Go 网络框架构建
- **完整中间件**: 日志记录、错误恢复、CORS、限流、安全头等
- **数据持久化**: MongoDB 存储 + Redis 缓存双重保障
- **RESTful API**: 完整的排行榜和分数管理接口
- **性能测试**: 支持高并发压力测试，QPS 可达 4000+
- **健康监控**: 完善的健康检查和系统指标监控
- **容器化部署**: Docker 支持，便于部署和扩展

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

- **网络框架**: netcore-go (高性能 Go 网络框架)
- **数据库**: MongoDB (文档存储)
- **缓存**: Redis (高性能缓存)
- **日志**: spoor v2.0.1 (高性能结构化日志库)
- **中间件**: 日志、恢复、CORS、限流、安全头、请求ID
- **容器化**: Docker + Docker Compose
- **测试**: 功能测试 + 压力测试工具

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MongoDB (本地或远程)
- Redis (本地或远程)

### 1. 克隆项目

```bash
git clone https://github.com/phuhao00/ranking.git
cd ranking
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置数据库

确保 MongoDB 和 Redis 服务正在运行：
```bash
# MongoDB (默认端口 27017)
mongod

# Redis (默认端口 6379)
redis-server
```

### 4. 启动服务

```bash
# 启动排行榜服务
go run cmd/server/main.go
```

### 5. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 查看系统指标
curl http://localhost:8080/api/v1/metrics
```

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

## 📚 API 接口

### 系统接口

#### 健康检查

```http
GET /health
```

响应示例：
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T10:30:00Z"
}
```

#### 系统指标

```http
GET /api/v1/metrics
```

### 排行榜管理

#### 创建排行榜

```http
POST /api/v1/leaderboard/create
Content-Type: application/json

{
  "name": "测试排行榜",
  "description": "这是一个测试排行榜",
  "game_id": "test_game_001",
  "type": "score",
  "order": "desc",
  "max_entries": 100
}
```

#### 获取排行榜列表

```http
GET /api/v1/leaderboard/list
```

### 分数管理

#### 提交分数

```http
POST /api/v1/score/submit
Content-Type: application/json

{
  "user_id": "user_001",
  "username": "玩家1",
  "score": 1000,
  "leaderboard_id": "test_leaderboard_id"
}
```

### 管理接口

```http
GET /admin/status
GET /admin/health
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
  level: "info"          # 日志级别 (基于spoor日志库)
  format: "json"         # 日志格式 (spoor支持json和console格式)
  output: "stdout"        # 日志输出 (spoor支持stdout和文件输出)
```

## 🚀 部署

### Docker 部署

```bash
# 构建镜像
docker build -t ranking-system .

# 运行容器
docker run -p 8080:8080 ranking-system
```

### 配置文件

主要配置文件 `configs/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

mongodb:
  uri: "mongodb://localhost:27017"
  database: "ranking"

redis:
  addr: "localhost:6379"
  db: 0

log:
  level: "info"
  format: "json"
```

## 📊 性能测试结果

基于实际压力测试的性能数据：

### 并发性能

| 并发数 | 接口类型 | QPS | 成功率 | 平均响应时间 |
|--------|----------|-----|--------|-------------|
| 10 | 健康检查 | 894.36 | 100% | 889µs |
| 50 | 健康检查 | 4466.54 | 100% | 924µs |
| 100 | 健康检查 | 3797.90 | 100% | 15.81ms |
| 50 | 指标接口 | 955.01 | 100% | 1.94ms |
| 50 | 混合场景 | 2708.06 | 100% | 4.37ms |

### 性能特点

- **最高 QPS**: 4466.54 (50并发健康检查)
- **推荐并发**: 50以下，保证100%成功率
- **响应时间**: 低并发下毫秒级响应
- **稳定性**: 中低并发下表现优异

## 🧪 测试

### 功能测试

运行排行榜功能测试：
```bash
go run examples/ranking_test.go
```

### 压力测试

运行压力测试（需要服务器运行）：
```bash
# 完整压力测试
go run examples/stress_benchmark.go

# 轻量级负载测试
go run examples/load_benchmark.go
```

### 中间件测试

测试中间件功能：
```bash
# 启动中间件演示服务
go run examples/middleware_demo.go

# 运行中间件测试脚本
powershell -ExecutionPolicy Bypass -File examples/test_middleware.ps1
```

## 📁 项目结构

```
ranking/
├── cmd/server/          # 服务器入口
├── internal/            # 内部代码
│   ├── app/            # 应用程序逻辑
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP 处理器
│   ├── middleware/     # 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   ├── server/         # 服务器配置
│   └── service/        # 业务逻辑层
├── pkg/logger/         # 日志工具
├── examples/           # 示例和测试
├── configs/            # 配置文件
└── scripts/            # 脚本文件
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 📄 许可证

本项目采用 MIT 许可证。

## 👥 作者

- **HHaou** - [GitHub](https://github.com/phuhao00)

## 🙏 致谢

- [netcore-go](https://github.com/phuhao00/netcore-go) - 高性能 Go 网络框架
- [spoor](https://github.com/phuhao00/spoor) - 高性能结构化日志库
- [MongoDB](https://www.mongodb.com/) - 文档数据库
- [Redis](https://redis.io/) - 内存数据库

---

**⭐ 如果这个项目对您有帮助，请给它一个星标！**