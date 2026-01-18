package graph

import (
	"database/sql"

	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

type Resolver struct {
	DB *sql.DB

	RequestsNotifier requestsuc.Notifier
	PublicBaseURL    string
}
