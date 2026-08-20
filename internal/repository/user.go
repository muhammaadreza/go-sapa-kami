package repository

import (
	"context"
	"fmt"
	"go-sapa-kami/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	// CRUD Operations
	Create(ctx context.Context, user *model.User) error
	CreateMany(ctx context.Context, users *[]model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userID uint) error
	GetAll(ctx context.Context, filter model.UserFilter) ([]model.User, int64, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmployeeID(ctx context.Context, employeeID uint) (*model.User, error)

	// Count Operations
	Count(ctx context.Context, filter model.UserFilter) (int64, error)
	CountActive(ctx context.Context, filter model.UserFilter) (int64, error)
	CountDaily(ctx context.Context, filter model.UserFilter) (map[string]int64, error)

	// Get Attachment Paths that will be removed by the User delete cascade
	GetAttachmentPaths(ctx context.Context, userID uint) ([]string, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

// Create a New User
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Create Many Users
func (r *userRepository) CreateMany(ctx context.Context, users *[]model.User) error {
	return r.db.WithContext(ctx).Create(users).Error
}

// Update an Existing User
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Updates(user)
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return gorm.ErrRecordNotFound }

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		First(user, user.ID).Error
}

// Delete a User
func (r *userRepository) Delete(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, userID)
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return gorm.ErrRecordNotFound }

	return nil
}

// Get All Users
func (r *userRepository) GetAll(ctx context.Context, filter model.UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Scopes(userFilter(filter))

	err := query.Count(&count).Error
	if err != nil { return nil, 0, err }

	if filter.Limit > 0 {
		page := filter.Page
		if page <= 0 { page = 1 }

		offset := (page - 1) * filter.Limit
		query = query.Limit(filter.Limit).Offset(offset)
	}

	sortKey := "name"
	if filter.SortKey != nil { sortKey = *filter.SortKey }

	sortOrder := "ASC"
	if filter.SortOrder != nil { sortOrder = *filter.SortOrder }

	err = query.Order(sortKey + " " + sortOrder).Find(&users).Error
	return users, count, err
}

// Get User by ID
func (r *userRepository) GetByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		First(&user, userID).Error

	return &user, err
}

// Get User by Employee ID
func (r *userRepository) GetByEmployeeID(ctx context.Context, employeeID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("employee_id = ?", employeeID).
		First(&user).Error

	return &user, err
}

// Count Users
func (r *userRepository) Count(ctx context.Context, filter model.UserFilter) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Scopes(userFilter(filter)).
		Count(&count).Error

	return count, err
}

// Count Active Users (Users who have sent aspirations)
func (r *userRepository) CountActive(ctx context.Context, filter model.UserFilter) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Joins("JOIN aspirations ON aspirations.sender_id = users.id").
		Scopes(activeUserFilter(filter)).
		Distinct("users.id").
		Count(&count).Error

	return count, err
}

// Count Daily Active Users
func (r *userRepository) CountDaily(ctx context.Context, filter model.UserFilter) (map[string]int64, error) {
	counts := make(map[string]int64)

	query := r.db.WithContext(ctx).
		Table("aspirations").
		Joins("JOIN users ON users.id = aspirations.sender_id").
		Select("DATE(aspirations.created_at) as date, COUNT(DISTINCT aspirations.sender_id) as total")

	if filter.Role != nil { query = query.Where("users.role = ?", *filter.Role) }
	if filter.StartDate != nil { query = query.Where("aspirations.created_at >= ?", *filter.StartDate+" 00:00:00") }
	if filter.EndDate != nil { query = query.Where("aspirations.created_at <= ?", *filter.EndDate+" 23:59:59") }

	rows, err := query.Group("DATE(aspirations.created_at)").Order("date ASC").Rows()
	if err != nil { return nil, err }
	defer rows.Close()

	for rows.Next() {
		var date string
		var total int64

		err := rows.Scan(&date, &total);
		if err != nil { return nil, fmt.Errorf("scan daily active user count: %w", err) }
		if len(date) < 10 { return nil, fmt.Errorf("scan daily active user count: invalid date %q", date) }

		counts[date[:10]] = total
	}

	err = rows.Err();
	if err != nil { return nil, fmt.Errorf("iterate daily active user counts: %w", err) }

	return counts, nil
}

// Get Attachment Paths that will be removed by the User delete cascade
func (r *userRepository) GetAttachmentPaths(ctx context.Context, userID uint) ([]string, error) {
	var aspirations []model.Aspiration
	err := r.db.WithContext(ctx).
		Select("id", "attachments").
		Where("sender_id = ?", userID).
		Preload("Discussions", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "aspiration_id", "attachments")
		}).Find(&aspirations).Error
	if err != nil { return nil, err }

	paths := make([]string, 0)
	for _, aspiration := range aspirations {
		paths = appendAttachmentPaths(paths, aspiration.Attachments)
		for _, discussion := range aspiration.Discussions {
			paths = appendAttachmentPaths(paths, discussion.Attachments)
		}
	}

	var discussions []model.Discussion
	err = r.db.WithContext(ctx).
		Select("discussions.id", "discussions.attachments").
		Joins("JOIN aspirations ON aspirations.id = discussions.aspiration_id").
		Where("discussions.sender_id = ? AND aspirations.sender_id <> ?", userID, userID).
		Find(&discussions).Error
	if err != nil { return nil, err }
	for _, discussion := range discussions {
		paths = appendAttachmentPaths(paths, discussion.Attachments)
	}

	return paths, nil
}

// Scopes for User Query
// User Filter
func userFilter(filter model.UserFilter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filter.Role != nil { db = db.Where("users.role = ?", *filter.Role) }
		if filter.Query != nil && *filter.Query != "" {
			like := "%" + *filter.Query + "%"
			db = db.Where("users.name ILIKE ? OR CAST(users.employee_id AS TEXT) ILIKE ?", like, like)
		}

		return db
	}
}

// Active User Filter
func activeUserFilter(filter model.UserFilter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if filter.Role != nil { db = db.Where("users.role = ?", *filter.Role) }
		if filter.StartDate != nil { db = db.Where("aspirations.created_at >= ?", *filter.StartDate+" 00:00:00") }
		if filter.EndDate != nil { db = db.Where("aspirations.created_at <= ?", *filter.EndDate+" 23:59:59") }

		return db
	}
}

// Helper function to append attachment paths from a slice of Attachments
func appendAttachmentPaths(paths []string, attachments model.Attachments) []string {
	for _, attachment := range attachments {
		if attachment.Path != "" { paths = append(paths, attachment.Path) }
	}

	return paths
}
