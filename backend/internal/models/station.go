package models

import (
	"time"
)

// Station status constants.
const (
	StationStatusClosed = 0
	StationStatusOpen   = 1
)

// Station maps the `stations` table.
type Station struct {
	ID            int64     `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Address       string    `db:"address" json:"address"`
	Latitude      float64   `db:"latitude" json:"latitude"`
	Longitude     float64   `db:"longitude" json:"longitude"`
	BusinessHours string    `db:"business_hours" json:"business_hours"`
	Status        int8      `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// StationDTO is the API representation of a station.
type StationDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	BusinessHours string `json:"business_hours"`
	Status        int8   `json:"status"`
	StatusText    string `json:"status_text"`
	CreatedAt     string `json:"created_at"`
}

func (s *Station) ToDTO() *StationDTO {
	statusText := "open"
	if s.Status == StationStatusClosed {
		statusText = "closed"
	}
	return &StationDTO{
		ID:            s.ID,
		Name:          s.Name,
		Address:       s.Address,
		Latitude:      s.Latitude,
		Longitude:     s.Longitude,
		BusinessHours: s.BusinessHours,
		Status:        s.Status,
		StatusText:    statusText,
		CreatedAt:     s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
