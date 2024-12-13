package storage

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)
//TODO Использовать миграции для создания схемы БД, Исправить батчинг `prepared statement`
//TODO Убрать логирование ошбики при проброссе наверх tx.Rollback() или убрать проброс tx.Rollback()
type DBStorage struct {
	DB *sqlx.DB
}

func NewDBStorage(db *sqlx.DB) *DBStorage {
	return &DBStorage{DB: db}
}

func (s *DBStorage) Init() error {
	schema := `
    CREATE TABLE IF NOT EXISTS url_records (
        uuid UUID NOT NULL,
        short_url VARCHAR(255) UNIQUE NOT NULL,
        original_url TEXT UNIQUE NOT NULL
    );`
	_, err := s.DB.Exec(schema)
	if err != nil {
		zap.L().Error("Failed to create table", zap.Error(err))
		return err
	}
	return nil
}

func (s *DBStorage) Get(shortURL string) (string, bool) {
	var originalURL string
	err := s.DB.Get(&originalURL, "SELECT original_url FROM url_records WHERE short_url=$1", shortURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (s *DBStorage) Put(shortURL string, originalURL string) (string, error) {
	newUUID := uuid.New().String()
	var existingShortURL string

	// Используем RETURNING short_url для получения short_url, если вставка прошла успешно
	err := s.DB.Get(&existingShortURL, `
        INSERT INTO url_records (uuid, short_url, original_url) 
        VALUES ($1, $2, $3) 
        ON CONFLICT (original_url) DO NOTHING 
        RETURNING short_url`, newUUID, shortURL, originalURL)

	if err != nil {
		if err == sql.ErrNoRows {
			existingShortURL, ok := s.GetShortURLByOriginalURL(originalURL)
			if !ok {
				zap.L().Error("URL exists but cannot retrieve short URL", zap.String("original_url", originalURL))
				return "", fmt.Errorf("url_exists_but_not_found")
			}
			return existingShortURL, fmt.Errorf("url_exists")
		}

		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == "original_url_unique" {
				existingShortURL, ok := s.GetShortURLByOriginalURL(originalURL)
				if !ok {
					zap.L().Error("URL exists but cannot retrieve short URL", zap.String("original_url", originalURL))
					return "", fmt.Errorf("url_exists_but_not_found")
				}
				return existingShortURL, fmt.Errorf("url_exists")
			}
		}

		zap.L().Error("Failed to insert into database", zap.Error(err))
		return "", err
	}
	return existingShortURL, nil
}

func (s *DBStorage) PutBatch(records []BatchRecord) ([]string, error) {
    tx, err := s.DB.Beginx()
    if err != nil {
        zap.L().Error("Failed to begin transaction", zap.Error(err))
        return nil, err
    }

    query := `
        INSERT INTO url_records (uuid, short_url, original_url) 
        VALUES (:uuid, :short_url, :original_url) 
        ON CONFLICT (original_url) DO NOTHING 
        RETURNING short_url`

    stmt, err := tx.PrepareNamed(query)
    if err != nil {
        tx.Rollback()
        zap.L().Error("Failed to prepare statement", zap.Error(err))
        return nil, err
    }
    defer stmt.Close()

    var shortURLs []string

    for i := range records {
        // Генерация UUID для каждой записи
        records[i].UUID = uuid.New().String()
    }

    zap.L().Info("Inserting batch records into database", zap.Int("count", len(records)))

    for _, record := range records {
        var insertedShortURL string
        err := stmt.Get(&insertedShortURL, record)
        if err != nil {
            if err == sql.ErrNoRows {
                // Конфликт возник, нужно получить существующий short_url
                existingShortURL, ok := s.GetShortURLByOriginalURL(record.OriginalURL)
                if !ok {
                    tx.Rollback()
                    zap.L().Error("URL exists but cannot retrieve short URL", zap.String("original_url", record.OriginalURL))
                    return nil, fmt.Errorf("url_exists_but_not_found")
                }
                shortURLs = append(shortURLs, existingShortURL)
                zap.L().Info("URL already exists", zap.String("original_url", record.OriginalURL), zap.String("existing_short_url", existingShortURL))
                continue
            }

            tx.Rollback()
            zap.L().Error("Failed to insert record", zap.Error(err))
            return nil, err
        }

        // Вставка успешна
        shortURLs = append(shortURLs, insertedShortURL)
    }

    if err := tx.Commit(); err != nil {
        zap.L().Error("Failed to commit transaction", zap.Error(err))
        return nil, err
    }

    zap.L().Info("Batch records inserted successfully", zap.Int("count", len(shortURLs)))
    return shortURLs, nil
}

func (s *DBStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	var shortURL string
	err := s.DB.Get(&shortURL, "SELECT short_url FROM url_records WHERE original_url=$1", originalURL)
	if err != nil {
		return "", false
	}
	return shortURL, true
}
