package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
)

type LocalStorage struct {
	BaseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{BaseDir: baseDir}
}

func (s *LocalStorage) Put(ctx context.Context, key string, r io.Reader) (attachmentsuc.PutResult, error) {
	_ = ctx
	if strings.TrimSpace(s.BaseDir) == "" {
		return attachmentsuc.PutResult{}, errors.New("base dir is empty")
	}
	fullPath, err := safeJoin(s.BaseDir, key)
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return attachmentsuc.PutResult{}, err
	}

	tmp, err := os.CreateTemp(filepath.Dir(fullPath), ".upload-*")
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	defer func() { _ = tmp.Close() }()

	h := sha256.New()
	w := io.MultiWriter(tmp, h)
	n, err := io.Copy(w, r)
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}
	if err := tmp.Close(); err != nil {
		return attachmentsuc.PutResult{}, err
	}

	if err := os.Rename(tmp.Name(), fullPath); err != nil {
		return attachmentsuc.PutResult{}, err
	}
	return attachmentsuc.PutResult{SizeBytes: n, SHA256Hex: hex.EncodeToString(h.Sum(nil))}, nil
}

func (s *LocalStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	_ = ctx
	if strings.TrimSpace(s.BaseDir) == "" {
		return nil, errors.New("base dir is empty")
	}
	fullPath, err := safeJoin(s.BaseDir, key)
	if err != nil {
		return nil, err
	}
	return os.Open(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	_ = ctx
	if strings.TrimSpace(s.BaseDir) == "" {
		return errors.New("base dir is empty")
	}
	fullPath, err := safeJoin(s.BaseDir, key)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalStorage) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	_ = ctx
	_ = key
	_ = expiration
	return "", errors.New("local storage does not support presigned URLs")
}

func safeJoin(baseDir, key string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("storage key is empty")
	}
	cleanKey := filepath.Clean(filepath.FromSlash(key))
	if cleanKey == "." || cleanKey == ".." {
		return "", errors.New("invalid storage key")
	}
	if filepath.IsAbs(cleanKey) {
		return "", errors.New("invalid storage key")
	}
	if strings.HasPrefix(cleanKey, ".."+string(os.PathSeparator)) {
		return "", errors.New("invalid storage key")
	}

	baseClean := filepath.Clean(baseDir)
	full := filepath.Join(baseClean, cleanKey)

	basePrefix := baseClean
	if !strings.HasSuffix(basePrefix, string(os.PathSeparator)) {
		basePrefix += string(os.PathSeparator)
	}
	fullClean := filepath.Clean(full)
	if fullClean == baseClean || !strings.HasPrefix(fullClean, basePrefix) {
		return "", errors.New("invalid storage key")
	}
	return fullClean, nil
}
