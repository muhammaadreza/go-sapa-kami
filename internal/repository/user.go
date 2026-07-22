package repository

import (
	"context"
	"go-sapa-kami/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	CreateMany(ctx context.Context, users *[]model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userID uint) error
	GetAll(ctx context.Context, filter model.UserFilterRequest) ([]model.User, int64, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmployeeID(ctx context.Context, employeeID uint) (*model.User, error)
	Count(ctx context.Context, filter model.UserFilterRequest) (int64, error)
	CountDaily(ctx context.Context, filter model.UserFilterRequest) (map[string]int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{ db }
}

// Create a New User
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		First(user, user.ID).Error
}

// Create Many Users
func (r *userRepository) CreateMany(ctx context.Context, users *[]model.User) error {
	err := r.db.WithContext(ctx).Create(users).Error
	if err != nil { return err }

	var ids []uint
	for _, user := range *users {
		ids = append(ids, user.ID)
	}

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id IN ?", ids).
		Find(users).Error
}

// Update an Existing User
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	err := r.db.WithContext(ctx).Updates(user).Error
	if err != nil { return err }

	return r.db.WithContext(ctx).
		Model(&model.User{}).
		First(user, user.ID).Error
}

// Delete a User
func (r *userRepository) Delete(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, userID).Error
}

// Get All Users
func (r *userRepository) GetAll(ctx context.Context, filter model.UserFilterRequest) ([]model.User, int64, error) {
	var users []model.User
	var count int64

	query := r.db.WithContext(ctx).Model(&model.User{})
	if filter.Role != nil { query = query.Where("role = ?", *filter.Role) }
	if filter.Query != nil && *filter.Query != "" {
        like := "%" + *filter.Query + "%"
        query = query.Where("name ILIKE ? OR CAST(employee_id AS TEXT) ILIKE ?", like, like)
    }

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

	switch sortKey {
		case "user_name": query = query.Joins("User").Order("User.name " + sortOrder)
		default: query = query.Order(sortKey + " " + sortOrder)
	}

	err = query.Find(&users).Error
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
func (r *userRepository) Count(ctx context.Context, filter model.UserFilterRequest) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).
		Model(&model.User{}).
		Joins("JOIN messages ON messages.user_id = users.id")

		if filter.Role != nil { query = query.Where("users.role = ?", *filter.Role) }
		if filter.StartDate != nil { query = query.Where("messages.created_at >= ?", *filter.StartDate+" 00:00:00") }
		if filter.EndDate != nil { query = query.Where("messages.created_at <= ?", *filter.EndDate+" 23:59:59") }

	err := query.Distinct("users.id").Count(&count).Error
	return count, err
}

// Count Daily Active Users
func (r *userRepository) CountDaily(ctx context.Context, filter model.UserFilterRequest) (map[string]int64, error) {
	counts := make(map[string]int64)

    query := r.db.WithContext(ctx).
        Table("messages").
        Select("DATE(created_at) as date, COUNT(DISTINCT user_id) as total")

	if filter.Role != nil { query = query.Where("role = ?", *filter.Role) }
    if filter.StartDate != nil { query = query.Where("created_at >= ?", *filter.StartDate+" 00:00:00") }
    if filter.EndDate != nil { query = query.Where("created_at <= ?", *filter.EndDate+" 23:59:59") }

    rows, err := query.Group("DATE(created_at)").Order("date ASC").Rows()
    if err != nil { return nil, err }
    defer rows.Close()

    for rows.Next() {
        var date string
        var total int64

        err := rows.Scan(&date, &total)
        if err == nil && len(date) >= 10 {
            counts[date[:10]] = total
        }
    }

    return counts, nil
}
