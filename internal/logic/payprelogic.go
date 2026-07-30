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

type PayPreLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayPreLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayPreLogic {
	return &PayPreLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PayPreLogic) PayPre(in *trade_itg_pb.PayPreReq) (*trade_itg_pb.PayPreRsp, error) {
	// 查询relation
	relationRsp, err := l.svcCtx.UserMgr.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetRelation")
	}

	tradeIdRsp, err := l.svcCtx.TradeIdMgr.GenTradeId(l.ctx, &trade_id_mgr_pb.GenTradeIdReq{
		SpId:    in.MerchantId,
		Uid:     relationRsp.Uid,
		SceneId: 1,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "TradeIdMgr.GenTradeId")
	}

	return &trade_itg_pb.PayPreRsp{
		UserId:        in.UserId,
		TransactionId: tradeIdRsp.TradeId,
	}, nil
}
