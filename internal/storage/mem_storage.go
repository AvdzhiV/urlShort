package storage

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type MemoryStorage struct {
	mu          sync.RWMutex
	urlMap      map[string]string // short_url -> original_url
	originalMap map[string]string // original_url -> short_url
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urlMap:      make(map[string]string),
		originalMap: make(map[string]string),
	}
}

func (s *MemoryStorage) Init() error {
	return nil
}

func (s *MemoryStorage) Get(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	origURL, ok := s.urlMap[shortURL]
	return origURL, ok
}

func (s *MemoryStorage) Put(shortURL string, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingShortURL, ok := s.originalMap[originalURL]; ok {
		return existingShortURL, fmt.Errorf("url_exists")
	}

	s.urlMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, nil
}

func (s *MemoryStorage) PutBatch(records []BatchRecord) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shortURLs []string

	for _, record := range records {
		if existingShortURL, ok := s.originalMap[record.OriginalURL]; ok {
			shortURLs = append(shortURLs, existingShortURL)
			zap.L().Info("URL already exists", zap.String("original_url", record.OriginalURL), zap.String("existing_short_url", existingShortURL))
		} else {
			s.urlMap[record.ShortURL] = record.OriginalURL
			s.originalMap[record.OriginalURL] = record.ShortURL
			shortURLs = append(shortURLs, record.ShortURL)
			zap.L().Info("Inserting new URL", zap.String("short_url", record.ShortURL), zap.String("original_url", record.OriginalURL))
		}
	}

	return shortURLs, nil
}

func (s *MemoryStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURL, ok := s.originalMap[originalURL]
	return shortURL, ok
}
