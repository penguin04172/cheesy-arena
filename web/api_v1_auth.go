// Copyright 2026 Team 254. All Rights Reserved.
//
// Authentication and CSRF protection for version 1 of the JSON API.

package web

import (
	"crypto/subtle"
	"github.com/google/uuid"
	"net/http"
)

const csrfTokenCookie = "csrf_token"

type apiV1Session struct {
	Authenticated bool    `json:"authenticated"`
	Admin         bool    `json:"admin"`
	Username      *string `json:"username"`
	CsrfToken     string  `json:"csrfToken"`
}

func (web *Web) apiV1SessionHandler(w http.ResponseWriter, r *http.Request) {
	csrfToken := ensureCsrfTokenCookie(w, r)
	session := web.getUserSessionFromCookie(r)
	response := apiV1Session{CsrfToken: csrfToken}
	if web.arena.EventSettings.AdminPassword == "" {
		response.Authenticated = true
		response.Admin = true
	} else if session != nil {
		response.Authenticated = true
		response.Admin = session.Username == adminUser
		response.Username = &session.Username
	}
	w.Header().Set("Cache-Control", "no-store")
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if web.arena.EventSettings.AdminPassword == "" {
			next.ServeHTTP(w, r)
			return
		}
		session := web.getUserSessionFromCookie(r)
		if session == nil {
			writeApiV1Error(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required.", nil)
			return
		}
		if session.Username != adminUser {
			writeApiV1Error(w, r, http.StatusForbidden, "admin_required", "Administrator access is required.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (web *Web) apiV1RequireCsrf(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(csrfTokenCookie)
		headerToken := r.Header.Get("X-CSRF-Token")
		if err != nil || cookie.Value == "" || headerToken == "" ||
			subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
			writeApiV1Error(w, r, http.StatusForbidden, "invalid_csrf_token", "A valid CSRF token is required.", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ensureCsrfTokenCookie(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(csrfTokenCookie); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	token := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    token,
		Path:     "/",
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return token
}
