package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/AvdzhiV/urlShort/internal/constants"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FileStorage struct {
	mu          *sync.Mutex
	urlMap      map[string]string
	originalMap map[string]string
	filePath    string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath:    filePath,
		urlMap:      make(map[string]string),
		originalMap: make(map[string]string),
		mu:          &sync.Mutex{},
	}
}

func (s *FileStorage) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("failed to open file %s: %w", s.filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			zap.L().Error("failed to close file", zap.Error(err))
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		err := json.Unmarshal(scanner.Bytes(), &record)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON record: %w", err)
		}
		s.urlMap[record.ShortURL] = record.OriginalURL
		s.originalMap[record.OriginalURL] = record.ShortURL
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}
	return nil
}

func (s *FileStorage) Get(shortURL string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	origURL, ok := s.urlMap[shortURL]
	return origURL, ok
}

func (s *FileStorage) Put(shortURL string, originalURL string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existingShortURL, ok := s.originalMap[originalURL]; ok {
		return existingShortURL, errors.New("url_exists")
	}

	s.urlMap[shortURL] = originalURL
	s.originalMap[originalURL] = shortURL
	return shortURL, s.save()
}

func (s *FileStorage) save() error {
	file, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", s.filePath, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			zap.L().Error("failed to close file", zap.Error(cerr))
		}
	}()

	writer := bufio.NewWriter(file)
	for shortURL, originalURL := range s.urlMap {
		record := URLRecord{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		}
		line, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON record: %w", err)
		}
		_, err = writer.Write(line)
		if err != nil {
			return fmt.Errorf("failed to write JSON line: %w", err)
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

func (s *FileStorage) PutBatch(records []BatchRecord) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var shortURLs []string

	for _, record := range records {
		if existingShortURL, ok := s.originalMap[record.OriginalURL]; ok {
			shortURLs = append(shortURLs, existingShortURL)
			zap.L().Info("URL already exists", zap.String(OriginalURLKey, record.OriginalURL),
				zap.String("existing_short_url", existingShortURL))
		} else {
			s.urlMap[record.ShortURL] = record.OriginalURL
			s.originalMap[record.OriginalURL] = record.ShortURL
			shortURLs = append(shortURLs, record.ShortURL)
			zap.L().Info("Inserting new URL", zap.String("short_url", record.ShortURL),
				zap.String(OriginalURLKey, record.OriginalURL))
		}
	}

	return shortURLs, s.save()
}

func (s *FileStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shortURL, ok := s.originalMap[originalURL]
	return shortURL, ok
}
