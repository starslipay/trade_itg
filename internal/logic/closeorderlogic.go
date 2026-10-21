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

type CloseOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseOrderLogic {
	return &CloseOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CloseOrderLogic) checkInputParams(in *trade_itg_pb.CloseOrderReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	return nil
}

// queryC2BBill 查询 C2B 转账单, 单号不存在时返回 error
func (l *CloseOrderLogic) queryC2BBill(in *trade_itg_pb.CloseOrderReq) (*account_mgr_pb.QueryC2BBillRsp, error) {
	queryC2BBillRsp, err := l.svcCtx.AccountMgr.QueryC2BBill(l.ctx, &account_mgr_pb.QueryC2BBillReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.QueryC2BBill")
	}
	if queryC2BBillRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "queryC2BBill TransactionId not match")
	}
	return queryC2BBillRsp, nil
}

func (l *CloseOrderLogic) getOrderInfo(in *trade_itg_pb.CloseOrderReq) (*order_mgr_pb.QueryOrderRsp, error) {
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

// closeOrder 调用 OrderMgr 关闭订单
func (l *CloseOrderLogic) closeOrder(in *trade_itg_pb.CloseOrderReq) error {
	closeOrderRsp, err := l.svcCtx.OrderMgr.CloseOrder(l.ctx, &order_mgr_pb.CloseOrderReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "OrderMgr.CloseOrder")
	}
	if closeOrderRsp.TransactionId != in.TransactionId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "closeOrder TransactionId not match")
	}
	return nil
}

// CloseOrder 关单: 先查询 C2B 转账单, 确认已关单后关闭订单
func (l *CloseOrderLogic) CloseOrder(in *trade_itg_pb.CloseOrderReq) (*trade_itg_pb.CloseOrderRsp, error) {
	// 校验入参
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询 C2B 转账单
	bill, err := l.queryC2BBill(in)
	if err != nil {
		return nil, err
	}

	// 转账单状态必须为已关单才能关闭订单
	if bill.State != consts.C2BBillStateClose {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "c2b bill state is not closed")
	}

	// 查询订单, 根据订单状态决定关单或返回
	orderRsp, err := l.getOrderInfo(in)
	if err != nil {
		return nil, err
	}

	switch orderRsp.OrderInfo.TradeState {
	case consts.OrderTradeStateInit:
		// 订单初始状态: 执行关单
		if err := l.closeOrder(in); err != nil {
			return nil, err
		}
		return &trade_itg_pb.CloseOrderRsp{
			TransactionId: in.TransactionId,
		}, nil
	case consts.OrderTradeStateClose:
		// 订单已是关单: 直接返回
		return &trade_itg_pb.CloseOrderRsp{
			TransactionId: in.TransactionId,
		}, nil
	case consts.OrderTradeStateSuccess:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order is success, cannot close")
	default:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order trade state is invalid")
	}
}
