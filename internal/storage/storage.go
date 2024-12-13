package storage

import (
	"fmt"

	"github.com/AvdzhiV/urlShort/configs"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage interface {
	Get(shortURL string) (string, bool)
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Put(shortURL string, originalURL string) (string, error)
	PutBatch(records []BatchRecord) ([]string, error)
	Init() error
}

type BatchRecord struct {
	UUID        string `db:"uuid"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
}

func StoreInit(cfg configs.Config, logger *zap.Logger) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sqlx.Connect("postgres", cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to the database: %w", err)
		}
		logger.Info("Successfully connected to the database")

		dbStore := NewDBStorage(db)
		if err := dbStore.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		logger.Info("Database initialized successfully")
		return dbStore, nil
	}

	if cfg.FileStoragePath != "" {
		fileStore := NewFileStorage(cfg.FileStoragePath)
		if err := fileStore.Init(); err != nil {
			return nil, fmt.Errorf("failed to initialize file storage: %w", err)
		}
		return fileStore, nil
	}

	logger.Warn("No storage configuration provided, using in-memory storage")
	memStore := NewMemoryStorage()
	if err := memStore.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize memory storage: %w", err)
	}
	return memStore, nil
}
