package public

import (
	"net/http"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PublicQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "not implemented", nil))
	}
}

func PublicBatchQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, types.NewResponse(200, "not implemented", nil))
	}
}
