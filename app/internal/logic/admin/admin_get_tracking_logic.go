package admin

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/pkg/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetTrackingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetTrackingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetTrackingLogic {
	return &AdminGetTrackingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetTrackingLogic) AdminGetTracking(req interface{}) (*types.Response, error) {
	return nil, errorx.NewError(errorx.InvalidParameter)
}
