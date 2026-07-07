package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
)

// ParcelRepo abstracts persistence for the parcels table.
type ParcelRepo interface {
	Create(ctx context.Context, db DBTX, p *models.Parcel) (int64, error)
	FindByID(ctx context.Context, db DBTX, id int64) (*models.Parcel, error)
	FindByTrackingNo(ctx context.Context, db DBTX, trackingNo string, stationID int64) (*models.Parcel, error)
	FindByPickupCode(ctx context.Context, db DBTX, pickupCode string, stationID int64) (*models.Parcel, error)
	ListByFilter(ctx context.Context, db DBTX, filter ParcelFilter) ([]*models.Parcel, int64, error)
	UpdateStatus(ctx context.Context, db DBTX, id int64, status int8) error
	UpdateStatusWithTime(ctx context.Context, db DBTX, id int64, status int8, timeField string) error
}

// ShelfRepo provides a minimal read/write interface for shelf assignment
// during parcel intake. The full shelf module is Phase 6.
type ShelfRepo interface {
	FindBestForStation(ctx context.Context, db DBTX, stationID int64) (*ShelfLayout, error)
	IncrementCapacity(ctx context.Context, db DBTX, id, version int64) error
	DecrementCapacity(ctx context.Context, db DBTX, id, version int64) error
}

// ShelfLayout maps a row from the shelf_layout table (minimal).
type ShelfLayout struct {
	ID              int64  `db:"id"`
	ShelfCode       string `db:"shelf_code"`
	CurrentCapacity int    `db:"current_capacity"`
	MaxCapacity     int    `db:"max_capacity"`
	Version         int64  `db:"version"`
}

// ParcelFilter holds optional filters for listing parcels.
type ParcelFilter struct {
	StationID      int64
	TrackingNo     string
	ReceiverPhone  string
	Status         *int8 // nil means no filter
	CourierCompany string
	ShelfCode      string
	ReceiverUserID int64 // for "my parcels" (ignored if 0)
	StorageStart   string // ISO datetime string, nullable
	StorageEnd     string
	Offset         int
	Limit          int
}

// mysqlParcelRepo implements ParcelRepo.
type mysqlParcelRepo struct{}

func NewParcelRepo() ParcelRepo { return &mysqlParcelRepo{} }

func (r *mysqlParcelRepo) Create(ctx context.Context, db DBTX, p *models.Parcel) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO parcels
		   (station_id, tracking_no, courier_company, shelf_code, pickup_code,
		    receiver_phone, receiver_user_id, receiver_name, weight, is_fragile,
		    remarks, status, storage_time, operator_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.StationID, p.TrackingNo, p.CourierCompany, p.ShelfCode,
		p.PickupCode, p.ReceiverPhone, p.ReceiverUserID, p.ReceiverName,
		p.Weight, p.IsFragile, p.Remarks, p.Status, p.StorageTime, p.OperatorID)
	if err != nil {
		return 0, fmt.Errorf("parcel_repo.Create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("parcel_repo.Create LastInsertId: %w", err)
	}
	return id, nil
}

func (r *mysqlParcelRepo) FindByID(ctx context.Context, db DBTX, id int64) (*models.Parcel, error) {
	var p models.Parcel
	err := db.GetContext(ctx, &p,
		`SELECT id, station_id, tracking_no, courier_company, shelf_code,
		        pickup_code, receiver_phone, receiver_user_id, receiver_name,
		        weight, is_fragile, remarks, status, storage_time, pickup_time,
		        return_time, last_notify_time, notify_count, operator_id, updated_at
		 FROM parcels WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("parcel_repo.FindByID: %w", err)
	}
	return &p, nil
}

func (r *mysqlParcelRepo) FindByTrackingNo(ctx context.Context, db DBTX, trackingNo string, stationID int64) (*models.Parcel, error) {
	var p models.Parcel
	err := db.GetContext(ctx, &p,
		`SELECT id, station_id, tracking_no, courier_company, shelf_code,
		        pickup_code, receiver_phone, receiver_user_id, receiver_name,
		        weight, is_fragile, remarks, status, storage_time, pickup_time,
		        return_time, last_notify_time, notify_count, operator_id, updated_at
		 FROM parcels WHERE tracking_no = ? AND station_id = ?`, trackingNo, stationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("parcel_repo.FindByTrackingNo: %w", err)
	}
	return &p, nil
}

func (r *mysqlParcelRepo) FindByPickupCode(ctx context.Context, db DBTX, pickupCode string, stationID int64) (*models.Parcel, error) {
	var p models.Parcel
	err := db.GetContext(ctx, &p,
		`SELECT id, station_id, tracking_no, courier_company, shelf_code,
		        pickup_code, receiver_phone, receiver_user_id, receiver_name,
		        weight, is_fragile, remarks, status, storage_time, pickup_time,
		        return_time, last_notify_time, notify_count, operator_id, updated_at
		 FROM parcels WHERE pickup_code = ? AND station_id = ? AND status = ?`,
		pickupCode, stationID, models.ParcelStatusPending)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("parcel_repo.FindByPickupCode: %w", err)
	}
	return &p, nil
}

func (r *mysqlParcelRepo) ListByFilter(ctx context.Context, db DBTX, filter ParcelFilter) ([]*models.Parcel, int64, error) {
	where, args := buildParcelWhere(filter)

	countQ := "SELECT COUNT(*) FROM parcels"
	if where != "" {
		countQ += " WHERE " + where
	}
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("parcel_repo.ListByFilter count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	listQ := `SELECT id, station_id, tracking_no, courier_company, shelf_code,
	                 pickup_code, receiver_phone, receiver_user_id, receiver_name,
	                 weight, is_fragile, remarks, status, storage_time, pickup_time,
	                 return_time, last_notify_time, notify_count, operator_id, updated_at
	          FROM parcels`
	if where != "" {
		listQ += " WHERE " + where
	}
	listQ += " ORDER BY storage_time DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, filter.Offset)

	var parcels []*models.Parcel
	if err := db.SelectContext(ctx, &parcels, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("parcel_repo.ListByFilter list: %w", err)
	}
	return parcels, total, nil
}

func (r *mysqlParcelRepo) UpdateStatus(ctx context.Context, db DBTX, id int64, status int8) error {
	_, err := db.ExecContext(ctx,
		`UPDATE parcels SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("parcel_repo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *mysqlParcelRepo) UpdateStatusWithTime(ctx context.Context, db DBTX, id int64, status int8, timeField string) error {
	var query string
	switch timeField {
	case "pickup_time":
		query = `UPDATE parcels SET status = ?, pickup_time = NOW() WHERE id = ?`
	case "return_time":
		query = `UPDATE parcels SET status = ?, return_time = NOW() WHERE id = ?`
	default:
		query = `UPDATE parcels SET status = ? WHERE id = ?`
	}
	_, err := db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("parcel_repo.UpdateStatusWithTime: %w", err)
	}
	return nil
}

// buildParcelWhere constructs WHERE clause for ListByFilter.
func buildParcelWhere(filter ParcelFilter) (string, []any) {
	var conds []string
	var args []any

	if filter.StationID > 0 {
		conds = append(conds, "station_id = ?")
		args = append(args, filter.StationID)
	}
	if kw := strings.TrimSpace(filter.TrackingNo); kw != "" {
		conds = append(conds, "tracking_no LIKE CONCAT('%', ?, '%')")
		args = append(args, kw)
	}
	if kw := strings.TrimSpace(filter.ReceiverPhone); kw != "" {
		conds = append(conds, "receiver_phone LIKE CONCAT('%', ?, '%')")
		args = append(args, kw)
	}
	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, *filter.Status)
	}
	if kw := strings.TrimSpace(filter.CourierCompany); kw != "" {
		conds = append(conds, "courier_company = ?")
		args = append(args, kw)
	}
	if kw := strings.TrimSpace(filter.ShelfCode); kw != "" {
		conds = append(conds, "shelf_code = ?")
		args = append(args, kw)
	}
	if filter.ReceiverUserID > 0 {
		conds = append(conds, "receiver_user_id = ?")
		args = append(args, filter.ReceiverUserID)
	}
	if s := strings.TrimSpace(filter.StorageStart); s != "" {
		conds = append(conds, "storage_time >= ?")
		args = append(args, s)
	}
	if s := strings.TrimSpace(filter.StorageEnd); s != "" {
		conds = append(conds, "storage_time <= ?")
		args = append(args, s)
	}

	if len(conds) == 0 {
		return "", nil
	}
	return strings.Join(conds, " AND "), args
}

// mysqlShelfRepo implements ShelfRepo for parcel intake.
type mysqlShelfRepo struct{}

func NewShelfRepo() ShelfRepo { return &mysqlShelfRepo{} }

func (r *mysqlShelfRepo) FindBestForStation(ctx context.Context, db DBTX, stationID int64) (*ShelfLayout, error) {
	var s ShelfLayout
	err := db.GetContext(ctx, &s,
		`SELECT id, shelf_code, current_capacity, max_capacity, version
		 FROM shelf_layout
		 WHERE station_id = ? AND current_capacity < max_capacity
		 ORDER BY (current_capacity / max_capacity) ASC
		 LIMIT 1`, stationID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("shelf_repo.FindBestForStation: %w", err)
	}
	return &s, nil
}

func (r *mysqlShelfRepo) IncrementCapacity(ctx context.Context, db DBTX, id, version int64) error {
	res, err := db.ExecContext(ctx,
		`UPDATE shelf_layout
		 SET current_capacity = current_capacity + 1, version = version + 1
		 WHERE id = ? AND version = ?`, id, version)
	if err != nil {
		return fmt.Errorf("shelf_repo.IncrementCapacity: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("shelf_repo.IncrementCapacity: optimistic lock conflict")
	}
	return nil
}

func (r *mysqlShelfRepo) DecrementCapacity(ctx context.Context, db DBTX, id, version int64) error {
	res, err := db.ExecContext(ctx,
		`UPDATE shelf_layout
		 SET current_capacity = current_capacity - 1, version = version + 1
		 WHERE id = ? AND version = ? AND current_capacity > 0`, id, version)
	if err != nil {
		return fmt.Errorf("shelf_repo.DecrementCapacity: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("shelf_repo.DecrementCapacity: optimistic lock conflict")
	}
	return nil
}

var (
	_ DBTX = (*sqlx.DB)(nil)
	_ DBTX = (*sqlx.Tx)(nil)
)
