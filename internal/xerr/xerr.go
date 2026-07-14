package xerr

import "fmt"

type CodeMsg struct {
	Code int
	Msg  string
}

func (c *CodeMsg) Error() string {
	return fmt.Sprintf("[%d]%s", c.Code, c.Msg)
}

// 错误码  10000 0000 ~~99999 9999
// 模块id  40000
// 错误码 = 模块id + 业务错误码
var (
	ModuleId        = 40000
	ModuleErrorBase = ModuleId * 10000
)

var (
	// 系统错误 0000-0999
	ErrServerInternal = newError(ModuleErrorBase+0, "server internal error")

	// 业务错误码 1000-1999
	ErrParam                                   = newError(ModuleErrorBase+1000, "param error")
	ErrUserNotExist                            = newError(ModuleErrorBase+1001, "user not exist")
	ErrPasswordWrong                           = newError(ModuleErrorBase+1002, "password wrong")
	ErrUserAlreadyRegistered                   = newError(ModuleErrorBase+1003, "user already registered")
	ErrRelationStateNotRegisteringOrRegistered = newError(ModuleErrorBase+1004, "relation state is not registering or registered")
)

func newError(code int, msg string) *CodeMsg {
	return &CodeMsg{
		Code: code,
		Msg:  msg,
	}
}

func NewParamError(msg string) *CodeMsg {
	return newError(ErrParam.Code, msg)
}

func NewServerInternalError(msg string) *CodeMsg {
	return newError(ErrServerInternal.Code, msg)
}
