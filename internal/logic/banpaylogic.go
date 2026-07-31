package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
)

type BanPayLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBanPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BanPayLogic {
	return &BanPayLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BanPayLogic) BanPay(in *trade_itg_pb.BanPayReq) (*trade_itg_pb.BanPayRsp, error) {
	// 查询商户信息
	merchantRsp, err := l.svcCtx.UserMgr.GetMerchantInfo(l.ctx, &user_mgr_pb.GetMerchantInfoReq{
		MerchantId: in.MerchantId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetMerchantInfo")
	}

	// 查询用户信息 并校验密码
	userRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}

	// 订单下单
	orderPreRsp, err := l.svcCtx.OrderMgr.BanPayPreOrder(l.ctx, &order_mgr_pb.BanPayPreOrderReq{
		TransactionId: in.TransactionId,
		OutOrderNo:    in.OutOrderNo,
		MerchantId:    in.MerchantId,
		MerchantUid:   merchantRsp.MerchantUid,
		MerchantName:  merchantRsp.MerchantName,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		Amount:        in.Amount,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.BanPayPreOrder")
	}
	// 校验orderPreRsp返回的transactionId是否与in.TransactionId一致
	if orderPreRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "orderPreRsp.TransactionId != in.TransactionId")
	}

	c2BRsp, err := l.svcCtx.AccountMgr.C2BFinal(l.ctx, &account_mgr_pb.C2BReq{
		TransactionId: in.TransactionId,
		Uid:           userRsp.Uid,
		UserId:        in.UserId,
		MerchantUid:   merchantRsp.MerchantUid,
		MerchantId:    in.MerchantId,
		Amount:        in.Amount,
		CurType:       1,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.C2BFinal")
	}

	// 订单更新成功
	orderSuccessRsp, err := l.svcCtx.OrderMgr.BanPaySuccessOrder(l.ctx, &order_mgr_pb.BanPaySuccessOrderReq{
		TransactionId: in.TransactionId,
		OutOrderNo:    in.OutOrderNo,
		MerchantId:    in.MerchantId,
		MerchantUid:   merchantRsp.MerchantUid,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		Amount:        in.Amount,
		PayTime:       c2BRsp.PayTime,
		DeductToken:   "", // TODO 待传
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.BanPaySuccessOrder")
	}

	// 校验orderSuccessRsp返回的transactionId是否与in.TransactionId一致
	if orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId")
	}

	// TODO 校验订单成功签名

	return &trade_itg_pb.BanPayRsp{
		UserId:            orderSuccessRsp.OrderInfo.UserId,
		TransactionId:     orderSuccessRsp.OrderInfo.TransactionId,
		OrderSuccessToken: orderSuccessRsp.OrderInfo.OrderSuccessToken,
	}, nil
}
