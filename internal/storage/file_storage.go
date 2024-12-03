package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/google/uuid"
)

type FileStorage struct {
	FilePath string
	mu       sync.RWMutex
	urlMap   map[string]string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		FilePath: filePath,
		urlMap:   make(map[string]string),
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

func (s *FileStorage) Put(shortURL string, originalURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urlMap[shortURL] = originalURL
	return s.save()
}

func (s *FileStorage) save() error {
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

func (s *FileStorage) PutBatch(records []BatchRecord) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    for _, record := range records {
        s.urlMap[record.ShortURL] = record.OriginalURL
    }
    return s.save()
}
