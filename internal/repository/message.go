package repository

import (
	"context"
	"go-sapa-kami/internal/model"

	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	Delete(ctx context.Context, messageID uint) error
	GetAll(ctx context.Context, filter model.MessageFilterRequest) ([]model.Message, int64, error)
	GetByID(ctx context.Context, messageID uint) (*model.Message, error)
	Count(ctx context.Context, filter model.MessageFilterRequest) (int64, error)
	CountDaily(ctx context.Context, filter model.MessageFilterRequest) (map[string]int64, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{ db }
}

// Create a New Message
func (r *messageRepository) Create(ctx context.Context, message *model.Message) error {
	err := r.db.WithContext(ctx).Create(message).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Preload("User").
		First(message, message.ID).Error
}

// Delete a Message
func (r *messageRepository) Delete(ctx context.Context, messageID uint) error {
	return r.db.WithContext(ctx).Delete(&model.Message{}, messageID).Error
}

// Get All Messages
func (r *messageRepository) GetAll(ctx context.Context, filter model.MessageFilterRequest) ([]model.Message, int64, error) {
	var messages []model.Message
	var count int64

	query := r.db.WithContext(ctx).Model(&model.Message{})
	if filter.Query != nil && *filter.Query != "" {
        like := "%" + *filter.Query + "%"
        query = query.Where("subject ILIKE ? OR user_id IN (SELECT id FROM users WHERE name ILIKE ?)", like, like)
    }

	if filter.Type != nil { query = query.Where("type = ?", *filter.Type) }
	if filter.UserID != nil { query = query.Where("user_id = ?", *filter.UserID) }
	if filter.StartDate != nil { query = query.Where("created_at >= ?", *filter.StartDate+" 00:00:00") }
    if filter.EndDate != nil { query = query.Where("created_at <= ?", *filter.EndDate+" 23:59:59") }

	err := query.Count(&count).Error
	if err != nil { return nil, 0, err }

	if filter.Limit > 0 {
		page := filter.Page
		if page <= 0 { page = 1 }
		offset := (page - 1) * filter.Limit
		query = query.Limit(filter.Limit).Offset(offset)
	}

	sortKey := "created_at"
    if filter.SortKey != nil { sortKey = *filter.SortKey }

    sortOrder := "DESC"
    if filter.SortOrder != nil { sortOrder = *filter.SortOrder }

	err = query.Preload("User").Order(sortKey + " " + sortOrder).Find(&messages).Error
	if err != nil { return nil, 0, err }

	return messages, count, nil
}

// Get Message by ID
func (r *messageRepository) GetByID(ctx context.Context, messageID uint) (*model.Message, error) {
	var message model.Message
	err := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Preload("User").
		First(&message, messageID).Error

	return &message, err
}

// Count Messages
func (r *messageRepository) Count(ctx context.Context, filter model.MessageFilterRequest) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Message{})
	if filter.Type != nil { query = query.Where("type = ?", *filter.Type) }
	if filter.StartDate != nil { query = query.Where("created_at >= ?", *filter.StartDate+" 00:00:00") }
	if filter.EndDate != nil { query = query.Where("created_at <= ?", *filter.EndDate+" 23:59:59") }

	err := query.Count(&count).Error
	if err != nil { return 0, err }

	return count, nil
}

// Counts Daily Messages
func (r *messageRepository) CountDaily(ctx context.Context, filter model.MessageFilterRequest) (map[string]int64, error) {
	counts := make(map[string]int64)

	query := r.db.WithContext(ctx).
		Model(&model.Message{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Group("DATE(created_at)")

	if filter.Type != nil { query = query.Where("type = ?", *filter.Type) }
	if filter.StartDate != nil { query = query.Where("created_at >= ?", *filter.StartDate+" 00:00:00") }
	if filter.EndDate != nil { query = query.Where("created_at <= ?", *filter.EndDate+" 23:59:59") }

	rows, err := query.Group("DATE(created_at)").Rows()
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		var date string
		var total int64

		rows.Scan(&date, &total)
		dateStr := date[:10]
		counts[dateStr] = total
	}

	return counts, nil
}
