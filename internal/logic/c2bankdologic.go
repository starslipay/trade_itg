package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type C2BankDoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2BankDoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2BankDoLogic {
	return &C2BankDoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2BankDoLogic) C2BankDo(in *trade_itg_pb.C2BankDoReq) (*trade_itg_pb.C2BankDoRsp, error) {
	checkPasswordRsp, err := l.svcCtx.UserMgrRpcClient.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}

	bank2CRsp, err := l.svcCtx.AccountMgrRpcClient.C2Bank(l.ctx, &account_mgr_pb.C2BankReq{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
		Uid:           checkPasswordRsp.Uid,
		BankType:      in.BankType,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          in.Desc,
	})
	if err != nil {
		return nil, err
	}
	return &trade_itg_pb.C2BankDoRsp{
		TransactionId: bank2CRsp.TransactionId,
		UserId:        bank2CRsp.UserId,
		IsRepeat:      0,
	}, nil
}
