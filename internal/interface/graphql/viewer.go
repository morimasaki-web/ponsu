package graphql

import (
	"context"

	"github.com/morimasaki-web/ponsu/internal/interface/graphqlctx"
)

type Viewer = graphqlctx.Viewer

func WithViewer(ctx context.Context, v Viewer) context.Context {
	return graphqlctx.WithViewer(ctx, v)
}

func ViewerFrom(ctx context.Context) (Viewer, bool) {
	return graphqlctx.ViewerFrom(ctx)
}
