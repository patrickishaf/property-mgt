package db

import (
	"log/slog"
	"time"

	"github.com/patrickishaf/property-mgt/config"
	"github.com/patrickishaf/property-mgt/internal/db/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect(cfg *config.DBConfig) *gorm.DB {
	pg, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	configureConnectionPooling(cfg, pg)

	db = pg
	err = db.AutoMigrate(
		&model.Account{},
		&model.Job{},
		&model.Lease{},
		&model.Payment{},
		&model.Property{},
		&model.PropertyUnit{},
		&model.Tenant{},
	)
	if err != nil {
		slog.Error("failed to migrate database", slog.String("error", err.Error()))
		panic(err)
	}
	return db
}

func CloseDB() {
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to close db", "error", err.Error())
	}
	if err := sqlDB.Close(); err != nil {
		slog.Error("failed to close db", "error", err.Error())
	} else {
		slog.Info("database connection closed successfully")
	}
}

func configureConnectionPooling(cfg *config.DBConfig, db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to configure db connection pooling")
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)

	s, err := time.ParseDuration(cfg.MaxIdleConnectionTime)
	if err == nil {
		sqlDB.SetConnMaxIdleTime(s)
	}
}
