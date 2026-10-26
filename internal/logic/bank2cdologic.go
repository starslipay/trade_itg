package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/bank_channel/bank_channel_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/consts"
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
	return nil
}

func (l *Bank2CDoLogic) checkPassword(in *trade_itg_pb.Bank2CDoReq) (*user_mgr_pb.CheckPasswordRsp, error) {
	userRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}
	if userRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "UserId not match")
	}

	return userRsp, nil
}

func (l *Bank2CDoLogic) Bank2C(in *trade_itg_pb.Bank2CDoReq, userRsp *user_mgr_pb.CheckPasswordRsp) (*account_mgr_pb.Bank2CRsp, error) {
	bank2CRsp, err := l.svcCtx.AccountMgr.Bank2C(l.ctx, &account_mgr_pb.Bank2CReq{
		TransactionId: in.TransactionId,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		BankType:      in.BankType,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          "bank2c",
		Memo:          in.Memo,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.Bank2C")
	}
	return bank2CRsp, nil
}

// GetRechargeWithdrawNo 向 bank_channel 获取充值提现渠道单号
func (l *Bank2CDoLogic) GetRechargeWithdrawNo(uid int64) (string, error) {
	rsp, err := l.svcCtx.BankChannel.GetRechargeWithdrawNo(l.ctx, &bank_channel_pb.GetRechargeWithdrawNoReq{
		Uid: uid,
	})
	if err != nil {
		return "", xerror.HandleRPCError(err, "BankChannel.GetRechargeWithdrawNo")
	}
	return rsp.TransactionId, nil
}

// CreateRechargeWithdrawOrder 在 order_mgr 创建充值提现单(初始状态)
func (l *Bank2CDoLogic) CreateRechargeWithdrawOrder(in *trade_itg_pb.Bank2CDoReq, userRsp *user_mgr_pb.CheckPasswordRsp, channelTransactionId string) error {
	_, err := l.svcCtx.OrderMgr.CreateRechargeWithdrawOrder(l.ctx, &order_mgr_pb.CreateRechargeWithdrawOrderReq{
		TransactionId:        in.TransactionId,
		ChannelTransactionId: channelTransactionId,
		UserId:               in.UserId,
		Uid:                  userRsp.Uid,
		OrderType:            consts.RechargeWithdrawOrderTypeRecharge,
		BankType:             in.BankType,
		Amount:               in.Amount,
		CurType:              1,
		Memo:                 in.Memo,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "OrderMgr.CreateRechargeWithdrawOrder")
	}
	return nil
}

// Recharge 调用 bank_channel 去银行扣款
func (l *Bank2CDoLogic) Recharge(in *trade_itg_pb.Bank2CDoReq, channelTransactionId string) error {
	_, err := l.svcCtx.BankChannel.Recharge(l.ctx, &bank_channel_pb.RechargeReq{
		TransactionId: channelTransactionId,
		Amount:        in.Amount,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "BankChannel.Recharge")
	}
	return nil
}

// UpdateRechargeWithdrawOrderState 更新充值提现单状态
func (l *Bank2CDoLogic) UpdateRechargeWithdrawOrderState(transactionId string, tradeState int32) error {
	_, err := l.svcCtx.OrderMgr.UpdateRechargeWithdrawOrderState(l.ctx, &order_mgr_pb.UpdateRechargeWithdrawOrderStateReq{
		TransactionId: transactionId,
		TradeState:    tradeState,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "OrderMgr.UpdateRechargeWithdrawOrderState")
	}
	return nil
}

func (l *Bank2CDoLogic) Bank2CDo(in *trade_itg_pb.Bank2CDoReq) (*trade_itg_pb.Bank2CDoRsp, error) {
	// 校验参数
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 校验密码 & 查询用户信息
	userRsp, err := l.checkPassword(in)
	if err != nil {
		return nil, err
	}

	// 向 bank_channel 获取充值提现渠道单号
	channelTransactionId, err := l.GetRechargeWithdrawNo(userRsp.Uid)
	if err != nil {
		return nil, err
	}

	// 创建充值提现单(初始状态)
	if err := l.CreateRechargeWithdrawOrder(in, userRsp, channelTransactionId); err != nil {
		return nil, err
	}

	// 去银行扣款
	if err := l.Recharge(in, channelTransactionId); err != nil {
		return nil, err
	}

	// 给用户C加钱
	bank2CRsp, err := l.Bank2C(in, userRsp)
	if err != nil {
		return nil, err
	}

	// 更新充值提现单为成功
	if err := l.UpdateRechargeWithdrawOrderState(in.TransactionId, consts.RechargeWithdrawTradeStateSuccess); err != nil {
		return nil, err
	}

	return &trade_itg_pb.Bank2CDoRsp{
		TransactionId: bank2CRsp.TransactionId,
		UserId:        bank2CRsp.UserId,
		IsRepeat:      0,
	}, nil
}
