package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	Env  string
	Host string
	Port int

	PostgresHost    string
	PostgresPort    int
	PostgresUser    string
	PostgresPassword string
	PostgresDB      string
	PostgresSSLMode string
}

func LoadFromEnv() (Config, error) {
	env := getEnv("PONSU_ENV", "dev")
	host := getEnv("PONSU_HOST", "127.0.0.1")
	port, err := getEnvInt("PONSU_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	pgHost := getEnv("PONSU_PG_HOST", "127.0.0.1")
	pgPort, err := getEnvInt("PONSU_PG_PORT", 5432)
	if err != nil {
		return Config{}, err
	}
	pgUser := getEnv("PONSU_PG_USER", "ponsu")
	pgPassword := getEnv("PONSU_PG_PASSWORD", "ponsu")
	pgDB := getEnv("PONSU_PG_DB", "ponsu")
	pgSSLMode := getEnv("PONSU_PG_SSLMODE", "disable")

	cfg := Config{
		Env:  env,
		Host: host,
		Port: port,

		PostgresHost:     pgHost,
		PostgresPort:     pgPort,
		PostgresUser:     pgUser,
		PostgresPassword: pgPassword,
		PostgresDB:       pgDB,
		PostgresSSLMode:  pgSSLMode,
	}
	return cfg, nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c Config) PostgresURL() string {
	if c.PostgresHost == "" || c.PostgresPort == 0 || c.PostgresUser == "" || c.PostgresDB == "" {
		return ""
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   fmt.Sprintf("%s:%d", c.PostgresHost, c.PostgresPort),
		Path:   "/" + c.PostgresDB,
	}
	q := u.Query()
	if c.PostgresSSLMode != "" {
		q.Set("sslmode", c.PostgresSSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
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
