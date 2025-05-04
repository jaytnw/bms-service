package models

import (
	"time"
)

type Status struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;index" json:"id"`
	DormID    string    `gorm:"type:varchar(100);not null" json:"dormId"`
	WasherID  string    `gorm:"type:varchar(100);not null;index:idx_washer_created" json:"washerId"`
	Status    string    `gorm:"type:varchar(50);not null" json:"status"`
	CreatedAt time.Time `gorm:"index:idx_washer_created" json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type StatusDTO struct {
	// WasherID  string    `json:"washerId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type WasherStatusHistory struct {
	WasherID string      `json:"washer_id"`
	History  []StatusDTO `json:"history"`
}

type DormStatusReport struct {
	DormID   string                `json:"dorm_id"`
	DormName string                `json:"dorm_name"`
	Machines []WasherStatusHistory `json:"machines"`
}

func ToStatusDTO(s Status) StatusDTO {
	return StatusDTO{
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
	}
}
