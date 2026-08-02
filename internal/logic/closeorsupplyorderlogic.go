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
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	ResultCodeOrderSuccess = 1
	ResultCodeOrderClosed  = 2
)

type CloseOrSupplyOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseOrSupplyOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseOrSupplyOrderLogic {
	return &CloseOrSupplyOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func checkInputParams(in *trade_itg_pb.CloseOrSupplyOrderReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	return nil
}

func (l *CloseOrSupplyOrderLogic) getMerchantInfo(in *trade_itg_pb.CloseOrSupplyOrderReq) (*user_mgr_pb.GetMerchantInfoRsp, error) {
	merchantRsp, err := l.svcCtx.UserMgr.GetMerchantInfo(l.ctx, &user_mgr_pb.GetMerchantInfoReq{
		MerchantId: in.MerchantId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetMerchantInfo")
	}
	if merchantRsp.MerchantId != in.MerchantId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "MerchantId not match")
	}

	return merchantRsp, nil
}

func (l *CloseOrSupplyOrderLogic) getUserInfo(in *trade_itg_pb.CloseOrSupplyOrderReq) (*user_mgr_pb.GetUserInfoRsp, error) {
	userRsp, err := l.svcCtx.UserMgr.GetUserInfo(l.ctx, &user_mgr_pb.GetUserInfoReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetUserInfo")
	}
	if userRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "UserId not match")
	}

	return userRsp, nil
}

func (l *CloseOrSupplyOrderLogic) getOrderInfo(in *trade_itg_pb.CloseOrSupplyOrderReq) (*order_mgr_pb.QueryOrderRsp, error) {
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

func (l *CloseOrSupplyOrderLogic) queryC2BBill(in *trade_itg_pb.CloseOrSupplyOrderReq) (*account_mgr_pb.QueryC2BBillRsp, error) {
	queryC2BBillRsp, err := l.svcCtx.AccountMgr.QueryC2BBill(l.ctx, &account_mgr_pb.QueryC2BBillReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		BizError, isSuccessParse := xerror.ParseBizError(err)
		if isSuccessParse && BizError.Code == 300001009 {
			return nil, nil
		} else {
			return nil, xerror.HandleRPCError(err, "AccountMgr.QueryC2BBill")
		}
	}

	if queryC2BBillRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "queryC2BBill TransactionId not match")
	}
	return queryC2BBillRsp, nil
}

func (l *CloseOrSupplyOrderLogic) closeOrder(in *trade_itg_pb.CloseOrSupplyOrderReq) error {
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

func (l *CloseOrSupplyOrderLogic) closeC2BBill(in *trade_itg_pb.CloseOrSupplyOrderReq) error {
	closeC2BBillRsp, err := l.svcCtx.AccountMgr.CloseC2BBill(l.ctx, &account_mgr_pb.CloseC2BBillReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return xerror.HandleRPCError(err, "AccountMgr.CloseC2BBill")
	}
	if closeC2BBillRsp.TransactionId != in.TransactionId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "closeC2BBill TransactionId not match")
	}
	return nil
}

func (l *CloseOrSupplyOrderLogic) banPaySuccessOrder(in *trade_itg_pb.CloseOrSupplyOrderReq, queryC2BBillRsp *account_mgr_pb.QueryC2BBillRsp,
	merchantRsp *user_mgr_pb.GetMerchantInfoRsp, userRsp *user_mgr_pb.GetUserInfoRsp) (*order_mgr_pb.BanPaySuccessOrderRsp, error) {
	orderSuccessRsp, err := l.svcCtx.OrderMgr.BanPaySuccessOrder(l.ctx, &order_mgr_pb.BanPaySuccessOrderReq{
		TransactionId: in.TransactionId,
		OutOrderNo:    in.OutOrderNo,
		MerchantId:    in.MerchantId,
		MerchantUid:   merchantRsp.MerchantUid,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		Amount:        in.Amount,
		PayTime:       queryC2BBillRsp.PayTime,
		DeductToken:   queryC2BBillRsp.DeductToken, // 传入扣款凭证
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.BanPaySuccessOrder")
	}
	if orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "banPaySuccessOrder TransactionId not match")
	}

	// TODO 校验订单成功token

	return orderSuccessRsp, nil
}

func (l *CloseOrSupplyOrderLogic) CloseOrSupplyOrder(in *trade_itg_pb.CloseOrSupplyOrderReq) (*trade_itg_pb.CloseOrSupplyOrderRsp, error) {
	if err := checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询商户信息
	merchantRsp, err := l.getMerchantInfo(in)
	if err != nil {
		return nil, err
	}

	// 查询用户信息
	userRsp, err := l.getUserInfo(in)
	if err != nil {
		return nil, err
	}

	// 查询订单信息
	queryOrderRsp, err := l.getOrderInfo(in)
	if err != nil {
		return nil, err
	}

	switch queryOrderRsp.OrderInfo.TradeState {
	case consts.OrderTradeStateInit:
		// 查询c2b bill
		queryC2BBillRsp, err := l.queryC2BBill(in)
		if err != nil {
			return nil, err
		}

		// c2b bill存在
		if queryC2BBillRsp != nil {
			if consts.C2BBillStateSuccess == queryC2BBillRsp.State {
				// 如果c2b bill状态为成功，正向补单, 更新订单状态为成功
				orderSuccessRsp, err := l.banPaySuccessOrder(in, queryC2BBillRsp, merchantRsp, userRsp)
				if err != nil {
					return nil, err
				}
				return &trade_itg_pb.CloseOrSupplyOrderRsp{
					ResultCode:        ResultCodeOrderSuccess,
					TransactionId:     in.TransactionId,
					OutOrderNo:        orderSuccessRsp.OrderInfo.OutOrderNo,
					UserId:            orderSuccessRsp.OrderInfo.UserId,
					MerchantId:        orderSuccessRsp.OrderInfo.MerchantId,
					Amount:            orderSuccessRsp.OrderInfo.Amount,
					PayTime:           orderSuccessRsp.OrderInfo.PayTime,
					OrderSuccessToken: orderSuccessRsp.OrderInfo.OrderSuccessToken,
				}, nil
			} else if queryC2BBillRsp.State == consts.C2BBillStateClose {
				// 如果c2b bill状态为关单，将订单也关单
				if err := l.closeOrder(in); err != nil {
					return nil, err
				}

				return &trade_itg_pb.CloseOrSupplyOrderRsp{
					ResultCode:    ResultCodeOrderClosed,
					TransactionId: in.TransactionId,
				}, nil
			} else {
				return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "c2b bill state is not init")
			}
		} else { // c2b bill不存在
			// 先关c2b bill, 保证无法扣除
			if err := l.closeC2BBill(in); err != nil {
				return nil, err
			}

			// 再关订单
			if err := l.closeOrder(in); err != nil {
				return nil, err
			}

			return &trade_itg_pb.CloseOrSupplyOrderRsp{
				ResultCode:    ResultCodeOrderClosed,
				TransactionId: queryOrderRsp.OrderInfo.TransactionId,
			}, nil
		}
	case consts.OrderTradeStateSuccess:
		// TODO 校验关键数据一致性

		return &trade_itg_pb.CloseOrSupplyOrderRsp{
			ResultCode:        ResultCodeOrderSuccess,
			TransactionId:     queryOrderRsp.OrderInfo.TransactionId,
			OutOrderNo:        queryOrderRsp.OrderInfo.OutOrderNo,
			UserId:            queryOrderRsp.OrderInfo.UserId,
			MerchantId:        queryOrderRsp.OrderInfo.MerchantId,
			Amount:            queryOrderRsp.OrderInfo.Amount,
			PayTime:           queryOrderRsp.OrderInfo.PayTime,
			OrderSuccessToken: queryOrderRsp.OrderInfo.OrderSuccessToken,
		}, nil
	case consts.OrderTradeStateClose:
		return &trade_itg_pb.CloseOrSupplyOrderRsp{
			ResultCode:    ResultCodeOrderClosed,
			TransactionId: queryOrderRsp.OrderInfo.TransactionId,
		}, nil
	default:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order trade state is not init")
	}
}
