package admin

import (
	"net/http"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminBatchQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "not implemented", nil))
	}
}

func AdminGetTrackingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTrackingReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(400, "invalid request", nil))
			return
		}

		if req.OrderNumber == "" {
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(400, "order number is required", nil))
			return
		}

		result, err := svcCtx.TrackingDetailSvc.QueryByOrderNumber(r.Context(), req.OrderNumber)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("query tracking failed: order_number=%s err=%v", req.OrderNumber, err)
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(500, "query failed", nil))
			return
		}

		if result == nil {
			httpx.OkJsonCtx(r.Context(), w, types.NewResponse(404, "tracking not found", nil))
			return
		}

		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "success", []interface{}{result}))
	}
}

func AdminHealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "not implemented", nil))
	}
}
