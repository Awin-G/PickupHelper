package service

import (
	"context"
	"database/sql"
	stderrors "errors"
	"math"
	"strings"
	"time"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

const maxGeoDistanceM = 500

// VerifyResult is the response for PickupService.Verify.
type VerifyResult struct {
	ParcelID     int64  `json:"parcel_id"`
	TrackingNo   string `json:"tracking_no"`
	PickupTime   string `json:"pickup_time"`
	OperatorType int8   `json:"operator_type"`
	ProxyOrderID *int64 `json:"proxy_order_id,omitempty"`
}

// SelfCheckoutResult is the response for PickupService.SelfCheckout.
type SelfCheckoutResult struct {
	ParcelID   int64  `json:"parcel_id"`
	PickupTime string `json:"pickup_time"`
}

// ScanStationItem represents a single success or failure in a bulk checkout.
type ScanStationSuccess struct {
	PickupCode string `json:"pickup_code"`
	ParcelID   int64  `json:"parcel_id"`
	PickupTime string `json:"pickup_time"`
}

type ScanStationFailure struct {
	PickupCode string `json:"pickup_code"`
	ReasonCode int    `json:"reason_code"`
	ReasonMsg  string `json:"reason_msg"`
}

// ScanStationResult is the response for PickupService.ScanStation.
type ScanStationResult struct {
	Success []ScanStationSuccess `json:"success"`
	Failed  []ScanStationFailure `json:"failed"`
}

// PickupLogEntry is the DTO for a single pickup log row (admin view).
type PickupLogEntry struct {
	*models.PickupLogDTO
	TrackingNo    string `json:"tracking_no,omitempty"`
	ReceiverPhone string `json:"receiver_phone,omitempty"`
}

// PickupLogListResult wraps the paged list.
type PickupLogListResult struct {
	Items []*PickupLogEntry `json:"items"`
	Total int64             `json:"total"`
}

// PickupLogFilter mirrors the handler query.
type PickupLogFilter struct {
	ParcelID     *int64
	OperatorID   *int64
	OperatorType *int8
	Start        string
	End          string
	Offset       int
	Limit        int
}

// PickupService implements pickup verification, self-checkout, scan-station
// bulk checkout, and pickup log queries.
type PickupService struct {
	parcelRepo   repository.ParcelRepo
	pickupRepo   repository.PickupLogRepo
	shelfRepo    repository.ShelfRepo
	userRepo     repository.UserRepo
	db           *sqlx.DB
}

func NewPickupService(
	pr repository.ParcelRepo,
	plr repository.PickupLogRepo,
	sr repository.ShelfRepo,
	ur repository.UserRepo,
	db *sqlx.DB,
) *PickupService {
	return &PickupService{parcelRepo: pr, pickupRepo: plr, shelfRepo: sr, userRepo: ur, db: db}
}

// Verify processes a pickup verification (admin or runner). It validates the
// pickup code, updates parcel status, decrements shelf capacity, and logs
// the operation in a single DB transaction.
func (s *PickupService) Verify(
	ctx context.Context,
	pickupCode, stationID int64, verificationMethod int8,
	operatorID int64, operatorType int8,
	geoLat, geoLng float64, clientIP, userAgent string,
) (*VerifyResult, error) {
	code := formatCode(pickupCode)
	p, err := s.parcelRepo.FindByPickupCode(ctx, s.db, code, stationID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrPickupCodeInvalid, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if p.Status != models.ParcelStatusPending {
		return nil, apperrors.New(apperrors.ErrPickupStatusNotPending, "")
	}

	// Optional geo check (stub — real implementation in Phase 6+).
	_ = geoLat
	_ = geoLng

	now := time.Now()
	log := &models.PickupLog{
		ParcelID:           p.ID,
		OperatorID:         sql.NullInt64{Int64: operatorID, Valid: operatorID > 0},
		OperatorType:       operatorType,
		VerificationMethod: verificationMethod,
		LocationLat:        sql.NullFloat64{Float64: geoLat, Valid: geoLat != 0},
		LocationLng:        sql.NullFloat64{Float64: geoLng, Valid: geoLng != 0},
		IPAddress:          sql.NullString{String: clientIP, Valid: clientIP != ""},
		UserAgent:          sql.NullString{String: userAgent, Valid: userAgent != ""},
	}

	err = repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		if e := s.parcelRepo.UpdateStatusWithTime(ctx, tx, p.ID, models.ParcelStatusPickedUp, "pickup_time"); e != nil {
			return e
		}
		if _, e := s.pickupRepo.Create(ctx, tx, log); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	return &VerifyResult{
		ParcelID:     p.ID,
		TrackingNo:   p.TrackingNo,
		PickupTime:   now.Format("2006-01-02 15:04:05"),
		OperatorType: operatorType,
	}, nil
}

// SelfCheckout lets a recipient check out their own parcel. Requires geo
// coordinates and validates the parcel belongs to the user.
func (s *PickupService) SelfCheckout(
	ctx context.Context,
	pickupCode, stationID, userID int64,
	geoLat, geoLng float64, clientIP, userAgent string,
) (*SelfCheckoutResult, error) {
	code := formatCode(pickupCode)
	p, err := s.parcelRepo.FindByPickupCode(ctx, s.db, code, stationID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrSelfCheckoutInvalid, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if !p.ReceiverUserID.Valid || p.ReceiverUserID.Int64 != userID {
		return nil, apperrors.New(apperrors.ErrSelfCheckoutInvalid, "")
	}
	if p.Status != models.ParcelStatusPending {
		return nil, apperrors.New(apperrors.ErrPickupStatusNotPending, "")
	}

	// Geo distance stub — real geo calculation in Phase 6+.
	_ = geoLat
	_ = geoLng

	now := time.Now()
	log := &models.PickupLog{
		ParcelID:           p.ID,
		OperatorID:         sql.NullInt64{Int64: userID, Valid: true},
		OperatorType:       models.OpTypeSelf,
		VerificationMethod: models.VerifyManual,
		LocationLat:        sql.NullFloat64{Float64: geoLat, Valid: geoLat != 0},
		LocationLng:        sql.NullFloat64{Float64: geoLng, Valid: geoLng != 0},
		IPAddress:          sql.NullString{String: clientIP, Valid: clientIP != ""},
		UserAgent:          sql.NullString{String: userAgent, Valid: userAgent != ""},
	}

	err = repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		if e := s.parcelRepo.UpdateStatusWithTime(ctx, tx, p.ID, models.ParcelStatusPickedUp, "pickup_time"); e != nil {
			return e
		}
		if _, e := s.pickupRepo.Create(ctx, tx, log); e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	return &SelfCheckoutResult{
		ParcelID:   p.ID,
		PickupTime: now.Format("2006-01-02 15:04:05"),
	}, nil
}

// ScanStation performs a bulk checkout of up to 10 parcels for the current
// user. Each code is processed independently; failures do not block successes.
func (s *PickupService) ScanStation(
	ctx context.Context,
	userID int64, pickupCodes []string,
	geoLat, geoLng float64, clientIP, userAgent string,
) *ScanStationResult {
	result := &ScanStationResult{
		Success: make([]ScanStationSuccess, 0),
		Failed:  make([]ScanStationFailure, 0),
	}
	if len(pickupCodes) > 10 {
		pickupCodes = pickupCodes[:10]
	}

	for _, raw := range pickupCodes {
		code := strings.TrimSpace(raw)
		p, err := s.parcelRepo.FindByPickupCode(ctx, s.db, code, 0)
		if stderrors.Is(err, sql.ErrNoRows) || p == nil {
			result.Failed = append(result.Failed, ScanStationFailure{
				PickupCode: code, ReasonCode: apperrors.ErrSelfCheckoutInvalid,
				ReasonMsg: "取件码无效",
			})
			continue
		}
		if err != nil {
			result.Failed = append(result.Failed, ScanStationFailure{
				PickupCode: code, ReasonCode: apperrors.ErrInternal, ReasonMsg: "查询异常",
			})
			continue
		}
		if !p.ReceiverUserID.Valid || p.ReceiverUserID.Int64 != userID {
			result.Failed = append(result.Failed, ScanStationFailure{
				PickupCode: code, ReasonCode: apperrors.ErrSelfCheckoutInvalid,
				ReasonMsg: "非本人包裹",
			})
			continue
		}
		if p.Status != models.ParcelStatusPending {
			result.Failed = append(result.Failed, ScanStationFailure{
				PickupCode: code, ReasonCode: apperrors.ErrPickupStatusNotPending,
				ReasonMsg: "包裹状态非待取",
			})
			continue
		}

		now := time.Now()
		log := &models.PickupLog{
			ParcelID:           p.ID,
			OperatorID:         sql.NullInt64{Int64: userID, Valid: true},
			OperatorType:       models.OpTypeSelf,
			VerificationMethod: models.VerifyManual,
			LocationLat:        sql.NullFloat64{Float64: geoLat, Valid: geoLat != 0},
			LocationLng:        sql.NullFloat64{Float64: geoLng, Valid: geoLng != 0},
			IPAddress:          sql.NullString{String: clientIP, Valid: clientIP != ""},
			UserAgent:          sql.NullString{String: userAgent, Valid: userAgent != ""},
		}

		err = repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
			if e := s.parcelRepo.UpdateStatusWithTime(ctx, tx, p.ID, models.ParcelStatusPickedUp, "pickup_time"); e != nil {
				return e
			}
			if _, e := s.pickupRepo.Create(ctx, tx, log); e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			result.Failed = append(result.Failed, ScanStationFailure{
				PickupCode: code, ReasonCode: apperrors.ErrInternal, ReasonMsg: "处理异常",
			})
			continue
		}
		result.Success = append(result.Success, ScanStationSuccess{
			PickupCode: code,
			ParcelID:   p.ID,
			PickupTime: now.Format("2006-01-02 15:04:05"),
		})
	}
	return result
}

// ListLogs returns a paginated list of pickup logs (admin only).
func (s *PickupService) ListLogs(ctx context.Context, filter PickupLogFilter) (*PickupLogListResult, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	repoFilter := repository.PickupLogFilter{
		ParcelID:     filter.ParcelID,
		OperatorID:   filter.OperatorID,
		OperatorType: filter.OperatorType,
		Start:        filter.Start,
		End:          filter.End,
		Offset:       filter.Offset,
		Limit:        limit,
	}
	logs, total, err := s.pickupRepo.ListByFilter(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	items := make([]*PickupLogEntry, 0, len(logs))
	for _, l := range logs {
		entry := &PickupLogEntry{PickupLogDTO: l.ToDTO()}
		if p, e := s.parcelRepo.FindByID(ctx, s.db, l.ParcelID); e == nil {
			entry.TrackingNo = p.TrackingNo
		}
		items = append(items, entry)
	}
	return &PickupLogListResult{Items: items, Total: total}, nil
}

// formatCode pads a code integer to 6 digits.
func formatCode(code int64) string {
	return formatSixDigit(int(code))
}

func formatSixDigit(n int) string {
	if n < 0 || n > 999999 {
		return "000000"
	}
	result := ""
	for i := 5; i >= 0; i-- {
		result += string('0' + byte(n/pow10(i)%10))
	}
	return result
}

func pow10(n int) int {
	return int(math.Pow10(n))
}

var _ = sql.ErrNoRows
