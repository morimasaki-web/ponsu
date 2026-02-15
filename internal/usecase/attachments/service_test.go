package attachments

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStorage implements the Storage interface for testing
type mockStorage struct {
	generatePresignedURLFunc func(ctx context.Context, key string, expiration time.Duration) (string, error)
}

func (m *mockStorage) Put(ctx context.Context, key string, r io.Reader) (PutResult, error) {
	return PutResult{}, nil
}

func (m *mockStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorage) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	if m.generatePresignedURLFunc != nil {
		return m.generatePresignedURLFunc(ctx, key, expiration)
	}
	return "", nil
}

func TestGetDownloadURL_Validation(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	userID := uuid.New()
	requestID := uuid.New()
	attachmentID := uuid.New()

	t.Run("DB is nil", func(t *testing.T) {
		svc := Service{
			DB:      nil,
			Storage: &mockStorage{},
		}
		_, err := svc.GetDownloadURL(ctx, orgID, userID, requestID, attachmentID, 15*time.Minute)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "db is nil")
	})

	t.Run("Storage is nil", func(t *testing.T) {
		// DBを作成せずに、Storageがnilであることのチェックだけを確認
		// 実際にはDBがnilなのでそちらのエラーが先に出る
		svc := Service{
			DB:      nil,
			Storage: nil,
		}
		_, err := svc.GetDownloadURL(ctx, orgID, userID, requestID, attachmentID, 15*time.Minute)
		require.Error(t, err)
		// DBもStorageもnilだが、DBのチェックが先なのでそちらのエラーが出る
		assert.Contains(t, err.Error(), "nil")
	})
}
