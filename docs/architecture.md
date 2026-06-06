# 系统架构说明

## 架构总览

```
┌────────────────────────────────────────────────────────────┐
│                       app/  应用层                          │
│  main.go  →  routes.go  →  handler  →  logic              │
│  cmd/worker/  →  Asynq consumer                           │
│  svc/ServiceContext  →  Config/DB/Redis/EventBus/Locker    │
└──────────────┬──────────────┬──────────────────────────────┘
               │ 调用          │ 调用
               ▼               ▼
┌──────────────────────┐  ┌──────────────────────────────────┐
│   domain/  领域层      │  │   infrastructure/  基础设施层     │
│                       │  │                                  │
│  entity/  实体模型     │  │  database/   GORM PostgreSQL     │
│  repo/    仓储接口 ────┼──▶  ├─ repo_impl/  仓储实现          │
│  service/ 领域服务     │  │  ├─ migration/  SQL 迁移          │
│  event/   事件接口     │  │  cache/      Redis 缓存           │
│  ├─ define/           │  │  lock/       Redis 分布式锁       │
│  ├─ dispatcher/       │  │  logger/     结构化日志            │
│  └─ listener/         │  │  event/      异步事件总线           │
│  queue/   队列配置     │  │  mq/         MQ 接口桩            │
│                       │  │  cron/       定时任务接口          │
│                       │  │  webhook/    HMAC 签名验证         │
└──────────┬───────────┘  └──────────────────────────────────┘
           │ 依赖
           ▼
┌──────────────────────┐
│     pkg/  公共层       │
│  errorx/    错误码     │
│  validator/ 参数校验   │
│  constant/  常量       │
│  toolx/     工具函数   │
└──────────────────────┘
```

## 依赖方向

```
app ──────────────▶ infrastructure ──▶ domain ──▶ pkg
 │                        │                │
 └────────────────────────┼────────────────┘
                          │
                          ▼
                       domain（app 也直接依赖 domain）
```

**禁止反向**：infrastructure 不能 import app；domain 不能 import infrastructure。

## HTTP 请求处理流程

```
1. HTTP Request
      │
2. Middleware Chain
   ├─ Recovery（panic 恢复 → 500）
   ├─ RequestID（注入 X-Request-ID）
   ├─ Timeout（30s 超时控制）
   ├─ CORS（跨域校验 + OPTIONS 预检）
   └─ RateLimit（按 IP 滑动窗口限流）
      │
3. Handler（第1层）
   ├─ 解析请求参数
   ├─ validator.Struct() 参数校验
   └─ 调用 Logic
      │
4. Logic（第2层）
   ├─ 业务编排（组合多个 Service）
   ├─ 事务控制
   └─ 返回 types.Response
      │
5. Service（第3层）
   ├─ 纯业务逻辑（无框架依赖）
   ├─ 业务规则校验
   ├─ 发布领域事件
   └─ 调用 Repo 接口
      │
6. Repo 实现（第4层）
   ├─ GORM CRUD
   └─ 错误转换为 errorx.CodeError
      │
7. PostgreSQL
```

## 事件驱动流程

```
┌──────────────────────┐
│ Domain Service        │
│ s.eventBus.Dispatch() │  ← 发布领域事件
└──────────┬───────────┘
           │ 调用 interface
           ▼
┌──────────────────────────────────────────┐
│ infrastructure/event/EventBus             │
│ ┌──────────────────────────────────────┐ │
│ │ eventCh (chan define.Event, 1024)    │ │
│ └──────────┬───────────────────────────┘ │
│            │ goroutine                    │
│            ▼                              │
│ ┌──────────────────────────────────────┐ │
│ │ listeners[eventType] → Handle(event) │ │
│ │ 每个 listener 带 panic recover       │ │
│ └──────────────────────────────────────┘ │
└──────────────────────────────────────────┘

同步版（InMemoryDispatcher）：domain/tracking/event/dispatcher/
异步版（EventBus）：infrastructure/event/bus.go
两者均实现 define.Dispatcher 接口，可互换。
```

## 异步任务流程

```
app 主进程                      Worker 进程（cmd/worker/）
─────────────────────────────────────────────────────────
config.NewAsynqClient()        config.NewAsynqServer()
  │                                │
asynqClient.Enqueue(task)      mux := asynq.NewServeMux()
  │                            RegisterAllHandlers(mux)
  ▼                            server.Run(mux) ──▶ 消费 Redis 队列
Redis ───────────────────────────────────────────────────
  (Asynq 基于 Redis)
```

## 模块间通信

```
Domain A (audit)         Domain B (task)
  │                          ▲
  │ 发布 AuditPassed          │ 监听
  ▼                          │
┌──────────────────────────────┐
│        EventBus               │
│  Register("audit.passed",    │
│    taskListener)             │
└──────────────────────────────┘
```

两个 domain 之间**零 import 依赖**，仅通过事件类型字符串关联。

## 技术选型理由

| 选型 | 理由 |
|------|------|
| go-zero | 微服务框架，配置驱动 + 内置限流/熔断/链路追踪 |
| GORM | Go 最成熟的 ORM，与 PostgreSQL 配合好 |
| Asynq | 基于 Redis，轻量免运维，适合中小规模异步任务 |
| Redis 分布式锁 | 自研（SET NX + Lua），避免引入 Redisson 等重量级依赖 |
| go-playground/validator | Go 社区标准校验库 |
| go.work workspace | Go 1.18+ 原生多模块管理，不需要 replace 指令 |
