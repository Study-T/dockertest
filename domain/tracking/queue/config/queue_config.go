package config

import "github.com/hibiken/asynq"

// AsynqConfig Asynq 队列配置
type AsynqConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
}

func NewAsynqClient(cfg AsynqConfig) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

func NewAsynqServer(cfg AsynqConfig) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				"default":  6,
				"critical": 10,
				"low":      1,
			},
		},
	)
}

func DefaultAsynqConfig() AsynqConfig {
	return AsynqConfig{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1,
		Concurrency:   10,
	}
}
