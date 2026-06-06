# 边界约束规则

> **重要**：以下规则为强制约束，AI 编程和人工开发均须遵守。违反任一条 → code review 阻断。

## 架构分层约束

```
app/              → 依赖 domain/ + infrastructure/ + pkg/
api/              → 独立模块，依赖 domain/ + pkg/
domain/           → 仅依赖 pkg/，零框架依赖
infrastructure/   → 依赖 domain/ + pkg/，实现领域接口
pkg/              → 零内部依赖（最底层）
```

## 强制规则

| # | 规则 | 说明 | 检查方式 |
|---|------|------|---------|
| 1 | **领域层纯业务** | domain/ 禁止 import GORM、go-zero rest、HTTP、Redis、MQ 等框架 | grep import domain/ |
| 2 | **基础设施纯技术** | infrastructure/ 只实现接口，不包含 `if status=="pending"` 等业务判断 | code review |
| 3 | **跨领域禁止直接 import** | domain/audit 不能 import domain/order → 必须通过 EventBus | golangci-lint depguard |
| 4 | **依赖单向** | app → infra → domain → pkg；禁止 infra → app、domain → infra | go mod graph |
| 5 | **接口下沉** | 仓储接口（Repo）、事件接口（Event/Dispatcher/Listener）定义在 domain | dir check |
| 6 | **pkg 零业务** | pkg/ 只放纯工具函数，不含任何业务字段名、业务常量 | code review |
| 7 | **按需引用** | 修改 domain/task 不应 require domain/order、domain/supplier | go.mod check |
| 8 | **配置不硬编码** | 超时时间、连接数、密钥等从 app.yaml + config 结构体读取 | grep hardcode |

## 违规示例与正确写法

### 1. 领域层依赖框架

```go
// ❌ 违规：domain 直接依赖 GORM
import "gorm.io/gorm"

type TaskRepo interface {
    FindByID(ctx context.Context, db *gorm.DB, id string) (*Task, error)
}

// ✅ 正确：domain 定义纯接口
type TaskRepo interface {
    FindByID(ctx context.Context, id string) (*Task, error)
}
```

### 2. 跨领域直接调用

```go
// ❌ 违规：domain/task 直接 import domain/order
import "ns-tracking-go/domain/order"

func (s *TaskService) CreateTask(...) {
    order := order.NewOrderService(...).GetOrder(...)  // 禁止
}

// ✅ 正确：通过事件总线解耦
func (s *TaskService) CreateTask(ctx context.Context, ...) error {
    // 1. 创建任务
    // 2. 发布事件
    s.eventBus.Dispatch(event.TaskCreated{TaskID: task.ID})
}

// 在 app 层注册监听器：
func (l *OrderListener) HandleTaskCreated(evt define.Event) {
    // 跨领域业务编排
}
```

### 3. 基础设施包含业务规则

```go
// ❌ 违规：repo 实现中包含业务判断
func (r *taskRepoImpl) Save(ctx context.Context, task *entity.Task) error {
    if task.Status == "pending" && task.Quantity > 1000 {  // 业务判断应放在 domain service
        task.Priority = "high"
    }
    return r.db.WithContext(ctx).Save(task).Error
}

// ✅ 正确：repo 只做持久化，业务判断在 domain service
// infrastructure/database/repo_impl/task_repo.go
func (r *taskRepoImpl) Save(ctx context.Context, task *entity.Task) error {
    return r.db.WithContext(ctx).Save(task).Error
}

// domain/task/service/task_service.go
func (s *TaskService) CreateTask(ctx context.Context, req CreateTaskReq) (*Task, error) {
    task := &Task{Status: "pending", Quantity: req.Quantity}
    if task.Quantity > 1000 {  // 业务规则在此
        task.Priority = "high"
    }
    return task, s.repo.Save(ctx, task)
}
```

### 4. 依赖方向错误

```
❌ infrastructure → app（禁止）
❌ domain → infrastructure（禁止）
❌ pkg → domain（禁止）

✅ app → infrastructure → domain → pkg
✅ pkg 零依赖
```

## go.mod 依赖检查清单

新增领域模块时，检查 `go.mod`：

```go
// domain/task/go.mod
// ✅ 正确：最小依赖
require (
    ns-tracking-go/pkg v0.0.0
    github.com/google/uuid v1.6.0
)

// ❌ 违规：依赖了不该依赖的
require (
    ns-tracking-go/domain/order v0.0.0      // 跨领域
    ns-tracking-go/infrastructure v0.0.0    // 下层反向依赖
    gorm.io/gorm v1.31.1                    // 框架代码
)
```
