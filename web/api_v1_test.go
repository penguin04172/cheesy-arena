// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestApiV1Event(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.Name = "API Test Event"
	web.arena.EventSettings.AdminPassword = "do-not-expose-admin-password"
	web.arena.EventSettings.TbaSecret = "do-not-expose-tba-secret"

	recorder := web.getHttpResponse("/api/v1/event")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
	assert.NotContains(t, recorder.Body.String(), "do-not-expose-admin-password")
	assert.NotContains(t, recorder.Body.String(), "do-not-expose-tba-secret")

	var response struct {
		Data apiV1PublicEvent `json:"data"`
		Meta apiV1Meta        `json:"meta"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "API Test Event", response.Data.Name)
	assert.Equal(t, "double_elimination", response.Data.PlayoffType)
	assert.NotEmpty(t, response.Meta.RequestId)
}

func TestApiV1SessionUnauthenticated(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.AdminPassword = "password"

	recorder := web.getHttpResponse("/api/v1/session")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))

	var response struct {
		Data apiV1Session `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.False(t, response.Data.Authenticated)
	assert.False(t, response.Data.Admin)
	assert.Nil(t, response.Data.Username)
	assert.NotEmpty(t, response.Data.CsrfToken)

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, csrfTokenCookie, cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
}

func TestApiV1SessionAuthenticated(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.AdminPassword = "password"
	session := model.UserSession{Token: "session-token", Username: adminUser, CreatedAt: time.Now()}
	require.NoError(t, web.arena.Database.CreateUserSession(&session))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/session", nil)
	request.AddCookie(&http.Cookie{Name: sessionTokenCookie, Value: session.Token})
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)

	var response struct {
		Data apiV1Session `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.True(t, response.Data.Authenticated)
	assert.True(t, response.Data.Admin)
	require.NotNil(t, response.Data.Username)
	assert.Equal(t, adminUser, *response.Data.Username)
}

func TestApiV1AdminAndCsrfMiddleware(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.AdminPassword = "password"
	protected := web.apiV1Middleware(web.apiV1RequireAdmin(web.apiV1RequireCsrf(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			writeApiV1Data(w, r, http.StatusOK, map[string]bool{"updated": true}, nil)
		},
	))))

	unauthenticated := httptest.NewRecorder()
	protected.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodPost, "/api/v1/test", nil))
	assert.Equal(t, http.StatusUnauthorized, unauthenticated.Code)
	assert.Contains(t, unauthenticated.Body.String(), "authentication_required")

	session := model.UserSession{Token: "session-token", Username: adminUser, CreatedAt: time.Now()}
	require.NoError(t, web.arena.Database.CreateUserSession(&session))
	missingCsrfRequest := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	missingCsrfRequest.AddCookie(&http.Cookie{Name: sessionTokenCookie, Value: session.Token})
	missingCsrf := httptest.NewRecorder()
	protected.ServeHTTP(missingCsrf, missingCsrfRequest)
	assert.Equal(t, http.StatusForbidden, missingCsrf.Code)
	assert.Contains(t, missingCsrf.Body.String(), "invalid_csrf_token")

	validRequest := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	validRequest.AddCookie(&http.Cookie{Name: sessionTokenCookie, Value: session.Token})
	validRequest.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: "csrf-token"})
	validRequest.Header.Set("X-CSRF-Token", "csrf-token")
	valid := httptest.NewRecorder()
	protected.ServeHTTP(valid, validRequest)
	assert.Equal(t, http.StatusOK, valid.Code)
}

func TestDecodeApiV1Json(t *testing.T) {
	for _, testCase := range []struct {
		name string
		body string
		ok   bool
	}{
		{name: "valid", body: `{"name":"value"}`, ok: true},
		{name: "unknown field", body: `{"unknown":true}`, ok: false},
		{name: "trailing value", body: `{"name":"value"} {}`, ok: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/test", strings.NewReader(testCase.body))
			recorder := httptest.NewRecorder()
			var destination struct {
				Name string `json:"name"`
			}
			err := decodeApiV1Json(recorder, request, &destination)
			if testCase.ok {
				assert.NoError(t, err)
				assert.Equal(t, "value", destination.Name)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
