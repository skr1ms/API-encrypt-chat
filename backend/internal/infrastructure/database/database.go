package database

import (
	"crypto-chat-backend/internal/domain/entities"
	"crypto-chat-backend/pkg/config"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

func New(cfg *config.DatabaseConfig) (*Database, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &Database{db}, nil
}

func (db *Database) Migrate() error {
	return db.AutoMigrate(
		&entities.User{},
		&entities.Chat{},
		&entities.Message{},
		&entities.ChatMember{},
		&entities.KeyExchange{},
		&entities.Session{},
	)
}

func (db *Database) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
