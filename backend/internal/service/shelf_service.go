package service

import (
	"context"
	"database/sql"
	stderrors "errors"
	"math"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

// ShelfDTO is the API-facing shelf representation.
type ShelfDTO struct {
	ID              int64  `json:"id"`
	StationID       int64  `json:"station_id"`
	ShelfCode       string `json:"shelf_code"`
	RowNum          int    `json:"row_num"`
	ColNum          int    `json:"col_num"`
	CurrentCapacity int    `json:"current_capacity"`
	MaxCapacity     int    `json:"max_capacity"`
	OccupancyRate   float64 `json:"occupancy_rate"`
	HeatLevel       int    `json:"heat_level"`
	Remark          string `json:"remark,omitempty"`
}

// ShelfListResult wraps the paged shelf list.
type ShelfListResult struct {
	Items []*ShelfDTO `json:"items"`
	Total int64       `json:"total"`
}

// OccupancyResult is the response for GET /shelves/occupancy.
type OccupancyResult struct {
	StationID int64       `json:"station_id"`
	Shelves   []*ShelfDTO `json:"shelves"`
	TotalUsed int         `json:"total_used"`
	TotalMax  int         `json:"total_max"`
}

// ShelfService implements shelf management.
type ShelfService struct {
	shelfRepo repository.ShelfRepo
	db        *sqlx.DB
}

func NewShelfService(sr repository.ShelfRepo, db *sqlx.DB) *ShelfService {
	return &ShelfService{shelfRepo: sr, db: db}
}

// ListShelf returns a paginated list of shelves for a station.
func (s *ShelfService) ListShelf(ctx context.Context, stationID int64, offset, limit int) (*ShelfListResult, error) {
	// We delegate to the existing shelf repo and wrap.
	type shelfRow struct {
		ID              int64   `db:"id"`
		StationID       int64   `db:"station_id"`
		ShelfCode       string  `db:"shelf_code"`
		RowNum          int     `db:"row_num"`
		ColNum          int     `db:"col_num"`
		CurrentCapacity int     `db:"current_capacity"`
		MaxCapacity     int     `db:"max_capacity"`
		Version         int64   `db:"version"`
		Remark          sql.NullString `db:"remark"`
	}

	var total int64
	if err := s.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM shelf_layout WHERE station_id = ?", stationID); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	if limit <= 0 {
		limit = 20
	}
	var rows []shelfRow
	if err := s.db.SelectContext(ctx, &rows,
		`SELECT id, station_id, shelf_code, row_num, col_num, current_capacity,
		        max_capacity, version, remark
		 FROM shelf_layout WHERE station_id = ?
		 ORDER BY shelf_code LIMIT ? OFFSET ?`,
		stationID, limit, offset); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	items := make([]*ShelfDTO, 0, len(rows))
	for _, r := range rows {
		dto := &ShelfDTO{
			ID: r.ID, StationID: r.StationID, ShelfCode: r.ShelfCode,
			RowNum: r.RowNum, ColNum: r.ColNum,
			CurrentCapacity: r.CurrentCapacity, MaxCapacity: r.MaxCapacity,
		}
		if r.MaxCapacity > 0 {
			dto.OccupancyRate = math.Round(float64(r.CurrentCapacity)/float64(r.MaxCapacity)*100) / 100
		}
		if r.Remark.Valid {
			dto.Remark = r.Remark.String
		}
		items = append(items, dto)
	}
	return &ShelfListResult{Items: items, Total: total}, nil
}

// CreateShelf adds a new shelf to a station.
func (s *ShelfService) CreateShelf(ctx context.Context, stationID int64, shelfCode string, rowNum, colNum, maxCapacity int, remark string) (*ShelfDTO, error) {
	if rowNum < 1 || rowNum > 99 || colNum < 1 || colNum > 99 || maxCapacity < 1 || maxCapacity > 9999 {
		return nil, apperrors.New(apperrors.ErrShelfCapacityInvalid, "")
	}

	// Check duplicate shelf code in station.
	var count int64
	if err := s.db.GetContext(ctx, &count,
		"SELECT COUNT(*) FROM shelf_layout WHERE station_id = ? AND shelf_code = ?",
		stationID, shelfCode); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if count > 0 {
		return nil, apperrors.New(apperrors.ErrShelfCodeExists, "")
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO shelf_layout (station_id, shelf_code, row_num, col_num, max_capacity, remark)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		stationID, shelfCode, rowNum, colNum, maxCapacity,
		sql.NullString{String: remark, Valid: remark != ""})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	id, _ := res.LastInsertId()

	return &ShelfDTO{
		ID: id, StationID: stationID, ShelfCode: shelfCode,
		RowNum: rowNum, ColNum: colNum, MaxCapacity: maxCapacity, Remark: remark,
	}, nil
}

// UpdateShelf updates shelf configuration.
func (s *ShelfService) UpdateShelf(ctx context.Context, shelfID int64, shelfCode string, rowNum, colNum, maxCapacity int, remark string) (*ShelfDTO, error) {
	type shelfRow struct {
		ID              int64           `db:"id"`
		StationID       int64           `db:"station_id"`
		ShelfCode       string          `db:"shelf_code"`
		RowNum          int             `db:"row_num"`
		ColNum          int             `db:"col_num"`
		CurrentCapacity int             `db:"current_capacity"`
		MaxCapacity     int             `db:"max_capacity"`
		Remark          sql.NullString  `db:"remark"`
	}

	var row shelfRow
	err := s.db.GetContext(ctx, &row,
		"SELECT id, station_id, shelf_code, row_num, col_num, current_capacity, max_capacity, remark FROM shelf_layout WHERE id = ?", shelfID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrShelfNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	if maxCapacity > 0 && maxCapacity < row.CurrentCapacity {
		return nil, apperrors.New(apperrors.ErrShelfMaxBelowCurrent, "")
	}

	// Check shelf code uniqueness if changed.
	if shelfCode != "" && shelfCode != row.ShelfCode {
		var count int64
		if err := s.db.GetContext(ctx, &count,
			"SELECT COUNT(*) FROM shelf_layout WHERE station_id = ? AND shelf_code = ? AND id != ?",
			row.StationID, shelfCode, shelfID); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
		}
		if count > 0 {
			return nil, apperrors.New(apperrors.ErrShelfCodeUsed, "")
		}
		row.ShelfCode = shelfCode
	}
	if rowNum > 0 {
		row.RowNum = rowNum
	}
	if colNum > 0 {
		row.ColNum = colNum
	}
	if maxCapacity > 0 {
		row.MaxCapacity = maxCapacity
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE shelf_layout SET shelf_code = ?, row_num = ?, col_num = ?, max_capacity = ?, remark = ?
		 WHERE id = ?`,
		row.ShelfCode, row.RowNum, row.ColNum, row.MaxCapacity,
		sql.NullString{String: remark, Valid: remark != ""}, shelfID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	dto := &ShelfDTO{
		ID: row.ID, StationID: row.StationID, ShelfCode: row.ShelfCode,
		RowNum: row.RowNum, ColNum: row.ColNum,
		CurrentCapacity: row.CurrentCapacity, MaxCapacity: row.MaxCapacity,
	}
	if row.MaxCapacity > 0 {
		dto.OccupancyRate = math.Round(float64(row.CurrentCapacity)/float64(row.MaxCapacity)*100) / 100
	}
	if remark != "" {
		dto.Remark = remark
	}
	return dto, nil
}

// Occupancy returns the heatmap data for a station.
func (s *ShelfService) Occupancy(ctx context.Context, stationID int64) (*OccupancyResult, error) {
	type shelfRow struct {
		ID              int64  `db:"id"`
		StationID       int64  `db:"station_id"`
		ShelfCode       string `db:"shelf_code"`
		RowNum          int    `db:"row_num"`
		ColNum          int    `db:"col_num"`
		CurrentCapacity int    `db:"current_capacity"`
		MaxCapacity     int    `db:"max_capacity"`
	}

	var rows []shelfRow
	if err := s.db.SelectContext(ctx, &rows,
		`SELECT id, station_id, shelf_code, row_num, col_num, current_capacity, max_capacity
		 FROM shelf_layout WHERE station_id = ?
		 ORDER BY shelf_code`, stationID); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	totalUsed := 0
	totalMax := 0
	shelves := make([]*ShelfDTO, 0, len(rows))
	for _, r := range rows {
		rate := 0.0
		if r.MaxCapacity > 0 {
			rate = math.Round(float64(r.CurrentCapacity)/float64(r.MaxCapacity)*100) / 100
		}
		heat := 0
		switch {
		case rate >= 0.9:
			heat = 4
		case rate >= 0.7:
			heat = 3
		case rate >= 0.5:
			heat = 2
		case rate >= 0.1:
			heat = 1
		}
		shelves = append(shelves, &ShelfDTO{
			ID: r.ID, StationID: r.StationID, ShelfCode: r.ShelfCode,
			RowNum: r.RowNum, ColNum: r.ColNum,
			CurrentCapacity: r.CurrentCapacity, MaxCapacity: r.MaxCapacity,
			OccupancyRate: rate, HeatLevel: heat,
		})
		totalUsed += r.CurrentCapacity
		totalMax += r.MaxCapacity
	}
	return &OccupancyResult{StationID: stationID, Shelves: shelves, TotalUsed: totalUsed, TotalMax: totalMax}, nil
}
