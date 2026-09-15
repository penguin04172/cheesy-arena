// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminDatabaseBackupIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/database/backups", "backup-1", "")
	assert.Equal(t, http.StatusCreated, first.Code)
	assert.Contains(t, first.Body.String(), `"created":true`)
	second := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/database/backups", "backup-1", "")
	assert.Equal(t, http.StatusOK, second.Code)
	assert.Contains(t, second.Body.String(), `"replayed":true`)
}

func TestApiV1AdminTournamentDataClear(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{Type: model.Qualification}
	require.NoError(t, web.arena.Database.CreateMatch(&match))
	require.NoError(t, web.arena.Database.CreateMatchResult(&model.MatchResult{MatchId: match.Id}))
	require.NoError(t, web.arena.Database.CreateScheduledBreak(&model.ScheduledBreak{MatchType: model.Qualification}))
	require.NoError(t, web.arena.Database.CreateRanking(&game.Ranking{TeamId: 254}))
	body := `{"eventId":` + strconv.Itoa(web.arena.EventSettings.Id) + `,"confirmation":"clear:qualification"}`
	first := apiV1AdminIdempotentRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/qualification", "clear-q-1", body)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.Contains(t, first.Body.String(), `"matches":1`)
	assert.Contains(t, first.Body.String(), `"rankings":1`)
	matches, err := web.arena.Database.GetMatchesByType(model.Qualification, true)
	require.NoError(t, err)
	assert.Empty(t, matches)

	replayed := apiV1AdminIdempotentRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/qualification", "clear-q-1", body)
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.Contains(t, replayed.Body.String(), `"matches":1`)
}

func TestApiV1AdminTournamentDataClearSafetyChecks(t *testing.T) {
	web := setupTestWeb(t)
	missingKey := apiV1AdminJsonRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/practice", `{"eventId":1,"confirmation":"clear:practice"}`)
	assert.Equal(t, http.StatusBadRequest, missingKey.Code)
	wrongConfirmation := apiV1AdminIdempotentRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/practice", "clear-p-1", `{"eventId":1,"confirmation":"clear:playoff"}`)
	assert.Equal(t, http.StatusUnprocessableEntity, wrongConfirmation.Code)
	invalidType := apiV1AdminIdempotentRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/test", "clear-test", `{"eventId":1,"confirmation":"clear:test"}`)
	assert.Equal(t, http.StatusBadRequest, invalidType.Code)
}

func TestApiV1AdminIdempotencyKeyCannotChangeOperation(t *testing.T) {
	web := setupTestWeb(t)
	backup := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/database/backups", "shared-key", "")
	require.Equal(t, http.StatusCreated, backup.Code)
	body := `{"eventId":` + strconv.Itoa(web.arena.EventSettings.Id) + `,"confirmation":"clear:practice"}`
	clear := apiV1AdminIdempotentRequest(web, http.MethodDelete, "/api/v1/admin/tournament-data/practice", "shared-key", body)
	assert.Equal(t, http.StatusConflict, clear.Code)
	assert.Contains(t, clear.Body.String(), "idempotency_key_reused")
}

func apiV1AdminIdempotentRequest(web *Web, method, path, key, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", "test-csrf-token")
	request.Header.Set("Idempotency-Key", key)
	request.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: "test-csrf-token"})
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	return recorder
}
