package svc

import (
	"strings"

	"ns-tracking-go/app/internal/config"
	"ns-tracking-go/domain/tracking/event/define"
	"ns-tracking-go/domain/tracking/repo"
	"ns-tracking-go/domain/tracking/service"
	"ns-tracking-go/infrastructure/cache"
	"ns-tracking-go/infrastructure/database"
	"ns-tracking-go/infrastructure/database/repo_impl"
	"ns-tracking-go/infrastructure/event"
	"ns-tracking-go/infrastructure/lock"
	"ns-tracking-go/infrastructure/logger"
	"ns-tracking-go/infrastructure/metrics"
	"ns-tracking-go/infrastructure/webhook"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config             config.Config
	Redis              *redis.Redis
	DB                 *gorm.DB
	Cache              *cache.ConfigCache
	TrackingCache      repo.TrackingCache
	EventBus           *event.EventBus
	Locker             *lock.RedisLock
	WebhookVerifier    *webhook.Verifier
	WebhookParser      *webhook.Parser
	RawEventRepo       repo.RawEventRepo
	TrackingDetailRepo repo.TrackingDetailRepo
	TrackingLogRepo    repo.TrackingLogRepo
	TrackingDetailSvc  *service.TrackingDetailService
	SyncService        *service.SyncService
	RetryService       *service.RetryService
	GrayscaleService   *service.GrayscaleService
}

type domainLogger struct{}

func (domainLogger) Infof(format string, v ...interface{})  { logx.Infof(format, v...) }
func (domainLogger) Errorf(format string, v ...interface{}) { logx.Errorf(format, v...) }

func NewServiceContext(c config.Config) *ServiceContext {
	logger.Init(c.Name)
	define.SetLogger(domainLogger{})

	if c.Webhook.EncryptKey == "" {
		logx.Error("WARNING: Webhook.EncryptKey is empty")
	}

	rds := redis.MustNewRedis(redis.RedisConf{
		Host: c.RedisConf.Host, Pass: c.RedisConf.Pass, Type: "node",
	})

	conn, err := database.NewDB(c.DataSource, c.Mode)
	if err != nil {
		panic(err)
	}

	metrics.Init()
	metrics.InitGrayscale()

	rawEventRepo := repo_impl.NewRawEventRepoImpl(conn)
	detailRepo := repo_impl.NewTrackingDetailRepoImpl(conn)
	logRepo := repo_impl.NewTrackingLogRepoImpl(conn)
	trackingCache := cache.NewTrackingCacheImpl(rds)
	normalizer := service.NewNormalizer()
	bus := event.NewEventBus()
	detailSvc := service.NewTrackingDetailService(detailRepo, logRepo, trackingCache, normalizer, bus)

	whitelist := strings.Split(c.Grayscale.Whitelist, ",")
	gs := service.NewGrayscaleService(service.GrayscaleConfig{
		Enabled:    c.Grayscale.Enabled,
		Mode:       c.Grayscale.Mode,
		Whitelist:  whitelist,
		Percentage: c.Grayscale.Percentage,
	})
	metrics.GrayscaleWhitelistSize.Set(float64(gs.WhitelistSize()))

	return &ServiceContext{
		Config:             c,
		Redis:              rds,
		DB:                 conn.DB,
		Cache:              cache.NewConfigCache(rds),
		TrackingCache:      trackingCache,
		EventBus:           bus,
		Locker:             lock.NewRedisLock(rds),
		WebhookVerifier:    webhook.NewVerifier(c.Webhook.EncryptKey, c.Webhook.ReplayWindow),
		WebhookParser:      webhook.NewParser(),
		RawEventRepo:       rawEventRepo,
		TrackingDetailRepo: detailRepo,
		TrackingLogRepo:    logRepo,
		TrackingDetailSvc:  detailSvc,
		SyncService:        service.NewSyncService(detailRepo, logRepo),
		RetryService:       service.NewRetryService(rawEventRepo, detailSvc),
		GrayscaleService:   gs,
	}
}

func (ctx *ServiceContext) Close() {
	if ctx.EventBus != nil {
		ctx.EventBus.Close()
	}
	if ctx.DB != nil {
		sqlDB, _ := ctx.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}
