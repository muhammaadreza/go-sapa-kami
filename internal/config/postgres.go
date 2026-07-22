package config

import (
	"fmt"
	"log"

	"go-sapa-kami/internal/model"
	"go-sapa-kami/internal/utils"

	"gorm.io/gorm"
	"gorm.io/driver/postgres"
)

func PostgresConfig(cfg *Postgres, defaults *Defaults) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)

	gormConfig := &gorm.Config{ TranslateError: true }
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil { log.Fatal("Failed to connect to PostgreSQL:", err) }

	enumSQL := `
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'role_enum') THEN
			CREATE TYPE role_enum AS ENUM ('user', 'admin', 'super_admin');
		END IF;

		IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'type_enum') THEN
			CREATE TYPE type_enum AS ENUM ('criticism', 'innovation');
		END IF;
	END
	$$;
	`

	err = db.Exec(enumSQL).Error;
	if err != nil { log.Fatal("Failed to create enums in PostgreSQL: ", err) }

	err = db.AutoMigrate(&model.User{}, &model.Message{})
	if err != nil { log.Fatal("Failed to migrate PostgreSQL: ", err) }

	seedDefaultAdmin(db, defaults)
	return db
}

func seedDefaultAdmin(db *gorm.DB, defaults *Defaults) {
	var count int64
	err := db.Model(&model.User{}).
		Where("role = ?", model.RoleSuperAdmin).
		Count(&count).Error

	if err != nil { log.Fatal("Failed to check for default admin user: ", err) }
	if count == 0 {
		hashedPassword, err := utils.HashPassword(defaults.Password)
		if err != nil { log.Fatal("Failed to hash default admin password: ", err) }

		user := &model.User{
			Name:       "Admin",
			Role:       model.RoleSuperAdmin,
			EmployeeID: defaults.AdminID,
			Password:   hashedPassword,
		}

		err = db.Create(user).Error
		if err != nil { log.Fatal("Failed to create default admin user: ", err) }
	}
}
