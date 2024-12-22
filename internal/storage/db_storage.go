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
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO Использовать миграции для создания схемы БД, Исправить батчинг `prepared statement`

//go:embed migrations/*.sql
var migrationsDir embed.FS

type DBStorage struct {
	Pool *pgxpool.Pool
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	pool, err := pgxpool.New(context.Background(), dsn)
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
		return err
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

func (dbs *DBStorage) Get(shortURL string) (string, bool) {
	var originalURL string
	err := dbs.Pool.
		QueryRow(context.Background(),
			"SELECT original_url FROM url_records WHERE short_url = $1",
			shortURL).
		Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (dbs *DBStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	var shortURL string
	err := dbs.Pool.
		QueryRow(context.Background(),
			"SELECT short_url FROM url_records WHERE original_url = $1",
			originalURL).
		Scan(&shortURL)
	if err != nil {
		return "", false
	}
	return shortURL, true
}

func (dbs *DBStorage) Put(shortURL string, originalURL string) (string, error) {
	// Сгенерим uuid сами или используем что-то ещё
	// Вставим запись
	_, err := dbs.Pool.Exec(context.Background(),
		"INSERT INTO url_records (uuid, short_url, original_url) VALUES (gen_random_uuid(), $1, $2)",
		shortURL, originalURL,
	)
	if err != nil {
		return "", fmt.Errorf("failed to insert record: %w", err)
	}
	return shortURL, nil
}

func (dbs *DBStorage) PutBatch(records []BatchRecord) ([]string, error) {
	inserted := make([]string, 0, len(records))
	for _, r := range records {
		// вставляем
		_, err := dbs.Pool.Exec(
			context.Background(),
			"INSERT INTO url_records (uuid, short_url, original_url) VALUES ($1, $2, $3)",
			r.UUID, r.ShortURL, r.OriginalURL,
		)
		if err != nil {
			return inserted, fmt.Errorf("failed to insert batch record: %w", err)
		}
		inserted = append(inserted, r.ShortURL)
	}
	return inserted, nil
}

func (db *DBStorage) Close() {
	db.Pool.Close()
}

/*
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
*/

/*
func (db *DB) PutEmployee(ctx context.Context, emp *model.Employee) error {
	tag, err := db.pool.Exec(
		ctx,
		`INSERT INTO employees(first_name, last_name, salary, position, email)
		VALUES ($1, $2, $3, (SELECT id FROM positions WHERE title=$4), $5)`,
		emp.FirstName, emp.LastName, emp.Salary, emp.Position, emp.Email,
	)
	if err != nil {

		return fmt.Errorf("failed to store employee: %w", err)
	}
	rowsAffectedCount := tag.RowsAffected()
	if rowsAffectedCount != 1 {
		return fmt.Errorf("expected one row to be affected, actually affected %d", rowsAffectedCount)
	}
	return nil
}
*/
