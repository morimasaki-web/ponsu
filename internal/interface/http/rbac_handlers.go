// Package http は PonSu のHTTPハンドラを提供する。
// このファイルは Cookie セッションの情報を使って、ログイン必須・ロール必須・org境界のチェックを行う。
package http

import (
	"net/http"
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

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title></head><body>" +
		"<h1>Template (admin)</h1>" +
		"<p>org_id: " + htmlEscape(orgID) + "</p>" +
		"<p>(MVP-011) admin only page placeholder.</p>" +
		"<p><a href=\"/\">Home</a></p>" +
		"</body></html>"))
}

// HandleAdminTemplatesNewShortcut は、ログイン中の org_id を使って org スコープのURLへリダイレクトする。
func (a *OIDCAuth) HandleAdminTemplatesNewShortcut(w http.ResponseWriter, r *http.Request, sess sessionData) {
	if sess.OrgID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/org/"+sess.OrgID+"/admin/templates/new", http.StatusFound)
}
