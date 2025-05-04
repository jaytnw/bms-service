package repository

import (
	"context"

	"github.com/jaytnw/bms-service/internal/models"
	"gorm.io/gorm"
)

type StatusRepository interface {
	FindAll(ctx context.Context) ([]models.Status, error)
	SaveStatus(ctx context.Context, status *models.Status) error
	FindLatestByWasherID(ctx context.Context, washerID string) (*models.Status, error)
	FindHistoryByWasherID(ctx context.Context, washerID string) ([]models.Status, error)
	FindHistoryByWasherIDs(ctx context.Context, washerIDs []string) ([]models.Status, error)
	FindLatest50HistoryByWasherIDs(ctx context.Context, washerIDs []string) ([]models.Status, error)
}

type statusRepo struct {
	conn *gorm.DB
}

func NewStatusRepo(conn *gorm.DB) StatusRepository {
	return &statusRepo{
		conn: conn,
	}
}

func (r *statusRepo) FindAll(ctx context.Context) ([]models.Status, error) {
	var statuses []models.Status
	result := r.conn.Find(&statuses)
	if result.Error != nil {
		return nil, result.Error
	}

	return statuses, nil
}

func (r *statusRepo) SaveStatus(ctx context.Context, status *models.Status) error {
	return r.conn.WithContext(ctx).Create(status).Error
}

func (r *statusRepo) FindLatestByWasherID(ctx context.Context, washerID string) (*models.Status, error) {
	var status models.Status
	err := r.conn.WithContext(ctx).
		Where("washer_id = ?", washerID).
		Order("created_at DESC").
		Find(&status).Error

	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *statusRepo) FindHistoryByWasherID(ctx context.Context, washerID string) ([]models.Status, error) {
	var statuses []models.Status
	err := r.conn.WithContext(ctx).
		Where("washer_id = ?", washerID).
		Order("created_at DESC").
		Find(&statuses).Error

	if err != nil {
		return nil, err
	}
	return statuses, nil
}

func (r *statusRepo) FindHistoryByWasherIDs(ctx context.Context, washerIDs []string) ([]models.Status, error) {
	var statuses []models.Status
	err := r.conn.WithContext(ctx).Where("washer_id IN ?", washerIDs).Find(&statuses).Error
	return statuses, err
}

func (r *statusRepo) FindLatest50HistoryByWasherIDs(ctx context.Context, washerIDs []string) ([]models.Status, error) {
	if len(washerIDs) == 0 {
		return nil, nil
	}

	var statuses []models.Status

	err := r.conn.WithContext(ctx).
		Raw(`
		SELECT washer_id, status, created_at
		FROM (
			SELECT *, 
			       ROW_NUMBER() OVER (PARTITION BY washer_id ORDER BY created_at DESC) AS rn
			FROM statuses
			WHERE washer_id IN ?
		) AS ranked
		WHERE rn <= 100
		ORDER BY washer_id, created_at DESC
	`, washerIDs).
		Scan(&statuses).Error

	return statuses, err
}
