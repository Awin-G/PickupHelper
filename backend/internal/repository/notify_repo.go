package repository

import (
	"context"
	"fmt"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
)

// NotifyRepo abstracts persistence for the notifications table.
type NotifyRepo interface {
	Create(ctx context.Context, db DBTX, n *models.Notification) (int64, error)
	ListByUser(ctx context.Context, db DBTX, userID int64, offset, limit int) ([]*models.Notification, int64, error)
	MarkAllRead(ctx context.Context, db DBTX, userID int64) error
}

type mysqlNotifyRepo struct{}

func NewNotifyRepo() NotifyRepo { return &mysqlNotifyRepo{} }

func (r *mysqlNotifyRepo) Create(ctx context.Context, db DBTX, n *models.Notification) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, parcel_id, title, content, type, is_read, send_status, channel)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.UserID, n.ParcelID, n.Title, n.Content, n.Type, n.IsRead, n.SendStatus, n.Channel)
	if err != nil {
		return 0, fmt.Errorf("notify_repo.Create: %w", err)
	}
	return res.LastInsertId()
}

func (r *mysqlNotifyRepo) ListByUser(ctx context.Context, db DBTX, userID int64, offset, limit int) ([]*models.Notification, int64, error) {
	var total int64
	if err := db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM notifications WHERE user_id = ?", userID); err != nil {
		return nil, 0, fmt.Errorf("notify_repo.ListByUser count: %w", err)
	}
	if limit <= 0 {
		limit = 20
	}
	var list []*models.Notification
	if err := db.SelectContext(ctx, &list,
		`SELECT id, user_id, parcel_id, title, content, type, is_read, send_status, channel, created_at
		 FROM notifications WHERE user_id = ?
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("notify_repo.ListByUser list: %w", err)
	}
	return list, total, nil
}

func (r *mysqlNotifyRepo) MarkAllRead(ctx context.Context, db DBTX, userID int64) error {
	_, err := db.ExecContext(ctx,
		"UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0", userID)
	if err != nil {
		return fmt.Errorf("notify_repo.MarkAllRead: %w", err)
	}
	return nil
}

var (
	_ DBTX = (*sqlx.DB)(nil)
	_ DBTX = (*sqlx.Tx)(nil)
)
