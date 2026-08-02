package util

import (
	"crypto/md5"
	"fmt"
	"strconv"

	"github.com/starslipay/paycomm/xerror"
	"github.com/starslipay/trade_itg/internal/xerr"
	"google.golang.org/grpc/codes"
)

func GenMD5(input string) string {
	return fmt.Sprintf("%X", md5.Sum([]byte(input)))
}

func GenOrderSuccessToken(transaction_id, out_order_no string, merchant_uid, uid, amount int64) string {
	md5Str := GenMD5(transaction_id + "|" + out_order_no + "|" + strconv.FormatInt(merchant_uid, 10) +
		"|" + strconv.FormatInt(uid, 10) + "|" + strconv.FormatInt(amount, 10))
	return md5Str
}

func CheckOrderSuccessToken(transaction_id, out_order_no string, merchant_uid, uid, amount int64, orderSuccessToken string) error {
	if orderSuccessToken != GenOrderSuccessToken(transaction_id, out_order_no, merchant_uid, uid, amount) {
		return xerror.NewBizError(codes.Internal, xerr.ErrCodeCheckOrderSuccessToken, "orderSuccessToken check failed")
	}
	return nil
}
