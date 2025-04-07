package storage

import (
	"github.com/AvdzhiV/urlShort/internal/constants"
	"go.uber.org/zap"
)

func ProcessRecords(records []BatchRecord, s *FileStorage) []string {
	var shortURLs []string
	for _, record := range records {
		if existingShortURL, ok := s.originalMap[record.OriginalURL]; ok {
			shortURLs = append(shortURLs, existingShortURL)
			zap.L().Info("URL already exists",
				zap.String(constants.OriginalURLKey, record.OriginalURL),
				zap.String("existing_short_url", existingShortURL))
		} else {
			s.urlMap[record.ShortURL] = record.OriginalURL
			s.originalMap[record.OriginalURL] = record.ShortURL
			shortURLs = append(shortURLs, record.ShortURL)
			zap.L().Info("Inserting new URL",
				zap.String("short_url", record.ShortURL),
				zap.String(constants.OriginalURLKey, record.OriginalURL))
		}
	}
	return shortURLs
}
