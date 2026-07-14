package svc

import (
	"github.com/starslipay/trade_itg/internal/config"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/trade_id_mgr/trade_id_mgr_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config              config.Config
	UserMgrRpcClient    user_mgr_pb.UserMgrClient
	AccountMgrRpcClient account_mgr_pb.AccountMgrClient
	TradeIdMgrRpcClient trade_id_mgr_pb.TradeIdMgrClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:              c,
		UserMgrRpcClient:    user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn()),
		AccountMgrRpcClient: account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn()),
		TradeIdMgrRpcClient: trade_id_mgr_pb.NewTradeIdMgrClient(zrpc.MustNewClient(c.TradeIdMgrRpcConfig).Conn()),
	}
}
