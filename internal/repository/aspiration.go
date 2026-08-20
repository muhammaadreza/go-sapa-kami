package repository

import (
	"fmt"
	"time"
	"context"

	"go-sapa-kami/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AspirationRepository interface {
	// CRUD Operations
	Create(ctx context.Context, aspiration *model.Aspiration) error
	Update(ctx context.Context, aspiration *model.Aspiration) error
	Delete(ctx context.Context, messageID uint) error
	GetAll(ctx context.Context, filter model.AspirationFilter) ([]model.Aspiration, int64, error)
	GetByID(ctx context.Context, messageID uint) (*model.Aspiration, error)
	GetExpired(ctx context.Context) ([]model.Aspiration, error)

	// Count Operations
	Count(ctx context.Context, filter model.AspirationFilter) (int64, error)
	CountDaily(ctx context.Context, filter model.AspirationFilter) (map[string]int64, error)

	// Vote and Discussion Operations
	Vote(ctx context.Context, vote *model.Vote) error
	GetVoteStats(ctx context.Context, aspirationID uint) (acceptCount int64, rejectCount int64, totalVoters int64, err error)
	Discussion(ctx context.Context, discussion *model.Discussion) error
}

type aspirationRepository struct {
	db *gorm.DB
}

func NewAspirationRepository(db *gorm.DB) AspirationRepository {
	return &aspirationRepository{db}
}

// Create a New Aspiration
func (r *aspirationRepository) Create(ctx context.Context, aspiration *model.Aspiration) error {
	err := r.db.WithContext(ctx).Create(aspiration).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Preload("Sender").
		First(aspiration, aspiration.ID).Error
}

// Update an Aspiration
func (r *aspirationRepository) Update(ctx context.Context, aspiration *model.Aspiration) error {
	result := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Where("id = ?", aspiration.ID).
		Update("status", aspiration.Status)

	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return gorm.ErrRecordNotFound }

	return nil
}

// Delete a Aspiration
func (r *aspirationRepository) Delete(ctx context.Context, messageID uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Aspiration{}, messageID)
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return gorm.ErrRecordNotFound }

	return nil
}

// Get All Messages
func (r *aspirationRepository) GetAll(ctx context.Context, filter model.AspirationFilter) ([]model.Aspiration, int64, error) {
	var messages []model.Aspiration
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Scopes(aspirationFilter(filter))

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

	switch sortKey {
	case "sender_name": query = query.Joins("Sender").Order("Sender.name " + sortOrder)
	default: query = query.Order(sortKey + " " + sortOrder)
	}

	err = query.
		Scopes(aspirationCountSummary).
		Preload("Sender").
		Find(&messages).Error
	if err != nil { return nil, 0, err }

	return messages, count, nil
}

// Get Aspiration by ID
func (r *aspirationRepository) GetByID(ctx context.Context, messageID uint) (*model.Aspiration, error) {
	var aspiration model.Aspiration
	err := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Scopes(aspirationCountSummary).
		Preload("Sender").
		Preload("Discussions.Sender").
		Preload("Votes").
		First(&aspiration, messageID).Error

	return &aspiration, err
}

func (r *aspirationRepository) GetExpired(ctx context.Context) ([]model.Aspiration, error) { // (Hapus param deadline int)
	var expired []model.Aspiration
	err := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Where("type = ? AND status = ? AND expired_at <= ?", model.TypeInnovation, model.StatusOpen, time.Now()).
		Find(&expired).Error

	return expired, err
}

// Count Messages
func (r *aspirationRepository) Count(ctx context.Context, filter model.AspirationFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Scopes(aspirationFilter(filter))

	err := query.Count(&count).Error
	if err != nil { return 0, err }

	return count, nil
}

// Counts Daily Messages
func (r *aspirationRepository) CountDaily(ctx context.Context, filter model.AspirationFilter) (map[string]int64, error) {
	counts := make(map[string]int64)

	query := r.db.WithContext(ctx).
		Model(&model.Aspiration{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Group("DATE(created_at)").
		Scopes(aspirationFilter(filter))

	rows, err := query.Rows()
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		var date string
		var total int64

		err := rows.Scan(&date, &total);
		if err != nil { return nil, fmt.Errorf("scan daily aspiration count: %w", err) }
		if len(date) < 10 { return nil, fmt.Errorf("scan daily aspiration count: invalid date %q", date) }

		counts[date[:10]] = total
	}
	err = rows.Err();
	if err != nil { return nil, fmt.Errorf("iterate daily aspiration counts: %w", err) }

	return counts, nil
}

// Vote and Discussion Functions
// Vote on an Aspiration
func (r *aspirationRepository) Vote(ctx context.Context, vote *model.Vote) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "aspiration_id"}, {Name: "voter_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).
		Create(vote).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.Vote{}).
		Preload("Voter").
		First(vote, vote.ID).Error
}

// Get Vote Statistics for an Aspiration
func (r *aspirationRepository) GetVoteStats(ctx context.Context, aspirationID uint) (accept, reject, total int64, err error) {
	type VoteStat struct {
		Accept int64
		Reject int64
		Total  int64
	}

	var stat VoteStat
	err = r.db.WithContext(ctx).
		Model(&model.Vote{}).
		Select(`
			COALESCE(SUM(CASE WHEN value = 'accept' THEN 1 ELSE 0 END), 0) as accept,
			COALESCE(SUM(CASE WHEN value = 'reject' THEN 1 ELSE 0 END), 0) as reject,
			COUNT(DISTINCT voter_id) as total
		`).
		Where("aspiration_id = ?", aspirationID).
		Scan(&stat).Error

	return stat.Accept, stat.Reject, stat.Total, err
}

// Create a New Discussion
func (r *aspirationRepository) Discussion(ctx context.Context, discussion *model.Discussion) error {
	err := r.db.WithContext(ctx).Create(discussion).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.Discussion{}).
		Preload("Sender").
		First(discussion, discussion.ID).Error
}

// Scope Functions
// Aspiration Filter
func aspirationFilter(filter model.AspirationFilter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filter.Type != nil { db = db.Where("type = ?", *filter.Type) }
		if filter.Status != nil { db = db.Where("status = ?", *filter.Status) }
		if filter.SenderID != nil { db = db.Where("sender_id = ?", *filter.SenderID) }
		if filter.StartDate != nil { db = db.Where("created_at >= ?", *filter.StartDate+" 00:00:00") }
		if filter.EndDate != nil { db = db.Where("created_at <= ?", *filter.EndDate+" 23:59:59") }

		if filter.Query != nil && *filter.Query != "" {
			like := "%" + *filter.Query + "%"
			db = db.Where("subject ILIKE ? OR (type = ? AND sender_id IN (SELECT id FROM users WHERE name ILIKE ?))", like, model.TypeInnovation, like)
		}

		return db
	}
}

// Aspiration Count Summary
func aspirationCountSummary(db *gorm.DB) *gorm.DB {
	return db.Select(`aspirations.*,
		COALESCE((SELECT COUNT(*) FROM votes WHERE votes.aspiration_id = aspirations.id AND votes.value = 'accept'), 0) AS accept_vote_count,
		COALESCE((SELECT COUNT(*) FROM votes WHERE votes.aspiration_id = aspirations.id AND votes.value = 'reject'), 0) AS reject_vote_count,
		COALESCE((SELECT COUNT(*) FROM discussions WHERE discussions.aspiration_id = aspirations.id), 0) AS discussion_count`)
}
