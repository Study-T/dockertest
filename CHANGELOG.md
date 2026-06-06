# Changelog

## [Unreleased]

### Added
- DDD 四层架构骨架（app / domain / infrastructure / pkg）+ go.work 工作区
- go-zero HTTP 服务框架集成（v1.10.2）
- Asynq 异步任务队列集成（v0.24.1，支持 critical/default/low 优先级）
- 事件系统：同步 InMemoryDispatcher + 异步 EventBus，接口统一在 define/
- Redis 分布式锁（SET NX + Lua 脚本原子释放）
- go-zero logx 结构化日志封装（dev/prod 模式 + context request-id 传递）
- 统一错误码框架 (pkg/errorx)：Code + CodeError + GetMessage + NewCodeError
- 参数校验封装 (pkg/validator)：go-playground/validator → CodeError
- 中间件链：RequestID / Timeout / Recovery / RateLimit（滑动窗口）/ CORS
- HMAC-SHA256 签名验证（hmac.Equal 防时序攻击）
- PostgreSQL + GORM 连接池（10 idle / 100 max / 1h lifetime）
- 健康检查端点（/health，检测 DB + Redis 连通性，返回 200/503）
- ServiceContext DI 容器（Config + DB + Redis + EventBus + RedisLock）
- 统一响应格式（types.Response + Success/Error 工厂方法）
- Worker 进程（Asynq Server + ServeMux 任务注册）
- Docker 多阶段构建（golang:1.25 → alpine:3.21 + HEALTHCHECK）
- docker-compose 本地基础设施（Postgres 14 + Redis 6 + healthcheck）
- .golangci.yml 代码规范配置（gofmt/errcheck/staticcheck/revive…）
- infrastructure 组件骨架：lock / logger / event bus / mq 接口 / cron 接口 / migration
- 文档体系：README / boundary-rules / ai-skills / architecture / CHANGELOG

### Changed
- 事件系统接口统一到 domain/tracking/event/define/（删除重复的 listener/interface.go）
- 消除 HMAC 签名代码重复（删除 app/internal/utils/signature.go）
- Go 版本统一到 1.25（四个子模块对齐）
- Logic 层返回值统一为 (*types.Response, error)，集成 errorx 错误码
- Handler 层响应统一使用 types.Response（废弃 map[string]interface{}）
- 中间件 RateLimit 从空壳补全为滑动窗口限流
- 路由注册从空函数补全为完整路由表 + 中间件链组装
- README 重写：补充完整项目结构、事件流程、请求流程、添加业务指引
