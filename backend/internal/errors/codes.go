// Package errors defines the unified error code scheme and AppError type
// used across all handlers and services.
package errors

import "net/http"

// Generic error codes (range 10001~10099) — see 详细设计文档/api详细设计.md 1.8.
const (
	Success             = 0
	ErrInvalidParam     = 10001
	ErrUnauthenticated  = 10002
	ErrForbidden        = 10003
	ErrNotFound         = 10004
	ErrConflict         = 10005
	ErrRateLimited      = 10006
	ErrPayloadTooLarge  = 10007
	ErrUnsupportedMedia = 10008
	ErrInternal         = 10009
	ErrIdempotencyKey   = 10010
	ErrUnknown          = 99999
)

// Module error code range starts (Phase 2+ fills in concrete values).
const (
	UserModuleStart   = 10100
	ParcelModuleStart = 10200
	PickupModuleStart = 10300
	ProxyModuleStart  = 10400
	ShelfModuleStart  = 10500
	NotifyModuleStart = 10600
)

// User module error codes (10101~10171) — see .planning/phases/02-user-module/02-CONTEXT.md.
const (
	ErrPhoneFormat      = 10101 // 手机号格式错误
	ErrSMSTooFrequent   = 10102 // 发送频率超限
	ErrSMSChannelFail   = 10103 // 短信通道异常
	ErrPhoneBlacklisted = 10104 // 手机号在黑名单中

	ErrCodeInvalid     = 10111 // 验证码错误或已过期
	ErrCodeUsed        = 10112 // 验证码已使用
	ErrUserBlacklisted = 10113 // 用户已被拉黑

	ErrRefreshInvalid = 10121 // refresh_token 无效
	ErrRefreshExpired = 10122 // refresh_token 已过期

	ErrNicknameTooLong = 10131 // 昵称超长
	ErrAvatarInvalid   = 10132 // 头像 URL 非法

	ErrRunnerDuplicate = 10141 // 已是跑腿员或审核中
	ErrIDCardInvalid   = 10142 // 证件照 URL 无效
	ErrCreditLow       = 10143 // 信用分不足

	ErrAppNotFound   = 10151 // 申请单不存在
	ErrAppNotPending = 10152 // 申请单状态非审核中
	ErrActionInvalid = 10153 // action 取值非法

	ErrAdminCredential = 10161 // 用户名或密码错误
	ErrAdminDisabled   = 10162 // 管理员已禁用

	ErrUserNotFound = 10171 // 用户不存在
)

// HTTPStatus maps a business code to its HTTP status code.
// Returns http.StatusInternalServerError for unknown codes.
func HTTPStatus(code int) int {
	switch code {
	case Success:
		return http.StatusOK
	case ErrInvalidParam, ErrIdempotencyKey:
		return http.StatusBadRequest
	case ErrUnauthenticated:
		return http.StatusUnauthorized
	case ErrForbidden:
		return http.StatusForbidden
	case ErrNotFound:
		return http.StatusNotFound
	case ErrConflict:
		return http.StatusConflict
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrPayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case ErrUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	case ErrInternal, ErrUnknown:
		return http.StatusInternalServerError
	// User module — 10101~10171
	case ErrPhoneFormat, ErrNicknameTooLong, ErrAvatarInvalid, ErrIDCardInvalid, ErrActionInvalid:
		return http.StatusBadRequest
	case ErrSMSTooFrequent, ErrPhoneBlacklisted, ErrUserBlacklisted, ErrCreditLow, ErrAdminDisabled:
		return http.StatusForbidden
	case ErrSMSChannelFail:
		return http.StatusInternalServerError
	case ErrCodeInvalid, ErrCodeUsed, ErrRefreshInvalid, ErrRefreshExpired, ErrAdminCredential:
		return http.StatusUnauthorized
	case ErrRunnerDuplicate, ErrAppNotPending:
		return http.StatusConflict
	case ErrAppNotFound, ErrUserNotFound:
		return http.StatusNotFound
	default:
		// Module-specific codes default to 500 unless they fall in a
		// known range. Phase 2+ can refine this as needed.
		return http.StatusInternalServerError
	}
}

// Msg returns the default Chinese description for a code.
// Returns "未知错误" for unknown codes.
func Msg(code int) string {
	switch code {
	case Success:
		return "success"
	case ErrInvalidParam:
		return "请求参数错误"
	case ErrUnauthenticated:
		return "未登录或 Token 失效"
	case ErrForbidden:
		return "无权限访问"
	case ErrNotFound:
		return "资源不存在"
	case ErrConflict:
		return "资源状态冲突"
	case ErrRateLimited:
		return "请求频率超限"
	case ErrPayloadTooLarge:
		return "上传文件过大"
	case ErrUnsupportedMedia:
		return "不支持的媒体类型"
	case ErrInternal:
		return "服务端内部错误"
	case ErrIdempotencyKey:
		return "Idempotency-Key 重复请求"
	case ErrUnknown:
		return "未知错误"
	// User module — 10101~10171
	case ErrPhoneFormat:
		return "手机号格式错误"
	case ErrSMSTooFrequent:
		return "短信发送频率超限"
	case ErrSMSChannelFail:
		return "短信通道异常"
	case ErrPhoneBlacklisted:
		return "手机号已被加入黑名单"
	case ErrCodeInvalid:
		return "验证码错误或已过期"
	case ErrCodeUsed:
		return "验证码已使用"
	case ErrUserBlacklisted:
		return "用户已被拉黑"
	case ErrRefreshInvalid:
		return "refresh_token 无效"
	case ErrRefreshExpired:
		return "refresh_token 已过期"
	case ErrNicknameTooLong:
		return "昵称长度超限"
	case ErrAvatarInvalid:
		return "头像 URL 非法"
	case ErrRunnerDuplicate:
		return "已是跑腿员或审核中"
	case ErrIDCardInvalid:
		return "证件照 URL 无效"
	case ErrCreditLow:
		return "信用分不足"
	case ErrAppNotFound:
		return "跑腿员申请单不存在"
	case ErrAppNotPending:
		return "申请单状态非审核中"
	case ErrActionInvalid:
		return "审核操作类型非法"
	case ErrAdminCredential:
		return "用户名或密码错误"
	case ErrAdminDisabled:
		return "管理员账号已禁用"
	case ErrUserNotFound:
		return "用户不存在"
	default:
		return "未知错误"
	}
}
