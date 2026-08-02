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

type C2cTransferDoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewC2cTransferDoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *C2cTransferDoLogic {
	return &C2cTransferDoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *C2cTransferDoLogic) checkInputParams(in *trade_itg_pb.C2CTransferDoReq) error {
	if in.TransactionId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "TransactionId is empty")
	}
	if in.BuyerUserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "BuyerUserId is empty")
	}
	if in.Password == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Password is empty")
	}
	if in.SellerUserId == "" {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "SellerUserId is empty")
	}
	if in.Amount <= 0 {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "Amount is empty")
	}
	return nil
}

func (l *C2cTransferDoLogic) checkBuyerPassword(in *trade_itg_pb.C2CTransferDoReq) (*user_mgr_pb.CheckPasswordRsp, error) {
	checkPasswordRsp, err := l.svcCtx.UserMgr.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.BuyerUserId,
		Password: in.Password,
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "UserMgr.CheckPassword")
	}
	if checkPasswordRsp.UserId != in.BuyerUserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "buyer userId not match")
	}

	return checkPasswordRsp, nil
}

func (l *C2cTransferDoLogic) getSellerRelation(in *trade_itg_pb.C2CTransferDoReq) (*user_mgr_pb.GetRelationRsp, error) {
	sellerRelationRsp, err := l.svcCtx.UserMgr.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.SellerUserId,
	})
	if err != nil {
		bizError, isSuccessParse := xerror.ParseBizError(err)
		if isSuccessParse && bizError.Code == xerr.UserMgrErrCodeUserNotExist {
			return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeSellerNotExist, "seller not exist")
		}
		return nil, xerror.HandleRPCError(err, "UserMgr.GetRelation")
	}
	if sellerRelationRsp.UserId != in.SellerUserId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "seller userId not match")
	}

	return sellerRelationRsp, nil
}

func (l *C2cTransferDoLogic) c2CLocal(in *trade_itg_pb.C2CTransferDoReq, checkPasswordRsp *user_mgr_pb.CheckPasswordRsp, sellerRelationRsp *user_mgr_pb.GetRelationRsp) (*account_mgr_pb.C2CRsp, error) {
	c2CLocalRsp, err := l.svcCtx.AccountMgr.C2CLocal(l.ctx, &account_mgr_pb.C2CReq{
		TransactionId: in.TransactionId,
		BuyerUid:      checkPasswordRsp.Uid,
		BuyerUserId:   in.BuyerUserId,
		SellerUid:     sellerRelationRsp.Uid,
		SellerUserId:  in.SellerUserId,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          "c2c transfer(local)",
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.C2CLocal")
	}

	if c2CLocalRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "transaction id not match")
	}
	return c2CLocalRsp, nil
}

func (l *C2cTransferDoLogic) c2CFinal(in *trade_itg_pb.C2CTransferDoReq, checkPasswordRsp *user_mgr_pb.CheckPasswordRsp, sellerRelationRsp *user_mgr_pb.GetRelationRsp) (*account_mgr_pb.C2CRsp, error) {
	c2CRsp, err := l.svcCtx.AccountMgr.C2CFinal(l.ctx, &account_mgr_pb.C2CReq{
		TransactionId: in.TransactionId,
		BuyerUid:      checkPasswordRsp.Uid,
		BuyerUserId:   in.BuyerUserId,
		SellerUid:     sellerRelationRsp.Uid,
		SellerUserId:  in.SellerUserId,
		Amount:        in.Amount,
		CurType:       1,
		Desc:          "c2c transfer(final)",
	})
	if err != nil {
		return nil, xerror.HandleRPCError(err, "AccountMgr.C2CFinal")
	}
	if c2CRsp.TransactionId != in.TransactionId {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeReqAndRspUnMatch, "transaction id not match")
	}
	return c2CRsp, nil
}

func (l *C2cTransferDoLogic) C2CTransferDo(in *trade_itg_pb.C2CTransferDoReq) (*trade_itg_pb.C2CTransferDoRsp, error) {
	// 校验参数
	err := l.checkInputParams(in)
	if err != nil {
		return nil, err
	}

	// 校验密码 & 校验买家是否存在 & 查询买家Uid
	checkPasswordRsp, err := l.checkBuyerPassword(in)
	if err != nil {
		return nil, err
	}

	// 校验卖家是否存在 & 查询卖家Uid
	sellerRelationRsp, err := l.getSellerRelation(in)
	if err != nil {
		return nil, err
	}

	var c2CRsp *account_mgr_pb.C2CRsp
	if 0 == in.Version {
		// c2c同步入账(c出和c入必须在同一个实例中)
		c2CRsp, err = l.c2CLocal(in, checkPasswordRsp, sellerRelationRsp)
		if err != nil {
			return nil, err
		}
	} else if 1 == in.Version {
		// c2c异步入账(c出和c入允许在不同的实例中)
		c2CRsp, err = l.c2CFinal(in, checkPasswordRsp, sellerRelationRsp)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "version invalid")
	}

	return &trade_itg_pb.C2CTransferDoRsp{
		TransactionId: c2CRsp.TransactionId,
		BuyerUserId:   c2CRsp.BuyerUserId,
		SellerUserId:  c2CRsp.SellerUserId,
		IsRepeat:      c2CRsp.IsRepeat,
	}, nil
}
