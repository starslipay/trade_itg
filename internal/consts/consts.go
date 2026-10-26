package consts

const (
	OrderTradeStateInit    = 1  // 初始状态
	OrderTradeStateSuccess = 2  // 成功
	OrderTradeStateRefund  = 3  // 已退款
	OrderTradeStateClose   = 99 // 关闭
)

const (
	C2BBillStateSuccess = 1  // 成功
	C2BBillStateClose   = 99 // 关闭
)

const (
	RechargeWithdrawOrderTypeRecharge = 1 // 充值(Bank2C)
	RechargeWithdrawOrderTypeWithdraw = 2 // 提现(C2Bank)
)

const (
	RechargeWithdrawTradeStateSuccess = 2 // 成功
	RechargeWithdrawTradeStateFail    = 3 // 失败
)
