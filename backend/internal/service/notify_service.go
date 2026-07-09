package service

import (
	"context"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

// NotifyListResult wraps paginated notification list.
type NotifyListResult struct {
	Items   []*models.NotificationDTO `json:"items"`
	Total   int64                     `json:"total"`
	Unread  int64                     `json:"unread"`
}

// NotifyService implements notification queries and management.
type NotifyService struct {
	notifyRepo repository.NotifyRepo
	db         *sqlx.DB
}

func NewNotifyService(nr repository.NotifyRepo, db *sqlx.DB) *NotifyService {
	return &NotifyService{notifyRepo: nr, db: db}
}

// ListNotifications returns the user's notification list with unread count.
func (s *NotifyService) ListNotifications(ctx context.Context, userID int64, offset, limit int) (*NotifyListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	list, total, err := s.notifyRepo.ListByUser(ctx, s.db, userID, offset, limit)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.NotificationDTO, 0, len(list))
	for _, n := range list {
		items = append(items, n.ToDTO())
	}
	// Count unread.
	var unread int64
	_ = s.db.GetContext(ctx, &unread,
		"SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0", userID)

	return &NotifyListResult{Items: items, Total: total, Unread: unread}, nil
}

// MarkAllRead marks all user's notifications as read.
func (s *NotifyService) MarkAllRead(ctx context.Context, userID int64) error {
	if err := s.notifyRepo.MarkAllRead(ctx, s.db, userID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return nil
}
