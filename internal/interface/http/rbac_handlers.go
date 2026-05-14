// Package http は PonSu のHTTPハンドラを提供する。
// このファイルは Cookie セッションの情報を使って、ログイン必須・ロール必須・org境界のチェックを行う。
package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

type RequestsHandler struct {
	cfg              config.Config
	db               *sql.DB
	requestsNotifier requestsuc.Notifier
}

func NewRequestsHandler(cfg config.Config, db *sql.DB, requestsNotifier requestsuc.Notifier) *RequestsHandler {
	return &RequestsHandler{cfg: cfg, db: db, requestsNotifier: requestsNotifier}
}

type AdminTemplatesHandler struct {
	db *sql.DB
}

func NewAdminTemplatesHandler(db *sql.DB) *AdminTemplatesHandler {
	return &AdminTemplatesHandler{db: db}
}

func (a *RequestsHandler) ensureDB() error {
	if a.db == nil {
		return errors.New("db is not configured")
	}
	return nil
}

func (a *AdminTemplatesHandler) ensureDB() error {
	if a.db == nil {
		return errors.New("db is not configured")
	}
	return nil
}

// HandleAdminTemplatesNew はテンプレ作成画面（プレースホルダ）を返す。
// orgID をURLから受け取り、セッションの org_id と一致しない場合は 403 を返す。
func (a *AdminTemplatesHandler) HandleAdminTemplatesNew(w http.ResponseWriter, r *http.Request, sess sessionData) {
	showAdminTemplatesNew(w, r, sess, adminTemplatesNewViewModel{})
}

// HandleAdminTemplatesNewShortcut は、ログイン中の org_id を使って org スコープのURLへリダイレクトする。
func (a *AdminTemplatesHandler) HandleAdminTemplatesNewShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		writeUnauthorized(w, r)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/admin/templates/new", http.StatusFound)
}

// --- MVP-041: minimal SSR for request creation with template selection ---

func (a *RequestsHandler) HandleRequestsIndexShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		writeUnauthorized(w, r)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/requests", http.StatusFound)
}

type requestsIndexViewModel struct {
	Error string
	Items []dbgen.Request
}

func (a *RequestsHandler) HandleRequestsIndex(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}

	limit := int32(50)
	offset := int32(0)
	q := dbgen.New(a.db)
	items, err := q.ListRequestsByOrg(r.Context(), dbgen.ListRequestsByOrgParams{OrgID: orgID, Limit: limit, Offset: offset})
	if err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	showRequestsIndex(w, sess, orgIDStr, requestsIndexViewModel{Items: items})
}

func showRequestsIndex(w http.ResponseWriter, sess sessionData, orgID string, vm requestsIndexViewModel) {
	type pageData struct {
		OrgID string
		Sess  sessionData
		VM    requestsIndexViewModel
	}
	renderHTML(w, "requests_index", pageData{OrgID: orgID, Sess: sess, VM: vm}, http.StatusOK)
}

func (a *RequestsHandler) HandleRequestsNewShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		writeUnauthorized(w, r)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/requests/new", http.StatusFound)
}

type requestsNewViewModel struct {
	Error      string
	Title      string
	TemplateID string
	Templates  []dbgen.WorkflowTemplate
}

func (a *RequestsHandler) HandleRequestsNew(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}

	q := dbgen.New(a.db)
	tpls, err := q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
	if err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	showRequestsNew(w, sess, orgIDStr, requestsNewViewModel{Templates: tpls})
}

func showRequestsNew(w http.ResponseWriter, sess sessionData, orgID string, vm requestsNewViewModel) {
	type pageData struct {
		OrgID string
		Sess  sessionData
		VM    requestsNewViewModel
	}
	renderHTML(w, "requests_new", pageData{OrgID: orgID, Sess: sess, VM: vm}, http.StatusOK)
}

func (a *RequestsHandler) HandleRequestsCreate(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, r, "bad form")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	tplIDStr := strings.TrimSpace(r.FormValue("workflow_template_id"))

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}
	actorUserID, err := uuid.Parse(sess.UserID)
	if err != nil {
		writeBadRequest(w, r, "invalid user id")
		return
	}
	vm := requestsNewViewModel{Title: title, TemplateID: tplIDStr}

	if title == "" {
		vm.Error = "title is required"
		q := dbgen.New(a.db)
		vm.Templates, _ = q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
		showRequestsNew(w, sess, orgIDStr, vm)
		return
	}
	if tplIDStr == "" {
		vm.Error = "workflow template is required"
		q := dbgen.New(a.db)
		vm.Templates, _ = q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
		showRequestsNew(w, sess, orgIDStr, vm)
		return
	}
	workflowTemplateID, err := uuid.Parse(tplIDStr)
	if err != nil {
		vm.Error = "invalid workflow template id"
		q := dbgen.New(a.db)
		vm.Templates, _ = q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
		showRequestsNew(w, sess, orgIDStr, vm)
		return
	}

	q := dbgen.New(a.db)
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	requestHash := sha256Hash(title + "\n" + tplIDStr)

	// 同一ユーザによる同一内容のリクエストが既に処理されていないか確認し、重複していれば既存のレスポンスを返す。
	if idempotencyKey != "" {
		rowsAffected, err := q.InsertIdempotencyIfNotExists(r.Context(), dbgen.InsertIdempotencyIfNotExistsParams{
			OrgID:          orgID,
			ActorUserID:    actorUserID,
			Action:         "create_request",
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
			StatusCode:     0,
			ResponseBody:   "",
		})
		if err != nil {
			writeInternalError(w, r, nil, err)
			return
		}

		if rowsAffected == 0 {
			idemRow, err := q.GetIdempotencyByKey(r.Context(), dbgen.GetIdempotencyByKeyParams{
				OrgID:          orgID,
				ActorUserID:    actorUserID,
				Action:         "create_request",
				IdempotencyKey: idempotencyKey,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "idempotency conflict", http.StatusConflict)
					return
				}
				writeInternalError(w, r, nil, err)
				return
			}

			if idemRow.RequestHash != requestHash {
				http.Error(w, "idempotency key reused with different request body", http.StatusConflict)
				return
			}

			if idemRow.ResponseBody != "" && idemRow.StatusCode == int32(http.StatusSeeOther) {
				http.Redirect(w, r, idemRow.ResponseBody, http.StatusSeeOther)
				return
			}

			http.Error(w, "idempotency key already in progress", http.StatusConflict)
			return
		}
	}

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	requestID, err := svc.CreateRequestWithTemplate(r.Context(), orgID, actorUserID, title, workflowTemplateID)
	if err != nil {
		vm.Error = err.Error()
		q := dbgen.New(a.db)
		vm.Templates, _ = q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
		showRequestsNew(w, sess, orgIDStr, vm)
		return
	}

	redirectURL := "/org/" + orgIDStr + "/requests/" + requestID.String()
	if idempotencyKey != "" {
		_ = q.UpdateIdempotencyResponse(r.Context(), dbgen.UpdateIdempotencyResponseParams{
			OrgID:          orgID,
			ActorUserID:    actorUserID,
			Action:         "create_request",
			IdempotencyKey: idempotencyKey,
			RequestHash:    requestHash,
			StatusCode:     int32(http.StatusSeeOther),
			ResponseBody:   redirectURL,
		})
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (a *RequestsHandler) HandleRequestsShow(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}
	requestIDStr := r.PathValue("requestID")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid request id")
		return
	}

	q := dbgen.New(a.db)
	reqRow, err := q.GetRequestByOrgAndID(r.Context(), dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeNotFound(w, r)
			return
		}
		writeInternalError(w, r, nil, err)
		return
	}
	steps, err := q.ListRequestSteps(r.Context(), requestID)
	if err != nil {
		writeInternalError(w, r, nil, err)
		return
	}

	audit, err := q.ListRequestAuditTrail(r.Context(), dbgen.ListRequestAuditTrailParams{OrgID: orgID, RequestID: requestID})
	if err != nil {
		writeInternalError(w, r, nil, err)
		return
	}

	flashErr := strings.TrimSpace(r.URL.Query().Get("err"))

	type stepVM struct {
		StepIndex int32
		Label     string
		Status    string
	}
	type auditVM struct {
		When   string
		Action string
		Actor  string
		Data   string
	}
	type pageData struct {
		OrgID                  string
		Sess                   sessionData
		Req                    dbgen.Request
		Steps                  []stepVM
		Audit                  []auditVM
		FlashErr               string
		CanSubmit              bool
		CanApproveRejectReturn bool
		CanResubmit            bool
	}

	stepsVM := make([]stepVM, 0, len(steps))
	for _, s := range steps {
		stepsVM = append(stepsVM, stepVM{StepIndex: s.StepIndex, Label: s.Label, Status: s.Status})
	}
	auditVMs := make([]auditVM, 0, len(audit))
	for _, a := range audit {
		when := a.OccurredAt.Local().Format("2006-01-02 15:04:05")
		actor := "-"
		if a.ActorUserID.Valid {
			actor = a.ActorUserID.UUID.String()
		}
		data := "{}"
		if len(a.Data) > 0 {
			data = string(a.Data)
		}
		auditVMs = append(auditVMs, auditVM{When: when, Action: a.Action, Actor: actor, Data: data})
	}

	canSubmit := reqRow.Status == "draft"
	canApproveRejectReturn := reqRow.Status == "submitted" && sess.Role == "admin"
	canResubmit := reqRow.Status == "returned" && reqRow.CreatedByUserID.Valid && sess.UserID == reqRow.CreatedByUserID.UUID.String()

	renderHTML(
		w,
		"requests_show",
		pageData{
			OrgID:                  orgIDStr,
			Sess:                   sess,
			Req:                    reqRow,
			Steps:                  stepsVM,
			Audit:                  auditVMs,
			FlashErr:               flashErr,
			CanSubmit:              canSubmit,
			CanApproveRejectReturn: canApproveRejectReturn,
			CanResubmit:            canResubmit,
		},
		http.StatusOK,
	)
}

func (a *RequestsHandler) HandleRequestsSubmit(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
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
		writeBadRequest(w, r, "invalid user id")
		return
	}

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	if err := svc.SubmitRequest(r.Context(), orgID, actorUserID, requestID); err != nil {
		msg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

func (a *RequestsHandler) HandleRequestsApprove(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
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
		writeBadRequest(w, r, "invalid user id")
		return
	}

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	if err := svc.ApproveRequest(r.Context(), orgID, actorUserID, requestID); err != nil {
		msg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

func (a *RequestsHandler) HandleRequestsReject(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, r, "bad form")
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
		writeBadRequest(w, r, "invalid user id")
		return
	}
	reason := strings.TrimSpace(r.FormValue("reason"))

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	if err := svc.RejectRequest(r.Context(), orgID, actorUserID, requestID, reason); err != nil {
		msg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

func (a *RequestsHandler) HandleRequestsReturn(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, r, "bad form")
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
		writeBadRequest(w, r, "invalid user id")
		return
	}
	reason := strings.TrimSpace(r.FormValue("reason"))
	if reason == "" {
		msg := url.QueryEscape("reason is required")
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	if err := svc.ReturnRequest(r.Context(), orgID, actorUserID, requestID, reason); err != nil {
		msg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

func (a *RequestsHandler) HandleRequestsResubmit(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
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
		writeBadRequest(w, r, "invalid user id")
		return
	}

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	if err := svc.ResubmitRequest(r.Context(), orgID, actorUserID, requestID); err != nil {
		msg := url.QueryEscape(err.Error())
		http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String()+"?err="+msg, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

type adminTemplatesNewViewModel struct {
	Error       string
	Name        string
	Description string
	Mode        string
	Approvers   string
	StepsText   string
	Definition  string
}

func mustOrgIDMatchSession(w http.ResponseWriter, r *http.Request, sess sessionData) (orgID string, ok bool) {
	orgID = r.PathValue("orgID")
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

func (a *RequestsHandler) mustOrgIDMatchSession(w http.ResponseWriter, r *http.Request, sess sessionData) (string, bool) {
	return mustOrgIDMatchSession(w, r, sess)
}

func (a *AdminTemplatesHandler) mustOrgIDMatchSession(w http.ResponseWriter, r *http.Request, sess sessionData) (string, bool) {
	return mustOrgIDMatchSession(w, r, sess)
}

func actorDisplayFromSession(sess sessionData) string {
	name := strings.TrimSpace(sess.Name)
	email := strings.TrimSpace(sess.Email)
	if name != "" && email != "" {
		return name + " <" + email + ">"
	}
	if email != "" {
		return email
	}
	if name != "" {
		return name
	}
	return ""
}

func defaultWorkflowTemplateDefinitionJSON() string {
	// MVP段階のため、スキーマはまだ固定しない（後続MVPで強化）。
	return "{\n  \"version\": 1,\n  \"steps\": [\n    {\n      \"type\": \"approval\",\n      \"approvers\": []\n    }\n  ]\n}\n"
}

func showAdminTemplatesNew(w http.ResponseWriter, r *http.Request, sess sessionData, vm adminTemplatesNewViewModel) {
	orgID := r.PathValue("orgID")
	if orgID == "" {
		writeBadRequest(w, r, "missing org id")
		return
	}
	if sess.OrgID == "" {
		writeUnauthorized(w, r)
		return
	}
	if sess.OrgID != orgID {
		writeForbidden(w, r)
		return
	}

	if strings.TrimSpace(vm.Definition) == "" {
		vm.Definition = defaultWorkflowTemplateDefinitionJSON()
	}
	if strings.TrimSpace(vm.Mode) == "" {
		vm.Mode = "builder"
	}
	if strings.TrimSpace(vm.StepsText) == "" && strings.TrimSpace(vm.Approvers) != "" {
		// 後方互換：旧approvers入力をsteps相当に見せる
		vm.StepsText = vm.Approvers
	}

	type pageData struct {
		OrgID string
		Sess  sessionData
		VM    adminTemplatesNewViewModel
	}
	renderHTML(w, "admin_templates_new", pageData{OrgID: orgID, Sess: sess, VM: vm}, http.StatusOK)
}

// HandleAdminTemplatesCreate はテンプレを作成して一覧へリダイレクトする。
func (a *AdminTemplatesHandler) HandleAdminTemplatesCreate(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, r, "bad form")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	desc := strings.TrimSpace(r.FormValue("description"))
	mode := strings.TrimSpace(r.FormValue("definition_mode"))
	approversText := r.FormValue("approvers")
	stepsText := r.FormValue("steps")
	def := strings.TrimSpace(r.FormValue("definition"))

	vm := adminTemplatesNewViewModel{Name: name, Description: desc, Mode: mode, Approvers: approversText, StepsText: stepsText, Definition: def}
	if name == "" {
		vm.Error = "name is required"
		showAdminTemplatesNew(w, r, sess, vm)
		return
	}

	if mode != "json" {
		mode = "builder"
	}
	vm.Mode = mode

	if mode == "builder" {
		// 後方互換: steps が空なら approvers を単一ステップとして扱う
		if strings.TrimSpace(stepsText) == "" {
			stepsText = approversText
			vm.StepsText = stepsText
		}

		steps := parseApprovalSteps(stepsText)
		if len(steps) == 0 {
			vm.Error = "steps is required (at least 1 approver)"
			showAdminTemplatesNew(w, r, sess, vm)
			return
		}
		generated, err := buildWorkflowTemplateDefinitionJSONFromApprovalSteps(steps)
		if err != nil {
			vm.Error = "failed to build definition"
			showAdminTemplatesNew(w, r, sess, vm)
			return
		}
		def = generated
		vm.Definition = def
	}

	if strings.TrimSpace(def) == "" {
		def = "{}"
		vm.Definition = def
	}
	if !json.Valid([]byte(def)) {
		vm.Error = "definition must be valid JSON"
		showAdminTemplatesNew(w, r, sess, vm)
		return
	}

	orgUUID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}

	var createdBy uuid.NullUUID
	if sess.UserID != "" {
		if u, err := uuid.Parse(sess.UserID); err == nil {
			createdBy = uuid.NullUUID{UUID: u, Valid: true}
		}
	}

	q := dbgen.New(a.db)
	created, err := q.CreateWorkflowTemplate(r.Context(), dbgen.CreateWorkflowTemplateParams{
		OrgID:           orgUUID,
		Name:            name,
		Description:     desc,
		Definition:      json.RawMessage(def),
		CreatedByUserID: createdBy,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// unique_violation
			if pgErr.Code == "23505" {
				vm.Error = "template name already exists"
				showAdminTemplatesNew(w, r, sess, vm)
				return
			}
		}
		writeInternalError(w, r, nil, err)
		return
	}

	redirect := "/org/" + orgIDStr + "/admin/templates?created=" + created.ID.String()
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

type approvalStep struct {
	Approvers []string
}

func parseApprovalSteps(text string) []approvalStep {
	// 空行区切りでグルーピングし、各グループ内は1行1人。
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	var steps []approvalStep
	var cur []string
	flush := func() {
		cur = uniqueKeepOrder(cur)
		if len(cur) == 0 {
			return
		}
		steps = append(steps, approvalStep{Approvers: cur})
		cur = nil
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		cur = append(cur, line)
	}
	flush()
	return steps
}

func uniqueKeepOrder(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it == "" {
			continue
		}
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

func sha256Hash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func buildWorkflowTemplateDefinitionJSONFromApprovalSteps(steps []approvalStep) (string, error) {
	type step struct {
		Type      string   `json:"type"`
		Approvers []string `json:"approvers"`
	}
	type def struct {
		Version int    `json:"version"`
		Steps   []step `json:"steps"`
	}

	outSteps := make([]step, 0, len(steps))
	for _, s := range steps {
		if len(s.Approvers) == 0 {
			continue
		}
		outSteps = append(outSteps, step{Type: "approval", Approvers: s.Approvers})
	}
	obj := def{Version: 1, Steps: outSteps}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

// HandleAdminTemplatesIndex はテンプレ一覧を返す。
func (a *AdminTemplatesHandler) HandleAdminTemplatesIndex(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		writeInternalError(w, r, nil, err)
		return
	}

	orgUUID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeBadRequest(w, r, "invalid org id")
		return
	}

	q := dbgen.New(a.db)
	items, err := q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{
		OrgID:  orgUUID,
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		writeInternalError(w, r, nil, err)
		return
	}

	createdID := strings.TrimSpace(r.URL.Query().Get("created"))

	renderHTML(w, "admin_templates_index", struct {
		OrgID     string
		CreatedID string
		Items     []dbgen.WorkflowTemplate
	}{OrgID: orgIDStr, CreatedID: createdID, Items: items}, http.StatusOK)
}
