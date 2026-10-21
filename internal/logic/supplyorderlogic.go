package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/consts"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/util"
	"github.com/starslipay/trade_itg/internal/xerr"
	"github.com/starslipay/trade_itg/trade_itg_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"google.golang.org/grpc/codes"

	"github.com/zeromicro/go-zero/core/logx"
)

type SupplyOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSupplyOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SupplyOrderLogic {
	return &SupplyOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SupplyOrderLogic) checkInputParams(in *trade_itg_pb.SupplyOrderReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	if in.MerchantId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "MerchantId is empty")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Amount is invalid")
	}
	return nil
}

// queryC2BBill 查询 C2B 转账单, 单号不存在时返回 error
func (l *SupplyOrderLogic) queryC2BBill(in *trade_itg_pb.SupplyOrderReq) (*account_mgr_pb.QueryC2BBillRsp, error) {
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

// checkBillMatch 校验入参和 C2B 转账单数据是否一致
func (l *SupplyOrderLogic) checkBillMatch(in *trade_itg_pb.SupplyOrderReq, bill *account_mgr_pb.QueryC2BBillRsp) error {
	if bill.UserId != in.UserId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId not match c2b bill")
	}
	if bill.MerchantId != in.MerchantId {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "MerchantId not match c2b bill")
	}
	if bill.Amount != in.Amount {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Amount not match c2b bill")
	}
	if bill.OutTradeNo != in.OutOrderNo {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "OutOrderNo not match c2b bill OutTradeNo")
	}
	return nil
}

func (l *SupplyOrderLogic) getMerchantInfo(in *trade_itg_pb.SupplyOrderReq) (*user_mgr_pb.GetMerchantInfoRsp, error) {
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

func (l *SupplyOrderLogic) getUserInfo(in *trade_itg_pb.SupplyOrderReq) (*user_mgr_pb.GetUserInfoRsp, error) {
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

func (l *SupplyOrderLogic) getOrderInfo(in *trade_itg_pb.SupplyOrderReq) (*order_mgr_pb.QueryOrderRsp, error) {
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

// banPaySuccessOrder 调用 OrderMgr 将订单更新为成功
func (l *SupplyOrderLogic) banPaySuccessOrder(in *trade_itg_pb.SupplyOrderReq, bill *account_mgr_pb.QueryC2BBillRsp,
	merchantRsp *user_mgr_pb.GetMerchantInfoRsp, userRsp *user_mgr_pb.GetUserInfoRsp) (*order_mgr_pb.BanPaySuccessOrderRsp, error) {
	orderSuccessRsp, err := l.svcCtx.OrderMgr.BanPaySuccessOrder(l.ctx, &order_mgr_pb.BanPaySuccessOrderReq{
		TransactionId: in.TransactionId,
		OutOrderNo:    in.OutOrderNo,
		MerchantId:    in.MerchantId,
		MerchantUid:   merchantRsp.MerchantUid,
		UserId:        in.UserId,
		Uid:           userRsp.Uid,
		Amount:        in.Amount,
		PayTime:       bill.PayTime,
		DeductToken:   bill.DeductToken,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.BanPaySuccessOrder")
	}
	if orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "banPaySuccessOrder TransactionId not match")
	}
	return orderSuccessRsp, nil
}

// SupplyOrder 补单: 先查询 C2B 转账单, 校验入参和转账单一致后, 将订单更新为成功
func (l *SupplyOrderLogic) SupplyOrder(in *trade_itg_pb.SupplyOrderReq) (*trade_itg_pb.SupplyOrderRsp, error) {
	// 校验入参
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询 C2B 转账单
	bill, err := l.queryC2BBill(in)
	if err != nil {
		return nil, err
	}

	// 校验入参和转账单数据一致性
	if err := l.checkBillMatch(in, bill); err != nil {
		return nil, err
	}

	// 转账单状态必须为成功才能补单
	if bill.State != consts.C2BBillStateSuccess {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "c2b bill state is not success")
	}

	// 查询商户和用户(补单需要 MerchantUid / Uid)
	merchantRsp, err := l.getMerchantInfo(in)
	if err != nil {
		return nil, err
	}
	userRsp, err := l.getUserInfo(in)
	if err != nil {
		return nil, err
	}

	// 查询订单, 根据订单状态决定补单或返回
	orderRsp, err := l.getOrderInfo(in)
	if err != nil {
		return nil, err
	}

	switch orderRsp.OrderInfo.TradeState {
	case consts.OrderTradeStateInit:
		// 订单初始状态: 执行补单, 更新为成功
		orderSuccessRsp, err := l.banPaySuccessOrder(in, bill, merchantRsp, userRsp)
		if err != nil {
			return nil, err
		}
		return &trade_itg_pb.SupplyOrderRsp{
			TransactionId:     orderSuccessRsp.OrderInfo.TransactionId,
			OutOrderNo:        orderSuccessRsp.OrderInfo.OutOrderNo,
			UserId:            orderSuccessRsp.OrderInfo.UserId,
			MerchantId:        orderSuccessRsp.OrderInfo.MerchantId,
			Amount:            orderSuccessRsp.OrderInfo.Amount,
			PayTime:           orderSuccessRsp.OrderInfo.PayTime,
			OrderSuccessToken: orderSuccessRsp.OrderInfo.OrderSuccessToken,
		}, nil
	case consts.OrderTradeStateSuccess:
		// 订单已是成功: 校验 token 后直接返回
		if err := util.CheckOrderSuccessToken(in.TransactionId, in.OutOrderNo,
			merchantRsp.MerchantUid, userRsp.Uid, in.Amount, orderRsp.OrderInfo.OrderSuccessToken); err != nil {
			return nil, err
		}
		return &trade_itg_pb.SupplyOrderRsp{
			TransactionId:     orderRsp.OrderInfo.TransactionId,
			OutOrderNo:        orderRsp.OrderInfo.OutOrderNo,
			UserId:            orderRsp.OrderInfo.UserId,
			MerchantId:        orderRsp.OrderInfo.MerchantId,
			Amount:            orderRsp.OrderInfo.Amount,
			PayTime:           orderRsp.OrderInfo.PayTime,
			OrderSuccessToken: orderRsp.OrderInfo.OrderSuccessToken,
		}, nil
	case consts.OrderTradeStateClose:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order is closed, cannot supply")
	default:
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "order trade state is invalid")
	}
}
