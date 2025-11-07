# MLqueue

<div align="center">

**高性能机器学习训练任务队列管理系统**

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![Python Version](https://img.shields.io/badge/Python-3.8%2B-3776AB?style=flat&logo=python)](https://python.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[功能特性](#功能特性) • [系统架构](#系统架构) • [快速开始](#快速开始) • [文档](#文档) • [API 参考](#api-参考)

</div>

---

## 项目简介

MLqueue 是一个生产级的高并发任务队列管理系统，专为机器学习训练工作流设计。它提供智能任务调度、优先级管理、资源协调和实时监控功能。

### 核心亮点

- **高性能**: 基于 Go 构建，支持 100+ 并发连接和工作池架构
- **双架构模式**: 同时支持云端调度（V1）和 Python 驱动（V2）执行模式
- **优先级队列**: 基于 Redis 的优先级队列，O(log N) 时间复杂度
- **实时监控**: 基于 Web 的任务和队列管理仪表板
- **Python SDK**: 简洁的装饰器风格 API，易于集成
- **生产就绪**: 速率限制、优雅关闭、Webhook 通知等企业级功能

---

## 功能特性

### 核心能力

- **任务管理**: 创建、排队、执行、取消和监控训练任务
- **优先级调度**: 动态调整优先级以优化资源利用
- **批量操作**: 高效处理多个任务
- **配置管理**: 将配置与执行逻辑分离
- **结果追踪**: 使用 JSONB 支持存储和检索训练结果
- **队列控制**: 暂停、恢复、重新排序队列操作
- **Webhook 通知**: 任务状态变更的异步回调

### 性能特性

- **连接池**: 100 个并发数据库/Redis 连接
- **工作池**: 可配置的 Goroutine 工作线程（默认：10）
- **速率限制**: 基于 Redis 的滑动窗口（100-1000 请求/分钟）
- **优雅关闭**: 确保运行中的任务在关闭前完成
- **心跳监控**: 追踪训练单元连接状态（V2）

---

## 系统架构

### 系统总览

```
┌─────────────────────────────────────────────────────────────┐
│                        MLqueue 系统                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │   前端界面   │◄────►│  Go 后端     │◄────►│   Redis   │ │
│  │   (Vue 3)    │      │  (Gin + GORM)│      │  (队列)   │ │
│  └──────────────┘      └──────┬───────┘      └───────────┘ │
│                               │                              │
│                               ▼                              │
│                        ┌─────────────┐                       │
│                        │ PostgreSQL  │                       │
│                        │  (存储)     │                       │
│                        └─────────────┘                       │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Python SDK (客户端)                     │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 后端架构 (Go)

```
backend/
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # PostgreSQL + Redis 连接
│   ├── models/          # 数据模型（Task、User、Group、Queue、Unit）
│   ├── queue/           # 队列管理和工作池
│   ├── handlers/        # HTTP 请求处理器（V1 & V2 API）
│   ├── middleware/      # 认证、CORS、速率限制
│   ├── routes/          # 路由定义
│   └── services/        # 业务逻辑（Webhook 等）
├── main.go              # 应用入口
├── go.mod               # Go 模块依赖
└── Dockerfile           # 容器镜像
```

**技术栈：**

- **Web 框架**: [Gin](https://gin-gonic.com/) - 高性能 HTTP 路由器
- **数据库**: [PostgreSQL](https://www.postgresql.org/) + [GORM](https://gorm.io/)
- **缓存/队列**: [Redis](https://redis.io/) 配合 [go-redis](https://github.com/redis/go-redis)
- **认证**: JWT Bearer Token
- **并发**: Goroutines + Channels

### 前端 (Vue 3)

```
frontend/
├── src/
│   ├── api/              # API 客户端（V1 & V2）
│   ├── components/       # Vue 组件
│   │   ├── TaskList.vue
│   │   ├── TaskCreate.vue
│   │   ├── QueueStatus.vue
│   │   └── v2/           # V2 架构组件
│   ├── App.vue           # 主应用
│   └── main.js           # 入口文件
└── vite.config.js
```

**技术栈：**

- **框架**: Vue 3 (Composition API) + Vite
- **HTTP 客户端**: Axios
- **样式**: 原生 CSS（响应式设计）

### Python SDK

```
python-sdk/
├── mlqueue/
│   ├── client.py         # V1 API 客户端
│   ├── v2_client.py      # V2 API 客户端
│   ├── trainer.py        # MLTrainer 封装
│   ├── task.py           # 任务模型
│   └── utils.py          # 装饰器和工具
└── examples/
    ├── v2_basic_usage.py
    └── v2_pytorch_example.py
```

---

## 双架构模式

### V1: 云端调度模式

**适用场景**: 使用云资源自动执行任务

```
Python 客户端 → 上传配置 → 后端队列 → 工作池 → 执行 → 存储结果
```

**最适合：**

- 简单的训练工作流
- 无服务器部署
- 自动化流水线
- 快速实验

### V2: Python 驱动模式

**适用场景**: 需要本地 GPU 控制的复杂训练

```
Web UI → 创建配置 → 后端存储 → Python 客户端同步 → 本地执行 → 上传结果
```

**最适合：**

- 需要特定 GPU 控制的复杂训练
- 交互式研究实验
- 本地资源利用
- 细粒度执行控制

---

## 快速开始

### 前置要求

- Go 1.24+
- PostgreSQL 12+
- Redis 6+
- Node.js 18+（前端）
- Python 3.8+（SDK）

### 1. 克隆仓库

```bash
git clone https://github.com/yourusername/mlqueue.git
cd mlqueue
```

### 2. 启动数据库服务

使用 Docker Compose：

```bash
cd backend
docker-compose up -d postgres redis
```

或手动启动：

```bash
# PostgreSQL
createdb mlqueue_db
psql -c "CREATE USER mlqueue WITH PASSWORD 'your_password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE mlqueue_db TO mlqueue;"

# Redis
redis-server
```

### 3. 配置后端

```bash
cd backend

# 复制环境变量模板
cp .env.example .env

# 编辑配置
vim .env
```

**最小 .env 配置：**

```env
# 服务器
SERVER_PORT=8080
ENV=development

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=mlqueue
DB_PASSWORD=your_password
DB_NAME=mlqueue_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=change-this-secret-key
```

### 4. 运行后端

```bash
# 安装依赖
go mod download

# 运行服务器
go run main.go
```

服务器启动在 `http://localhost:8080`

### 5. 运行前端（可选）

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端地址 `http://localhost:5173`

### 6. 安装 Python SDK（可选）

```bash
cd python-sdk
pip install -e .
```

---

## 使用示例

### Python SDK - V2 模式（推荐）

```python
from mlqueue import MLQueueV2Client

# 初始化客户端
client = MLQueueV2Client(
    api_url="http://localhost:8080",
    api_key="your-api-key"
)

# 创建项目结构
group = client.create_group(
    name="ResNet 实验",
    description="ImageNet 分类实验"
)

# 创建训练单元
unit = client.create_training_unit(
    group_id=group.group_id,
    name="ResNet50",
    config={"model": "resnet50", "dataset": "imagenet"}
)

# 添加训练队列（可从 Web UI 添加）
queue = client.create_queue(
    unit_id=unit.unit_id,
    parameters={"lr": 0.001, "batch_size": 64, "epochs": 100}
)

# 同步并执行队列
client.sync_unit(unit.unit_id)
queues = client.list_queues(unit.unit_id, status="pending")

for q in queues:
    # 开始执行
    client.start_queue(q.queue_id)

    # 你的训练逻辑
    result = train_model(q.parameters)

    # 完成并提交结果
    client.complete_queue(
        queue_id=q.queue_id,
        result=result,
        metrics={"accuracy": 0.95, "loss": 0.123}
    )
```

### Python SDK - V1 模式

```python
from mlqueue import MLTrainer

trainer = MLTrainer(
    api_url="http://localhost:8080/v1",
    api_key="your-api-key"
)

# 批量创建训练配置
configs = [
    {"name": "lr_0.001", "config": {"lr": 0.001}, "priority": 1},
    {"name": "lr_0.01", "config": {"lr": 0.01}, "priority": 2},
]

tasks = trainer.add_configs(configs)

# 执行训练
with trainer.start_training("experiment", config) as ctx:
    cfg = ctx.get_config()
    result = train_model(**cfg)
    ctx.log_result(result)
```

### 后端 API - 直接 HTTP 调用

```bash
# 创建任务
curl -X POST http://localhost:8080/v1/tasks \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "训练任务",
    "config": {"lr": 0.001, "epochs": 100},
    "priority": 1
  }'

# 获取队列状态
curl http://localhost:8080/v1/queue/status \
  -H "Authorization: Bearer YOUR_API_KEY"

# 列出任务
curl "http://localhost:8080/v1/tasks?status=completed&limit=10" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

## API 参考

### V1 API（云端调度）

| 端点                       | 方法    | 描述     |
|--------------------------|-------|--------|
| `/v1/tasks`              | POST  | 创建任务   |
| `/v1/tasks/batch`        | POST  | 批量创建任务 |
| `/v1/tasks`              | GET   | 列出任务   |
| `/v1/tasks/:id`          | GET   | 获取任务详情 |
| `/v1/tasks/:id/priority` | PATCH | 更新优先级  |
| `/v1/tasks/:id/cancel`   | POST  | 取消任务   |
| `/v1/queue/status`       | GET   | 队列状态   |
| `/v1/queue/pause`        | POST  | 暂停队列   |
| `/v1/queue/resume`       | POST  | 恢复队列   |

### V2 API（Python 驱动）

| 端点                        | 方法   | 描述     |
|---------------------------|------|--------|
| `/v2/groups`              | POST | 创建组    |
| `/v2/groups`              | GET  | 列出组    |
| `/v2/groups/:id/units`    | POST | 创建训练单元 |
| `/v2/units/:id`           | GET  | 获取单元详情 |
| `/v2/units/:id/sync`      | POST | 同步配置   |
| `/v2/units/:id/heartbeat` | POST | 更新心跳   |
| `/v2/units/:id/queues`    | POST | 创建队列   |
| `/v2/queues`              | GET  | 列出队列   |
| `/v2/queues/:id/start`    | POST | 开始执行   |
| `/v2/queues/:id/complete` | POST | 提交结果完成 |

**完整 API 文档**: 参见 `backend/API_V2.md`

---

## 配置说明

### 后端环境变量

```env
# 服务器配置
SERVER_PORT=8080              # HTTP 服务器端口
SERVER_HOST=0.0.0.0           # 绑定地址
ENV=production                # 环境（development/production）

# 数据库（PostgreSQL）
DB_HOST=localhost
DB_PORT=5432
DB_USER=mlqueue
DB_PASSWORD=your_password
DB_NAME=mlqueue_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=100         # 最大连接数
DB_MAX_IDLE_CONNS=10          # 空闲连接数

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=100           # 连接池大小

# 认证
JWT_SECRET=your-secret-key    # 生产环境请修改！
JWT_EXPIRY_HOURS=24

# 速率限制
RATE_LIMIT_STANDARD=100       # 每分钟请求数
RATE_LIMIT_PREMIUM=1000
RATE_LIMIT_BATCH=10

# 队列配置
QUEUE_WORKER_COUNT=10         # 并发工作线程
QUEUE_MAX_SIZE=10000          # 队列最大容量

# Webhooks
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_RETRY_COUNT=3
```

### 前端配置

创建 `frontend/.env`：

```env
VITE_API_URL=http://localhost:8080/v1
VITE_API_KEY=your-api-key
```

---

## 部署

### Docker 部署（推荐）

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend

# 停止服务
docker-compose down
```

### 生产环境部署

1. **构建后端**：

```bash
cd backend
go build -o mlqueue-server main.go
```

2. **构建前端**：

```bash
cd frontend
npm run build
# 使用 Nginx 或其他 Web 服务器部署 dist/ 目录
```

3. **数据库设置**：

- 启用 PostgreSQL 连接池（pgbouncer）
- 配置 Redis 持久化（AOF + RDB）
- 设置定期数据库备份

4. **安全性**：

- 使用 HTTPS/TLS 加密
- 定期轮换 JWT 密钥
- 配置防火墙规则
- 使用基于环境的密钥管理

5. **监控**：

- 设置 Prometheus 指标收集
- 配置 Grafana 仪表板
- 启用应用日志（ELK Stack）

---

## 性能优化

### 数据库层

- 连接池（最大 100 个连接）
- GORM 预编译语句缓存
- 索引列：`task_id`、`user_id`、`status`、`priority`

### Redis 层

- 有序集合实现 O(log N) 优先级队列操作
- Pipeline 批处理减少网络往返
- 连接池（100 个连接）

### 应用层

- Goroutine 工作池（可配置并发度）
- 基于 Channel 的任务分发
- 基于 Context 的超时和取消
- WaitGroup 实现优雅关闭

---

## 开发

### 运行测试

```bash
# 后端测试
cd backend
go test ./...

# Python SDK 测试
cd python-sdk
pytest tests/
```

### 代码结构

- **后端**: 遵循清洁架构原则
- **前端**: 基于组件的 Vue 3 结构
- **Python SDK**: 需要类型提示和文档字符串

### 添加新功能

1. **新 API 端点**：
    - 在 `internal/handlers/` 添加处理器
    - 在 `internal/routes/` 注册路由
    - 更新 API 文档

2. **新前端组件**：
    - 在 `frontend/src/components/` 创建
    - 在 `App.vue` 中引入

3. **Python SDK 扩展**：
    - 向 `v2_client.py` 或 `client.py` 添加方法
    - 在 `examples/` 中更新示例

---

## 故障排除

### 常见问题

**数据库连接错误**

- 验证 PostgreSQL 是否运行：`pg_isready`
- 检查 `.env` 中的凭据
- 确保数据库存在：`psql -l`

**Redis 连接错误**

- 检查 Redis 状态：`redis-cli ping`
- 验证 Redis 配置
- 确保端口 6379 可访问

**任务未执行**

- 检查队列状态：`GET /v1/queue/status`
- 验证配置中的工作线程数
- 查看后端日志中的错误

**前端 API 错误**

- 验证后端运行在正确端口
- 检查 CORS 配置
- 验证前端 `.env` 中的 API 密钥

---

## 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 开启 Pull Request

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 致谢

使用以下技术构建：

- [Go](https://golang.org/) - 后端语言
- [Gin](https://gin-gonic.com/) - Web 框架
- [GORM](https://gorm.io/) - ORM 库
- [Vue.js](https://vuejs.org/) - 前端框架
- [Redis](https://redis.io/) - 队列和缓存
- [PostgreSQL](https://www.postgresql.org/) - 数据库

---

## 支持

- **文档**: 参见 `backend/README.md`、`frontend/README.md`、`python-sdk/README.md`
- **问题**: [GitHub Issues](https://github.com/yourusername/mlqueue/issues)
- **讨论**: [GitHub Discussions](https://github.com/yourusername/mlqueue/discussions)

---

<div align="center">

**MLqueue** - 智能 ML 训练任务管理

为 ML 社区用心打造

</div>
