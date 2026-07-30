package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	UserMgrRpcConfig    zrpc.RpcClientConf
	AccountMgrRpcConfig zrpc.RpcClientConf
	TradeIdMgrRpcConfig zrpc.RpcClientConf
	OrderMgrRpcConfig   zrpc.RpcClientConf
}
