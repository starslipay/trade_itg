package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  40000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId        = int64(455902)
	ModuleErrorBase = ModuleId * 100
)

var (
	// 系统错误 0000-0999
	ErrCodeSystem = ModuleErrorBase + 0

	// 业务错误码 1000-1999
	ErrCodeSellerNotExist         = ModuleErrorBase + 100 // 卖方不存在
	ErrCodeParams                 = ModuleErrorBase + 101 // 参数错误
	ErrCodeReqAndRspUnMatch       = ModuleErrorBase + 102 // 请求参数和响应参数不匹配
	ErrCodeCheckOrderSuccessToken = ModuleErrorBase + 103 // 订单成功token校验失败
)

var (
	// user_mgr 错误码用户不存在
	UserMgrErrCodeUserNotExist = int64(455904101)
)
