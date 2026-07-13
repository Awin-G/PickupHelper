package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"pickup-helper/internal/models"
)

// StationRepo abstracts persistence for the stations table.
type StationRepo interface {
	List(ctx context.Context, db DBTX, filter StationFilter) ([]*models.Station, int64, error)
	FindByID(ctx context.Context, db DBTX, id int64) (*models.Station, error)
	Create(ctx context.Context, db DBTX, s *models.Station) (int64, error)
	Update(ctx context.Context, db DBTX, id int64, cols []string, args []any) error
}

// StationFilter holds optional filters for listing stations.
type StationFilter struct {
	Keyword string
	Status  *int8
	Offset  int
	Limit   int
}

type mysqlStationRepo struct{}

func NewStationRepo() StationRepo { return &mysqlStationRepo{} }

func (r *mysqlStationRepo) List(ctx context.Context, db DBTX, filter StationFilter) ([]*models.Station, int64, error) {
	where, args := buildStationWhere(filter)
	countQ := "SELECT COUNT(*) FROM stations"
	if where != "" {
		countQ += " WHERE " + where
	}
	var total int64
	if err := db.GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("station_repo.List count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	listQ := `SELECT id, name, address, latitude, longitude, business_hours, status, created_at, updated_at
	          FROM stations`
	if where != "" {
		listQ += " WHERE " + where
	}
	listQ += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, limit, filter.Offset)

	var out []*models.Station
	if err := db.SelectContext(ctx, &out, listQ, listArgs...); err != nil {
		return nil, 0, fmt.Errorf("station_repo.List list: %w", err)
	}
	return out, total, nil
}

func (r *mysqlStationRepo) FindByID(ctx context.Context, db DBTX, id int64) (*models.Station, error) {
	var s models.Station
	err := db.GetContext(ctx, &s,
		`SELECT id, name, address, latitude, longitude, business_hours, status, created_at, updated_at
		 FROM stations WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("station_repo.FindByID: %w", err)
	}
	return &s, nil
}

func (r *mysqlStationRepo) Create(ctx context.Context, db DBTX, s *models.Station) (int64, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO stations (name, address, latitude, longitude, business_hours, status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.Name, s.Address, s.Latitude, s.Longitude, s.BusinessHours, s.Status)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return 0, fmt.Errorf("station_repo.Create: %w", err)
		}
		return 0, fmt.Errorf("station_repo.Create: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("station_repo.Create LastInsertId: %w", err)
	}
	return id, nil
}

func (r *mysqlStationRepo) Update(ctx context.Context, db DBTX, id int64, cols []string, args []any) error {
	if len(cols) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString("UPDATE stations SET ")
	for i, col := range cols {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(col)
		sb.WriteString(" = ?")
	}
	sb.WriteString(" WHERE id = ?")
	args = append(args, id)
	if _, err := db.ExecContext(ctx, sb.String(), args...); err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return fmt.Errorf("station_repo.Update: %w", err)
		}
		return fmt.Errorf("station_repo.Update: %w", err)
	}
	return nil
}

func buildStationWhere(filter StationFilter) (string, []any) {
	var conds []string
	var args []any
	if kw := strings.TrimSpace(filter.Keyword); kw != "" {
		like := "%" + kw + "%"
		conds = append(conds, "(name LIKE ? OR address LIKE ?)")
		args = append(args, like, like)
	}
	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, *filter.Status)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return strings.Join(conds, " AND "), args
}
