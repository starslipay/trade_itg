package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type Bank2CDoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBank2CDoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Bank2CDoLogic {
	return &Bank2CDoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Bank2CDoLogic) Bank2CDo(in *trade_itg_pb.Bank2CDoReq) (*trade_itg_pb.Bank2CDoRsp, error) {
	// 查询relation
	relationRsp, err := l.svcCtx.UserMgrRpcClient.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerr.NewServerInternalError(err.Error())
	}

	bank2CRsp, err := l.svcCtx.AccountMgrRpcClient.Bank2C(l.ctx, &account_mgr_pb.Bank2CReq{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
		Uid:           relationRsp.Uid,
		BankType:      in.BankType,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          in.Desc,
	})
	if err != nil {
		return nil, xerr.NewServerInternalError(err.Error())
	}
	return &trade_itg_pb.Bank2CDoRsp{
		TransactionId: bank2CRsp.TransactionId,
		UserId:        bank2CRsp.UserId,
		IsRepeat:      false,
	}, nil
}
