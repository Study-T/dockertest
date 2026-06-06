package admin

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/pkg/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminBatchQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminBatchQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminBatchQueryLogic {
	return &AdminBatchQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminBatchQueryLogic) AdminBatchQuery(req interface{}) (*types.Response, error) {
	return nil, errorx.NewError(errorx.InvalidParameter)
}
