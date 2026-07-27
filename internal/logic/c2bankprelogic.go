package logic

import (
	"context"

	"github.com/starslipay/trade_id_mgr/trade_id_mgr_pb"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2BankPreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2BankPreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2BankPreLogic {
	return &C2BankPreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2BankPreLogic) C2BankPre(in *trade_itg_pb.C2BankPreReq) (*trade_itg_pb.C2BankPreRsp, error) {
	// 查询relation
	relationRsp, err := l.svcCtx.UserMgrRpcClient.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}

	tradeIdRsp, err := l.svcCtx.TradeIdMgrRpcClient.GenTradeId(l.ctx, &trade_id_mgr_pb.GenTradeIdReq{
		SpId:    "1000000000",
		Uid:     relationRsp.Uid,
		SceneId: 1,
	})
	if err != nil {
		return nil, xerr.ParseRPCError(err)
	}

	return &trade_itg_pb.C2BankPreRsp{
		UserId:        relationRsp.UserId,
		TransactionId: tradeIdRsp.TradeId,
	}, nil
}
