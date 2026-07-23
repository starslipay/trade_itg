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
	ErrCodeParam                                   = ModuleErrorBase + 1000
	ErrCodeUserNotExist                            = ModuleErrorBase + 1001
	ErrCodePasswordWrong                           = ModuleErrorBase + 1002
	ErrCodeUserAlreadyRegistered                   = ModuleErrorBase + 1003
	ErrCodeRelationStateNotRegisteringOrRegistered = ModuleErrorBase + 1004
)
