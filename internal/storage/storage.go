package storage

import (
	"context"
	"fmt"

	"github.com/AvdzhiV/urlShort/configs"

	"go.uber.org/zap"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	GetShortURLByOriginalURL(ctx context.Context, originalURL string) (string, bool)
	Put(ctx context.Context, shortURL string, originalURL string) (string, error)
	PutBatch(ctx context.Context, records []BatchRecord) ([]string, error)
}

type BatchRecord struct {
	UUID        string `db:"uuid"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
}

func NewStore(ctx context.Context, cfg configs.Config, logger *zap.Logger) (Storage, error) {
	if cfg.DatabaseDSN != "" {
		dbStore, err := NewDBStorage(ctx,  cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect/init db with migrations: %w", err)
		}
		logger.Info("Database initialized successfully")
		return dbStore, nil
	}

	if cfg.FileStoragePath != "" {
		fileStore := NewFileStorage(cfg.FileStoragePath)
		if err := fileStore.Init(ctx); err != nil {
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
