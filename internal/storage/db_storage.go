package storage

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type DBStorage struct {
	DB *sqlx.DB
}

func NewDBStorage(db *sqlx.DB) *DBStorage {
	return &DBStorage{DB: db}
}

func (s *DBStorage) Init() error {
	schema := `
    CREATE TABLE IF NOT EXISTS url_records (
        uuid TEXT PRIMARY KEY,
        short_url VARCHAR(255) UNIQUE NOT NULL,
        original_url TEXT NOT NULL
    );`
	_, err := s.DB.Exec(schema)
	return err
}

func (s *DBStorage) Get(shortURL string) (string, bool) {
	var originalURL string
	err := s.DB.Get(&originalURL, "SELECT original_url FROM url_records WHERE short_url=$1", shortURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (s *DBStorage) Put(shortURL string, originalURL string) error {

	newUUID := uuid.New().String()
	_, err := s.DB.Exec("INSERT INTO url_records (uuid, short_url, original_url) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING", newUUID, shortURL, originalURL)
	if err != nil {
		zap.L().Error("Failed to insert into database", zap.Error(err))
		return err
	}
	return nil
}

func (s *DBStorage) PutBatch(records []BatchRecord) error {
	tx, err := s.DB.Beginx()
	if err != nil {
		return err
	}
	//newUUID := uuid.New().String()
	query := `INSERT INTO url_records (uuid, short_url, original_url) VALUES (:uuid, :short_url, :original_url) ON CONFLICT (short_url) DO NOTHING`
	stmt, err := tx.PrepareNamed(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i, record := range records {
        records[i].UUID = uuid.New().String()

		_, err := stmt.Exec(record)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
