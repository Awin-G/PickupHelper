package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
)

// StationListFilter is the service-layer filter for listing stations.
type StationListFilter struct {
	Keyword  string
	Status   *int8
	Page     int
	PageSize int
}

// StationListResult bundles page + total for ListStations.
type StationListResult struct {
	Items []*models.StationDTO `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// CreateStationRequest is the input for creating a station.
type CreateStationRequest struct {
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	BusinessHours string  `json:"business_hours"`
	Status        int8    `json:"status"`
}

// UpdateStationRequest is the input for updating a station.
type UpdateStationRequest struct {
	Name          *string  `json:"name,omitempty"`
	Address       *string  `json:"address,omitempty"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	BusinessHours *string  `json:"business_hours,omitempty"`
	Status        *int8    `json:"status,omitempty"`
}

// StationService implements station CRUD operations.
type StationService struct {
	stationRepo repository.StationRepo
	db          *sqlx.DB
}

func NewStationService(sr repository.StationRepo, db *sqlx.DB) *StationService {
	return &StationService{stationRepo: sr, db: db}
}

// ListStations returns a paginated list of stations.
func (s *StationService) ListStations(ctx context.Context, filter StationListFilter) (*StationListResult, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	repoFilter := repository.StationFilter{
		Keyword: strings.TrimSpace(filter.Keyword),
		Status:  filter.Status,
		Offset:  (page - 1) * pageSize,
		Limit:   pageSize,
	}
	stations, total, err := s.stationRepo.List(ctx, s.db, repoFilter)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*models.StationDTO, 0, len(stations))
	for _, st := range stations {
		items = append(items, st.ToDTO())
	}
	return &StationListResult{Items: items, Total: total, Page: page, Size: pageSize}, nil
}

// GetStation returns a single station.
func (s *StationService) GetStation(ctx context.Context, id int64) (*models.StationDTO, error) {
	st, err := s.stationRepo.FindByID(ctx, s.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrNotFound, "驿站不存在")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return st.ToDTO(), nil
}

// CreateStation creates a new station.
func (s *StationService) CreateStation(ctx context.Context, req CreateStationRequest) (*models.StationDTO, error) {
	if req.Name == "" {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "name 必填")
	}
	if len([]rune(req.Name)) > 100 {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "name 最长 100 字符")
	}
	if req.Address == "" {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "address 必填")
	}
	biz := req.BusinessHours
	if biz == "" {
		biz = "09:00-20:00"
	}
	status := req.Status
	if status != models.StationStatusClosed && status != models.StationStatusOpen {
		status = models.StationStatusOpen
	}

	st := &models.Station{
		Name:          req.Name,
		Address:       req.Address,
		Latitude:      sql.NullFloat64{Float64: req.Latitude, Valid: req.Latitude != 0},
		Longitude:     sql.NullFloat64{Float64: req.Longitude, Valid: req.Longitude != 0},
		BusinessHours: biz,
		Status:        status,
	}
	id, err := s.stationRepo.Create(ctx, s.db, st)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, apperrors.New(apperrors.ErrConflict, "驿站名称已存在")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return s.GetStation(ctx, id)
}

// UpdateStation updates a station.
func (s *StationService) UpdateStation(ctx context.Context, id int64, req UpdateStationRequest) (*models.StationDTO, error) {
	if _, err := s.stationRepo.FindByID(ctx, s.db, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.New(apperrors.ErrNotFound, "驿站不存在")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	var cols []string
	var args []any

	if req.Name != nil {
		if len([]rune(*req.Name)) > 100 {
			return nil, apperrors.New(apperrors.ErrInvalidParam, "name 最长 100 字符")
		}
		cols = append(cols, "name")
		args = append(args, *req.Name)
	}
	if req.Address != nil {
		cols = append(cols, "address")
		args = append(args, *req.Address)
	}
	if req.Latitude != nil {
		cols = append(cols, "latitude")
		args = append(args, *req.Latitude)
	}
	if req.Longitude != nil {
		cols = append(cols, "longitude")
		args = append(args, *req.Longitude)
	}
	if req.BusinessHours != nil {
		cols = append(cols, "business_hours")
		args = append(args, *req.BusinessHours)
	}
	if req.Status != nil {
		cols = append(cols, "status")
		args = append(args, *req.Status)
	}

	if len(cols) == 0 {
		return s.GetStation(ctx, id)
	}

	if err := s.stationRepo.Update(ctx, s.db, id, cols, args); err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, apperrors.New(apperrors.ErrConflict, "驿站名称已存在")
		}
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return s.GetStation(ctx, id)
}
