package storage

import (
	"errors"
	"sync"
)

type MemoryStorage struct {
	urlMap      map[string]string // short_url -> original_url
	originalMap map[string]string // original_url -> short_url
	mu          sync.RWMutex
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
		return existingShortURL, errors.New("url_exists")
	}

	s.urlMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, nil
}

func (s *MemoryStorage) PutBatch(records []BatchRecord) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := &FileStorage{}
	shortURLs := ProcessRecords(records, m)

	return shortURLs, nil
}

func (s *MemoryStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shortURL, ok := s.originalMap[originalURL]
	return shortURL, ok
}
