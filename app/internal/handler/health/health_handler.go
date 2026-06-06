package health

import (
	"net/http"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbOK := checkDB(svcCtx)
		redisOK := checkRedis(svcCtx, r)

		status := "ok"
		code := 200
		if !dbOK || !redisOK {
			status = "degraded"
			code = 503
		}

		httpx.OkJsonCtx(r.Context(), w, types.Response{
			Code:    code,
			Message: status,
			Data: map[string]interface{}{
				"status": status,
				"db":     dbOK,
				"redis":  redisOK,
			},
		})
	}
}

func checkDB(svcCtx *svc.ServiceContext) bool {
	if svcCtx.DB == nil {
		return false
	}
	sqlDB, err := svcCtx.DB.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

func checkRedis(svcCtx *svc.ServiceContext, r *http.Request) bool {
	if svcCtx.Redis == nil {
		return false
	}
	return svcCtx.Redis.PingCtx(r.Context())
}
