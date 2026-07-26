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

func (l *C2cTransferDoLogic) C2CTransferDo(in *trade_itg_pb.C2CTransferDoReq) (*trade_itg_pb.C2CTransferDoRsp, error) {
	checkPasswordRsp, err := l.svcCtx.UserMgrRpcClient.CheckPassword(l.ctx, &user_mgr_pb.CheckPasswordReq{
		UserId:   in.BuyerUserId,
		Password: in.Password,
	})
	if err != nil {
		err = xerr.ParseRPCError(err)
		return nil, err
	}

	sellerRelationRsp, err := l.svcCtx.UserMgrRpcClient.GetRelation(l.ctx, &user_mgr_pb.GetRelationReq{
		UserId: in.SellerUserId,
	})
	if err != nil {
		bizError, isSuccessParse := xerror.ParseBizError(err)
		if isSuccessParse {
			if bizError.Code == xerr.UserMgrErrCodeUserNotExist {
				return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeSellerNotExist, "seller not exist")
			} else {
				return nil, xerror.NewBizError(codes.Internal, bizError.Code, bizError.Message)
			}
		}
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeCallRpc, err.Error())
	}

	var c2CLocalRsp *account_mgr_pb.C2CRsp
	if 0 == in.Version {
		c2CLocalRsp, err = l.svcCtx.AccountMgrRpcClient.C2CLocal(l.ctx, &account_mgr_pb.C2CReq{
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
			err = xerr.ParseRPCError(err)
			return nil, err
		}
	} else if 1 == in.Version {
		c2CLocalRsp, err = l.svcCtx.AccountMgrRpcClient.C2CFinal(l.ctx, &account_mgr_pb.C2CReq{
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
			err = xerr.ParseRPCError(err)
			return nil, err
		}
	} else {
		return nil, xerror.NewBizError(codes.Internal, xerr.ErrCodeParams, "version invalid")
	}

	return &trade_itg_pb.C2CTransferDoRsp{
		TransactionId: c2CLocalRsp.TransactionId,
		BuyerUserId:   c2CLocalRsp.BuyerUserId,
		SellerUserId:  c2CLocalRsp.SellerUserId,
		IsRepeat:      0,
	}, nil
}
