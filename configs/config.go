package configs

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	addressFlag         = flag.String("a", "localhost:8080", "Host for the server")
	baseURLFlag         = flag.String("b", "http://localhost:8080", "Base URL for the short links")
	fileStoragePathFlag = flag.String("f", "data.json", "Path to the file storage")
	databaseDSNFlag     = flag.String("d", "", "DATABASE_DSN")
)

type Config struct {
	Host            string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	Port            int
}

func (a Config) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *Config) Set(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("cannot parse port: %w", err)
	}
	a.Host = parts[0]
	a.Port = port
	return nil
}

func ParseParts() *Config {
	flag.Parse()
	var ok bool
	cfg := &Config{}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = *addressFlag
	}

	cfg.BaseURL, ok = os.LookupEnv("BASE_URL")
	if !ok {
		cfg.BaseURL = *baseURLFlag
	}

	databaseDSN := os.Getenv("DATABASE_DSN")
	if databaseDSN == "" {
		databaseDSN = *databaseDSNFlag
	}
	cfg.DatabaseDSN = databaseDSN

	fileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if fileStoragePath == "" {
		fileStoragePath = *fileStoragePathFlag
	}
	cfg.FileStoragePath = fileStoragePath

	err := cfg.Set(serverAddress)
	if err != nil {
		return nil
	}
	return cfg
}
