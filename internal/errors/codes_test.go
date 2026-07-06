package errors

import (
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		code   int
		expect int
	}{
		{Success, http.StatusOK},
		{ErrInvalidParam, http.StatusBadRequest},
		{ErrUnauthenticated, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrNotFound, http.StatusNotFound},
		{ErrConflict, http.StatusConflict},
		{ErrRateLimited, http.StatusTooManyRequests},
		{ErrPayloadTooLarge, http.StatusRequestEntityTooLarge},
		{ErrUnsupportedMedia, http.StatusUnsupportedMediaType},
		{ErrInternal, http.StatusInternalServerError},
		{ErrIdempotencyKey, http.StatusBadRequest},
		{ErrUnknown, http.StatusInternalServerError},
		{999999, http.StatusInternalServerError}, // unknown code
		// User module
		{ErrPhoneFormat, http.StatusBadRequest},
		{ErrSMSTooFrequent, http.StatusForbidden},
		{ErrSMSChannelFail, http.StatusInternalServerError},
		{ErrPhoneBlacklisted, http.StatusForbidden},
		{ErrCodeInvalid, http.StatusUnauthorized},
		{ErrCodeUsed, http.StatusUnauthorized},
		{ErrUserBlacklisted, http.StatusForbidden},
		{ErrRefreshInvalid, http.StatusUnauthorized},
		{ErrRefreshExpired, http.StatusUnauthorized},
		{ErrNicknameTooLong, http.StatusBadRequest},
		{ErrAvatarInvalid, http.StatusBadRequest},
		{ErrRunnerDuplicate, http.StatusConflict},
		{ErrIDCardInvalid, http.StatusBadRequest},
		{ErrCreditLow, http.StatusForbidden},
		{ErrAppNotFound, http.StatusNotFound},
		{ErrAppNotPending, http.StatusConflict},
		{ErrActionInvalid, http.StatusBadRequest},
		{ErrAdminCredential, http.StatusUnauthorized},
		{ErrAdminDisabled, http.StatusForbidden},
		{ErrUserNotFound, http.StatusNotFound},
	}
	for _, c := range cases {
		got := HTTPStatus(c.code)
		if got != c.expect {
			t.Errorf("HTTPStatus(%d) = %d, want %d", c.code, got, c.expect)
		}
	}
}

func TestMsg(t *testing.T) {
	cases := []struct {
		code   int
		expect string
	}{
		{Success, "success"},
		{ErrInvalidParam, "请求参数错误"},
		{ErrUnauthenticated, "未登录或 Token 失效"},
		{ErrForbidden, "无权限访问"},
		{ErrNotFound, "资源不存在"},
		{ErrConflict, "资源状态冲突"},
		{ErrRateLimited, "请求频率超限"},
		{ErrInternal, "服务端内部错误"},
		{ErrUnknown, "未知错误"},
		{999999, "未知错误"}, // unknown code
		// User module — sampling a few representative codes
		{ErrPhoneFormat, "手机号格式错误"},
		{ErrSMSTooFrequent, "短信发送频率超限"},
		{ErrUserBlacklisted, "用户已被拉黑"},
		{ErrRefreshExpired, "refresh_token 已过期"},
		{ErrRunnerDuplicate, "已是跑腿员或审核中"},
		{ErrAppNotFound, "跑腿员申请单不存在"},
		{ErrAdminCredential, "用户名或密码错误"},
		{ErrUserNotFound, "用户不存在"},
	}
	for _, c := range cases {
		got := Msg(c.code)
		if got != c.expect {
			t.Errorf("Msg(%d) = %q, want %q", c.code, got, c.expect)
		}
	}
}
