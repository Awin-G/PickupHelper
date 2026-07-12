package service

import (
	"context"
	"database/sql"
	"testing"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStationList_Success(t *testing.T) {
	sr := &mockStationRepo{ListFn: func(_ context.Context, _ repository.DBTX, _ repository.StationFilter) ([]*models.Station, int64, error) {
		return []*models.Station{{ID: 1, Name: "站A", Address: "地址1", Status: 1}}, 1, nil
	}}
	svc := NewStationService(sr, nil)
	res, err := svc.ListStations(context.Background(), StationListFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.Total)
	require.Len(t, res.Items, 1)
	assert.Equal(t, "站A", res.Items[0].Name)
}

func TestStationList_DefaultPage(t *testing.T) {
	var captured repository.StationFilter
	sr := &mockStationRepo{ListFn: func(_ context.Context, _ repository.DBTX, f repository.StationFilter) ([]*models.Station, int64, error) {
		captured = f
		return nil, 0, nil
	}}
	svc := NewStationService(sr, nil)
	_, err := svc.ListStations(context.Background(), StationListFilter{})
	require.NoError(t, err)
	assert.Equal(t, 20, captured.Limit)
	assert.Equal(t, 0, captured.Offset)
}

func TestStationGet_Success(t *testing.T) {
	sr := &mockStationRepo{FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.Station, error) {
		return &models.Station{ID: id, Name: "站B", Address: "地址2", Status: 1}, nil
	}}
	svc := NewStationService(sr, nil)
	dto, err := svc.GetStation(context.Background(), 5)
	require.NoError(t, err)
	assert.Equal(t, int64(5), dto.ID)
	assert.Equal(t, "站B", dto.Name)
}

func TestStationGet_NotFound(t *testing.T) {
	svc := NewStationService(&mockStationRepo{}, nil)
	_, err := svc.GetStation(context.Background(), 999)
	requireAppErr(t, err, apperrors.ErrNotFound)
}

func TestStationCreate_Success(t *testing.T) {
	sr := &mockStationRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.Station, error) {
			return &models.Station{ID: id, Name: "站C", Address: "addr", Latitude: 30.5, Longitude: 114.3, Status: 1}, nil
		},
	}
	svc := NewStationService(sr, nil)
	dto, err := svc.CreateStation(context.Background(), CreateStationRequest{
		Name: "站C", Address: "addr", Latitude: 30.5, Longitude: 114.3,
	})
	require.NoError(t, err)
	assert.Equal(t, "站C", dto.Name)
}

func TestStationCreate_EmptyName(t *testing.T) {
	svc := NewStationService(&mockStationRepo{}, nil)
	_, err := svc.CreateStation(context.Background(), CreateStationRequest{Address: "addr"})
	requireAppErr(t, err, apperrors.ErrInvalidParam)
}

func TestStationCreate_Duplicate(t *testing.T) {
	sr := &mockStationRepo{CreateFn: func(_ context.Context, _ repository.DBTX, _ *models.Station) (int64, error) {
		return 0, sql.ErrNoRows // Simulate duplicate via wrapping; real impl uses Contains
	}}
	_ = sr
	// Duplicate detection relies on error string matching. Mock by setting CreateFn to return a wrapped error.
	sr2 := &mockStationRepo{CreateFn: func(_ context.Context, _ repository.DBTX, _ *models.Station) (int64, error) {
		return 0, apperrors.Wrap(assert.AnError, apperrors.ErrConflict, "Duplicate")
	}}
	svc := NewStationService(sr2, nil)
	_, err := svc.CreateStation(context.Background(), CreateStationRequest{Name: "dup", Address: "addr"})
	require.Error(t, err)
}

func TestStationUpdate_Success(t *testing.T) {
	current := &models.Station{ID: 1, Name: "旧名", Address: "旧地址", Status: 1}
	sr := &mockStationRepo{
		FindByIDFn: func(_ context.Context, _ repository.DBTX, id int64) (*models.Station, error) {
			return current, nil
		},
		UpdateFn: func(_ context.Context, _ repository.DBTX, _ int64, cols []string, args []any) error {
			for i, col := range cols {
				switch col {
				case "name":
					current.Name = args[i].(string)
				}
			}
			return nil
		},
	}
	svc := NewStationService(sr, nil)
	newName := "新名"
	dto, err := svc.UpdateStation(context.Background(), 1, UpdateStationRequest{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "新名", dto.Name)
}

func TestStationUpdate_NotFound(t *testing.T) {
	svc := NewStationService(&mockStationRepo{}, nil)
	name := "x"
	_, err := svc.UpdateStation(context.Background(), 999, UpdateStationRequest{Name: &name})
	requireAppErr(t, err, apperrors.ErrNotFound)
}
