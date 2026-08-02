package logic

import (
	"context"

	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_id_mgr/trade_id_mgr_pb"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"google.golang.org/grpc/codes"

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

func (l *PayPreLogic) checkInputParams(in *trade_itg_pb.PayPreReq) error {
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	return nil
}

func (l *PayPreLogic) getRelation(in *trade_itg_pb.PayPreReq) (*user_mgr_pb.GetRelationRsp, error) {
	relationRsp, err := l.svcCtx.UserMgr.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetRelation")
	}
	if relationRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "GetRelation UserId not match")
	}
	return relationRsp, nil
}

func (l *PayPreLogic) GenTradeId(in *trade_itg_pb.PayPreReq, relationRsp *user_mgr_pb.GetRelationRsp) (*trade_id_mgr_pb.GenTradeIdRsp, error) {
	tradeIdRsp, err := l.svcCtx.TradeIdMgr.GenTradeId(l.ctx, &trade_id_mgr_pb.GenTradeIdReq{
		SpId:    in.MerchantId,
		Uid:     relationRsp.Uid,
		SceneId: 1,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "TradeIdMgr.GenTradeId")
	}
	return tradeIdRsp, nil
}

func (l *PayPreLogic) PayPre(in *trade_itg_pb.PayPreReq) (*trade_itg_pb.PayPreRsp, error) {
	// 校验参数
	err := l.checkInputParams(in)
	if err != nil {
		return nil, err
	}

	// 查询relation
	relationRsp, err := l.getRelation(in)
	if err != nil {
		return nil, err
	}

	// 生成交易ID
	tradeIdRsp, err := l.GenTradeId(in, relationRsp)
	if err != nil {
		return nil, err
	}

	return &trade_itg_pb.PayPreRsp{
		UserId:        in.UserId,
		TransactionId: tradeIdRsp.TradeId,
	}, nil
}
