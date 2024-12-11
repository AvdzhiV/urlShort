package storage

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage interface {
	Get(shortURL string) (string, bool)
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Put(shortURL string, originalURL string) (string, error)
	PutBatch(records []BatchRecord) ([]string, error)
	Init() error
}

type BatchRecord struct {
	UUID        string `db:"uuid"`
	ShortURL    string `db:"short_url"`
	OriginalURL string `db:"original_url"`
}
