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

// PublishResult is the response for Publish.
type PublishResult struct {
	OrderID   int64  `json:"order_id"`
	Status    int8   `json:"status"`
	CreatedAt string `json:"created_at"`
}

// AcceptResult is the response for Accept.
type AcceptResult struct {
	OrderID        int64             `json:"order_id"`
	Status         int8              `json:"status"`
	TempPickupCode string            `json:"temp_pickup_code"`
	Parcel         *models.ParcelDTO `json:"parcel"`
}

// DeliveryResult is the response for RequestDelivery.
type DeliveryResult struct {
	OrderID      int64  `json:"order_id"`
	Status       int8   `json:"status"`
	DeliveryTime string `json:"delivery_time"`
}

// ConfirmResult is the response for ConfirmDelivery.
type ConfirmResult struct {
	OrderID int64 `json:"order_id"`
	Status  int8  `json:"status"`
}

// CancelResult is the response for CancelOrder.
type CancelResult struct {
	OrderID int64 `json:"order_id"`
	Status  int8  `json:"status"`
}

// ProxyListResult wraps a paginated proxy order list.
type ProxyListResult struct {
	Items []*models.ProxyOrderDTO `json:"items"`
	Total int64                   `json:"total"`
}

// ProxyMyOrderFilter is the service-layer filter for my orders.
type ProxyMyOrderFilter struct {
	UserID int64
	Role   string // "publisher" / "taker" / ""
	Status *int8
	Offset int
	Limit  int
}

// ProxyService implements proxy order publishing, acceptance, delivery,
// confirmation, cancellation, and listing.
type ProxyService struct {
	proxyRepo  repository.ProxyOrderRepo
	parcelRepo repository.ParcelRepo
	userRepo   repository.UserRepo
	db         *sqlx.DB
}

func NewProxyService(
	pr repository.ProxyOrderRepo,
	par repository.ParcelRepo,
	ur repository.UserRepo,
	db *sqlx.DB,
) *ProxyService {
	return &ProxyService{proxyRepo: pr, parcelRepo: par, userRepo: ur, db: db}
}

// Publish creates a new proxy order for the user's pending parcel.
func (s *ProxyService) Publish(ctx context.Context, userID int64, parcelID int64, rewardAmount float64, deadline, remark string) (*PublishResult, error) {
	if rewardAmount < 0.01 || rewardAmount > 500.00 {
		return nil, apperrors.New(apperrors.ErrProxyRewardOutOfRange, "")
	}

	dl, err := time.Parse("2006-01-02 15:04:05", deadline)
	if err != nil {
		return nil, apperrors.New(apperrors.ErrProxyDeadlineInvalid, "")
	}
	if dl.Before(time.Now().Add(30 * time.Minute)) {
		return nil, apperrors.New(apperrors.ErrProxyDeadlineInvalid, "")
	}

	p, err := s.parcelRepo.FindByID(ctx, s.db, parcelID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyParcelNotOwner, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if !p.ReceiverUserID.Valid || p.ReceiverUserID.Int64 != userID {
		return nil, apperrors.New(apperrors.ErrProxyParcelNotOwner, "")
	}
	if p.Status != models.ParcelStatusPending {
		return nil, apperrors.New(apperrors.ErrProxyParcelNotPending, "")
	}

	// Check for existing active proxy order.
	if _, err := s.proxyRepo.FindByParcelID(ctx, s.db, parcelID,
		models.ProxyStatusPending, models.ProxyStatusDelivering, models.ProxyStatusConfirm); err == nil {
		return nil, apperrors.New(apperrors.ErrProxyDuplicateOrder, "")
	} else if !stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	order := &models.ProxyOrder{
		StationID:    p.StationID,
		ParcelID:     parcelID,
		PublisherID:  userID,
		RewardAmount: rewardAmount,
		Deadline:     dl,
		Status:       models.ProxyStatusPending,
	}
	id, err := s.proxyRepo.Create(ctx, s.db, order)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	return &PublishResult{
		OrderID:   id,
		Status:    models.ProxyStatusPending,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// ListTasks returns the task hall list for runners.
func (s *ProxyService) ListTasks(ctx context.Context, stationID int64, minReward float64, sort string, offset, limit int) (*ProxyListResult, error) {
	filter := repository.ProxyTaskFilter{
		StationID: stationID,
		MinReward: minReward,
		Sort:      sort,
		Offset:    offset,
		Limit:     limit,
	}
	orders, total, err := s.proxyRepo.ListTasks(ctx, s.db, filter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.ProxyOrderDTO, 0, len(orders))
	for _, o := range orders {
		dto := o.ToDTO()
		dto.ParcelID = o.ParcelID
		dto.StationID = o.StationID
		items = append(items, dto)
	}
	return &ProxyListResult{Items: items, Total: total}, nil
}

// Accept lets a runner accept a pending proxy order.
func (s *ProxyService) Accept(ctx context.Context, orderID, userID int64) (*AcceptResult, error) {
	user, err := s.userRepo.FindByID(ctx, s.db, userID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyNotRunner, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if user.UserType != models.UserTypeRunner || user.RunnerStatus != models.RunnerStatusApproved {
		return nil, apperrors.New(apperrors.ErrProxyNotRunner, "")
	}
	if user.IsBlacklistedBool() {
		return nil, apperrors.New(apperrors.ErrProxyNotRunner, "")
	}

	order, err := s.proxyRepo.FindByID(ctx, s.db, orderID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyOrderNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if order.PublisherID == userID {
		return nil, apperrors.New(apperrors.ErrProxySelfAccept, "")
	}
	if order.Status != models.ProxyStatusPending {
		return nil, apperrors.New(apperrors.ErrProxyAlreadyTaken, "")
	}

	tempCode, _ := generateRandomCode(6)
	rows, err := s.proxyRepo.UpdateAccepted(ctx, s.db, orderID, userID, tempCode)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if rows == 0 {
		return nil, apperrors.New(apperrors.ErrProxyAlreadyTaken, "")
	}

	p, _ := s.parcelRepo.FindByID(ctx, s.db, order.ParcelID)
	var parcel *models.ParcelDTO
	if p != nil {
		parcel = p.ToDTO()
	}

	return &AcceptResult{
		OrderID:        orderID,
		Status:         models.ProxyStatusDelivering,
		TempPickupCode: tempCode,
		Parcel:         parcel,
	}, nil
}

// RequestDelivery marks the order as "待确认" by the runner.
func (s *ProxyService) RequestDelivery(ctx context.Context, orderID, userID int64, photos []string, remark string) (*DeliveryResult, error) {
	_ = remark
	if len(photos) < 1 || len(photos) > 5 {
		return nil, apperrors.New(apperrors.ErrProxyPhotoInvalid, "")
	}
	for _, url := range photos {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return nil, apperrors.New(apperrors.ErrProxyPhotoInvalid, "")
		}
	}

	order, err := s.proxyRepo.FindByID(ctx, s.db, orderID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyNotTaker, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if !order.TakerID.Valid || order.TakerID.Int64 != userID {
		return nil, apperrors.New(apperrors.ErrProxyNotTaker, "")
	}
	if order.Status != models.ProxyStatusDelivering {
		return nil, apperrors.New(apperrors.ErrProxyNotDelivering, "")
	}

	if err := s.proxyRepo.UpdateDelivery(ctx, s.db, orderID, models.ProxyStatusConfirm); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	return &DeliveryResult{
		OrderID:      orderID,
		Status:       models.ProxyStatusConfirm,
		DeliveryTime: time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// ConfirmDelivery lets the publisher confirm or reject delivery.
func (s *ProxyService) ConfirmDelivery(ctx context.Context, orderID, userID int64, accepted bool, reason string) (*ConfirmResult, error) {
	if !accepted && strings.TrimSpace(reason) == "" {
		return nil, apperrors.New(apperrors.ErrProxyRejectNoReason, "")
	}

	order, err := s.proxyRepo.FindByID(ctx, s.db, orderID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyNotPublisher, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if order.PublisherID != userID {
		return nil, apperrors.New(apperrors.ErrProxyNotPublisher, "")
	}
	if order.Status != models.ProxyStatusConfirm {
		return nil, apperrors.New(apperrors.ErrProxyNotPendingConfirm, "")
	}

	newStatus := int8(models.ProxyStatusDone)
	if !accepted {
		newStatus = int8(models.ProxyStatusFailed)
	}
	if err := s.proxyRepo.UpdateConfirm(ctx, s.db, orderID, newStatus); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	_ = reason

	return &ConfirmResult{OrderID: orderID, Status: newStatus}, nil
}

// CancelOrder cancels a proxy order (publisher or runner only).
func (s *ProxyService) CancelOrder(ctx context.Context, orderID, userID int64, reason string) (*CancelResult, error) {
	order, err := s.proxyRepo.FindByID(ctx, s.db, orderID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrProxyOrderNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if order.PublisherID != userID && (!order.TakerID.Valid || order.TakerID.Int64 != userID) {
		return nil, apperrors.New(apperrors.ErrProxyNotTaker, "")
	}
	if order.Status != models.ProxyStatusPending && order.Status != models.ProxyStatusDelivering {
		return nil, apperrors.New(apperrors.ErrProxyCancelNotAllowed, "")
	}

	newStatus := int8(models.ProxyStatusCancelled)
	if userID == order.PublisherID && order.Status == models.ProxyStatusDelivering {
		newStatus = int8(models.ProxyStatusFailed)
	}

	if err := s.proxyRepo.Cancel(ctx, s.db, orderID, newStatus, reason); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return &CancelResult{OrderID: orderID, Status: newStatus}, nil
}

// ListMyOrders returns orders where the user is publisher or taker.
func (s *ProxyService) ListMyOrders(ctx context.Context, filter ProxyMyOrderFilter) (*ProxyListResult, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	repoFilter := repository.ProxyMyOrderFilter{
		UserID: filter.UserID,
		Role:   filter.Role,
		Status: filter.Status,
		Offset: filter.Offset,
		Limit:  limit,
	}
	orders, total, err := s.proxyRepo.ListMyOrders(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.ProxyOrderDTO, 0, len(orders))
	for _, o := range orders {
		dto := o.ToDTO()
		dto.ParcelID = o.ParcelID
		items = append(items, dto)
	}
	return &ProxyListResult{Items: items, Total: total}, nil
}

func generateRandomCode(length int) (string, error) {
	maxVal := big.NewInt(int64(powInt10(length)) - int64(powInt10(length-1)))
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", err
	}
	code := int64(powInt10(length-1)) + n.Int64()
	return fmt.Sprintf("%0*d", length, code), nil
}

func powInt10(n int) int {
	v := 1
	for i := 0; i < n; i++ {
		v *= 10
	}
	return v
}
