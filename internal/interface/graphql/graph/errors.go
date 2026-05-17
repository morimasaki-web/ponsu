package graph

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

const (
	errCodeNotFound     = "NOT_FOUND"
	errCodeUnauthorized = "UNAUTHORIZED"
	errCodeForbidden    = "FORBIDDEN"
	errCodeInternal     = "INTERNAL_ERROR"
)

// NewErrorPresenter はドメインエラーをコード付きGraphQLエラーに変換する。
// 内部エラーの詳細はクライアントに漏らさず、コードとユーザー向けメッセージのみ返す。
func NewErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		var gqlErr *gqlerror.Error
		if errors.As(err, &gqlErr) {
			return gqlErr
		}

		code, message := classify(err)
		return &gqlerror.Error{
			Message: message,
			Extensions: map[string]any{
				"code": code,
			},
		}
	}
}

func classify(err error) (code, message string) {
	switch {
	case errors.Is(err, errUnauthorized):
		return errCodeUnauthorized, "認証が必要です"
	case errors.Is(err, errForbidden),
		errors.Is(err, requestsuc.ErrForbidden),
		errors.Is(err, attachmentsuc.ErrForbidden):
		return errCodeForbidden, "この操作を行う権限がありません"
	case errors.Is(err, requestsuc.ErrNotFound),
		errors.Is(err, attachmentsuc.ErrNotFound):
		return errCodeNotFound, "リソースが見つかりません"
	default:
		return errCodeInternal, "内部エラーが発生しました"
	}
}
