package graph

import (
	"context"
	"errors"
	"fmt"
	"testing"

	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

func TestNewErrorPresenter(t *testing.T) {
	presenter := NewErrorPresenter()
	ctx := context.Background()

	tests := []struct {
		name        string
		err         error
		wantCode    string
		wantMessage string
	}{
		{
			name:        "errUnauthorized → UNAUTHORIZED",
			err:         errUnauthorized,
			wantCode:    errCodeUnauthorized,
			wantMessage: "認証が必要です",
		},
		{
			name:        "errForbidden → FORBIDDEN",
			err:         errForbidden,
			wantCode:    errCodeForbidden,
			wantMessage: "この操作を行う権限がありません",
		},
		{
			name:        "requestsuc.ErrForbidden → FORBIDDEN",
			err:         requestsuc.ErrForbidden,
			wantCode:    errCodeForbidden,
			wantMessage: "この操作を行う権限がありません",
		},
		{
			name:        "requestsuc.ErrNotFound → NOT_FOUND",
			err:         requestsuc.ErrNotFound,
			wantCode:    errCodeNotFound,
			wantMessage: "リソースが見つかりません",
		},
		{
			name:        "attachmentsuc.ErrForbidden → FORBIDDEN",
			err:         attachmentsuc.ErrForbidden,
			wantCode:    errCodeForbidden,
			wantMessage: "この操作を行う権限がありません",
		},
		{
			name:        "attachmentsuc.ErrNotFound → NOT_FOUND",
			err:         attachmentsuc.ErrNotFound,
			wantCode:    errCodeNotFound,
			wantMessage: "リソースが見つかりません",
		},
		{
			name:        "ラップされたErrNotFound → NOT_FOUND",
			err:         fmt.Errorf("operation failed: %w", requestsuc.ErrNotFound),
			wantCode:    errCodeNotFound,
			wantMessage: "リソースが見つかりません",
		},
		{
			name:        "未知のエラー → INTERNAL_ERROR",
			err:         errors.New("unexpected db error"),
			wantCode:    errCodeInternal,
			wantMessage: "内部エラーが発生しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := presenter(ctx, tt.err)

			if got.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			code, ok := got.Extensions["code"]
			if !ok {
				t.Fatal("extensions[code] が存在しません")
			}
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}
