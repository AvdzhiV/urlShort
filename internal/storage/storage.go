package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	FilePath string
	mu       sync.RWMutex
	urlMap   map[string]string
	db       *sqlx.DB
}

func NewStorage(filePath string, db *sqlx.DB) *Storage {
	return &Storage{
		FilePath: filePath,
		urlMap:   make(map[string]string),
		db:       db,
	}
}

func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Если файла нет, ничего не загружаем
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		err := json.Unmarshal(scanner.Bytes(), &record)
		if err != nil {
			return err
		}
		s.urlMap[record.ShortURL] = record.OriginalURL
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *Storage) Save() error {
	file, err := os.Create(s.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for shortURL, originalURL := range s.urlMap {
		record := URLRecord{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		}
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		_, err = file.Write(line)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) Get(shortURL string) (string, bool) {
	if s.db != nil {
		var originalURL string
		err := s.db.Get(&originalURL, "SELECT original_url FROM url_records WHERE short_url=$1", shortURL)
		if err != nil {
			return "", false
		}
		return originalURL, true
	} else {
		s.mu.RLock()
		defer s.mu.RUnlock()
		origURL, exists := s.urlMap[shortURL]
		return origURL, exists
	}
}

func (s *Storage) Put(shortURL string, originalURL string) error {
    if s.db != nil {
        newUUID := uuid.New().String()
        _, err := s.db.Exec("INSERT INTO url_records (uuid, short_url, original_url) VALUES ($1, $2, $3) ON CONFLICT (short_url) DO NOTHING", newUUID, shortURL, originalURL)
        if err != nil {
            zap.L().Error("Failed to insert into database", zap.Error(err))
            return err
        }
        return nil
    } else {
        s.mu.Lock()
        defer s.mu.Unlock()
        s.urlMap[shortURL] = originalURL
        return s.Save()
    }
}

func (s *Storage) InitDB() error {
	if s.db == nil {
		return nil
	}
	schema := `
	CREATE TABLE IF NOT EXISTS url_records (
		uuid TEXT PRIMARY KEY,
		short_url VARCHAR(255) UNIQUE NOT NULL,
		original_url TEXT NOT NULL
    );`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) DB() *sqlx.DB {
	return s.db
}
