# AI 编程规范

> 目标：让 AI 生成的代码符合项目 DDD 架构约束，减少 code review 返工。

## 代码生成规则

| 维度 | 规则 | 违规示例 |
|------|------|---------|
| 目录结构 | 严格 DDD 分层：Handler → Logic → Service → Repo | ❌ Handler 直接调 Repo |
| 业务逻辑 | 领域服务纯业务，不包含 DB/HTTP/Cache 操作 | ❌ service 中写 `db.Where(...)` |
| 错误处理 | 统一使用 `pkg/errorx` 错误码，禁止 `errors.New("xxx")` | ❌ `return nil, errors.New("task not found")` |
| 配置管理 | 所有配置走 `app.yaml` + `config` 结构体，禁止硬编码 | ❌ `timeout := 30 * time.Second` |
| 参数校验 | Handler 层用 `pkg/validator.Struct()`，禁止手写 if 校验 | ❌ `if req.Name == "" { ... }` |
| 响应格式 | 统一使用 `types.Response`，禁止直接写 `map[string]interface{}` | ❌ `httpx.OkJson(w, map[string]...{})` |
| 日志 | 使用 `logx.WithContext(ctx)`，带 request-id 链路追踪 | ❌ `log.Println("error")` |
| 依赖注入 | 通过 `svc.ServiceContext` 获取依赖，禁止全局变量 | ❌ `var db *gorm.DB` |
| 跨领域通信 | 通过 `event.EventBus`，禁止直接 import 其他 domain | ❌ `import "ns-tracking-go/domain/order"` |
| 仓储接口 | domain/repo 中定义，infrastructure/database/repo_impl 中实现 | ❌ domain 中 import GORM |

## 提示词模板

### 创建新领域模块

```
在 ns-tracking-go 的 domain/{name}/ 下创建新领域模块：

要求：
1. go.mod：模块路径 ns-tracking-go/domain/{name}，Go 1.25，仅依赖 pkg
2. entity/：定义 GORM 实体模型（json tag + gorm tag）
3. repo/：定义仓储接口（仅接口，使用 entity 类型，不 import GORM）
4. service/：实现领域服务（纯业务逻辑，通过接口调用 repo）
5. event/：定义该领域的事件类型（实现 define.Event 接口）
6. 添加单元测试 {name}_test.go（表格驱动测试）

参考 domain/tracking/ 的结构。
```

### 创建新 API 端点

```
在 ns-tracking-go 中添加新的 HTTP API：

1. app/internal/types/types.go — 添加 XxxRequest / XxxResponse 结构体
2. app/internal/handler/{module}/xxx_handler.go — 创建 Handler
   - func XxxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc
   - 用 validator.Struct(req) 校验参数
   - 调用 Logic 层
   - 返回 types.Success(data) 或 types.Error(err)
3. app/internal/logic/{module}/xxx_logic.go — 创建 Logic
   - func (l *XxxLogic) Xxx(req *types.XxxRequest) (*types.Response, error)
   - 业务编排，调用 domain service
   - 错误返回 errorx.NewError(code)
4. app/internal/handler/routes.go — 注册路由
5. 编译验证：cd app && go build ./...
```

### 创建仓储实现

```
在 infrastructure/database/repo_impl/ 实现 domain 定义的仓储接口：

要求：
1. 结构体嵌入 *gorm.DB
2. 构造函数 func NewXxxRepo(db *gorm.DB) repo.XxxRepo
3. 每个方法使用 db.WithContext(ctx)
4. 不包含业务判断（status=="pending" 等放在 domain service）
5. 错误转换为 errorx.CodeError（DatabaseError / DatabaseQueryFailed）
```

### 添加事件监听器

```
1. domain/{name}/event/ 定义事件结构体（实现 define.Event 接口）
2. app 层或 domain/listener 创建监听器（实现 define.Listener 接口）
3. app 初始化时通过 svcCtx.EventBus.Register(eventType, listener) 注册
4. 监听器 Handle 方法中调用 domain service 编排跨领域逻辑
```

### 添加异步任务

```
1. domain/tracking/queue/tasks/ 定义任务类型常量
2. domain/tracking/queue/handler/ 创建任务处理器（实现 asynq.Handler 或 ProcessTask）
3. 在 cmd/worker/routes.go 的 RegisterAllHandlers 中注册
4. 发布任务：asynqClient.Enqueue(task)
```

### 添加分布式锁

```
import "ns-tracking-go/infrastructure/lock"

token, err := svcCtx.Locker.Lock(ctx, "key:xxx", 30*time.Second)
if err != nil {
    return err  // 获取锁失败
}
defer svcCtx.Locker.Unlock(ctx, "key:xxx", token)
// 临界区操作...
```

## AI 辅助开发工作流

```
1. 理解阶段
   - 读取 docs/boundary-rules.md（边界约束）
   - 读取 docs/architecture.md（系统架构）
   - 分析 domain/{module}/ 现有代码

2. 设计阶段
   - 设计领域实体（entity）和仓储接口（repo）
   - 设计领域服务（service）——纯业务逻辑
   - 确定是否需要事件 / 队列

3. 生成阶段
   - 按分层顺序生成代码（entity → repo → service → handler → logic）
   - 每层生成后独立编译验证

4. 验证阶段
   - go build ./...（编译检查）
   - go vet ./...（静态分析）
   - cd app && go run main.go（启动验证）
```
