package logic

import (
	"context"

	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_id_mgr/trade_id_mgr_pb"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Bank2CPreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBank2CPreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Bank2CPreLogic {
	return &Bank2CPreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Bank2CPreLogic) Bank2CPre(in *trade_itg_pb.Bank2CPreReq) (*trade_itg_pb.Bank2CPreRsp, error) {
	// 查询relation
	relationRsp, err := l.svcCtx.UserMgr.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetRelation")
	}

	tradeIdRsp, err := l.svcCtx.TradeIdMgr.GenTradeId(l.ctx, &trade_id_mgr_pb.GenTradeIdReq{
		SpId:    "1000000000",
		Uid:     relationRsp.Uid,
		SceneId: 1,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "TradeIdMgr.GenTradeId")
	}

	return &trade_itg_pb.Bank2CPreRsp{
		UserId:        relationRsp.UserId,
		TransactionId: tradeIdRsp.TradeId,
	}, nil
}
