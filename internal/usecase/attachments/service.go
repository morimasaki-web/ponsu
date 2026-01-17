package attachments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not found")
)

type Service struct {
	DB      *sql.DB
	Storage Storage
}

func (s Service) Upload(
	ctx context.Context,
	orgID uuid.UUID,
	actorUserID uuid.UUID,
	requestID uuid.UUID,
	filename string,
	contentType string,
	r io.Reader,
) (dbgen.RequestAttachment, error) {
	if s.DB == nil {
		return dbgen.RequestAttachment{}, errors.New("db is nil")
	}
	if s.Storage == nil {
		return dbgen.RequestAttachment{}, errors.New("storage is nil")
	}
	if strings.TrimSpace(filename) == "" {
		return dbgen.RequestAttachment{}, errors.New("filename is required")
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	attachmentID := uuid.New()
	storageKey := fmt.Sprintf("org/%s/requests/%s/%s", orgID.String(), requestID.String(), attachmentID.String())

	tx, err := s.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return dbgen.RequestAttachment{}, err
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)
	if _, err := q.GetMembershipByOrgAndUserID(ctx, dbgen.GetMembershipByOrgAndUserIDParams{OrgID: orgID, UserID: actorUserID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbgen.RequestAttachment{}, ErrForbidden
		}
		return dbgen.RequestAttachment{}, err
	}
	if _, err := q.GetRequestByOrgAndID(ctx, dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbgen.RequestAttachment{}, ErrNotFound
		}
		return dbgen.RequestAttachment{}, err
	}

	putRes, err := s.Storage.Put(ctx, storageKey, r)
	if err != nil {
		return dbgen.RequestAttachment{}, err
	}
	if strings.TrimSpace(putRes.SHA256Hex) == "" {
		_ = s.Storage.Delete(ctx, storageKey)
		return dbgen.RequestAttachment{}, errors.New("storage did not return sha256")
	}

	row, err := q.CreateRequestAttachment(ctx, dbgen.CreateRequestAttachmentParams{
		OrgID:            orgID,
		RequestID:        requestID,
		Filename:         filename,
		ContentType:      contentType,
		SizeBytes:        putRes.SizeBytes,
		Sha256:           putRes.SHA256Hex,
		StorageKey:       storageKey,
		UploadedByUserID: uuid.NullUUID{UUID: actorUserID, Valid: true},
	})
	if err != nil {
		_ = s.Storage.Delete(ctx, storageKey)
		return dbgen.RequestAttachment{}, err
	}

	if err := tx.Commit(); err != nil {
		_ = s.Storage.Delete(ctx, storageKey)
		return dbgen.RequestAttachment{}, err
	}
	return row, nil
}

func (s Service) ListByRequest(
	ctx context.Context,
	orgID uuid.UUID,
	actorUserID uuid.UUID,
	requestID uuid.UUID,
) ([]dbgen.RequestAttachment, error) {
	if s.DB == nil {
		return nil, errors.New("db is nil")
	}

	q := dbgen.New(s.DB)
	if _, err := q.GetMembershipByOrgAndUserID(ctx, dbgen.GetMembershipByOrgAndUserIDParams{OrgID: orgID, UserID: actorUserID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrForbidden
		}
		return nil, err
	}

	rows, err := q.ListRequestAttachmentsByOrgAndRequestID(ctx, dbgen.ListRequestAttachmentsByOrgAndRequestIDParams{OrgID: orgID, RequestID: requestID})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s Service) Open(
	ctx context.Context,
	orgID uuid.UUID,
	actorUserID uuid.UUID,
	requestID uuid.UUID,
	attachmentID uuid.UUID,
) (dbgen.RequestAttachment, io.ReadCloser, error) {
	if s.DB == nil {
		return dbgen.RequestAttachment{}, nil, errors.New("db is nil")
	}
	if s.Storage == nil {
		return dbgen.RequestAttachment{}, nil, errors.New("storage is nil")
	}

	q := dbgen.New(s.DB)
	if _, err := q.GetMembershipByOrgAndUserID(ctx, dbgen.GetMembershipByOrgAndUserIDParams{OrgID: orgID, UserID: actorUserID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbgen.RequestAttachment{}, nil, ErrForbidden
		}
		return dbgen.RequestAttachment{}, nil, err
	}

	meta, err := q.GetRequestAttachmentByOrgRequestAndID(ctx, dbgen.GetRequestAttachmentByOrgRequestAndIDParams{OrgID: orgID, RequestID: requestID, ID: attachmentID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbgen.RequestAttachment{}, nil, ErrNotFound
		}
		return dbgen.RequestAttachment{}, nil, err
	}

	rc, err := s.Storage.Open(ctx, meta.StorageKey)
	if err != nil {
		return dbgen.RequestAttachment{}, nil, err
	}
	return meta, rc, nil
}
