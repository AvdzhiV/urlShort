package storage

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	mu     sync.RWMutex
	urlMap map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urlMap: make(map[string]string),
	}
}

func (s *MemoryStorage) Init() error {
	return nil
}

func (s *MemoryStorage) Get(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	origURL, exists := s.urlMap[shortURL]
	return origURL, exists
}

func (s *MemoryStorage) Put(shortURL string, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, есть ли уже такой originalURL
	for _, url := range s.urlMap {
		if url == originalURL {
			return fmt.Errorf("url_exists")
		}
	}

	s.urlMap[shortURL] = originalURL
	return nil
}

func (s *MemoryStorage) PutBatch(records []BatchRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, record := range records {
		s.urlMap[record.ShortURL] = record.OriginalURL
	}
	return nil
}

func (s *MemoryStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    for shortURL, url := range s.urlMap {
        if url == originalURL {
            return shortURL, true
        }
    }
    return "", false
}
