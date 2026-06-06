package public

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/pkg/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicBatchQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicBatchQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicBatchQueryLogic {
	return &PublicBatchQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublicBatchQueryLogic) PublicBatchQuery(req interface{}) (*types.Response, error) {
	return nil, errorx.NewError(errorx.InvalidParameter)
}
