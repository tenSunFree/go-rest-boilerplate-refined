package seeders

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tenSunFree/luma-lang-go/internal/constants"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/records"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

type Seeder interface {
	UserSeeder(userData []records.Users) (err error)
	LiveCourseSeeder() (err error)
}

type seeder struct {
	db *sqlx.DB
}

func NewSeeder(db *sqlx.DB) Seeder {
	return &seeder{db: db}
}

func (s *seeder) UserSeeder(userData []records.Users) (err error) {
	if len(userData) == 0 {
		return errors.New("users data is empty")
	}

	logger.Info("inserting users data...", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, user := range userData {
		user.CreatedAt = time.Now().UTC()
		if _, err = tx.NamedExec(`
    		INSERT INTO users(id, username, email, password, active, role_id, created_at)
    		VALUES (uuid_generate_v4(), :username, :email, :password, :active, :role_id, :created_at)
    		ON CONFLICT (email) WHERE deleted_at IS NULL DO NOTHING
    	`, user); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	logger.Info("users data inserted successfully", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	return
}
