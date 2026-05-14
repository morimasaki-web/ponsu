package graph

import (
	"database/sql"

	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

type Resolver struct {
	DB *sql.DB

	RequestsNotifier requestsuc.Notifier
	PublicBaseURL    string

	// Storage for attachments
	Storage attachmentsuc.Storage
}
