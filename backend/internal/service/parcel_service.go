package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	stderrors "errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

// ParcelListFilter mirrors the handler-level filter.
type ParcelListFilter struct {
	StationID      int64
	TrackingNo     string
	ReceiverPhone  string
	Status         *int8
	CourierCompany string
	ShelfCode      string
	ReceiverUserID int64
	StorageStart   string
	StorageEnd     string
	Offset         int
	Limit          int
}

// ParcelListResult bundles the paged list of admin-facing DTOs.
type ParcelListResult struct {
	Items []*models.ParcelDTO `json:"items"`
	Total int64               `json:"total"`
}

// ScanInRequest is the input for ParcelService.ScanIn.
type ScanInRequest struct {
	StationID      int64   `json:"station_id"`
	TrackingNo     string  `json:"tracking_no"`
	CourierCompany string  `json:"courier_company"`
	ReceiverPhone  string  `json:"receiver_phone"`
	ReceiverName   string  `json:"receiver_name"`
	ShelfCode      string  `json:"shelf_code"`
	Weight         float64 `json:"weight"`
	IsFragile      bool    `json:"is_fragile"`
	Remarks        string  `json:"remarks"`
	OperatorID     int64   `json:"operator_id"`
}

// ScanInResult is the response for ScanIn.
type ScanInResult struct {
	ParcelID    int64  `json:"parcel_id"`
	PickupCode  string `json:"pickup_code"`
	ShelfCode   string `json:"shelf_code"`
	StorageTime string `json:"storage_time"`
}

// UpdateStatusRequest is the input for ParcelService.UpdateStatus.
type UpdateStatusRequest struct {
	Status int8   `json:"status"`
	Reason string `json:"reason"`
}

// ParcelService implements parcel intake, listing, status management.
type ParcelService struct {
	parcelRepo repository.ParcelRepo
	shelfRepo  repository.ShelfRepo
	userRepo   repository.UserRepo
	db         *sqlx.DB
}

func NewParcelService(pr repository.ParcelRepo, sr repository.ShelfRepo, ur repository.UserRepo, db *sqlx.DB) *ParcelService {
	return &ParcelService{parcelRepo: pr, shelfRepo: sr, userRepo: ur, db: db}
}

// ScanIn performs parcel intake. It validates input, generates a unique
// pickup code, auto-assigns a shelf if not specified, and inserts the parcel
// + updates shelf capacity in a single transaction.
func (s *ParcelService) ScanIn(ctx context.Context, req ScanInRequest) (*ScanInResult, error) {
	trackingNo := strings.TrimSpace(req.TrackingNo)
	if trackingNo == "" || len(trackingNo) > 64 {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "tracking_no 非法")
	}
	courier := strings.TrimSpace(req.CourierCompany)
	if courier == "" {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "courier_company 必填")
	}
	receiverPhone := strings.TrimSpace(req.ReceiverPhone)
	if !models.IsValidPhone(receiverPhone) {
		return nil, apperrors.New(apperrors.ErrReceiverPhoneInvalid, "")
	}

	// Check for duplicate tracking_no in the same station.
	if _, err := s.parcelRepo.FindByTrackingNo(ctx, s.db, trackingNo, req.StationID); err == nil {
		return nil, apperrors.New(apperrors.ErrParcelDuplicate, "")
	} else if !stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	// Resolve shelf: specified shelf takes priority, else auto-assign.
	shelfCode := strings.TrimSpace(req.ShelfCode)
	if shelfCode == "" {
		sl, err := s.shelfRepo.FindBestForStation(ctx, s.db, req.StationID)
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.New(apperrors.ErrShelfNotFoundOrFull, "")
		}
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
		}
		shelfCode = sl.ShelfCode
	}

	// Generate a unique pickup code (retry up to 10 times).
	pickupCode := ""
	for i := 0; i < 10; i++ {
		code, err := generatePickupCode()
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
		}
		if _, err := s.parcelRepo.FindByPickupCode(ctx, s.db, code, req.StationID); stderrors.Is(err, sql.ErrNoRows) {
			pickupCode = code
			break
		}
	}
	if pickupCode == "" {
		return nil, apperrors.New(apperrors.ErrPickupCodeGenFail, "")
	}

	// Look up receiver_user_id by phone.
	var receiverUserID sql.NullInt64
	if u, err := s.userRepo.FindByPhone(ctx, s.db, receiverPhone); err == nil {
		receiverUserID = sql.NullInt64{Int64: u.ID, Valid: true}
	}

	now := time.Now()
	isFragile := int8(0)
	if req.IsFragile {
		isFragile = 1
	}

	parcel := &models.Parcel{
		StationID:      req.StationID,
		TrackingNo:     trackingNo,
		CourierCompany: courier,
		ShelfCode:      nullableStr(shelfCode),
		PickupCode:     pickupCode,
		ReceiverPhone:  receiverPhone,
		ReceiverUserID: receiverUserID,
		ReceiverName:   nullableStr(strings.TrimSpace(req.ReceiverName)),
		Weight:         sql.NullFloat64{Float64: req.Weight, Valid: req.Weight > 0},
		IsFragile:      isFragile,
		Remarks:        nullableStr(strings.TrimSpace(req.Remarks)),
		Status:         models.ParcelStatusPending,
		StorageTime:    now,
		OperatorID:     sql.NullInt64{Int64: req.OperatorID, Valid: req.OperatorID > 0},
	}

	var parcelID int64
	err := repository.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		id, e := s.parcelRepo.Create(ctx, tx, parcel)
		if e != nil {
			return e
		}
		parcelID = id
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	return &ScanInResult{
		ParcelID:    parcelID,
		PickupCode:  pickupCode,
		ShelfCode:   shelfCode,
		StorageTime: now.Format("2006-01-02 15:04:05"),
	}, nil
}

// ListParcels returns a paginated admin-facing parcel list.
func (s *ParcelService) ListParcels(ctx context.Context, filter ParcelListFilter) (*ParcelListResult, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	repoFilter := repository.ParcelFilter{
		StationID:      filter.StationID,
		TrackingNo:     filter.TrackingNo,
		ReceiverPhone:  filter.ReceiverPhone,
		Status:         filter.Status,
		CourierCompany: filter.CourierCompany,
		ShelfCode:      filter.ShelfCode,
		StorageStart:   filter.StorageStart,
		StorageEnd:     filter.StorageEnd,
		Offset:         filter.Offset,
		Limit:          limit,
	}
	parcels, total, err := s.parcelRepo.ListByFilter(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.ParcelDTO, 0, len(parcels))
	for _, p := range parcels {
		items = append(items, p.ToAdminDTO())
	}
	return &ParcelListResult{Items: items, Total: total}, nil
}

// ListMyParcels returns the current user's own parcel list.
func (s *ParcelService) ListMyParcels(ctx context.Context, userID int64, status *int8, keyword string, offset, limit int) (*ParcelListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	repoFilter := repository.ParcelFilter{
		ReceiverUserID: userID,
		Status:         status,
		Offset:         offset,
		Limit:          limit,
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		repoFilter.TrackingNo = kw
	}
	parcels, total, err := s.parcelRepo.ListByFilter(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.ParcelDTO, 0, len(parcels))
	for _, p := range parcels {
		items = append(items, p.ToMyDTO())
	}
	return &ParcelListResult{Items: items, Total: total}, nil
}

// GetParcel returns a parcel DTO. The caller indicates whether the request
// is from the owner (in which case pickup_code is returned) or admin.
func (s *ParcelService) GetParcel(ctx context.Context, parcelID int64, userID int64, isAdmin bool) (*models.ParcelDTO, error) {
	p, err := s.parcelRepo.FindByID(ctx, s.db, parcelID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrParcelNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	if isAdmin {
		return p.ToAdminDTO(), nil
	}

	if p.ReceiverUserID.Valid && p.ReceiverUserID.Int64 == userID {
		dto := p.ToMyDTO()
		dto.PickupCode = p.PickupCode
		return dto, nil
	}
	return nil, apperrors.New(apperrors.ErrParcelNoPermission, "")
}

// UpdateStatus changes a parcel's status. Only admin may change to
// Detained (3), Returned (4), or Abnormal (5).
func (s *ParcelService) UpdateStatus(ctx context.Context, parcelID int64, req UpdateStatusRequest) (*models.ParcelDTO, error) {
	if req.Status != models.ParcelStatusDetained &&
		req.Status != models.ParcelStatusReturned &&
		req.Status != models.ParcelStatusAbnormal {
		return nil, apperrors.New(apperrors.ErrParcelStatusReadonly, "")
	}

	p, err := s.parcelRepo.FindByID(ctx, s.db, parcelID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrParcelNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	// Validate status transitions.
	switch req.Status {
	case models.ParcelStatusDetained:
		if p.Status != models.ParcelStatusPending {
			return nil, apperrors.New(apperrors.ErrParcelStatusInvalid, "")
		}
	case models.ParcelStatusReturned:
		if p.Status != models.ParcelStatusDetained {
			return nil, apperrors.New(apperrors.ErrParcelStatusInvalid, "")
		}
	case models.ParcelStatusAbnormal:
		if p.Status == models.ParcelStatusPickedUp {
			return nil, apperrors.New(apperrors.ErrParcelStatusInvalid, "")
		}
	}

	timeField := ""
	if req.Status == models.ParcelStatusReturned {
		timeField = "return_time"
	}

	err = s.parcelRepo.UpdateStatusWithTime(ctx, s.db, parcelID, req.Status, timeField)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	p, err = s.parcelRepo.FindByID(ctx, s.db, parcelID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return p.ToAdminDTO(), nil
}

// GetPickupCode returns the pickup_code for a parcel owned by the current user.
func (s *ParcelService) GetPickupCode(ctx context.Context, parcelID, userID int64) (*models.ParcelDTO, error) {
	p, err := s.parcelRepo.FindByID(ctx, s.db, parcelID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrParcelNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if !p.ReceiverUserID.Valid || p.ReceiverUserID.Int64 != userID {
		return nil, apperrors.New(apperrors.ErrParcelNotOwner, "")
	}
	if p.Status != models.ParcelStatusPending {
		return nil, apperrors.New(apperrors.ErrParcelNotPending, "")
	}
	dto := p.ToMyDTO()
	dto.PickupCode = p.PickupCode
	return dto, nil
}

// generatePickupCode produces a cryptographically random 6-digit string.
func generatePickupCode() (string, error) {
	maxVal := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", fmt.Errorf("generatePickupCode: %w", err)
	}
	code := 100000 + n.Int64()
	return fmt.Sprintf("%06d", code), nil
}


