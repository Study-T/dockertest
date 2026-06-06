// Package cron 定时任务抽象接口。
// TODO: 实现 go-zero cron 调度器适配此接口。
// TODO: 实现 robfig/cron 调度器适配此接口。
package cron

import "context"

type Task interface {
	Name() string
	CronExpr() string
	Execute(ctx context.Context) error
}

type Scheduler interface {
	Register(task Task) error
	Start() error
	Stop() error
}
