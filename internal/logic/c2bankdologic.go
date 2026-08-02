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

func (l *C2BankDoLogic) checkInputParams(in *trade_itg_pb.C2BankDoReq) error {
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	if in.Password == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Password is empty")
	}
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
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

func (l *C2BankDoLogic) C2Bank(in *trade_itg_pb.C2BankDoReq, userRsp *user_mgr_pb.CheckPasswordRsp) (*account_mgr_pb.C2BankRsp, error) {
	c2BankRsp, err := l.svcCtx.AccountMgr.C2Bank(l.ctx, &account_mgr_pb.C2BankReq{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		BankType:      in.BankType,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          in.Desc,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.C2Bank")
	}
	return c2BankRsp, nil
}

func (l *C2BankDoLogic) checkPassword(in *trade_itg_pb.C2BankDoReq) (*user_mgr_pb.CheckPasswordRsp, error) {
	userRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}
	if userRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId not match")
	}

	return userRsp, nil
}

func (l *C2BankDoLogic) C2BankDo(in *trade_itg_pb.C2BankDoReq) (*trade_itg_pb.C2BankDoRsp, error) {
	// 校验参数
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 校验密码
	userRsp, err := l.checkPassword(in)
	if err != nil {
		return nil, err
	}

	// 调用C2Bank
	c2BankRsp, err := l.C2Bank(in, userRsp)
	if err != nil {
		return nil, err
	}

	// TODO 去银行给用户加钱

	return &trade_itg_pb.C2BankDoRsp{
		TransactionId: c2BankRsp.TransactionId,
		UserId:        c2BankRsp.UserId,
		IsRepeat:      0,
	}, nil
}
