package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"google.golang.org/grpc/codes"

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

func (l *Bank2CDoLogic) checkInputParams(in *trade_itg_pb.Bank2CDoReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	if in.Password == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Password is empty")
	}
	if in.BankType == 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "BankType is empty")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Amount is empty")
	}
	if in.Desc == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Desc is empty")
	}
	return nil
}

func (l *Bank2CDoLogic) checkPassword(in *trade_itg_pb.Bank2CDoReq) (*user_mgr_pb.CheckPasswordRsp, error) {
	checkPasswordRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}
	if checkPasswordRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "UserId not match")
	}

	return checkPasswordRsp, nil
}

func (l *Bank2CDoLogic) Bank2C(in *trade_itg_pb.Bank2CDoReq, checkPasswordRsp *user_mgr_pb.CheckPasswordRsp) (*account_mgr_pb.Bank2CRsp, error) {
	bank2CRsp, err := l.svcCtx.AccountMgr.Bank2C(l.ctx, &account_mgr_pb.Bank2CReq{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
		Uid:           checkPasswordRsp.Uid,
		BankType:      in.BankType,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          in.Desc,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.Bank2C")
	}
	return bank2CRsp, nil
}

func (l *Bank2CDoLogic) Bank2CDo(in *trade_itg_pb.Bank2CDoReq) (*trade_itg_pb.Bank2CDoRsp, error) {
	// 校验参数
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 校验密码 & 查询用户信息
	checkPasswordRsp, err := l.checkPassword(in)
	if err != nil {
		return nil, err
	}

	// TODO 去银行扣款

	// 给用户C加钱
	bank2CRsp, err := l.Bank2C(in, checkPasswordRsp)
	if err != nil {
		return nil, err
	}

	return &trade_itg_pb.Bank2CDoRsp{
		TransactionId: bank2CRsp.TransactionId,
		UserId:        bank2CRsp.UserId,
		IsRepeat:      0,
	}, nil
}
