package main

import (
	"flag"

	"ns-tracking-go/domain/tracking/queue/config"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()

	var cfg config.AsynqConfig
	conf.MustLoad(*configFile, &cfg)

	server := config.NewAsynqServer(cfg)
	mux := asynq.NewServeMux()

	RegisterAllHandlers(mux)

	logx.Infof("Worker starting, concurrency=%d", cfg.Concurrency)
	if err := server.Run(mux); err != nil {
		logx.Errorf("Worker failed: %v", err)
	}
}
