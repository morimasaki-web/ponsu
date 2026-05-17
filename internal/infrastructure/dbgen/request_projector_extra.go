package dbgen

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const setRequestReturned = `
UPDATE public.requests
SET status = 'returned',
    updated_at = $3
WHERE org_id = $1 AND id = $2
`

type SetRequestReturnedParams struct {
	OrgID     uuid.UUID `json:"org_id"`
	ID        uuid.UUID `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (q *Queries) SetRequestReturned(ctx context.Context, arg SetRequestReturnedParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, setRequestReturned, arg.OrgID, arg.ID, arg.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const setRequestResubmitted = `
UPDATE public.requests
SET status = 'submitted',
    submitted_at = $3,
    decided_at = NULL,
    updated_at = $3
WHERE org_id = $1 AND id = $2
`

type SetRequestResubmittedParams struct {
	OrgID       uuid.UUID `json:"org_id"`
	ID          uuid.UUID `json:"id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

func (q *Queries) SetRequestResubmitted(ctx context.Context, arg SetRequestResubmittedParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, setRequestResubmitted, arg.OrgID, arg.ID, arg.SubmittedAt)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
