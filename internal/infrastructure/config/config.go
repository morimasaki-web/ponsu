// Package config は環境変数からアプリ設定を読み込み、接続情報などを組み立てる。
// MVP段階ではHTTPサーバ設定とPostgreSQL/OIDC/セッション関連の最小構成を扱う。
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env  string
	Host string
	Port int

	AttachmentsStorage          string
	AttachmentsLocalDir         string
	AttachmentsMaxBytes         int64
	AttachmentsS3Endpoint       string
	AttachmentsS3Region         string
	AttachmentsS3Bucket         string
	AttachmentsS3AccessKey      string
	AttachmentsS3SecretKey      string
	AttachmentsS3ForcePathStyle bool

	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string

	OIDCIssuerURL     string
	OIDCClientID      string
	OIDCClientSecret  string
	OIDCRedirectURL   string
	OIDCScopes        string
	OIDCAllowedEmails string

	SessionHashKey  string
	SessionBlockKey string
}

func LoadFromEnv() (Config, error) {
	env := getEnv("PONSU_ENV", "dev")
	host := getEnv("PONSU_HOST", "127.0.0.1")
	port, err := getEnvInt("PONSU_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	attachmentsStorage := getEnv("PONSU_ATTACHMENTS_STORAGE", "local")
	attachmentsLocalDir := getEnv("PONSU_ATTACHMENTS_LOCAL_DIR", "./.data/attachments")
	attachmentsMaxBytes, err := getEnvInt64("PONSU_ATTACHMENTS_MAX_BYTES", 10*1024*1024)
	if err != nil {
		return Config{}, err
	}
	attachmentsS3Endpoint := getEnv("PONSU_ATTACHMENTS_S3_ENDPOINT", "")
	if strings.TrimSpace(attachmentsS3Endpoint) == "" {
		minioPort := getEnv("PONSU_MINIO_PORT", "9000")
		attachmentsS3Endpoint = "http://127.0.0.1:" + minioPort
	}
	attachmentsS3Region := getEnv("PONSU_ATTACHMENTS_S3_REGION", "us-east-1")
	attachmentsS3Bucket := getEnv("PONSU_ATTACHMENTS_S3_BUCKET", getEnv("PONSU_MINIO_BUCKET", "ponsu"))
	attachmentsS3AccessKey := getEnv("PONSU_ATTACHMENTS_S3_ACCESS_KEY", getEnv("PONSU_MINIO_ROOT_USER", ""))
	attachmentsS3SecretKey := getEnv("PONSU_ATTACHMENTS_S3_SECRET_KEY", getEnv("PONSU_MINIO_ROOT_PASSWORD", ""))
	attachmentsS3ForcePathStyle, err := getEnvBool("PONSU_ATTACHMENTS_S3_FORCE_PATH_STYLE", true)
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

	oidcIssuer := getEnv("PONSU_OIDC_ISSUER_URL", "")
	oidcClientID := getEnv("PONSU_OIDC_CLIENT_ID", "")
	oidcClientSecret := getEnv("PONSU_OIDC_CLIENT_SECRET", "")
	oidcRedirectURL := getEnv("PONSU_OIDC_REDIRECT_URL", "")
	oidcScopes := getEnv("PONSU_OIDC_SCOPES", "")
	oidcAllowedEmails := getEnv("PONSU_OIDC_ALLOWED_EMAILS", "")

	sessionHashKey := getEnv("PONSU_SESSION_HASH_KEY", "")
	sessionBlockKey := getEnv("PONSU_SESSION_BLOCK_KEY", "")

	cfg := Config{
		Env:  env,
		Host: host,
		Port: port,

		AttachmentsStorage:          attachmentsStorage,
		AttachmentsLocalDir:         attachmentsLocalDir,
		AttachmentsMaxBytes:         attachmentsMaxBytes,
		AttachmentsS3Endpoint:       attachmentsS3Endpoint,
		AttachmentsS3Region:         attachmentsS3Region,
		AttachmentsS3Bucket:         attachmentsS3Bucket,
		AttachmentsS3AccessKey:      attachmentsS3AccessKey,
		AttachmentsS3SecretKey:      attachmentsS3SecretKey,
		AttachmentsS3ForcePathStyle: attachmentsS3ForcePathStyle,

		PostgresHost:     pgHost,
		PostgresPort:     pgPort,
		PostgresUser:     pgUser,
		PostgresPassword: pgPassword,
		PostgresDB:       pgDB,
		PostgresSSLMode:  pgSSLMode,

		OIDCIssuerURL:     oidcIssuer,
		OIDCClientID:      oidcClientID,
		OIDCClientSecret:  oidcClientSecret,
		OIDCRedirectURL:   oidcRedirectURL,
		OIDCScopes:        oidcScopes,
		OIDCAllowedEmails: oidcAllowedEmails,

		SessionHashKey:  sessionHashKey,
		SessionBlockKey: sessionBlockKey,
	}
	return cfg, nil
}

// HTTPAddr はHTTPサーバの待ち受けアドレス（host:port）を返す。
func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// PostgresURL は PostgreSQL の接続URLを返す（必須項目が欠けている場合は空文字を返す）。
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

// getEnv は環境変数を読み、未設定/空の場合はデフォルト値を返す。
func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// getEnvInt は環境変数を int として読み、未設定/空の場合はデフォルト値を返す。
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

func getEnvInt64(key string, def int64) (int64, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid int64 env %s=%q: %w", key, v, err)
	}
	return parsed, nil
}

func getEnvBool(key string, def bool) (bool, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return false, fmt.Errorf("invalid bool env %s=%q: %w", key, v, err)
	}
	return parsed, nil
}
