# ns-tracking-go

云途物流 Webhook 接收服务 — Go 语言实现

## 架构

```
云途 Webhook → Go 接收 → 标准化 → 写入 tracking_details → Ruby 读缓存
```

**技术栈：** go-zero + GORM + PostgreSQL + Redis + Prometheus

**架构模式：** DDD 四模块（app / domain / infrastructure / pkg）

## 目录结构

```
├── app/                    # 应用层（HTTP + Worker）
├── domain/tracking/        # 领域层（实体 + 服务 + 接口）
├── infrastructure/         # 基础设施层（DB + Cache + Cron + Metrics）
├── pkg/                    # 公共层（错误码 + 工具）
├── docs/                   # 文档
└── test/                   # 集成测试
```

## 核心功能

- **Webhook 接收**：三种信封格式（标准/直接/AES 加密）
- **签名验证**：SHA256(timestamp + encryptKey + body) + 重放窗口
- **标准化**：44 节点映射（39 英文 + 5 中文）+ 海关 80→20
- **缓存写入**：tracking_details（Ruby fetch_cache 兼容）
- **幂等处理**：ON CONFLICT DO NOTHING（每条 track_event 独立）
- **失败重试**：退避 30s→120s→600s→3600s，5 次后死信
- **兜底同步**：每小时扫描未更新 4h 的单号
- **灰度开关**：whitelist / percentage / all 三种模式
- **监控**：Prometheus 指标 + 告警规则

## 快速开始

```bash
# 1. 初始化数据库
psql -f infrastructure/database/migration/001_create_raw_events.sql

# 2. 配置环境变量
export DB_SOURCE="postgres://..."
export WEBHOOK_ENCRYPT_KEY="..."

# 3. 启动
go run app/main.go -f app/etc/app.yaml
```

## 测试

```bash
# 单元测试
go test ./domain/tracking/service/... -cover
go test ./infrastructure/webhook/... -cover

# 集成测试
go test ./test/integration/... -cover

# 代码检查
go vet ./...
```

## 文档

- [部署文档](docs/deployment.md)
- [API 文档](docs/api.md)
- [告警规则](docs/alerts.yaml)
- [Ruby 侧适配](../报告/Ruby侧适配方案.md)
