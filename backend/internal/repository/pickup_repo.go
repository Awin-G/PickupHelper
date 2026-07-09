package repository

import (
	"context"
	"fmt"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
)

// PickupLogRepo abstracts persistence for the pickup_logs table.
type PickupLogRepo interface {
	Create(ctx context.Context, db DBTX, log *models.PickupLog) (int64, error)
	ListByFilter(ctx context.Context, db DBTX, filter PickupLogFilter) ([]*models.PickupLog, int64, error)
}

// PickupLogFilter holds optional filters for listing pickup logs.
type PickupLogFilter struct {
	ParcelID     *int64
	OperatorID   *int64
	OperatorType *int8
	Start        string
	End          string
	Offset       int
	Limit        int
}

type mysqlPickupLogRepo struct{}

func NewPickupLogRepo() PickupLogRepo { return &mysqlPickupLogRepo{} }

func (r *mysqlPickupLogRepo) Create(ctx context.Context, db DBTX, log *models.PickupLog) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO pickup_logs
		   (parcel_id, operator_id, operator_type, verification_method,
		    location_lat, location_lng, ip_address, user_agent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ParcelID, log.OperatorID, log.OperatorType, log.VerificationMethod,
		log.LocationLat, log.LocationLng, log.IPAddress, log.UserAgent)
	if err != nil {
		return 0, fmt.Errorf("pickup_log_repo.Create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("pickup_log_repo.Create LastInsertId: %w", err)
	}
	return id, nil
}

func (r *mysqlPickupLogRepo) ListByFilter(ctx context.Context, db DBTX, filter PickupLogFilter) ([]*models.PickupLog, int64, error) {
	where, args := buildPickupLogWhere(filter)

	countQ := "SELECT COUNT(*) FROM pickup_logs"
	if where != "" {
		countQ += " WHERE " + where
	}
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("pickup_log_repo.ListByFilter count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	listQ := `SELECT id, parcel_id, operator_id, operator_type, verification_method,
	                 location_lat, location_lng, ip_address, user_agent, created_at
	          FROM pickup_logs`
	if where != "" {
		listQ += " WHERE " + where
	}
	listQ += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, filter.Offset)

	var logs []*models.PickupLog
	if err := db.SelectContext(ctx, &logs, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("pickup_log_repo.ListByFilter list: %w", err)
	}
	return logs, total, nil
}

func buildPickupLogWhere(filter PickupLogFilter) (string, []any) {
	var conds []string
	var args []any
	if filter.ParcelID != nil {
		conds = append(conds, "parcel_id = ?")
		args = append(args, *filter.ParcelID)
	}
	if filter.OperatorID != nil {
		conds = append(conds, "operator_id = ?")
		args = append(args, *filter.OperatorID)
	}
	if filter.OperatorType != nil {
		conds = append(conds, "operator_type = ?")
		args = append(args, *filter.OperatorType)
	}
	if filter.Start != "" {
		conds = append(conds, "created_at >= ?")
		args = append(args, filter.Start)
	}
	if filter.End != "" {
		conds = append(conds, "created_at <= ?")
		args = append(args, filter.End)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return joinConds(conds), args
}

// joinConds joins condition fragments with " AND ".
func joinConds(conds []string) string {
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
