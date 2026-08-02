package logic

import (
	"context"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/svc"
	"github.com/starslipay/trade_itg/internal/util"
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

func (l *BanPayLogic) checkInputParams(in *trade_itg_pb.BanPayReq) error {
	if in.UserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "UserId is empty")
	}
	if in.Password == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Password is empty")
	}
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	if in.OutOrderNo == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "OutOrderNo is empty")
	}
	if in.MerchantId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "MerchantId is empty")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Amount is <= 0")
	}

	return nil
}

func (l *BanPayLogic) getMerchantInfo(in *trade_itg_pb.BanPayReq) (*user_mgr_pb.GetMerchantInfoRsp, error) {
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

func (l *BanPayLogic) checkPassword(in *trade_itg_pb.BanPayReq) (*user_mgr_pb.CheckPasswordRsp, error) {
	userRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}
	if userRsp.UserId != in.UserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "CheckPassword UserId not match")
	}
	return userRsp, nil
}

func (l *BanPayLogic) banPayPreOrder(in *trade_itg_pb.BanPayReq, userRsp *user_mgr_pb.CheckPasswordRsp, merchantRsp *user_mgr_pb.GetMerchantInfoRsp) (*order_mgr_pb.BanPayPreOrderRsp, error) {
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
	return orderPreRsp, nil
}

func (l *BanPayLogic) C2BFinal(in *trade_itg_pb.BanPayReq, userRsp *user_mgr_pb.CheckPasswordRsp, merchantRsp *user_mgr_pb.GetMerchantInfoRsp) (*account_mgr_pb.C2BRsp, error) {
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
	if c2BRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "c2BRsp.TransactionId != in.TransactionId")
	}
	return c2BRsp, nil
}

func (l *BanPayLogic) banPaySuccessOrder(in *trade_itg_pb.BanPayReq, c2BRsp *account_mgr_pb.C2BRsp, merchantRsp *user_mgr_pb.GetMerchantInfoRsp, userRsp *user_mgr_pb.CheckPasswordRsp) (*order_mgr_pb.BanPaySuccessOrderRsp, error) {
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
		DeductToken:   c2BRsp.DeductToken,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "OrderMgr.BanPaySuccessOrder")
	}

	// 校验orderSuccessRsp返回的transactionId是否与in.TransactionId一致
	if orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "orderSuccessRsp.OrderInfo.TransactionId != in.TransactionId")
	}

	return orderSuccessRsp, nil
}

func (l *BanPayLogic) BanPay(in *trade_itg_pb.BanPayReq) (*trade_itg_pb.BanPayRsp, error) {
	// 校验输入参数
	if err := l.checkInputParams(in); err != nil {
		return nil, err
	}

	// 查询商户信息
	merchantRsp, err := l.getMerchantInfo(in)
	if err != nil {
		return nil, err
	}

	// 查询用户信息 并校验密码
	userRsp, err := l.checkPassword(in)
	if err != nil {
		return nil, err
	}

	// 订单下单
	_, err = l.banPayPreOrder(in, userRsp, merchantRsp)
	if err != nil {
		return nil, err
	}

	// 去账户服务扣款
	c2BRsp, err := l.C2BFinal(in, userRsp, merchantRsp)
	if err != nil {
		return nil, err
	}

	// 订单更新成功
	orderSuccessRsp, err := l.banPaySuccessOrder(in, c2BRsp, merchantRsp, userRsp)
	if err != nil {
		return nil, err
	}

	// 校验订单成功签名
	if err := util.CheckOrderSuccessToken(in.TransactionId, in.OutOrderNo, merchantRsp.MerchantUid, userRsp.Uid,
		in.Amount, orderSuccessRsp.OrderInfo.OrderSuccessToken); err != nil {
		return nil, err
	}

	return &trade_itg_pb.BanPayRsp{
		TransactionId:     orderSuccessRsp.OrderInfo.TransactionId,
		OutOrderNo:        orderSuccessRsp.OrderInfo.OutOrderNo,
		UserId:            orderSuccessRsp.OrderInfo.UserId,
		MerchantId:        orderSuccessRsp.OrderInfo.MerchantId,
		Amount:            orderSuccessRsp.OrderInfo.Amount,
		PayTime:           orderSuccessRsp.OrderInfo.PayTime,
		OrderSuccessToken: orderSuccessRsp.OrderInfo.OrderSuccessToken,
	}, nil
}
