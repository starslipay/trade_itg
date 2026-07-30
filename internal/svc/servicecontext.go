package svc

import (
	"github.com/starslipay/trade_itg/internal/config"

	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/order_mgr/order_mgr_pb"
	"github.com/starslipay/trade_id_mgr/trade_id_mgr_pb"
	"github.com/starslipay/user_mgr/user_mgr_pb"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	UserMgr    user_mgr_pb.UserMgrClient
	AccountMgr account_mgr_pb.AccountMgrClient
	TradeIdMgr trade_id_mgr_pb.TradeIdMgrClient
	OrderMgr   order_mgr_pb.OrderMgrClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		UserMgr:    user_mgr_pb.NewUserMgrClient(zrpc.MustNewClient(c.UserMgrRpcConfig).Conn()),
		AccountMgr: account_mgr_pb.NewAccountMgrClient(zrpc.MustNewClient(c.AccountMgrRpcConfig).Conn()),
		TradeIdMgr: trade_id_mgr_pb.NewTradeIdMgrClient(zrpc.MustNewClient(c.TradeIdMgrRpcConfig).Conn()),
		OrderMgr:   order_mgr_pb.NewOrderMgrClient(zrpc.MustNewClient(c.OrderMgrRpcConfig).Conn()),
	}
}
