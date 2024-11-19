package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
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
}

func NewStorage(filePath string) *Storage {
	return &Storage{
		FilePath: filePath,
		urlMap:   make(map[string]string),
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	origURL, exists := s.urlMap[shortURL]
	return origURL, exists
}

func (s *Storage) Put(shortURL string, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urlMap[shortURL] = originalURL
	return s.Save()
}
