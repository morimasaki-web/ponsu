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
)

func NewServer(db *sql.DB) *handler.Server {
	schema := generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DB: db}})
	srv := handler.New(schema)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	return srv
}

func PlaygroundHandler(graphQLEndpointPath string) http.Handler {
	return playground.Handler("PonSu GraphQL Playground", graphQLEndpointPath)
}
