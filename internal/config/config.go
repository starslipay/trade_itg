package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	UserMgrRpcConfig    zrpc.RpcClientConf
	AccountMgrRpcConfig zrpc.RpcClientConf
	TradeIdMgrRpcConfig zrpc.RpcClientConf
	OrderMgrRpcConfig   zrpc.RpcClientConf

	// AccessLog 请求/响应日志脱敏配置
	AccessLog AccessLogConf
	// Debug 用于测试流程中断的 mock 断点开关
	Debug DebugConf
}

// AccessLogConf 访问日志配置
type AccessLogConf struct {
	Enable          bool     `json:",default=true"`
	SensitiveFields []string `json:",optional"`
}

// DebugConf 测试用 mock 断点配置, 打开后 BanPay 会在指定位置中断返回 mock 错误,
// 用于演练 C2B 扣款成功但订单未更新、订单已更新但响应未返回等中断场景下补单/对账的行为
type DebugConf struct {
	// BreakBeforePreOrder 在下单前中断
	BreakBeforePreOrder bool `json:",optional"`
	// BreakBeforeC2B 在调用 C2B 扣款前中断
	BreakBeforeC2B bool `json:",optional"`
	// BreakBeforeOrderSuccess 在订单更新成功前中断(C2B 已成功扣款但订单未更新)
	BreakBeforeOrderSuccess bool `json:",optional"`
	// BreakAfterOrderSuccess 在订单更新成功后中断(订单已更新成功但响应未返回)
	BreakAfterOrderSuccess bool `json:",optional"`
}
