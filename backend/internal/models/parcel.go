package models

import (
	"database/sql"
	"time"
)

// Parcel status constants (matching parcels.status).
const (
	ParcelStatusPending  = 1 // 待取
	ParcelStatusPickedUp = 2 // 已取
	ParcelStatusDetained = 3 // 滞留
	ParcelStatusReturned = 4 // 已退件
	ParcelStatusAbnormal = 5 // 异常
)

// Parcel maps the `parcels` table.
type Parcel struct {
	ID             int64          `db:"id" json:"id"`
	StationID      int64          `db:"station_id" json:"station_id"`
	TrackingNo     string         `db:"tracking_no" json:"tracking_no"`
	CourierCompany string         `db:"courier_company" json:"courier_company"`
	ShelfCode      sql.NullString `db:"shelf_code" json:"shelf_code,omitempty"`
	PickupCode     string         `db:"pickup_code" json:"pickup_code"`
	ReceiverPhone  string         `db:"receiver_phone" json:"receiver_phone"`
	ReceiverUserID sql.NullInt64  `db:"receiver_user_id" json:"receiver_user_id,omitempty"`
	ReceiverName   sql.NullString `db:"receiver_name" json:"receiver_name,omitempty"`
	Weight         sql.NullFloat64 `db:"weight" json:"weight,omitempty"`
	IsFragile      int8           `db:"is_fragile" json:"is_fragile"`
	Remarks        sql.NullString `db:"remarks" json:"remarks,omitempty"`
	Status         int8           `db:"status" json:"status"`
	StorageTime    time.Time      `db:"storage_time" json:"storage_time"`
	PickupTime     sql.NullTime   `db:"pickup_time" json:"pickup_time,omitempty"`
	ReturnTime     sql.NullTime   `db:"return_time" json:"return_time,omitempty"`
	LastNotifyTime sql.NullTime   `db:"last_notify_time" json:"last_notify_time,omitempty"`
	NotifyCount    int            `db:"notify_count" json:"notify_count"`
	OperatorID     sql.NullInt64  `db:"operator_id" json:"operator_id,omitempty"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

// ParcelDTO is the API-facing parcel representation. Sensitive/additional
// fields like pickup_code or receiver_phone are conditionally exposed.
type ParcelDTO struct {
	ID             int64   `json:"id"`
	StationID      int64   `json:"station_id,omitempty"`
	TrackingNo     string  `json:"tracking_no"`
	CourierCompany string  `json:"courier_company"`
	ShelfCode      string  `json:"shelf_code,omitempty"`
	PickupCode     string  `json:"pickup_code,omitempty"`
	ReceiverPhone  string  `json:"receiver_phone,omitempty"`
	ReceiverName   string  `json:"receiver_name,omitempty"`
	Weight         float64 `json:"weight,omitempty"`
	IsFragile      bool    `json:"is_fragile"`
	Remarks        string  `json:"remarks,omitempty"`
	Status         int8    `json:"status"`
	StatusText     string  `json:"status_text"`
	StorageTime    string  `json:"storage_time"`
	PickupTime     string  `json:"pickup_time,omitempty"`
	ReturnTime     string  `json:"return_time,omitempty"`
	NotifyCount    int     `json:"notify_count"`
}

// ToDTO converts a Parcel to its DTO. PickupCode and ReceiverPhone are
// always masked or hidden by default; callers decide whether to surface them
// through additional per-field logic.
func (p *Parcel) ToDTO() *ParcelDTO {
	dto := &ParcelDTO{
		ID:             p.ID,
		StationID:      p.StationID,
		TrackingNo:     p.TrackingNo,
		CourierCompany: p.CourierCompany,
		IsFragile:      p.IsFragile == 1,
		Status:         p.Status,
		StatusText:     ParcelStatusText(p.Status),
		StorageTime:    p.StorageTime.Format("2006-01-02 15:04:05"),
		NotifyCount:    p.NotifyCount,
	}
	if p.ShelfCode.Valid {
		dto.ShelfCode = p.ShelfCode.String
	}
	if p.ReceiverName.Valid {
		dto.ReceiverName = p.ReceiverName.String
	}
	if p.Weight.Valid {
		dto.Weight = p.Weight.Float64
	}
	if p.Remarks.Valid {
		dto.Remarks = p.Remarks.String
	}
	return dto
}

// ToAdminDTO converts a Parcel to an admin-facing DTO that includes the
// full pickup_code and masked phone.
func (p *Parcel) ToAdminDTO() *ParcelDTO {
	dto := p.ToDTO()
	dto.PickupCode = p.PickupCode
	dto.ReceiverPhone = MaskPhone(p.ReceiverPhone)
	return dto
}

// ToMyDTO converts a Parcel to an owner-facing DTO with pickup_code but
// no phone exposure.
func (p *Parcel) ToMyDTO() *ParcelDTO {
	dto := p.ToDTO()
	dto.PickupCode = p.PickupCode
	dto.StationID = 0
	return dto
}

// ParcelStatusText returns a human-readable status string.
func ParcelStatusText(s int8) string {
	switch s {
	case ParcelStatusPending:
		return "待取"
	case ParcelStatusPickedUp:
		return "已取"
	case ParcelStatusDetained:
		return "滞留"
	case ParcelStatusReturned:
		return "已退件"
	case ParcelStatusAbnormal:
		return "异常"
	default:
		return "未知"
	}
}

// Copy nullable helpers from the user model for convenience in the repository.
func nullableStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v > 0}
}

func nullableTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}
