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

func (l *CloseOrSupplyOrderLogic) CloseOrSupplyOrder(in *trade_itg_pb.CloseOrSupplyOrderReq) (*trade_itg_pb.CloseOrSupplyOrderRsp, error) {
	// 查询商户信息
	merchantRsp, err := l.svcCtx.UserMgr.GetMerchantInfo(l.ctx, &user_mgr_pb.GetMerchantInfoReq{
		MerchantId: in.MerchantId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetMerchantInfo")
	}

	// 查询用户信息
	userRsp, err := l.svcCtx.UserMgr.GetUserInfo(l.ctx, &user_mgr_pb.GetUserInfoReq{
		UserId: in.UserId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.GetUserInfo")
	}

	queryOrderRsp, err := l.svcCtx.OrderMgr.QueryOrder(l.ctx, &order_mgr_pb.QueryOrderReq{
		TransactionId: in.TransactionId,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.QueryOrder")
	}

	switch queryOrderRsp.OrderInfo.TradeState {
	case consts.OrderTradeStateInit:
		isExistC2BBill := true
		queryC2BBillRsp, err := l.svcCtx.AccountMgr.QueryC2BBill(l.ctx, &account_mgr_pb.QueryC2BBillReq{
			TransactionId: in.TransactionId,
		})
		if err != nil {
			BizError, isSuccessParse := xerror.ParseBizError(err)
			if isSuccessParse {
				// c2b bill不存在
				if BizError.Code == 300001009 {
					isExistC2BBill = false
				}
			}
			return nil, xerror.HandleRPCError(err, "AccountMgr.QueryC2BBill")
		}

		if isExistC2BBill {
			// 如果c2b bill状态为成功，更新订单状态为成功
			if consts.C2BBillStateSuccess == queryC2BBillRsp.State {
				orderSuccessRsp, err := l.svcCtx.OrderMgr.BanPaySuccessOrder(l.ctx, &order_mgr_pb.BanPaySuccessOrderReq{
					TransactionId: in.TransactionId,
					OutOrderNo:    in.OutOrderNo,
					MerchantId:    in.MerchantId,
					MerchantUid:   merchantRsp.MerchantUid,
					UserId:        in.UserId,
					Uid:           userRsp.Uid,
					Amount:        in.Amount,
					PayTime:       queryC2BBillRsp.PayTime,
					DeductToken:   "", // TODO 待传
				})
				if err != nil {
					return nil, xerror.HandleRPCError(err, "OrderMgr.BanPaySuccessOrder")
				}
				return &trade_itg_pb.CloseOrSupplyOrderRsp{
					ResultCode:        ResultCodeOrderSuccess,
					TransactionId:     in.TransactionId,
					OutOrderNo:        orderSuccessRsp.OrderInfo.OutOrderNo,
					UserId:            orderSuccessRsp.OrderInfo.UserId,
					MerchantId:        orderSuccessRsp.OrderInfo.MerchantId,
					PayTime:           orderSuccessRsp.OrderInfo.PayTime,
					OrderSuccessToken: orderSuccessRsp.OrderInfo.OrderSuccessToken,
				}, nil
			} else if queryC2BBillRsp.State == consts.C2BBillStateClose {
				// 如果c2b bill状态为关单，更新订单状态为关单
				_, err := l.svcCtx.OrderMgr.CloseOrder(l.ctx, &order_mgr_pb.CloseOrderReq{
					TransactionId: in.TransactionId,
				})
				if err != nil {
					return nil, xerror.HandleRPCError(err, "OrderMgr.CloseOrder")
				}

				// TODO 校验关键数据一致性
				return &trade_itg_pb.CloseOrSupplyOrderRsp{
					ResultCode:    ResultCodeOrderClosed,
					TransactionId: in.TransactionId,
				}, nil
			} else {
				return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "c2b bill state is not init")
			}
		} else {
			// TODO 先关c2b bill

			// 如果c2b bill不存在，更新订单状态为关单，同时更新c2b bill状态为关单
			_, err := l.svcCtx.OrderMgr.CloseOrder(l.ctx, &order_mgr_pb.CloseOrderReq{
				TransactionId: in.TransactionId,
			})
			if err != nil {
				return nil, xerror.HandleRPCError(err, "OrderMgr.BanPayCloseOrder")
			}

			// TODO 校验关键数据一致性

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
			PayTime:           queryOrderRsp.OrderInfo.PayTime,
			OrderSuccessToken: queryOrderRsp.OrderInfo.OrderSuccessToken,
		}, nil
	case consts.OrderTradeStateClose:
		return &trade_itg_pb.CloseOrSupplyOrderRsp{
			ResultCode:    ResultCodeOrderClosed,
			TransactionId: in.TransactionId,
		}, nil
	default:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order trade state is not init")
	}
}
