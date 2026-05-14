package http

import "net/http"

type authedHandler func(http.ResponseWriter, *http.Request, sessionData)

// RequireLogin はログイン済みセッションを要求し、未ログインの場合は 401 を返す。
func (a *OIDCAuth) RequireLogin(next authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := a.readSession(r)
		if !ok || sess.UserID == "" {
			writeUnauthorized(w, r)
			return
		}
		next(w, r, sess)
	}
}

// RequireRole はログイン済みであることに加え、指定ロールを要求する。
func (a *OIDCAuth) RequireRole(role string, next authedHandler) http.HandlerFunc {
	return a.RequireLogin(func(w http.ResponseWriter, r *http.Request, sess sessionData) {
		if sess.Role != role {
			writeForbidden(w, r)
			return
		}
		next(w, r, sess)
	})
}
