package graphqlctx

import (
	"context"

	"github.com/google/uuid"
)

type Viewer struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Role   string
	Name   string
	Email  string
}

type viewerContextKey struct{}

func WithViewer(ctx context.Context, v Viewer) context.Context {
	return context.WithValue(ctx, viewerContextKey{}, v)
}

func ViewerFrom(ctx context.Context) (Viewer, bool) {
	v, ok := ctx.Value(viewerContextKey{}).(Viewer)
	return v, ok
}
