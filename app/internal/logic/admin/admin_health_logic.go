package admin

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/pkg/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminHealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminHealthLogic {
	return &AdminHealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminHealthLogic) AdminHealth(req interface{}) (*types.Response, error) {
	return nil, errorx.NewError(errorx.InvalidParameter)
}
