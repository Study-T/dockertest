package handler

import (
	"net/http"
	"time"

	"ns-tracking-go/app/internal/handler/admin"
	"ns-tracking-go/app/internal/handler/health"
	"ns-tracking-go/app/internal/handler/public"
	"ns-tracking-go/app/internal/handler/tracking"
	"ns-tracking-go/app/internal/handler/webhook"
	"ns-tracking-go/app/internal/middleware"
	"ns-tracking-go/app/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, svcCtx *svc.ServiceContext) {
	mws := chainMiddleware(svcCtx)

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/health",
		Handler: health.HealthHandler(svcCtx),
	})

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/admin/batch-query",
		Handler: mws(admin.AdminBatchQueryHandler(svcCtx)),
	})
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/admin/tracking/:order_number",
		Handler: mws(admin.AdminGetTrackingHandler(svcCtx)),
	})
	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/admin/health",
		Handler: mws(admin.AdminHealthHandler(svcCtx)),
	})

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/public/query",
		Handler: mws(public.PublicQueryHandler(svcCtx)),
	})
	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/public/batch-query",
		Handler: mws(public.PublicBatchQueryHandler(svcCtx)),
	})

	server.AddRoute(rest.Route{
		Method:  "GET",
		Path:    "/tracking/:order_number",
		Handler: mws(tracking.GetTrackingHandler(svcCtx)),
	})

	server.AddRoute(rest.Route{
		Method:  "POST",
		Path:    "/webhook",
		Handler: mws(webhook.WebhookHandler(svcCtx)),
	})
}

func chainMiddleware(svcCtx *svc.ServiceContext) func(http.HandlerFunc) http.HandlerFunc {
	recovery := middleware.RecoveryMiddleware()
	reqID := middleware.RequestIDMiddleware()
	timeout := middleware.TimeoutMiddleware(30 * time.Second)
	rateLimit := middleware.RateLimitMiddleware(100)
	cors := middleware.CORSMiddleware(
		svcCtx.Config.CORS.AllowedOrigins,
		svcCtx.Config.CORS.AllowedMethods,
	)

	return func(handler http.HandlerFunc) http.HandlerFunc {
		return recovery(reqID(timeout(rateLimit(cors(handler)))))
	}
}
