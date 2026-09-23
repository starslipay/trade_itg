package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  40000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId = int64(455902)
)

var (
	// 系统错误 0000-0999
	ErrCodeSystem = 455902000

	// 业务错误码 1000-1999
	ErrCodeSellerNotExist         = int64(455902100) // 卖方不存在
	ErrCodeParams                 = int64(455902101) // 参数错误
	ErrCodeReqAndRspUnMatch       = int64(455902102) // 请求参数和响应参数不匹配
	ErrCodeCheckOrderSuccessToken = int64(455902103) // 订单成功token校验失败
)

var (
	// user_mgr 错误码用户不存在
	UserMgrErrCodeUserNotExist = int64(455904101)
)
