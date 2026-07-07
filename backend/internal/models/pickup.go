package models

import (
	"database/sql"
	"time"
)

// Operator type constants (pickup_logs.operator_type).
const (
	OpTypeAdmin   = 1 // 管理员
	OpTypeKiosk   = 2 // 自助机
	OpTypeRunner  = 3 // 跑腿员
	OpTypeSelf    = 4 // 本人
)

// Verification method constants (pickup_logs.verification_method).
const (
	VerifyScan   = 1 // 扫码
	VerifyManual = 2 // 手动输入
	VerifyFace   = 3 // 人脸
)

// PickupLog maps the `pickup_logs` table.
type PickupLog struct {
	ID                 int64           `db:"id" json:"id"`
	ParcelID           int64           `db:"parcel_id" json:"parcel_id"`
	OperatorID         sql.NullInt64   `db:"operator_id" json:"operator_id,omitempty"`
	OperatorType       int8            `db:"operator_type" json:"operator_type"`
	VerificationMethod int8            `db:"verification_method" json:"verification_method"`
	LocationLat        sql.NullFloat64 `db:"location_lat" json:"location_lat,omitempty"`
	LocationLng        sql.NullFloat64 `db:"location_lng" json:"location_lng,omitempty"`
	IPAddress          sql.NullString  `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent          sql.NullString  `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt          time.Time       `db:"created_at" json:"created_at"`
}

// PickupLogDTO is the API-facing representation.
type PickupLogDTO struct {
	ID                 int64   `json:"id"`
	ParcelID           int64   `json:"parcel_id"`
	TrackingNo         string  `json:"tracking_no,omitempty"`
	ReceiverPhone      string  `json:"receiver_phone,omitempty"`
	OperatorID         int64   `json:"operator_id,omitempty"`
	OperatorType       int8    `json:"operator_type"`
	VerificationMethod int8    `json:"verification_method"`
	LocationLat        float64 `json:"location_lat,omitempty"`
	LocationLng        float64 `json:"location_lng,omitempty"`
	IPAddress          string  `json:"ip_address,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

// ToDTO converts a PickupLog to its DTO. Extra fields (tracking_no etc.)
// are populated by the service layer when needed.
func (pl *PickupLog) ToDTO() *PickupLogDTO {
	dto := &PickupLogDTO{
		ID:                 pl.ID,
		ParcelID:           pl.ParcelID,
		OperatorType:       pl.OperatorType,
		VerificationMethod: pl.VerificationMethod,
		CreatedAt:          pl.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if pl.OperatorID.Valid {
		dto.OperatorID = pl.OperatorID.Int64
	}
	if pl.LocationLat.Valid {
		dto.LocationLat = pl.LocationLat.Float64
	}
	if pl.LocationLng.Valid {
		dto.LocationLng = pl.LocationLng.Float64
	}
	if pl.IPAddress.Valid {
		dto.IPAddress = pl.IPAddress.String
	}
	return dto
}
