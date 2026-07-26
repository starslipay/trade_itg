package xerr

// 错误码  10000 0000 ~~99999 9999
// 模块id  40000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId        = int64(40000)
	ModuleErrorBase = ModuleId * 10000
)

var (
	// 系统错误 0000-0999
	ErrCodeSystem = ModuleErrorBase + 0

	// 业务错误码 1000-1999
	ErrCodeSellerNotExist = ModuleErrorBase + 1000 // 卖方不存在
	ErrCodeParams         = ModuleErrorBase + 1001 // 参数错误
)

var (
	// user_mgr 错误码
	UserMgrErrCodeUserNotExist = int64(200001001)
)
