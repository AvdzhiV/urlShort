package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var (
	hostFlag = flag.String("host", "localhost", "Host for the server")
	portFlag = flag.Int("port", 8080, "Port for the server")
)

type Config struct {
	Host string
	Port int
}

func (a Config) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}
func (a *Config) Set(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return errors.New("Need address in a form host:port")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	a.Host = parts[0]
	a.Port = port
	return nil
}

func Init() *Config {
	flag.Parse()

	return &Config{
		Host: *hostFlag,
		Port: *portFlag,
	}
}
