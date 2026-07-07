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

	// Parcel module error codes (10201~10299) — see 详细设计文档/api详细设计.md 3.1~3.7.
	ErrParcelDuplicate       = 10201 // 快递单号已存在（同驿站内）
	ErrShelfNotFoundOrFull   = 10202 // 货架编号不存在或已满
	ErrReceiverPhoneInvalid  = 10203 // 收件人手机号格式错误
	ErrCourierCompanyInvalid = 10204 // 快递公司不在字典中
	ErrPickupCodeGenFail     = 10205 // 取件码生成冲突重试失败
	ErrBatchFileInvalid      = 10211 // 文件格式不支持
	ErrBatchFileTooLarge     = 10212 // 文件过大
	ErrBatchTemplateInvalid  = 10213 // 模板字段缺失或非法
	ErrParcelNotFound        = 10221 // 包裹不存在
	ErrParcelNoPermission    = 10222 // 无权查看该包裹
	ErrParcelStatusInvalid   = 10232 // 状态流转非法
	ErrParcelStatusReadonly  = 10233 // 目标状态不支持手工变更
	ErrParcelNotOwner        = 10242 // 非本人包裹
	ErrParcelNotPending      = 10243 // 包裹状态非待取

	// Pickup module error codes (10301~10399) — see 详细设计文档/api详细设计.md 4.1~4.4.
	ErrPickupCodeInvalid    = 10301 // 取件码不存在或已失效
	ErrPickupStatusNotPending = 10302 // 包裹状态非「待取」
	ErrPickupGeoAbnormal    = 10303 // 地理位置异常
	ErrPickupUserBlacklisted = 10304 // 用户在黑名单中
	ErrPickupFreqLimit      = 10305 // 取件频次超限
	ErrSelfCheckoutInvalid  = 10311 // 取件码无效
	ErrSelfCheckoutGeoFar   = 10312 // 地理位置超距
	ErrSelfCheckoutRisk     = 10313 // 风控触发

	// Proxy module error codes (10401~10499) — see 详细设计文档/api详细设计.md 5.1~5.7.
	ErrProxyParcelNotOwner    = 10401 // 包裹不存在或非本人
	ErrProxyParcelNotPending  = 10402 // 包裹状态非待取
	ErrProxyDuplicateOrder    = 10403 // 包裹已有进行中的代取订单
	ErrProxyRewardOutOfRange  = 10404 // 悬赏金额超出范围
	ErrProxyDeadlineInvalid   = 10405 // 截止时间不合法
	ErrProxyOrderNotFound     = 10411 // 订单不存在
	ErrProxyAlreadyTaken      = 10412 // 订单已被接单或已取消
	ErrProxyNotRunner         = 10413 // 当前用户无跑腿员资质
	ErrProxySelfAccept        = 10414 // 接单人与发布者为同一人
	ErrProxyNotTaker          = 10421 // 订单不存在或无权操作
	ErrProxyNotDelivering     = 10422 // 订单状态非配送中
	ErrProxyPhotoInvalid      = 10423 // 照片数量超限或URL非法
	ErrProxyNotPublisher      = 10431 // 订单不存在或无权操作
	ErrProxyNotPendingConfirm = 10432 // 订单状态非待确认
	ErrProxyRejectNoReason    = 10433 // 拒绝时未提供原因
	ErrProxyCancelNotAllowed  = 10443 // 订单状态不允许取消

	// Shelf module error codes (10501~10599) — see 详细设计文档/api详细设计.md 6.
	ErrShelfCodeExists    = 10501 // 货架编号已存在
	ErrShelfCapacityInvalid = 10502 // 容量或排/列数值非法
	ErrShelfNotFound      = 10511 // 货架不存在
	ErrShelfMaxBelowCurrent= 10512 // max_capacity 小于当前占用
	ErrShelfCodeUsed      = 10513 // 货架编号已被占用

	// Notification module error codes (10601~10699).
	ErrNotifyNotFound = 10601 // 通知不存在
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
	// Parcel module — 10201~10243
	case ErrParcelDuplicate, ErrParcelStatusInvalid, ErrParcelStatusReadonly, ErrParcelNotPending:
		return http.StatusConflict
	case ErrShelfNotFoundOrFull, ErrReceiverPhoneInvalid, ErrCourierCompanyInvalid, ErrPickupCodeGenFail,
		ErrBatchFileInvalid, ErrBatchFileTooLarge, ErrBatchTemplateInvalid, ErrParcelNotOwner:
		return http.StatusBadRequest
	case ErrParcelNotFound:
		return http.StatusNotFound
	case ErrParcelNoPermission:
		return http.StatusForbidden
	// Pickup module — 10301~10313
	case ErrPickupCodeInvalid, ErrSelfCheckoutInvalid:
		return http.StatusBadRequest
	case ErrPickupStatusNotPending, ErrPickupGeoAbnormal, ErrPickupUserBlacklisted,
		ErrSelfCheckoutGeoFar, ErrSelfCheckoutRisk:
		return http.StatusForbidden
	case ErrPickupFreqLimit:
		return http.StatusTooManyRequests
	// Proxy module — 10401~10443
	case ErrProxyParcelNotOwner, ErrProxyParcelNotPending, ErrProxyDuplicateOrder,
		ErrProxyAlreadyTaken, ErrProxySelfAccept, ErrProxyNotTaker, ErrProxyNotDelivering,
		ErrProxyNotPublisher, ErrProxyNotPendingConfirm, ErrProxyCancelNotAllowed:
		return http.StatusConflict
	case ErrProxyRewardOutOfRange, ErrProxyDeadlineInvalid, ErrProxyPhotoInvalid, ErrProxyRejectNoReason:
		return http.StatusBadRequest
	case ErrProxyOrderNotFound:
		return http.StatusNotFound
	case ErrProxyNotRunner:
		return http.StatusForbidden
	// Shelf module — 10501~10513
	case ErrShelfCodeExists, ErrShelfCodeUsed, ErrShelfMaxBelowCurrent:
		return http.StatusConflict
	case ErrShelfCapacityInvalid:
		return http.StatusBadRequest
	case ErrShelfNotFound:
		return http.StatusNotFound
	case ErrNotifyNotFound:
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
	// Parcel module — 10201~10243
	case ErrParcelDuplicate:
		return "快递单号已存在"
	case ErrShelfNotFoundOrFull:
		return "货架编号不存在或已满"
	case ErrReceiverPhoneInvalid:
		return "收件人手机号格式错误"
	case ErrCourierCompanyInvalid:
		return "快递公司不在字典中"
	case ErrPickupCodeGenFail:
		return "取件码生成失败"
	case ErrBatchFileInvalid:
		return "文件格式不支持"
	case ErrBatchFileTooLarge:
		return "文件过大"
	case ErrBatchTemplateInvalid:
		return "模板字段缺失或非法"
	case ErrParcelNotFound:
		return "包裹不存在"
	case ErrParcelNoPermission:
		return "无权查看该包裹"
	case ErrParcelStatusInvalid:
		return "状态流转非法"
	case ErrParcelStatusReadonly:
		return "目标状态不支持手工变更"
	case ErrParcelNotOwner:
		return "非本人包裹"
	case ErrParcelNotPending:
		return "包裹状态非待取"
	// Pickup module — 10301~10313
	case ErrPickupCodeInvalid:
		return "取件码不存在或已失效"
	case ErrPickupStatusNotPending:
		return "包裹状态非待取"
	case ErrPickupGeoAbnormal:
		return "地理位置异常"
	case ErrPickupUserBlacklisted:
		return "用户在黑名单中"
	case ErrPickupFreqLimit:
		return "取件频次超限"
	case ErrSelfCheckoutInvalid:
		return "取件码无效"
	case ErrSelfCheckoutGeoFar:
		return "地理位置超距"
	case ErrSelfCheckoutRisk:
		return "风控触发，需管理员介入"
	// Proxy module — 10401~10443
	case ErrProxyParcelNotOwner:
		return "包裹不存在或非本人"
	case ErrProxyParcelNotPending:
		return "包裹状态非待取"
	case ErrProxyDuplicateOrder:
		return "包裹已有进行中的代取订单"
	case ErrProxyRewardOutOfRange:
		return "悬赏金额超出范围"
	case ErrProxyDeadlineInvalid:
		return "截止时间不合法"
	case ErrProxyOrderNotFound:
		return "订单不存在"
	case ErrProxyAlreadyTaken:
		return "订单已被接单或已取消"
	case ErrProxyNotRunner:
		return "当前用户无跑腿员资质"
	case ErrProxySelfAccept:
		return "接单人与发布者为同一人"
	case ErrProxyNotTaker:
		return "订单不存在或无权操作"
	case ErrProxyNotDelivering:
		return "订单状态非配送中"
	case ErrProxyPhotoInvalid:
		return "照片数量超限或URL非法"
	case ErrProxyNotPublisher:
		return "订单不存在或无权操作"
	case ErrProxyNotPendingConfirm:
		return "订单状态非待确认"
	case ErrProxyRejectNoReason:
		return "拒绝时未提供原因"
	case ErrProxyCancelNotAllowed:
		return "订单状态不允许取消"
	// Shelf module — 10501~10513
	case ErrShelfCodeExists:
		return "货架编号已存在"
	case ErrShelfCapacityInvalid:
		return "容量或排/列数值非法"
	case ErrShelfNotFound:
		return "货架不存在"
	case ErrShelfMaxBelowCurrent:
		return "最大容量小于当前占用"
	case ErrShelfCodeUsed:
		return "货架编号已被占用"
	case ErrNotifyNotFound:
		return "通知不存在"
	default:
		return "未知错误"
	}
}
