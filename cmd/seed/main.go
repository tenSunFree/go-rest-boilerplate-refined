package main

import (
	_ "github.com/lib/pq"
	"github.com/tenSunFree/luma-lang-go/cmd/seed/seeders"
	"github.com/tenSunFree/luma-lang-go/internal/config"
	"github.com/tenSunFree/luma-lang-go/internal/constants"
	"github.com/tenSunFree/luma-lang-go/internal/datasources/drivers"
	"github.com/tenSunFree/luma-lang-go/pkg/logger"
)

func init() {
	if err := config.InitializeAppConfig(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
	}
	logger.Info("configuration loaded", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
}

func main() {
	db, err := drivers.SetupSQLXPostgres()
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}
	defer func() { _ = db.Close() }()

	logger.Info("seeding...", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})

	seeder := seeders.NewSeeder(db)
	err = seeder.UserSeeder(seeders.UserData)
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	err = seeder.LiveCourseSeeder()
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	logger.Info("seeding success!", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
}
