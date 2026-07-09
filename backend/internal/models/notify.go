package models

import (
	"database/sql"
	"time"
)

// Notification type constants.
const (
	NotifyTypeInbound      int8 = 1 // 入库
	NotifyTypeRemind       int8 = 2 // 催取
	NotifyTypeProxyStatus  int8 = 3 // 代取状态
	NotifyTypeSystem       int8 = 4 // 系统
)

// Send status constants.
const (
	SendStatusPending int8 = 0 // 待发送
	SendStatusSent    int8 = 1 // 已发送
	SendStatusFailed  int8 = 2 // 发送失败
)

// Channel constants.
const (
	ChannelWechat int8 = 1 // 微信
	ChannelSMS    int8 = 2 // 短信
)

// Notification maps the `notifications` table.
type Notification struct {
	ID         int64          `db:"id" json:"id"`
	UserID     int64          `db:"user_id" json:"user_id"`
	ParcelID   sql.NullInt64  `db:"parcel_id" json:"parcel_id,omitempty"`
	Title      string         `db:"title" json:"title"`
	Content    string         `db:"content" json:"content"`
	Type       int8           `db:"type" json:"type"`
	IsRead     int8           `db:"is_read" json:"is_read"`
	SendStatus int8           `db:"send_status" json:"send_status"`
	Channel    int8           `db:"channel" json:"channel"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

// NotificationDTO is the API-facing representation.
type NotificationDTO struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Type      int8   `json:"type"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

// ToDTO converts a Notification to its DTO.
func (n *Notification) ToDTO() *NotificationDTO {
	return &NotificationDTO{
		ID:        n.ID,
		Title:     n.Title,
		Content:   n.Content,
		Type:      n.Type,
		IsRead:    n.IsRead == 1,
		CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
