package models

import (
	"database/sql"
	"time"
)

// ProxyOrder status constants (matching proxy_orders.status).
const (
	ProxyStatusPending    int8 = 1 // 待接单
	ProxyStatusDelivering = 2 // 配送中
	ProxyStatusConfirm   = 3 // 待确认
	ProxyStatusDone      = 4 // 已完成
	ProxyStatusCancelled = 5 // 已取消
	ProxyStatusFailed    = 6 // 取件失败
)

// ProxyOrder maps the `proxy_orders` table.
type ProxyOrder struct {
	ID              int64          `db:"id" json:"id"`
	StationID       int64          `db:"station_id" json:"station_id"`
	ParcelID        int64          `db:"parcel_id" json:"parcel_id"`
	PublisherID     int64          `db:"publisher_id" json:"publisher_id"`
	TakerID         sql.NullInt64  `db:"taker_id" json:"taker_id,omitempty"`
	RewardAmount    float64        `db:"reward_amount" json:"reward_amount"`
	TempPickupCode  sql.NullString `db:"temp_pickup_code" json:"temp_pickup_code,omitempty"`
	Deadline        time.Time      `db:"deadline" json:"deadline"`
	Status          int8           `db:"status" json:"status"`
	CancelReason    sql.NullString `db:"cancel_reason" json:"cancel_reason,omitempty"`
	DeliveryTime    sql.NullTime   `db:"delivery_time" json:"delivery_time,omitempty"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

// ProxyOrderDTO is the API-facing representation.
type ProxyOrderDTO struct {
	ID                int64   `json:"id"`
	ParcelID          int64   `json:"parcel_id,omitempty"`
	StationID         int64   `json:"station_id,omitempty"`
	StationName       string  `json:"station_name,omitempty"`
	PublisherID       int64   `json:"publisher_id,omitempty"`
	PublisherNickname string  `json:"publisher_nickname,omitempty"`
	TakerID           int64   `json:"taker_id,omitempty"`
	TakerNickname     string  `json:"taker_nickname,omitempty"`
	RewardAmount      float64 `json:"reward_amount"`
	Status            int8    `json:"status"`
	StatusText        string  `json:"status_text"`
	TempPickupCode    string  `json:"temp_pickup_code,omitempty"`
	Deadline          string  `json:"deadline,omitempty"`
	DeliveryTime      string  `json:"delivery_time,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

// ToDTO converts a ProxyOrder to its basic DTO.
func (po *ProxyOrder) ToDTO() *ProxyOrderDTO {
	dto := &ProxyOrderDTO{
		ID:           po.ID,
		ParcelID:     po.ParcelID,
		StationID:    po.StationID,
		PublisherID:  po.PublisherID,
		RewardAmount: po.RewardAmount,
		Status:       po.Status,
		StatusText:   ProxyStatusText(po.Status),
		CreatedAt:    po.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if po.TakerID.Valid {
		dto.TakerID = po.TakerID.Int64
	}
	if po.TempPickupCode.Valid {
		dto.TempPickupCode = po.TempPickupCode.String
	}
	if po.DeliveryTime.Valid {
		dto.DeliveryTime = po.DeliveryTime.Time.Format("2006-01-02 15:04:05")
	}
	if !po.Deadline.IsZero() {
		dto.Deadline = po.Deadline.Format("2006-01-02 15:04:05")
	}
	return dto
}

// ProxyStatusText returns a human-readable status string.
func ProxyStatusText(s int8) string {
	switch s {
	case ProxyStatusPending:
		return "待接单"
	case ProxyStatusDelivering:
		return "配送中"
	case ProxyStatusConfirm:
		return "待确认"
	case ProxyStatusDone:
		return "已完成"
	case ProxyStatusCancelled:
		return "已取消"
	case ProxyStatusFailed:
		return "取件失败"
	default:
		return "未知"
	}
}
