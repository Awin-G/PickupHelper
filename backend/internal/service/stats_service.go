package service

import (
	"context"

	apperrors "pickup-helper/internal/errors"

	"github.com/jmoiron/sqlx"
)

// DashboardResult is the response for GET /stats/dashboard.
type DashboardResult struct {
	TodayInbound    int64   `json:"today_inbound"`
	TodayOutbound   int64   `json:"today_outbound"`
	PendingCount    int64   `json:"pending_count"`
	DelayedCount    int64   `json:"delayed_count"`
	AbnormalCount   int64   `json:"abnormal_count"`
	ProxyActive     int64   `json:"proxy_active"`
	ShelfUsageRate  float64 `json:"shelf_usage_rate"`
}

// TrendPoint is a single data point for GET /stats/trend.
type TrendPoint struct {
	Date     string `json:"date"`
	Inbound  int64  `json:"inbound"`
	Outbound int64  `json:"outbound"`
	Delayed  int64  `json:"delayed"`
}

// TrendResult wraps trend data.
type TrendResult struct {
	Granularity string        `json:"granularity"`
	Points      []*TrendPoint `json:"points"`
}

// ProxyFinanceResult is the response for GET /stats/proxy-finance.
type ProxyFinanceResult struct {
	TotalOrders    int64   `json:"total_orders"`
	CompletedOrders int64  `json:"completed_orders"`
	TotalAmount    float64 `json:"total_amount"`
	AvgAmount      float64 `json:"avg_amount"`
}

// StatsService implements dashboard, trend, and proxy-finance queries.
type StatsService struct {
	db *sqlx.DB
}

func NewStatsService(db *sqlx.DB) *StatsService {
	return &StatsService{db: db}
}

// Dashboard returns real-time dashboard stats for a station.
func (s *StatsService) Dashboard(ctx context.Context, stationID int64) (*DashboardResult, error) {
	r := &DashboardResult{}

	// Today inbound (count of parcels stored today).
	_ = s.db.GetContext(ctx, &r.TodayInbound,
		"SELECT COUNT(*) FROM parcels WHERE station_id = ? AND DATE(storage_time) = CURDATE()", stationID)

	// Today outbound (count of parcels picked up today).
	_ = s.db.GetContext(ctx, &r.TodayOutbound,
		"SELECT COUNT(*) FROM parcels WHERE station_id = ? AND DATE(pickup_time) = CURDATE()", stationID)

	// Pending.
	_ = s.db.GetContext(ctx, &r.PendingCount,
		"SELECT COUNT(*) FROM parcels WHERE station_id = ? AND status = 1", stationID)

	// Delayed (detained, status=3).
	_ = s.db.GetContext(ctx, &r.DelayedCount,
		"SELECT COUNT(*) FROM parcels WHERE station_id = ? AND status = 3", stationID)

	// Abnormal.
	_ = s.db.GetContext(ctx, &r.AbnormalCount,
		"SELECT COUNT(*) FROM parcels WHERE station_id = ? AND status = 5", stationID)

	// Active proxy orders (status 1-3).
	_ = s.db.GetContext(ctx, &r.ProxyActive,
		"SELECT COUNT(*) FROM proxy_orders WHERE station_id = ? AND status IN (1,2,3)", stationID)

	// Shelf usage rate.
	var used, total int64
	_ = s.db.GetContext(ctx, &used,
		"SELECT COALESCE(SUM(current_capacity), 0) FROM shelf_layout WHERE station_id = ?", stationID)
	_ = s.db.GetContext(ctx, &total,
		"SELECT COALESCE(SUM(max_capacity), 0) FROM shelf_layout WHERE station_id = ?", stationID)
	if total > 0 {
		r.ShelfUsageRate = float64(used) / float64(total)
	}

	return r, nil
}

// Trend returns trend data for the given granularity.
func (s *StatsService) Trend(ctx context.Context, stationID int64, granularity string) (*TrendResult, error) {
	type row struct {
		Date     string `db:"date_label"`
		Inbound  int64  `db:"inbound"`
		Outbound int64  `db:"outbound"`
		Dtcount  int64  `db:"dtcount"`
	}

	var rows []row
	query := "SELECT DATE(storage_time) AS date_label, COUNT(*) AS inbound, " +
		"COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS outbound, " +
		"COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) AS dtcount " +
		"FROM parcels WHERE station_id = ? " +
		"GROUP BY DATE(storage_time) ORDER BY DATE(storage_time) DESC LIMIT 30"

	if err := s.db.SelectContext(ctx, &rows, query, stationID); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	points := make([]*TrendPoint, len(rows))
	for i, r := range rows {
		points[i] = &TrendPoint{Date: r.Date, Inbound: r.Inbound, Outbound: r.Outbound, Delayed: r.Dtcount}
	}
	return &TrendResult{Granularity: granularity, Points: points}, nil
}

// CourierCheckResult is the response for GET /stats/courier-check.
type CourierCheckItem struct {
	CourierCompany  string  `json:"courier_company"`
	InboundCount    int64   `json:"inbound_count"`
	OutboundCount   int64   `json:"outbound_count"`
	DelayedCount    int64   `json:"delayed_count"`
	ReturnedCount   int64   `json:"returned_count"`
	AvgStorageHours float64 `json:"avg_storage_hours"`
}

// CourierCheck returns per-courier-company parcel statistics.
func (s *StatsService) CourierCheck(ctx context.Context, stationID int64, start, end string) ([]*CourierCheckItem, error) {
	type row struct {
		CourierCompany  string  `db:"courier_company"`
		InboundCount    int64   `db:"inbound_count"`
		OutboundCount   int64   `db:"outbound_count"`
		DelayedCount    int64   `db:"delayed_count"`
		ReturnedCount   int64   `db:"returned_count"`
		AvgStorageHours float64 `db:"avg_storage_hours"`
	}

	q := `SELECT courier_company,
	             COUNT(*) AS inbound_count,
	             COALESCE(SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END), 0) AS outbound_count,
	             COALESCE(SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END), 0) AS delayed_count,
	             COALESCE(SUM(CASE WHEN status = 4 THEN 1 ELSE 0 END), 0) AS returned_count,
	             COALESCE(AVG(CASE WHEN pickup_time IS NOT NULL
	                 THEN TIMESTAMPDIFF(HOUR, storage_time, pickup_time)
	                 ELSE TIMESTAMPDIFF(HOUR, storage_time, NOW()) END), 0) AS avg_storage_hours
	      FROM parcels
	      WHERE station_id = ?`
	var args []any
	args = append(args, stationID)

	if start != "" {
		q += " AND DATE(storage_time) >= ?"
		args = append(args, start)
	}
	if end != "" {
		q += " AND DATE(storage_time) <= ?"
		args = append(args, end)
	}
	q += " GROUP BY courier_company ORDER BY inbound_count DESC"

	var rows []row
	if err := s.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	items := make([]*CourierCheckItem, len(rows))
	for i, r := range rows {
		items[i] = &CourierCheckItem{
			CourierCompany:  r.CourierCompany,
			InboundCount:    r.InboundCount,
			OutboundCount:   r.OutboundCount,
			DelayedCount:    r.DelayedCount,
			ReturnedCount:   r.ReturnedCount,
			AvgStorageHours: r.AvgStorageHours,
		}
	}
	return items, nil
}

// ProxyFinance returns proxy order financial summary.
func (s *StatsService) ProxyFinance(ctx context.Context, stationID int64) (*ProxyFinanceResult, error) {
	r := &ProxyFinanceResult{}

	type row struct {
		Total     int64   `db:"total"`
		Completed int64   `db:"completed"`
		Amount    float64 `db:"amount"`
	}

	var row1 row
	_ = s.db.GetContext(ctx, &row1,
		"SELECT COUNT(*) AS total, COALESCE(SUM(reward_amount), 0) AS amount FROM proxy_orders WHERE station_id = ?", stationID)
	r.TotalOrders = row1.Total
	r.TotalAmount = row1.Amount

	_ = s.db.GetContext(ctx, &r.CompletedOrders,
		"SELECT COUNT(*) FROM proxy_orders WHERE station_id = ? AND status = 4", stationID)

	if r.TotalOrders > 0 {
		r.AvgAmount = r.TotalAmount / float64(r.TotalOrders)
	}

	return r, nil
}
