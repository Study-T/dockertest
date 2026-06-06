package public

import (
	"context"

	"ns-tracking-go/app/internal/svc"
	"ns-tracking-go/app/internal/types"
	"ns-tracking-go/pkg/errorx"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublicQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicQueryLogic {
	return &PublicQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PublicQueryLogic) PublicQuery(req interface{}) (*types.Response, error) {
	return nil, errorx.NewError(errorx.InvalidParameter)
}
