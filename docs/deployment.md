# 部署文档

> ns-tracking-go 云途 Webhook 服务

---

## 1. 环境变量

| 变量 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `DB_SOURCE` | 是 | - | PostgreSQL 连接字符串 |
| `REDIS_HOST` | 是 | - | Redis 地址 |
| `REDIS_PASS` | 否 | - | Redis 密码 |
| `WEBHOOK_ENCRYPT_KEY` | 是 | - | 云途 Webhook 加密密钥 |
| `GE_YUN_EXPRESS_TOKEN` | 是 | - | GE 云途 Token |
| `YUN_EXPRESS_TOKEN` | 否 | - | 普通云途 Token（fallback） |
| `WEBHOOK_GO_ENABLED` | 否 | `false` | Go 服务总开关 |
| `WEBHOOK_GO_GRAYSCALE_MODE` | 否 | `whitelist` | 灰度模式 |
| `WEBHOOK_GO_WHITELIST` | 否 | - | 白名单单号（逗号分隔） |
| `WEBHOOK_GO_PERCENTAGE` | 否 | `10` | 百分比模式比例 |
| `YUN_EXPRESS_FALLBACK_PULL` | 否 | `true` | Ruby Pull 兜底开关 |
| `QUEUE_KEY` | 否 | `queue:yun_express_webhook_track` | Redis 队列 key |

## 2. 数据库初始化

```bash
# 执行建表 SQL
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f infrastructure/database/migration/001_create_raw_events.sql
```

## 3. 启动服务

```bash
# HTTP 服务
go run app/main.go -f app/etc/app.yaml

# Worker（定时任务）
go run app/cmd/worker/main.go -f app/cmd/worker/etc/worker.yaml

# Queue Worker（Redis 队列监听）
go run app/cmd/queue/main.go -f app/cmd/queue/etc/worker.yaml
```

## 4. 健康检查

```bash
curl http://localhost:8080/health
```

## 5. 灰度上线步骤

1. 部署 Go 服务，设置 `WEBHOOK_GO_ENABLED=true`
2. 设置 `WEBHOOK_GO_GRAYSCALE_MODE=whitelist`
3. 设置 `WEBHOOK_GO_WHITELIST=测试单号`
4. 云途后台配置 Webhook 推送地址为 Go 服务
5. 触发测试单号的轨迹更新
6. 验证 Go 写入数据与 Ruby Pull 数据一致
7. 逐步扩大白名单或切换到 percentage 模式
8. 全量切换后持续观测 1 周

## 6. 回滚方案

| 场景 | 操作 | 耗时 |
|------|------|------|
| 快速回滚 | `WEBHOOK_GO_ENABLED=false`，重启 Go 服务 | < 1 分钟 |
| Ruby 切回 Pull | `YUN_EXPRESS_FALLBACK_PULL=true` | < 1 分钟 |
| 部分回滚 | 从白名单移除问题单号 | < 5 分钟 |
| 数据回滚 | 按 `tracking_details.synced_at` 时间戳批量回滚 | 需准备 SQL |

## 7. 监控

```bash
# Prometheus 指标
curl http://localhost:8080/metrics

# 关键指标
# - webhook_requests_total{status="success|error"}
# - webhook_latency_seconds
# - raw_events_by_status{status="pending|processed|failed|dead_lettered"}
# - grayscale_decisions_total{mode="...", result="processed|skipped"}
# - sync_job_total{result="success|error"}
```

## 8. 告警阈值

| 指标 | 阈值 | 级别 |
|------|------|------|
| Webhook 失败率 | > 5% (5min) | P1 |
| Webhook P95 延迟 | > 10s | P2 |
| 缓存写入失败率 | > 1% (5min) | P1 |
| 服务宕机 | 3 次健康检查失败 | P0 |
