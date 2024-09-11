package configs

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
)

var (
	addressFlag = flag.String("a", "localhost:8080", "Host for the server")
	baseURLFlag = flag.String("b", "http://localhost:8080", "Base URL for the short links")
)

type Config struct {
	Host    string
	Port    int
	BaseURL string
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
		return err
	}
	a.Host = parts[0]
	a.Port = port
	return nil
}

func ParseParts() *Config {
	flag.Parse()

	cfg := &Config{}

	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = *addressFlag
	}

	cfg.BaseURL = os.Getenv("BASE_URL")
	if cfg.BaseURL == "" {
		cfg.BaseURL = *baseURLFlag
	}

	err := cfg.Set(serverAddress)
	if err != nil {
		return nil
	}
	return cfg
}
