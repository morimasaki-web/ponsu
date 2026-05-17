package graphql

import (
	"database/sql"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/morimasaki-web/ponsu/internal/interface/graphql/graph"
	"github.com/morimasaki-web/ponsu/internal/interface/graphql/graph/generated"
	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

func NewServer(db *sql.DB, notifier requestsuc.Notifier, publicBaseURL string, storage attachmentsuc.Storage) *handler.Server {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{
			DB:               db,
			RequestsNotifier: notifier,
			PublicBaseURL:    publicBaseURL,
			Storage:          storage,
		},
	})
	srv := handler.New(schema)
	srv.SetErrorPresenter(graph.NewErrorPresenter())
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.MultipartForm{}) // POSTより先に追加
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	return srv
}

func PlaygroundHandler(graphQLEndpointPath string) http.Handler {
	return playground.Handler("PonSu GraphQL Playground", graphQLEndpointPath)
}
