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
	}
	for _, c := range cases {
		got := Msg(c.code)
		if got != c.expect {
			t.Errorf("Msg(%d) = %q, want %q", c.code, got, c.expect)
		}
	}
}
