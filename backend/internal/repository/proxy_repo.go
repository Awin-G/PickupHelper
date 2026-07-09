package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
)

// ProxyOrderRepo abstracts persistence for the proxy_orders table.
type ProxyOrderRepo interface {
	Create(ctx context.Context, db DBTX, o *models.ProxyOrder) (int64, error)
	FindByID(ctx context.Context, db DBTX, id int64) (*models.ProxyOrder, error)
	FindByParcelID(ctx context.Context, db DBTX, parcelID int64, status ...int8) (*models.ProxyOrder, error)
	ListTasks(ctx context.Context, db DBTX, filter ProxyTaskFilter) ([]*models.ProxyOrder, int64, error)
	ListMyOrders(ctx context.Context, db DBTX, filter ProxyMyOrderFilter) ([]*models.ProxyOrder, int64, error)
	UpdateAccepted(ctx context.Context, db DBTX, id, takerID int64, code string) (int64, error)
	UpdateDelivery(ctx context.Context, db DBTX, id int64, status int8) error
	UpdateConfirm(ctx context.Context, db DBTX, id int64, status int8) error
	Cancel(ctx context.Context, db DBTX, id int64, status int8, reason string) error
}

// ProxyTaskFilter holds filters for the task hall list.
type ProxyTaskFilter struct {
	StationID int64
	MinReward float64
	Sort      string // "reward_amount", "-reward_amount", "deadline"
	Offset    int
	Limit     int
}

// ProxyMyOrderFilter holds filters for "my orders" list.
type ProxyMyOrderFilter struct {
	UserID    int64
	Role      string // "publisher" / "taker" / ""
	Status    *int8
	Offset    int
	Limit     int
}

type mysqlProxyOrderRepo struct{}

func NewProxyOrderRepo() ProxyOrderRepo { return &mysqlProxyOrderRepo{} }

func (r *mysqlProxyOrderRepo) Create(ctx context.Context, db DBTX, o *models.ProxyOrder) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO proxy_orders
		   (station_id, parcel_id, publisher_id, reward_amount, deadline, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		o.StationID, o.ParcelID, o.PublisherID, o.RewardAmount, o.Deadline, o.Status)
	if err != nil {
		return 0, fmt.Errorf("proxy_repo.Create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("proxy_repo.Create LastInsertId: %w", err)
	}
	return id, nil
}

func (r *mysqlProxyOrderRepo) FindByID(ctx context.Context, db DBTX, id int64) (*models.ProxyOrder, error) {
	var o models.ProxyOrder
	err := db.GetContext(ctx, &o,
		`SELECT id, station_id, parcel_id, publisher_id, taker_id,
		        reward_amount, temp_pickup_code, deadline, status,
		        cancel_reason, delivery_time, created_at, updated_at
		 FROM proxy_orders WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("proxy_repo.FindByID: %w", err)
	}
	return &o, nil
}

func (r *mysqlProxyOrderRepo) FindByParcelID(ctx context.Context, db DBTX, parcelID int64, status ...int8) (*models.ProxyOrder, error) {
	query := `SELECT id, station_id, parcel_id, publisher_id, taker_id,
	                  reward_amount, temp_pickup_code, deadline, status,
	                  cancel_reason, delivery_time, created_at, updated_at
	           FROM proxy_orders WHERE parcel_id = ?`
	args := []any{parcelID}
	if len(status) > 0 {
		query += " AND status IN (" + placeholders(len(status)) + ")"
		for _, s := range status {
			args = append(args, s)
		}
	}
	var o models.ProxyOrder
	err := db.GetContext(ctx, &o, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("proxy_repo.FindByParcelID: %w", err)
	}
	return &o, nil
}

func (r *mysqlProxyOrderRepo) ListTasks(ctx context.Context, db DBTX, filter ProxyTaskFilter) ([]*models.ProxyOrder, int64, error) {
	var conds []string
	var args []any
	conds = append(conds, "status = ?")
	args = append(args, models.ProxyStatusPending)
	if filter.StationID > 0 {
		conds = append(conds, "station_id = ?")
		args = append(args, filter.StationID)
	}
	if filter.MinReward > 0 {
		conds = append(conds, "reward_amount >= ?")
		args = append(args, filter.MinReward)
	}
	where := joinCond(conds)

	countQ := "SELECT COUNT(*) FROM proxy_orders WHERE " + where
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("proxy_repo.ListTasks count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	orderBy := "created_at DESC"
	switch filter.Sort {
	case "reward_amount":
		orderBy = "reward_amount ASC"
	case "-reward_amount":
		orderBy = "reward_amount DESC"
	case "deadline":
		orderBy = "deadline ASC"
	}

	listQ := `SELECT id, station_id, parcel_id, publisher_id, taker_id,
	                 reward_amount, temp_pickup_code, deadline, status,
	                 cancel_reason, delivery_time, created_at, updated_at
	          FROM proxy_orders WHERE ` + where +
		` ORDER BY ` + orderBy + ` LIMIT ? OFFSET ?`
	listArgs := append(args, limit, filter.Offset)

	var orders []*models.ProxyOrder
	if err := db.SelectContext(ctx, &orders, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("proxy_repo.ListTasks list: %w", err)
	}
	return orders, total, nil
}

func (r *mysqlProxyOrderRepo) ListMyOrders(ctx context.Context, db DBTX, filter ProxyMyOrderFilter) ([]*models.ProxyOrder, int64, error) {
	var conds []string
	var args []any
	if filter.UserID > 0 {
		if filter.Role == "publisher" {
			conds = append(conds, "publisher_id = ?")
			args = append(args, filter.UserID)
		} else if filter.Role == "taker" {
			conds = append(conds, "taker_id = ?")
			args = append(args, filter.UserID)
		} else {
			conds = append(conds, "(publisher_id = ? OR taker_id = ?)")
			args = append(args, filter.UserID, filter.UserID)
		}
	}
	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, *filter.Status)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + joinCond(conds)
	}

	countQ := "SELECT COUNT(*) FROM proxy_orders" + where
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("proxy_repo.ListMyOrders count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	listQ := `SELECT id, station_id, parcel_id, publisher_id, taker_id,
	                 reward_amount, temp_pickup_code, deadline, status,
	                 cancel_reason, delivery_time, created_at, updated_at
	          FROM proxy_orders` + where +
		` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	listArgs := append(args, limit, filter.Offset)

	var orders []*models.ProxyOrder
	if err := db.SelectContext(ctx, &orders, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("proxy_repo.ListMyOrders list: %w", err)
	}
	return orders, total, nil
}

func (r *mysqlProxyOrderRepo) UpdateAccepted(ctx context.Context, db DBTX, id, takerID int64, code string) (int64, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE proxy_orders
		 SET taker_id = ?, status = ?, temp_pickup_code = ?
		 WHERE id = ? AND status = ? AND taker_id IS NULL`,
		takerID, models.ProxyStatusDelivering, code, id, models.ProxyStatusPending)
	if err != nil {
		return 0, fmt.Errorf("proxy_repo.UpdateAccepted: %w", err)
	}
	return res.RowsAffected()
}

func (r *mysqlProxyOrderRepo) UpdateDelivery(ctx context.Context, db DBTX, id int64, status int8) error {
	_, err := db.ExecContext(ctx,
		`UPDATE proxy_orders SET status = ?, delivery_time = NOW()
		 WHERE id = ? AND status = ?`,
		status, id, models.ProxyStatusDelivering)
	if err != nil {
		return fmt.Errorf("proxy_repo.UpdateDelivery: %w", err)
	}
	return nil
}

func (r *mysqlProxyOrderRepo) UpdateConfirm(ctx context.Context, db DBTX, id int64, status int8) error {
	_, err := db.ExecContext(ctx,
		`UPDATE proxy_orders SET status = ? WHERE id = ? AND status = ?`,
		status, id, models.ProxyStatusConfirm)
	if err != nil {
		return fmt.Errorf("proxy_repo.UpdateConfirm: %w", err)
	}
	return nil
}

func (r *mysqlProxyOrderRepo) Cancel(ctx context.Context, db DBTX, id int64, status int8, reason string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE proxy_orders
		 SET status = ?, cancel_reason = ?
		 WHERE id = ? AND status IN (?, ?)`,
		status, reason, id, models.ProxyStatusPending, models.ProxyStatusDelivering)
	if err != nil {
		return fmt.Errorf("proxy_repo.Cancel: %w", err)
	}
	return nil
}

func placeholders(n int) string {
	b := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	return string(b)
}

func joinCond(conds []string) string {
	var result string
	for i, c := range conds {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}

var (
	_ DBTX = (*sqlx.DB)(nil)
	_ DBTX = (*sqlx.Tx)(nil)
)
