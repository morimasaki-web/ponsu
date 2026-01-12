package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env  string
	Host string
	Port int
}

func LoadFromEnv() (Config, error) {
	env := getEnv("PONSU_ENV", "dev")
	host := getEnv("PONSU_HOST", "127.0.0.1")
	port, err := getEnvInt("PONSU_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Env:  env,
		Host: host,
		Port: port,
	}
	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid int env %s=%q: %w", key, v, err)
	}
	return parsed, nil
}
