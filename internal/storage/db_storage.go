package storage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrURLExists = errors.New("original_url already exists")

//go:embed migrations/*.sql
var migrationsDir embed.FS

type DBStorage struct {
	Pool *pgxpool.Pool
}

func NewDBStorage(ctx context.Context,  dsn string) (*DBStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}
	return &DBStorage{
		Pool: pool,
	}, nil
}

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		log.Fatal(err)
		return fmt.Errorf("не удалось создать новый драйвер iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func (dbs *DBStorage) Get(ctx context.Context, shortURL string) (string, bool) {
	var originalURL string
	err := dbs.Pool.
		QueryRow(ctx,
			"SELECT original_url FROM url_records WHERE short_url = $1",
			shortURL).
		Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (dbs *DBStorage) GetShortURLByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	var shortURL string
	err := dbs.Pool.
		QueryRow(ctx,
			"SELECT short_url FROM url_records WHERE original_url = $1",
			originalURL).
		Scan(&shortURL)
	if err != nil {
		return "", false
	}
	return shortURL, true
}

func (dbs *DBStorage) Put(ctx context.Context,shortURL string, originalURL string) (string, error) {
	query := `
		INSERT INTO url_records (uuid, short_url, original_url)
		VALUES (gen_random_uuid(), $1, $2)
		ON CONFLICT (original_url)
		DO NOTHING RETURNING short_url
	`
	// Вставим запись
	err := dbs.Pool.QueryRow(ctx, query, shortURL, originalURL).Scan(&shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existingShortURL, exists := dbs.GetShortURLByOriginalURL(ctx,originalURL)
			if !exists {
				return "", ErrURLExists
			}
			return existingShortURL, ErrURLExists
		}
		return "", fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	return shortURL, nil
}

func (dbs *DBStorage) PutBatch(ctx context.Context, records []BatchRecord) ([]string, error) {
	inserted := make([]string, 0, len(records))

	tx, err := dbs.Pool.Begin(ctx)
	if err != nil {
		log.Fatalf("Unable to start transaction: %v\n", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			log.Printf("Error during transaction rollback: %v\n", rollbackErr)
		} // Откат транзакции в случае ошибки
	}()
	// Выполнение batch-запросов
	for _, record := range records {
		_, err := tx.Exec(
			ctx,
			"INSERT INTO url_records (uuid, short_url, original_url) VALUES (gen_random_uuid(), $1, $2)",
			record.ShortURL, record.OriginalURL,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, ok := dbs.GetShortURLByOriginalURL(ctx, record.OriginalURL)
				if !ok {
					return nil, ErrURLExists
				}
				return nil, ErrURLExists
			}
			return inserted, fmt.Errorf("error executing query: %w", err)
		}
		inserted = append(inserted, record.ShortURL)
	}
	if err := tx.Commit(context.Background()); err != nil {
		return inserted, fmt.Errorf("error committing transaction: %w", err)
	}
	return inserted, nil
}
func (dbs *DBStorage) Close() {
	dbs.Pool.Close()
}
