package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
)

type attachmentJSON struct {
	ID          string `json:"id"`
	OrgID       string `json:"org_id"`
	RequestID   string `json:"request_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	CreatedAt   string `json:"created_at"`
	DownloadURL string `json:"download_url"`
}

func attachmentToJSON(r *http.Request, a dbgen.RequestAttachment) attachmentJSON {
	orgID := r.PathValue("orgID")
	requestID := r.PathValue("requestID")
	downloadURL := "/org/" + orgID + "/requests/" + requestID + "/attachments/" + a.ID.String()
	return attachmentJSON{
		ID:          a.ID.String(),
		OrgID:       a.OrgID.String(),
		RequestID:   a.RequestID.String(),
		Filename:    a.Filename,
		ContentType: a.ContentType,
		SizeBytes:   a.SizeBytes,
		SHA256:      a.Sha256,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		DownloadURL: downloadURL,
	}
}

func handleAttachmentsUpload(svc attachmentsuc.Service, maxBytes int64) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		orgIDStr, ok := mustOrgIDMatchSessionForAPI(w, r, sess)
		if !ok {
			return
		}
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeBadRequest(w, r, "invalid org id")
			return
		}
		requestID, err := uuid.Parse(r.PathValue("requestID"))
		if err != nil {
			writeBadRequest(w, r, "invalid request id")
			return
		}
		actorUserID, err := uuid.Parse(sess.UserID)
		if err != nil {
			writeUnauthorized(w, r)
			return
		}

		if maxBytes <= 0 {
			maxBytes = 10 * 1024 * 1024
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		file, header, err := r.FormFile("file")
		if err != nil {
			if errors.Is(err, http.ErrMissingFile) {
				writeBadRequest(w, r, "missing file")
				return
			}
			writeBadRequest(w, r, "invalid multipart form")
			return
		}
		defer func() { _ = file.Close() }()

		filename := sanitizeFilename(header.Filename)
		contentType := header.Header.Get("Content-Type")
		if strings.TrimSpace(contentType) == "" {
			contentType = "application/octet-stream"
		}

		created, err := svc.Upload(r.Context(), orgID, actorUserID, requestID, filename, contentType, file)
		if err != nil {
			writeAttachmentError(w, r, err)
			return
		}

		writeJSON(w, http.StatusCreated, attachmentToJSON(r, created))
	}
}

func handleAttachmentsList(svc attachmentsuc.Service) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		orgIDStr, ok := mustOrgIDMatchSessionForAPI(w, r, sess)
		if !ok {
			return
		}
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeBadRequest(w, r, "invalid org id")
			return
		}
		requestID, err := uuid.Parse(r.PathValue("requestID"))
		if err != nil {
			writeBadRequest(w, r, "invalid request id")
			return
		}
		actorUserID, err := uuid.Parse(sess.UserID)
		if err != nil {
			writeUnauthorized(w, r)
			return
		}

		rows, err := svc.ListByRequest(r.Context(), orgID, actorUserID, requestID)
		if err != nil {
			writeAttachmentError(w, r, err)
			return
		}
		out := make([]attachmentJSON, 0, len(rows))
		for _, a := range rows {
			out = append(out, attachmentToJSON(r, a))
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func handleAttachmentsDownload(svc attachmentsuc.Service) authedHandler {
	return func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		orgIDStr, ok := mustOrgIDMatchSessionForAPI(w, r, sess)
		if !ok {
			return
		}
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			writeBadRequest(w, r, "invalid org id")
			return
		}
		requestID, err := uuid.Parse(r.PathValue("requestID"))
		if err != nil {
			writeBadRequest(w, r, "invalid request id")
			return
		}
		attachmentID, err := uuid.Parse(r.PathValue("attachmentID"))
		if err != nil {
			writeBadRequest(w, r, "invalid attachment id")
			return
		}
		actorUserID, err := uuid.Parse(sess.UserID)
		if err != nil {
			writeUnauthorized(w, r)
			return
		}

		meta, rc, err := svc.Open(r.Context(), orgID, actorUserID, requestID, attachmentID)
		if err != nil {
			writeAttachmentError(w, r, err)
			return
		}
		defer func() { _ = rc.Close() }()

		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": meta.Filename}))
		if strings.TrimSpace(meta.ContentType) != "" {
			w.Header().Set("Content-Type", meta.ContentType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		if meta.SizeBytes > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(meta.SizeBytes, 10))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, rc)
	}
}

func mustOrgIDMatchSessionForAPI(w http.ResponseWriter, r *http.Request, sess sessionData) (string, bool) {
	orgID := r.PathValue("orgID")
	if orgID == "" {
		writeBadRequest(w, r, "missing org id")
		return "", false
	}
	if sess.OrgID == "" {
		writeUnauthorized(w, r)
		return "", false
	}
	if sess.OrgID != orgID {
		writeForbidden(w, r)
		return "", false
	}
	return orgID, true
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "file"
	}
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.ReplaceAll(name, "\u0000", "")
	if name == "." || name == "/" {
		return "file"
	}
	return name
}

func writeAttachmentError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, attachmentsuc.ErrForbidden) {
		writeForbidden(w, r)
		return
	}
	if errors.Is(err, attachmentsuc.ErrNotFound) {
		writeNotFound(w, r)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		writeNotFound(w, r)
		return
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
		return
	}
	if strings.Contains(err.Error(), "invalid storage key") {
		writeBadRequest(w, r, "invalid storage key")
		return
	}
	writeInternalError(w, r, nil, err)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}
