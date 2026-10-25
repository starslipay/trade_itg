package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/consts"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundOrderLogic {
	return &RefundOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefundOrderLogic) checkInputParams(in *trade_itg_pb.RefundOrderReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	return nil
}

func (l *RefundOrderLogic) getOrderInfo(in *trade_itg_pb.RefundOrderReq) (*order_mgr_pb.QueryOrderRsp, error) {
	queryOrderRsp, err := l.svcCtx.OrderMgr.QueryOrder(l.ctx, &order_mgr_pb.QueryOrderReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.QueryOrder")
	}
	if queryOrderRsp.OrderInfo.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "OrderTransactionId not match")
	}
	return queryOrderRsp, nil
}

// b2cRefund 调用 B2CFinal 退款, 重入时 IsRepeat=1 继续流程
func (l *RefundOrderLogic) b2cRefund(in *trade_itg_pb.RefundOrderReq, orderRsp *order_mgr_pb.QueryOrderRsp) error {
	b2cRsp, err := l.svcCtx.AccountMgr.B2CFinal(l.ctx, &account_mgr_pb.B2CReq{
		TransactionId: in.TransactionId,
		OutTradeNo:    orderRsp.OrderInfo.OutOrderNo,
		Uid:           orderRsp.OrderInfo.Uid,
		UserId:        orderRsp.OrderInfo.UserId,
		MerchantUid:   orderRsp.OrderInfo.MerchantUid,
		MerchantId:    orderRsp.OrderInfo.MerchantId,
		Amount:        orderRsp.OrderInfo.Amount,
		CurType:       1,
		BizType:       5, // 退款
		Desc:          "refund",
		Memo:          "order refund",
	})
	if err != nil {
		return xerror.HandleRPCError(err, "AccountMgr.B2CFinal")
	}
	if b2cRsp.TransactionId != in.TransactionId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "b2cRefund TransactionId not match")
	}
	return nil
}

// updateOrderToRefund 调用 OrderMgr 将订单状态更新为已退款
func (l *RefundOrderLogic) updateOrderToRefund(in *trade_itg_pb.RefundOrderReq) error {
	refundRsp, err := l.svcCtx.OrderMgr.RefundOrder(l.ctx, &order_mgr_pb.RefundOrderReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "OrderMgr.RefundOrder")
	}
	if refundRsp.TransactionId != in.TransactionId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "updateOrderToRefund TransactionId not match")
	}
	return nil
}

// RefundOrder 退款: 仅将已成功订单退款(B2C退款 + 订单状态改为已退款)
func (l *RefundOrderLogic) RefundOrder(in *trade_itg_pb.RefundOrderReq) (*trade_itg_pb.RefundOrderRsp, error) {
	// 校验入参
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询订单
	orderRsp, err := l.getOrderInfo(in)
	if err != nil {
		return nil, err
	}

	// 校验订单状态: 只有成功状态才能退款, 已退款可重入
	switch orderRsp.OrderInfo.TradeState {
	case consts.OrderTradeStateSuccess:
		// 成功状态: 执行退款流程
	case consts.OrderTradeStateRefund:
		// 已退款: 幂等重入, 直接返回
		return &trade_itg_pb.RefundOrderRsp{
			TransactionId: in.TransactionId,
		}, nil
	default:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order trade state is not success, cannot refund")
	}

	// B2C 退款(商户扣款, 用户异步入账)
	if err := l.b2cRefund(in, orderRsp); err != nil {
		return nil, err
	}

	// 更新订单状态为已退款
	if err := l.updateOrderToRefund(in); err != nil {
		return nil, err
	}

	return &trade_itg_pb.RefundOrderRsp{
		TransactionId: in.TransactionId,
	}, nil
}
