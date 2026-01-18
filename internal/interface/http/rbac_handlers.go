// Package http は PonSu のHTTPハンドラを提供する。
// このファイルは Cookie セッションの情報を使って、ログイン必須・ロール必須・org境界のチェックを行う。
package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	requestsuc "github.com/morimasaki-web/ponsu/internal/usecase/requests"
)

type authedHandler func(http.ResponseWriter, *http.Request, sessionData)

// RequireLogin はログイン済みセッションを要求し、未ログインの場合は 401 を返す。
func (a *OIDCAuth) RequireLogin(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := a.readSession(r)
		if !ok || sess.UserID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, sess)
	}
}

// RequireRole はログイン済みであることに加え、指定ロールを要求する。
func (a *OIDCAuth) RequireRole(role string, next authedHandler) http.HandlerFunc {
	return a.RequireLogin(func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		if sess.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r, sess)
	})
}

// HandleAdminTemplatesNew はテンプレ作成画面（プレースホルダ）を返す。
// orgID をURLから受け取り、セッションの org_id と一致しない場合は 403 を返す。
func (a *OIDCAuth) HandleAdminTemplatesNew(w http.ResponseWriter, r *http.Request, sess sessionData) {
	showAdminTemplatesNew(w, r, sess, adminTemplatesNewViewModel{})
}

// HandleAdminTemplatesNewShortcut は、ログイン中の org_id を使って org スコープのURLへリダイレクトする。
func (a *OIDCAuth) HandleAdminTemplatesNewShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/admin/templates/new", http.StatusFound)
}

// --- MVP-041: minimal SSR for request creation with template selection ---

func (a *OIDCAuth) HandleRequestsIndexShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/requests", http.StatusFound)
}

type requestsIndexViewModel struct {
	Error string
	Items []dbgen.Request
}

func (a *OIDCAuth) HandleRequestsIndex(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	limit := int32(50)
	offset := int32(0)
	q := dbgen.New(a.db)
	items, err := q.ListRequestsByOrg(r.Context(), dbgen.ListRequestsByOrgParams{OrgID: orgID, Limit: limit, Offset: offset})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list requests: %v", err), http.StatusInternalServerError)
		return
	}
	showRequestsIndex(w, sess, orgIDStr, requestsIndexViewModel{Items: items})
}

func showRequestsIndex(w http.ResponseWriter, sess sessionData, orgID string, vm requestsIndexViewModel) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;max-width:1100px}table{width:100%;border-collapse:collapse;margin-top:12px}th,td{border-bottom:1px solid rgba(127,127,127,.25);padding:8px;text-align:left;vertical-align:top}th{font-weight:700} .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:12px} .btn{padding:8px 12px;border:1px solid rgba(127,127,127,.35);border-radius:6px;background:#fff;color:#111;text-decoration:none;display:inline-block} .muted{opacity:.8} .err{background:#fff1f1;border:1px solid #ffb4b4;padding:10px;border-radius:8px;margin:12px 0}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>Requests</h1>")
	b.WriteString("<p class=\"muted\">org_id: " + htmlEscape(orgID) + "</p>")

	if strings.TrimSpace(vm.Error) != "" {
		b.WriteString("<div class=\"err\">" + htmlEscape(vm.Error) + "</div>")
	}

	b.WriteString("<div class=\"row\">")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgID) + "/requests/new\">New Request</a>")
	b.WriteString("<a class=\"btn\" href=\"/\">Home</a>")
	b.WriteString("</div>")

	if len(vm.Items) == 0 {
		b.WriteString("<p>No requests yet.</p>")
		b.WriteString("</body></html>")
		_, _ = w.Write([]byte(b.String()))
		return
	}

	b.WriteString("<table><thead><tr><th>Title</th><th>Status</th><th>Created</th><th>ID</th></tr></thead><tbody>")
	for _, it := range vm.Items {
		createdAtStr := it.CreatedAt.Local().Format("2006-01-02 15:04:05")
		if it.CreatedAt.IsZero() {
			createdAtStr = "-"
		}
		b.WriteString("<tr>")
		b.WriteString("<td><a href=\"/org/" + htmlEscape(orgID) + "/requests/" + htmlEscape(it.ID.String()) + "\">" + htmlEscape(it.Title) + "</a></td>")
		b.WriteString("<td>" + htmlEscape(it.Status) + "</td>")
		b.WriteString("<td>" + htmlEscape(createdAtStr) + "</td>")
		b.WriteString("<td><code>" + htmlEscape(it.ID.String()) + "</code></td>")
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")
	b.WriteString("</body></html>")
	_, _ = w.Write([]byte(b.String()))
}

func (a *OIDCAuth) HandleRequestsNewShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

func (a *OIDCAuth) HandleRequestsNew(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	q := dbgen.New(a.db)
	tpls, err := q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}
	showRequestsNew(w, sess, orgIDStr, requestsNewViewModel{Templates: tpls})
}

func showRequestsNew(w http.ResponseWriter, sess sessionData, orgID string, vm requestsNewViewModel) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;max-width:900px}label{display:block;font-weight:600;margin:12px 0 6px}input,select{width:100%;padding:8px;border:1px solid rgba(127,127,127,.35);border-radius:6px;font:inherit} .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:12px} .btn{padding:8px 12px;border:1px solid rgba(127,127,127,.35);border-radius:6px;background:#fff;color:#111;text-decoration:none;display:inline-block} .btn.primary{background:#111;color:#fff;border-color:#111} .err{background:#fff1f1;border:1px solid #ffb4b4;padding:10px;border-radius:8px;margin:12px 0} .muted{opacity:.8}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>New Request</h1>")
	b.WriteString("<p class=\"muted\">org_id: " + htmlEscape(orgID) + "</p>")

	if strings.TrimSpace(vm.Error) != "" {
		b.WriteString("<div class=\"err\">" + htmlEscape(vm.Error) + "</div>")
	}

	if len(vm.Templates) == 0 {
		b.WriteString("<p>No workflow templates available.</p>")
		if sess.Role == "admin" {
			b.WriteString("<p><a class=\"btn\" href=\"/admin/templates/new\">Create template (admin)</a></p>")
		}
		b.WriteString("<p><a class=\"btn\" href=\"/\">Home</a></p>")
		b.WriteString("</body></html>")
		_, _ = w.Write([]byte(b.String()))
		return
	}

	action := "/org/" + orgID + "/requests"
	b.WriteString("<form method=\"post\" action=\"" + htmlEscape(action) + "\">")
	b.WriteString("<label for=\"title\">Title</label>")
	b.WriteString("<input id=\"title\" name=\"title\" required value=\"" + htmlEscape(vm.Title) + "\" />")
	b.WriteString("<label for=\"workflow_template_id\">Workflow Template</label>")
	b.WriteString("<select id=\"workflow_template_id\" name=\"workflow_template_id\" required>")
	b.WriteString("<option value=\"\" disabled")
	if strings.TrimSpace(vm.TemplateID) == "" {
		b.WriteString(" selected")
	}
	b.WriteString(">Select a template</option>")
	for _, t := range vm.Templates {
		id := t.ID.String()
		b.WriteString("<option value=\"" + htmlEscape(id) + "\"")
		if id == vm.TemplateID {
			b.WriteString(" selected")
		}
		b.WriteString(">" + htmlEscape(t.Name) + "</option>")
	}
	b.WriteString("</select>")
	b.WriteString("<div class=\"row\">")
	b.WriteString("<button class=\"btn primary\" type=\"submit\">Create</button>")
	b.WriteString("<a class=\"btn\" href=\"/\">Home</a>")
	b.WriteString("</div>")
	b.WriteString("</form>")
	b.WriteString("</body></html>")
	_, _ = w.Write([]byte(b.String()))
}

func (a *OIDCAuth) HandleRequestsCreate(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	tplIDStr := strings.TrimSpace(r.FormValue("workflow_template_id"))

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}
	actorUserID, err := uuid.Parse(sess.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
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

	svc := requestsuc.Service{DB: a.db, Notifier: a.requestsNotifier, PublicBaseURL: a.cfg.PublicBaseURLForLinks(), ActorDisplay: actorDisplayFromSession(sess)}
	requestID, err := svc.CreateRequestWithTemplate(r.Context(), orgID, actorUserID, title, workflowTemplateID)
	if err != nil {
		vm.Error = err.Error()
		q := dbgen.New(a.db)
		vm.Templates, _ = q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{OrgID: orgID, Limit: 200, Offset: 0})
		showRequestsNew(w, sess, orgIDStr, vm)
		return
	}

	http.Redirect(w, r, "/org/"+orgIDStr+"/requests/"+requestID.String(), http.StatusSeeOther)
}

func (a *OIDCAuth) HandleRequestsShow(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}
	requestIDStr := r.PathValue("requestID")
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}

	q := dbgen.New(a.db)
	reqRow, err := q.GetRequestByOrgAndID(r.Context(), dbgen.GetRequestByOrgAndIDParams{OrgID: orgID, ID: requestID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get request: %v", err), http.StatusInternalServerError)
		return
	}
	steps, err := q.ListRequestSteps(r.Context(), requestID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list steps: %v", err), http.StatusInternalServerError)
		return
	}

	audit, err := q.ListRequestAuditTrail(r.Context(), dbgen.ListRequestAuditTrailParams{OrgID: orgID, RequestID: requestID})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list audit: %v", err), http.StatusInternalServerError)
		return
	}

	flashErr := strings.TrimSpace(r.URL.Query().Get("err"))

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;max-width:1000px}table{width:100%;border-collapse:collapse;margin-top:12px}th,td{border-bottom:1px solid rgba(127,127,127,.25);padding:8px;text-align:left;vertical-align:top}th{font-weight:700} .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:12px} .btn{padding:8px 12px;border:1px solid rgba(127,127,127,.35);border-radius:6px;background:#fff;color:#111;text-decoration:none;display:inline-block} .btn.danger{border-color:#b42318;color:#b42318} .btn.primary{background:#111;color:#fff;border-color:#111} .muted{opacity:.8} .err{background:#fff1f1;border:1px solid #ffb4b4;padding:10px;border-radius:8px;margin:12px 0} textarea{width:100%;padding:8px;border:1px solid rgba(127,127,127,.35);border-radius:6px;font:inherit;min-height:90px}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>Request</h1>")
	b.WriteString("<p class=\"muted\">org_id: " + htmlEscape(orgIDStr) + "</p>")
	if flashErr != "" {
		b.WriteString("<div class=\"err\">" + htmlEscape(flashErr) + "</div>")
	}
	b.WriteString("<p><strong>Title</strong>: " + htmlEscape(reqRow.Title) + "</p>")
	b.WriteString("<p><strong>Status</strong>: " + htmlEscape(reqRow.Status) + "</p>")
	b.WriteString("<p><strong>ID</strong>: <code>" + htmlEscape(reqRow.ID.String()) + "</code></p>")

	b.WriteString("<h2>Actions</h2>")
	b.WriteString("<div class=\"row\">")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgIDStr) + "/requests\">Back to list</a>")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgIDStr) + "/requests/new\">New Request</a>")
	b.WriteString("</div>")

	if reqRow.Status == "draft" {
		b.WriteString("<form method=\"post\" action=\"/org/" + htmlEscape(orgIDStr) + "/requests/" + htmlEscape(reqRow.ID.String()) + "/submit\" style=\"margin-top:12px\">")
		b.WriteString("<button class=\"btn primary\" type=\"submit\">Submit</button>")
		b.WriteString("</form>")
	}

	if reqRow.Status == "submitted" && sess.Role == "admin" {
		b.WriteString("<div class=\"row\" style=\"margin-top:12px\">")
		b.WriteString("<form method=\"post\" action=\"/org/" + htmlEscape(orgIDStr) + "/requests/" + htmlEscape(reqRow.ID.String()) + "/approve\">")
		b.WriteString("<button class=\"btn primary\" type=\"submit\">Approve</button>")
		b.WriteString("</form>")
		b.WriteString("</div>")

		b.WriteString("<form method=\"post\" action=\"/org/" + htmlEscape(orgIDStr) + "/requests/" + htmlEscape(reqRow.ID.String()) + "/reject\" style=\"margin-top:12px\">")
		b.WriteString("<label for=\"reason\" style=\"display:block;font-weight:600;margin:12px 0 6px\">Reject reason</label>")
		b.WriteString("<textarea id=\"reason\" name=\"reason\" placeholder=\"Why reject?\"></textarea>")
		b.WriteString("<button class=\"btn danger\" type=\"submit\" style=\"margin-top:8px\">Reject</button>")
		b.WriteString("</form>")
	}

	b.WriteString("<h2>Steps</h2>")
	if len(steps) == 0 {
		b.WriteString("<p>No steps.</p>")
	} else {
		b.WriteString("<table><thead><tr><th>#</th><th>Label</th><th>Status</th></tr></thead><tbody>")
		for _, s := range steps {
			b.WriteString("<tr><td>" + htmlEscape(fmt.Sprintf("%d", s.StepIndex)) + "</td><td>" + htmlEscape(s.Label) + "</td><td>" + htmlEscape(s.Status) + "</td></tr>")
		}
		b.WriteString("</tbody></table>")
	}

	b.WriteString("<h2>Audit</h2>")
	if len(audit) == 0 {
		b.WriteString("<p>No audit entries.</p>")
	} else {
		b.WriteString("<table><thead><tr><th>When</th><th>Action</th><th>Actor</th><th>Data</th></tr></thead><tbody>")
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
			b.WriteString("<tr>")
			b.WriteString("<td>" + htmlEscape(when) + "</td>")
			b.WriteString("<td>" + htmlEscape(a.Action) + "</td>")
			b.WriteString("<td><code>" + htmlEscape(actor) + "</code></td>")
			b.WriteString("<td><code>" + htmlEscape(data) + "</code></td>")
			b.WriteString("</tr>")
		}
		b.WriteString("</tbody></table>")
	}

	b.WriteString("<div class=\"row\">")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgIDStr) + "/requests\">Requests</a>")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgIDStr) + "/requests/new\">New Request</a>")
	b.WriteString("<a class=\"btn\" href=\"/\">Home</a>")
	b.WriteString("</div>")
	b.WriteString("</body></html>")
	_, _ = w.Write([]byte(b.String()))
}

func (a *OIDCAuth) HandleRequestsSubmit(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}
	requestID, err := uuid.Parse(r.PathValue("requestID"))
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}
	actorUserID, err := uuid.Parse(sess.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
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

func (a *OIDCAuth) HandleRequestsApprove(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}
	requestID, err := uuid.Parse(r.PathValue("requestID"))
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}
	actorUserID, err := uuid.Parse(sess.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
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

func (a *OIDCAuth) HandleRequestsReject(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}
	requestID, err := uuid.Parse(r.PathValue("requestID"))
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}
	actorUserID, err := uuid.Parse(sess.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
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

type adminTemplatesNewViewModel struct {
	Error       string
	Name        string
	Description string
	Mode        string
	Approvers   string
	StepsText   string
	Definition  string
}

func (a *OIDCAuth) mustOrgIDMatchSession(w http.ResponseWriter, r *http.Request, sess sessionData) (orgID string, ok bool) {
	orgID = r.PathValue("orgID")
	if orgID == "" {
		http.Error(w, "missing org id", http.StatusBadRequest)
		return "", false
	}
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	if sess.OrgID != orgID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return "", false
	}
	return orgID, true
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
		http.Error(w, "missing org id", http.StatusBadRequest)
		return
	}
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if sess.OrgID != orgID {
		http.Error(w, "forbidden", http.StatusForbidden)
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

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;max-width:900px}label{display:block;font-weight:600;margin:12px 0 6px}input,textarea,select{width:100%;padding:8px;border:1px solid rgba(127,127,127,.35);border-radius:6px;font:inherit}textarea{min-height:220px;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}textarea.small{min-height:100px} .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:12px} .btn{padding:8px 12px;border:1px solid rgba(127,127,127,.35);border-radius:6px;background:#fff;color:#111;text-decoration:none;display:inline-block} .btn.primary{background:#111;color:#fff;border-color:#111} .err{background:#fff1f1;border:1px solid #ffb4b4;padding:10px;border-radius:8px;margin:12px 0} .muted{opacity:.8} .card{border:1px solid rgba(127,127,127,.25);border-radius:10px;padding:12px;margin-top:12px}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>New Workflow Template (admin)</h1>")
	b.WriteString("<p class=\"muted\">org_id: " + htmlEscape(orgID) + "</p>")

	if strings.TrimSpace(vm.Error) != "" {
		b.WriteString("<div class=\"err\">" + htmlEscape(vm.Error) + "</div>")
	}

	action := "/org/" + orgID + "/admin/templates"
	b.WriteString("<form method=\"post\" action=\"" + htmlEscape(action) + "\">")
	b.WriteString("<label for=\"name\">Name</label>")
	b.WriteString("<input id=\"name\" name=\"name\" required value=\"" + htmlEscape(vm.Name) + "\" />")
	b.WriteString("<label for=\"description\">Description</label>")
	b.WriteString("<input id=\"description\" name=\"description\" value=\"" + htmlEscape(vm.Description) + "\" />")

	b.WriteString("<label for=\"definition_mode\">Definition input</label>")
	b.WriteString("<select id=\"definition_mode\" name=\"definition_mode\">")
	if vm.Mode == "json" {
		b.WriteString("<option value=\"builder\">Builder (approvers)</option><option value=\"json\" selected>Raw JSON</option>")
	} else {
		b.WriteString("<option value=\"builder\" selected>Builder (approvers)</option><option value=\"json\">Raw JSON</option>")
	}
	b.WriteString("</select>")

	b.WriteString("<div class=\"card\">")
	b.WriteString("<div style=\"font-weight:700\">Builder: approvers</div>")
	b.WriteString("<div class=\"muted\" style=\"margin-top:4px\">空行でステップ区切り。各ステップは1行1人（メール or ユーザーID）。builder選択時はこの入力からdefinition(JSON)を自動生成します。</div>\n")
	b.WriteString("<label for=\"steps\">Steps</label>\n")
	b.WriteString(`<textarea class="small" id="steps" name="steps" spellcheck="false" placeholder="# Step 1
alice@example.com
bob@example.com

# Step 2
carol@example.com
">`)
	b.WriteString(htmlEscape(vm.StepsText))
	b.WriteString("</textarea>\n")
	b.WriteString("<div class=\"muted\" style=\"margin-top:6px\">※ 先頭の <code>#</code> 行はコメントとして無視します。</div>\n")
	b.WriteString("</div>")

	b.WriteString("<div class=\"card\">")
	b.WriteString("<div style=\"font-weight:700\">Raw JSON</div>")
	b.WriteString("<div class=\"muted\" style=\"margin-top:4px\">json選択時はこの内容をそのまま保存します。</div>")
	b.WriteString("<label for=\"definition\">Definition (JSON)</label>")
	b.WriteString("<textarea id=\"definition\" name=\"definition\" spellcheck=\"false\">" + htmlEscape(vm.Definition) + "</textarea>")
	b.WriteString("</div>")

	b.WriteString("<div class=\"row\">")
	b.WriteString("<button class=\"btn primary\" type=\"submit\">Create</button>")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgID) + "/admin/templates\">Back to list</a>")
	b.WriteString("<a class=\"btn\" href=\"/\">Home</a>")
	b.WriteString("</div>")
	b.WriteString("</form>")
	b.WriteString("</body></html>")

	_, _ = w.Write([]byte(b.String()))
}

// HandleAdminTemplatesCreate はテンプレを作成して一覧へリダイレクトする。
func (a *OIDCAuth) HandleAdminTemplatesCreate(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
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
		http.Error(w, "invalid org id", http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("failed to create template: %v", err), http.StatusInternalServerError)
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
func (a *OIDCAuth) HandleAdminTemplatesIndex(w http.ResponseWriter, r *http.Request, sess sessionData) {
	orgIDStr, ok := a.mustOrgIDMatchSession(w, r, sess)
	if !ok {
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orgUUID, err := uuid.Parse(orgIDStr)
	if err != nil {
		http.Error(w, "invalid org id", http.StatusBadRequest)
		return
	}

	q := dbgen.New(a.db)
	items, err := q.ListWorkflowTemplatesByOrg(r.Context(), dbgen.ListWorkflowTemplatesByOrgParams{
		OrgID:  orgUUID,
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list templates: %v", err), http.StatusInternalServerError)
		return
	}

	createdID := strings.TrimSpace(r.URL.Query().Get("created"))

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title>")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">")
	b.WriteString("<style>body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;margin:24px;max-width:1100px}table{width:100%;border-collapse:collapse;margin-top:12px}th,td{border-bottom:1px solid rgba(127,127,127,.25);padding:8px;text-align:left;vertical-align:top}th{font-weight:700} .row{display:flex;gap:8px;flex-wrap:wrap;margin-top:12px} .btn{padding:8px 12px;border:1px solid rgba(127,127,127,.35);border-radius:6px;background:#fff;color:#111;text-decoration:none;display:inline-block} .note{background:#f3f7ff;border:1px solid #b7d2ff;padding:10px;border-radius:8px;margin:12px 0} .muted{opacity:.8}</style>")
	b.WriteString("</head><body>")
	b.WriteString("<h1>Workflow Templates (admin)</h1>")
	b.WriteString("<p class=\"muted\">org_id: " + htmlEscape(orgIDStr) + "</p>")

	if createdID != "" {
		b.WriteString("<div class=\"note\">Created: " + htmlEscape(createdID) + "</div>")
	}

	b.WriteString("<div class=\"row\">")
	b.WriteString("<a class=\"btn\" href=\"/org/" + htmlEscape(orgIDStr) + "/admin/templates/new\">New</a>")
	b.WriteString("<a class=\"btn\" href=\"/\">Home</a>")
	b.WriteString("</div>")

	if len(items) == 0 {
		b.WriteString("<p>No templates yet.</p>")
		b.WriteString("</body></html>")
		_, _ = w.Write([]byte(b.String()))
		return
	}

	b.WriteString("<table><thead><tr><th>Name</th><th>Description</th><th>Created</th><th>ID</th></tr></thead><tbody>")
	for _, it := range items {
		createdAt := it.CreatedAt
		// MVP用：見た目を分かりやすくするためローカル表示。
		createdAtStr := createdAt.Local().Format("2006-01-02 15:04:05")
		if createdAt.IsZero() {
			createdAtStr = "-"
		}
		b.WriteString("<tr>")
		b.WriteString("<td>" + htmlEscape(it.Name) + "</td>")
		b.WriteString("<td>" + htmlEscape(it.Description) + "</td>")
		b.WriteString("<td>" + htmlEscape(createdAtStr) + "</td>")
		b.WriteString("<td><code>" + htmlEscape(it.ID.String()) + "</code></td>")
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")
	b.WriteString("</body></html>")
	_, _ = w.Write([]byte(b.String()))
}
