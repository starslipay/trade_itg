package logic

import (
	"context"

	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/trade_itg_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2cTransferDoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cTransferDoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cTransferDoLogic {
	return &C2cTransferDoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cTransferDoLogic) C2CTransferDo(in *trade_itg_pb.C2CTransferDoReq) (*trade_itg_pb.C2CTransferDoRsp, error) {
	return &trade_itg_pb.C2CTransferDoRsp{}, nil
}
