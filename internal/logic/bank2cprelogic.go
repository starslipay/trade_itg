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

func (l *Bank2CPreLogic) checkInputParams(in *trade_itg_pb.Bank2CPreReq) error {
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	return nil
}

func (l *Bank2CPreLogic) GetRelation(in *trade_itg_pb.Bank2CPreReq) (*user_mgr_pb.GetRelationRsp, error) {
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

func (l *Bank2CPreLogic) GenTradeId(in *trade_itg_pb.Bank2CPreReq, relationRsp *user_mgr_pb.GetRelationRsp) (*trade_id_mgr_pb.GenTradeIdRsp, error) {
	tradeIdRsp, err := l.svcCtx.TradeIdMgr.GenTradeId(l.ctx, &trade_id_mgr_pb.GenTradeIdReq{
		SpId:    "1000000000",
		Uid:     relationRsp.Uid,
		SceneId: 1,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "TradeIdMgr.GenTradeId")
	}
	return tradeIdRsp, nil
}

func (l *Bank2CPreLogic) Bank2CPre(in *trade_itg_pb.Bank2CPreReq) (*trade_itg_pb.Bank2CPreRsp, error) {
	// 校验参数
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询relation
	relationRsp, err := l.GetRelation(in)
	if err != nil {
		return nil, err
	}

	// 生成交易ID
	tradeIdRsp, err := l.GenTradeId(in, relationRsp)
	if err != nil {
		return nil, err
	}

	return &trade_itg_pb.Bank2CPreRsp{
		UserId:        relationRsp.UserId,
		TransactionId: tradeIdRsp.TradeId,
	}, nil
}
