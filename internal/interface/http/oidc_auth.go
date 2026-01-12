// Package http は PonSu のHTTPハンドラを提供する。
// このファイルは OIDC ログインと、署名/暗号化Cookieによる簡易セッションを担当する。
package http

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/securecookie"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/config"
	"github.com/morimasaki-web/ponsu/internal/infrastructure/dbgen"
	"golang.org/x/oauth2"
)

const (
	sessionCookieName = "ponsu_session"
	stateCookieName   = "ponsu_oidc_state"
)

type OIDCAuth struct {
	cfg config.Config
	db  *sql.DB

	logger *slog.Logger

	sc *securecookie.SecureCookie

	once     sync.Once
	initErr  error
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

// NewOIDCAuth は OIDC 認証と Cookie セッションを扱うハンドラを生成する。
// db が nil の場合、ログイン後のRBAC確定（ユーザー/所属の保存）は失敗する。
func NewOIDCAuth(cfg config.Config, logger *slog.Logger, db *sql.DB) *OIDCAuth {
	if logger == nil {
		logger = slog.Default()
	}

	hashKey := []byte(cfg.SessionHashKey)
	blockKey := []byte(cfg.SessionBlockKey)

	// Allow running without explicit keys (dev only). It will reset on restart.
	if len(hashKey) == 0 {
		hashKey = mustRandomBytes(32)
		logger.Warn("PONSU_SESSION_HASH_KEY is empty; using ephemeral dev key")
	}
	if len(blockKey) == 0 {
		blockKey = nil
	}

	sc := securecookie.New(hashKey, blockKey)
	// Keep max length reasonable; default is 4096.
	sc.MaxLength(4096)

	return &OIDCAuth{
		cfg:    cfg,
		db:     db,
		logger: logger,
		sc:     sc,
	}
}

// ensureDB はRBACのためのDB依存が満たされているかを確認する。
func (a *OIDCAuth) ensureDB() error {
	if a.db == nil {
		return errors.New("db is not configured (set postgres env and run migrations)")
	}
	return nil
}

// ensureInit は OIDC Provider のディスカバリと検証器の初期化を1回だけ行う。
func (a *OIDCAuth) ensureInit(ctx context.Context) error {
	a.once.Do(func() {
		if a.cfg.OIDCIssuerURL == "" || a.cfg.OIDCClientID == "" || a.cfg.OIDCClientSecret == "" {
			a.initErr = errors.New("OIDC is not configured (set PONSU_OIDC_ISSUER_URL / PONSU_OIDC_CLIENT_ID / PONSU_OIDC_CLIENT_SECRET)")
			return
		}

		provider, err := oidc.NewProvider(ctx, a.cfg.OIDCIssuerURL)
		if err != nil {
			a.initErr = fmt.Errorf("oidc discovery failed: %w", err)
			return
		}

		a.provider = provider
		a.verifier = provider.Verifier(&oidc.Config{ClientID: a.cfg.OIDCClientID})
	})

	return a.initErr
}

// HandleHome はログイン状態の簡易表示（MVP用）を返す。
func (a *OIDCAuth) HandleHome(w http.ResponseWriter, r *http.Request) {
	sess, ok := a.readSession(r)

	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>PonSu</title></head><body>")
	b.WriteString("<h1>PonSu</h1>")

	if ok {
		b.WriteString("<p>Logged in.</p>")
		b.WriteString("<ul>")
		b.WriteString("<li>sub: " + htmlEscape(sess.Sub) + "</li>")
		if sess.UserID != "" {
			b.WriteString("<li>user_id: " + htmlEscape(sess.UserID) + "</li>")
		}
		if sess.OrgID != "" {
			b.WriteString("<li>org_id: " + htmlEscape(sess.OrgID) + "</li>")
		}
		if sess.Role != "" {
			b.WriteString("<li>role: " + htmlEscape(sess.Role) + "</li>")
		}
		if sess.Email != "" {
			b.WriteString("<li>email: " + htmlEscape(sess.Email) + "</li>")
		}
		if sess.Name != "" {
			b.WriteString("<li>name: " + htmlEscape(sess.Name) + "</li>")
		}
		b.WriteString("</ul>")
		if sess.Role == "admin" {
			b.WriteString("<p><a href=\"/admin/templates/new\">Admin: New Template</a></p>")
		}
		b.WriteString("<p><a href=\"/auth/logout\">Logout</a></p>")
	} else {
		b.WriteString("<p>Not logged in.</p>")
		b.WriteString("<p><a href=\"/auth/login\">Login with OIDC</a></p>")

		if a.cfg.OIDCIssuerURL == "" || a.cfg.OIDCClientID == "" {
			b.WriteString("<p><small>OIDC not configured. Set PONSU_OIDC_ISSUER_URL / PONSU_OIDC_CLIENT_ID / PONSU_OIDC_CLIENT_SECRET.</small></p>")
		}
	}

	b.WriteString("</body></html>")
	_, _ = w.Write([]byte(b.String()))
}

// HandleLogin は OIDC の認可エンドポイントへリダイレクトしてログインを開始する。
func (a *OIDCAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := a.ensureInit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL := a.cfg.OIDCRedirectURL
	if redirectURL == "" {
		redirectURL = inferExternalURL(r) + "/auth/callback"
	}

	scopes := defaultScopes(a.cfg.OIDCScopes)
	oc := oauth2.Config{
		ClientID:     a.cfg.OIDCClientID,
		ClientSecret: a.cfg.OIDCClientSecret,
		Endpoint:     a.provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}

	state := randomToken(32)
	nonce := randomToken(32)

	if err := a.writeStateCookie(w, r, oidcState{State: state, Nonce: nonce}); err != nil {
		a.logger.Error("failed to write state cookie", "error", err)
		http.Error(w, "failed to start login", http.StatusInternalServerError)
		return
	}

	url := oc.AuthCodeURL(state, oidc.Nonce(nonce))
	http.Redirect(w, r, url, http.StatusFound)
}

// HandleCallback は OIDC のコールバックを処理し、IDトークン検証後にセッションを確立する。
func (a *OIDCAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if err := a.ensureInit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := a.ensureDB(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code/state", http.StatusBadRequest)
		return
	}

	saved, ok := a.readStateCookie(r)
	if !ok {
		http.Error(w, "missing login state (retry login)", http.StatusBadRequest)
		return
	}
	if saved.State != state {
		http.Error(w, "invalid state (retry login)", http.StatusBadRequest)
		return
	}

	redirectURL := a.cfg.OIDCRedirectURL
	if redirectURL == "" {
		redirectURL = inferExternalURL(r) + "/auth/callback"
	}
	oc := oauth2.Config{
		ClientID:     a.cfg.OIDCClientID,
		ClientSecret: a.cfg.OIDCClientSecret,
		Endpoint:     a.provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       defaultScopes(a.cfg.OIDCScopes),
	}

	tok, err := oc.Exchange(r.Context(), code)
	if err != nil {
		a.logger.Error("token exchange failed", "error", err)
		http.Error(w, "token exchange failed", http.StatusBadRequest)
		return
	}

	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		http.Error(w, "missing id_token", http.StatusBadRequest)
		return
	}

	idToken, err := a.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		a.logger.Error("id_token verify failed", "error", err)
		http.Error(w, "id_token verify failed", http.StatusBadRequest)
		return
	}

	if saved.Nonce != "" && idToken.Nonce != saved.Nonce {
		http.Error(w, "invalid nonce", http.StatusBadRequest)
		return
	}

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "failed to parse claims", http.StatusBadRequest)
		return
	}

	email := claims.Email
	if email == "" {
		email = claims.PreferredUsername
	}

	userID, orgID, role, err := a.ensureUserAndDefaultMembership(r.Context(), a.cfg.OIDCIssuerURL, claims.Sub, email, claims.Name)
	if err != nil {
		a.logger.Error("failed to ensure user/membership", "error", err)
		http.Error(w, "failed to create user/membership", http.StatusInternalServerError)
		return
	}

	sess := sessionData{
		Sub:    claims.Sub,
		Email:  email,
		Name:   claims.Name,
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		Iat:    time.Now().Unix(),
	}
	if err := a.writeSession(w, r, sess); err != nil {
		a.logger.Error("failed to write session", "error", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	a.clearStateCookie(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleLogout はセッションを破棄してホームに戻す。
func (a *OIDCAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	a.clearSession(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

type oidcState struct {
	State string
	Nonce string
}

type sessionData struct {
	Sub    string
	UserID string
	OrgID  string
	Role   string
	Email  string
	Name   string
	Iat    int64
}

// ensureUserAndDefaultMembership はOIDCのID情報からユーザーを upsert し、所属が無ければデフォルト組織を作って admin として所属させる。
func (a *OIDCAuth) ensureUserAndDefaultMembership(ctx context.Context, issuer, sub, email, name string) (userID string, orgID string, role string, err error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return "", "", "", err
	}
	defer func() { _ = tx.Rollback() }()

	q := dbgen.New(tx)
	usr, err := q.UpsertUserFromOIDC(ctx, dbgen.UpsertUserFromOIDCParams{
		OidcIssuer: issuer,
		OidcSub:    sub,
		Email:      email,
		Name:       name,
	})
	if err != nil {
		return "", "", "", err
	}

	mem, err := q.GetAnyMembershipByUserID(ctx, usr.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", "", err
		}

		orgName := defaultOrgName(email, name)
		org, err := q.CreateOrganization(ctx, orgName)
		if err != nil {
			return "", "", "", err
		}

		up, err := q.UpsertMembership(ctx, dbgen.UpsertMembershipParams{
			OrgID:  org.ID,
			UserID: usr.ID,
			Role:   "admin",
		})
		if err != nil {
			return "", "", "", err
		}

		mem.OrgID = up.OrgID
		mem.UserID = up.UserID
		mem.Role = up.Role
	}

	if err := tx.Commit(); err != nil {
		return "", "", "", err
	}

	return usr.ID.String(), mem.OrgID.String(), mem.Role, nil
}

// defaultOrgName は新規作成するデフォルト組織名を決める。
func defaultOrgName(email, name string) string {
	label := strings.TrimSpace(name)
	if label == "" {
		label = strings.TrimSpace(email)
	}
	if label == "" {
		label = "Personal"
	}
	return label + " Organization"
}

// writeStateCookie は state/nonce を一時Cookieとして保存する。
func (a *OIDCAuth) writeStateCookie(w http.ResponseWriter, r *http.Request, st oidcState) error {
	value := map[string]string{"state": st.State, "nonce": st.Nonce}
	encoded, err := a.sc.Encode(stateCookieName, value)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     stateCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   10 * 60,
	}
	http.SetCookie(w, cookie)
	return nil
}

// readStateCookie は state/nonce の一時Cookieを読み出す。
func (a *OIDCAuth) readStateCookie(r *http.Request) (oidcState, bool) {
	c, err := r.Cookie(stateCookieName)
	if err != nil {
		return oidcState{}, false
	}

	var value map[string]string
	if err := a.sc.Decode(stateCookieName, c.Value, &value); err != nil {
		return oidcState{}, false
	}

	return oidcState{State: value["state"], Nonce: value["nonce"]}, true
}

// clearStateCookie は state/nonce の一時Cookieを削除する。
func (a *OIDCAuth) clearStateCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// writeSession はセッション情報を署名/暗号化してCookieに書き込む。
func (a *OIDCAuth) writeSession(w http.ResponseWriter, r *http.Request, sess sessionData) error {
	value := map[string]string{
		"sub":     sess.Sub,
		"user_id": sess.UserID,
		"org_id":  sess.OrgID,
		"role":    sess.Role,
		"email":   sess.Email,
		"name":    sess.Name,
		"iat":     fmt.Sprintf("%d", sess.Iat),
	}
	encoded, err := a.sc.Encode(sessionCookieName, value)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60,
	}
	http.SetCookie(w, cookie)
	return nil
}

// readSession はセッションCookieを読み取り、復号/検証して返す。
func (a *OIDCAuth) readSession(r *http.Request) (sessionData, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return sessionData{}, false
	}

	var value map[string]string
	if err := a.sc.Decode(sessionCookieName, c.Value, &value); err != nil {
		return sessionData{}, false
	}

	var iat int64
	_, _ = fmt.Sscanf(value["iat"], "%d", &iat)

	return sessionData{
		Sub:    value["sub"],
		UserID: value["user_id"],
		OrgID:  value["org_id"],
		Role:   value["role"],
		Email:  value["email"],
		Name:   value["name"],
		Iat:    iat,
	}, true
}

// clearSession はセッションCookieを削除する。
func (a *OIDCAuth) clearSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// defaultScopes は設定値から OIDC のスコープ配列を作る。
func defaultScopes(v string) []string {
	if strings.TrimSpace(v) == "" {
		return []string{"openid", "profile", "email"}
	}
	parts := strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

// randomToken は URL-safe なランダムトークンを生成する。
func randomToken(n int) string {
	b := mustRandomBytes(n)
	return base64.RawURLEncoding.EncodeToString(b)
}

// mustRandomBytes は暗号学的に安全な乱数を生成し、失敗時は panic する。
func mustRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// isHTTPS はリクエストが HTTPS 相当かどうかを判定する（TLS または X-Forwarded-Proto）。
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	return strings.EqualFold(proto, "https")
}

// inferExternalURL はリクエスト情報から外部向けベースURLを推測する。
func inferExternalURL(r *http.Request) string {
	scheme := "http"
	if isHTTPS(r) {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// htmlEscape は最小限のHTMLエスケープを行う（MVP用）。
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
