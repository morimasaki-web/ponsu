package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/morimasaki-web/ponsu/internal/interface/graphql/graph"
	"github.com/morimasaki-web/ponsu/internal/interface/graphql/graph/generated"
)

func NewServer() *handler.Server {
	schema := generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})
	return handler.NewDefaultServer(schema)
}

func PlaygroundHandler(graphQLEndpointPath string) http.Handler {
	return playground.Handler("PonSu GraphQL Playground", graphQLEndpointPath)
}
