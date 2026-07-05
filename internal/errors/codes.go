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
	UserModuleStart    = 10100
	ParcelModuleStart  = 10200
	PickupModuleStart  = 10300
	ProxyModuleStart   = 10400
	ShelfModuleStart   = 10500
	NotifyModuleStart  = 10600
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
	default:
		return "未知错误"
	}
}
