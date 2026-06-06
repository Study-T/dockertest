package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/cache"
	"ns-tracking-go/infrastructure/database"
	"ns-tracking-go/infrastructure/database/repo_impl"
	"ns-tracking-go/infrastructure/event"
	"ns-tracking-go/infrastructure/logger"
	"ns-tracking-go/infrastructure/queue"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()

	var c Config
	conf.MustLoad(*configFile, &c)
	logger.Init(c.Name + "-queue-worker")

	// Initialize database
	conn, err := database.NewDB(c.DataSource, "")
	if err != nil {
		logx.Errorf("connect database failed: %v", err)
		return
	}

	// Initialize Redis
	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.RedisConf.Host, Pass: c.RedisConf.Pass, Type: "node",
	})

	// Initialize repos
	detailRepo := repo_impl.NewTrackingDetailRepoImpl(conn)
	logRepo := repo_impl.NewTrackingLogRepoImpl(conn)
	rawEventRepo := repo_impl.NewRawEventRepoImpl(conn)
	trackingCache := cache.NewTrackingCacheImpl(rds)

	// Initialize queue
	queueConfig := queue.QueueConfig{
		Key:     c.Queue.Key,
		Timeout: c.Queue.Timeout,
	}
	queueRepo := queue.NewRedisQueue(c.RedisConf.Host, c.RedisConf.Pass, c.RedisConf.Db, queueConfig)

	// Initialize services
	normalizer := service.NewNormalizer()
	eventBus := event.NewEventBus()
	detailSvc := service.NewTrackingDetailService(detailRepo, logRepo, trackingCache, normalizer, eventBus)
	queueSvc := service.NewQueueService(queueRepo, rawEventRepo, detailSvc)

	// Start worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logx.Info("received shutdown signal")
		cancel()
	}()

	logx.Info("queue worker starting...")
	if err := queueSvc.StartListening(ctx); err != nil {
		logx.Errorf("queue worker stopped: %v", err)
	}
	logx.Info("queue worker stopped")
}
