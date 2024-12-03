package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FileStorage struct {
	FilePath    string
	mu          sync.RWMutex
	urlMap      map[string]string
	originalMap map[string]string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		FilePath:    filePath,
		urlMap:      make(map[string]string),
		originalMap: make(map[string]string),
	}
}

func (s *FileStorage) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.FilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
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
		s.originalMap[record.OriginalURL] = record.ShortURL
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (s *FileStorage) Get(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	origURL, exists := s.urlMap[shortURL]
	return origURL, exists
}

func (s *FileStorage) Put(shortURL string, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingShortURL, exists := s.originalMap[originalURL]; exists {
		return existingShortURL, fmt.Errorf("url_exists")
	}

	s.urlMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, s.save()
}

func (s *FileStorage) save() error {
    file, err := os.Create(s.FilePath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
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
        _, err = writer.Write(line)
        if err != nil {
            return err
        }
        _, err = writer.Write([]byte("\n"))
        if err != nil {
            return err
        }
    }
    return writer.Flush()
}

func (s *FileStorage) PutBatch(records []BatchRecord) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shortURLs []string

	for _, record := range records {
		if existingShortURL, exists := s.originalMap[record.OriginalURL]; exists {
			shortURLs = append(shortURLs, existingShortURL)
			zap.L().Info("URL already exists", zap.String("original_url", record.OriginalURL), zap.String("existing_short_url", existingShortURL))
		} else {
			s.urlMap[record.ShortURL] = record.OriginalURL
			s.originalMap[record.OriginalURL] = record.ShortURL
			shortURLs = append(shortURLs, record.ShortURL)
			zap.L().Info("Inserting new URL", zap.String("short_url", record.ShortURL), zap.String("original_url", record.OriginalURL))
		}
	}

	return shortURLs, s.save()
}

func (s *FileStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURL, exists := s.originalMap[originalURL]
	return shortURL, exists
}
