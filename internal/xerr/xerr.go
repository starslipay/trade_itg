package xerr

import (
	"github.com/starslipay/paycomm/xerror"
	"google.golang.org/grpc/codes"
)

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
	ErrCodeCallRpc        = ModuleErrorBase + 1002 // rpc错误
)

var (
	// user_mgr 错误码
	UserMgrErrCodeUserNotExist = int64(200001001)
)

func ParseRPCError(err error) error {
	// 解析下游业务错误
	bizError, isSuccessParse := xerror.ParseBizError(err)
	if isSuccessParse {
		return xerror.NewBizError(codes.Internal, bizError.Code, bizError.Message)
	}

	// 如果没有解析到业务错误，返回rpc错误码
	return xerror.NewBizError(codes.Internal, ErrCodeCallRpc, err.Error())
}
