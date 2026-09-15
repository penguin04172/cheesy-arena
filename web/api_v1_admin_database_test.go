// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"bytes"
	"io"
	"mime/multipart"
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

func TestApiV1AdminDatabaseRestoreAndReplay(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.Name = "Restored Event"
	require.NoError(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))
	backup := new(bytes.Buffer)
	require.NoError(t, web.arena.Database.WriteBackup(backup))
	web.arena.EventSettings.Name = "Current Event"
	require.NoError(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))
	require.NoError(t, web.arena.LoadSettings())

	fields := map[string]string{"eventId": strconv.Itoa(web.arena.EventSettings.Id), "confirmation": "restore"}
	restored := apiV1AdminRestoreRequest(web, "restore-1", fields, backup.Bytes())
	require.Equal(t, http.StatusOK, restored.Code, restored.Body.String())
	assert.Equal(t, "Restored Event", web.arena.EventSettings.Name)
	assert.Contains(t, restored.Body.String(), `"restored":true`)

	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/database/restore", "restore-1", "")
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
}

func TestApiV1AdminDatabaseRestoreRejectsInvalidUpload(t *testing.T) {
	web := setupTestWeb(t)
	fields := map[string]string{"eventId": strconv.Itoa(web.arena.EventSettings.Id), "confirmation": "restore"}
	recorder := apiV1AdminRestoreRequest(web, "restore-invalid", fields, []byte("invalid"))
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid_database_backup")
}

func TestApiV1AdminDatabaseRestoreRequiresConfirmation(t *testing.T) {
	web := setupTestWeb(t)
	backup := new(bytes.Buffer)
	require.NoError(t, web.arena.Database.WriteBackup(backup))
	recorder := apiV1AdminRestoreRequest(web, "restore-confirm", map[string]string{"eventId": "1"}, backup.Bytes())
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "confirmation_mismatch")
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

func apiV1AdminRestoreRequest(web *Web, key string, fields map[string]string, contents []byte) *httptest.ResponseRecorder {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for name, value := range fields {
		_ = writer.WriteField(name, value)
	}
	part, _ := writer.CreateFormFile("databaseFile", "backup.db")
	_, _ = io.Copy(part, bytes.NewReader(contents))
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/database/restore", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", "test-csrf-token")
	request.Header.Set("Idempotency-Key", key)
	request.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: "test-csrf-token"})
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	return recorder
}
