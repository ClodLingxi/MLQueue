# MLQueue - 高并发ML训练任务队列系统

一个基于Go和React的高并发ML训练任务队列管理系统，支持任务调度、优先级管理、实时监控和Webhook通知。

## 架构概述

### 后端架构（Go）

**技术栈：**
- **Web框架**: Gin（高性能HTTP路由）
- **数据库**: PostgreSQL（任务持久化）
- **缓存/队列**: Redis（高速队列 + 速率限制）
- **认证**: JWT Bearer Token
- **并发**: Goroutines + Channels工作池

**高并发设计特性：**
1. **连接池管理**: 数据库和Redis连接池（默认100个连接）
2. **工作池模式**: 可配置的Goroutine工作池（默认10个worker）
3. **优先级队列**: Redis Sorted Set实现的优先级队列
4. **速率限制**: 基于Redis的滑动窗口速率限制
5. **优雅关闭**: 支持优雅关闭，确保正在处理的任务完成
6. **Webhook异步通知**: 异步发送任务状态变更通知

**项目结构：**
```
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接（PostgreSQL + Redis）
│   ├── models/          # 数据模型
│   ├── queue/           # 队列管理系统
│   ├── handlers/        # API处理器
│   ├── middleware/      # 认证、CORS、速率限制中间件
│   ├── routes/          # 路由配置
│   └── services/        # 业务服务（Webhook等）
├── web/                 # React前端
└── main.go              # 程序入口
```

### 前端架构（React + TypeScript）

**技术栈：**
- **框架**: React 18 + TypeScript + Vite
- **样式**: TailwindCSS
- **状态管理**: Zustand
- **HTTP客户端**: Axios
- **图标**: Lucide React

**功能模块：**
1. **任务管理**: 创建、查看、取消、过滤任务
2. **队列监控**: 实时查看队列状态、运行中任务
3. **统计仪表板**: 任务统计信息（待开发）
4. **自动刷新**: 每3-5秒自动更新数据

## 快速开始

### 前置要求

- Go 1.25+
- Node.js 18+
- PostgreSQL 12+
- Redis 6+

### 1. 数据库设置

**PostgreSQL:**
```bash
# 创建数据库
createdb mlqueue_db

# 创建用户
psql -c "CREATE USER mlqueue WITH PASSWORD 'your_password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE mlqueue_db TO mlqueue;"
```

**Redis:**
```bash
# 启动Redis（默认端口6379）
redis-server
```

### 2. 后端设置

```bash
# 复制环境变量配置
cp .env.example .env

# 编辑.env文件，配置数据库连接
vim .env

# 安装依赖
go mod download

# 运行服务器
go run main.go
```

服务器将在 `http://localhost:8080` 启动

### 3. 前端设置

```bash
cd web

# 安装依赖
npm install

# 复制环境变量
cp .env.example .env

# 编辑.env文件，设置API URL和API Key
vim .env

# 启动开发服务器
npm run dev
```

前端将在 `http://localhost:5173` 启动

## 配置说明

### 后端环境变量 (.env)

```env
# 服务器配置
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
ENV=development

# 数据库配置（PostgreSQL）
DB_HOST=localhost
DB_PORT=5432
DB_USER=mlqueue
DB_PASSWORD=your_password
DB_NAME=mlqueue_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=100    # 最大连接数（高并发）
DB_MAX_IDLE_CONNS=10     # 空闲连接数

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=100      # 连接池大小（高并发）

# JWT配置
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRY_HOURS=24

# 速率限制
RATE_LIMIT_STANDARD=100   # 标准用户：100请求/分钟
RATE_LIMIT_PREMIUM=1000   # 高级用户：1000请求/分钟
RATE_LIMIT_BATCH=10       # 批量操作：10请求/分钟

# 队列配置
QUEUE_WORKER_COUNT=10     # Worker数量（并发处理任务数）
QUEUE_MAX_SIZE=10000      # 队列最大容量

# Webhook配置
WEBHOOK_TIMEOUT_SECONDS=30
WEBHOOK_RETRY_COUNT=3
```

### 前端环境变量 (web/.env)

```env
VITE_API_URL=http://localhost:8080/v1
VITE_API_KEY=your-api-key-here
```

## API文档

详细API文档请参考 `CLOUD_API.md`

### 主要端点

**任务管理:**
- `POST /v1/tasks` - 创建任务
- `POST /v1/tasks/batch` - 批量创建任务
- `GET /v1/tasks` - 列出任务
- `GET /v1/tasks/:id` - 获取任务详情
- `PATCH /v1/tasks/:id/priority` - 更新优先级
- `POST /v1/tasks/:id/cancel` - 取消任务

**队列管理:**
- `GET /v1/queue/status` - 获取队列状态
- `POST /v1/queue/pause` - 暂停队列
- `POST /v1/queue/resume` - 恢复队列
- `POST /v1/queue/reorder` - 重新排序

**统计:**
- `GET /v1/statistics/tasks` - 获取统计信息
- `GET /v1/tasks/:id/logs` - 获取任务日志

## 高并发性能优化

### 1. 数据库层
- **连接池**: 最大100个连接，避免连接耗尽
- **预编译语句**: GORM PrepareStmt缓存
- **索引优化**: task_id, user_id, status, priority字段建立索引

### 2. Redis层
- **连接池**: 100个连接池大小
- **Sorted Set**: 实现O(log N)优先级队列
- **Pipeline**: 批量操作减少网络往返

### 3. 应用层
- **Goroutine工作池**: 限制并发数，防止资源耗尽
- **Channel**: 用于worker通信，实现高效调度
- **Context**: 支持超时和取消操作
- **优雅关闭**: WaitGroup确保任务完成

### 4. 网络层
- **HTTP超时**: ReadTimeout, WriteTimeout, IdleTimeout
- **速率限制**: 防止API滥用
- **CORS**: 配置跨域支持

## 部署指南

### Docker部署（推荐）

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: mlqueue_db
      POSTGRES_USER: mlqueue
      POSTGRES_PASSWORD: your_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  backend:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis

  frontend:
    build: ./web
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
```

### 生产环境建议

1. **数据库优化**:
   - 启用PostgreSQL连接池（pgbouncer）
   - 配置Redis持久化（AOF + RDB）
   - 定期备份数据库

2. **负载均衡**:
   - 使用Nginx/HAProxy进行负载均衡
   - 多实例部署后端服务
   - Redis Cluster用于高可用

3. **监控**:
   - Prometheus + Grafana监控
   - 日志聚合（ELK Stack）
   - APM工具（Jaeger/DataDog）

4. **安全**:
   - HTTPS/TLS加密
   - 定期轮换API Key
   - 防火墙规则
   - 限制数据库访问

## 开发指南

### 添加新的API端点

1. 在 `internal/handlers/` 创建处理函数
2. 在 `internal/routes/routes.go` 注册路由
3. 更新前端 `web/src/api/client.ts`
4. 在 `CLOUD_API.md` 更新文档

### 扩展队列功能

修改 `internal/queue/queue.go` 中的处理逻辑：

```go
func (qm *QueueManager) processTask(workerID int, taskID string) {
    // 你的任务处理逻辑
}
```

### 添加新的UI组件

在 `web/src/components/` 创建新组件，并在App.tsx中引用。

## 故障排除

### 常见问题

**1. 数据库连接失败**
- 检查PostgreSQL是否运行
- 验证.env中的数据库凭据
- 确认防火墙允许5432端口

**2. Redis连接失败**
- 检查Redis服务状态
- 验证Redis配置
- 检查6379端口是否开放

**3. 前端无法连接后端**
- 确认后端服务运行在正确端口
- 检查CORS配置
- 验证API Key是否正确

**4. 任务未执行**
- 检查队列是否暂停
- 查看worker日志
- 确认Redis队列中有任务

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License

## 联系方式

如有问题，请提交Issue或联系开发团队。
