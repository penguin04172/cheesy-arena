// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllianceSelectionCommandTransitions(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.executeAllianceSelectionCommand("setTimer", float64(3)))
	require.NoError(t, web.executeAllianceSelectionCommand("startTimer", nil))
	assert.True(t, web.arena.AllianceSelectionShowTimer)
	assert.Equal(t, 3, web.arena.AllianceSelectionTimeRemainingSec)
	require.NoError(t, web.executeAllianceSelectionCommand("startTimer", nil))
	assert.Equal(t, 3, web.arena.AllianceSelectionTimeRemainingSec)
	require.NoError(t, web.executeAllianceSelectionCommand("stopTimer", nil))
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, 3, web.arena.AllianceSelectionTimeRemainingSec)
	require.NoError(t, web.executeAllianceSelectionCommand("restartTimer", nil))
	assert.Equal(t, 3, web.arena.AllianceSelectionTimeRemainingSec)
	require.NoError(t, web.executeAllianceSelectionCommand("startTimer", nil))
	require.Eventually(t, func() bool { return web.arena.AllianceSelectionTimeRemainingSec == 2 }, 2*time.Second, 20*time.Millisecond)
	require.NoError(t, web.executeAllianceSelectionCommand("hideTimer", nil))
	assert.False(t, web.arena.AllianceSelectionShowTimer)
	assert.Zero(t, web.arena.AllianceSelectionTimeRemainingSec)
}

func TestAllianceSelectionCommandValidation(t *testing.T) {
	web := setupTestWeb(t)
	for _, value := range []any{float64(-1), float64(3601), float64(1.5), "5", nil} {
		assert.ErrorIs(t, web.executeAllianceSelectionCommand("setTimer", value), errInvalidAllianceSelectionValue)
	}
	assert.ErrorIs(t, web.executeAllianceSelectionCommand("startTimer", "unexpected"), errInvalidAllianceSelectionValue)
	assert.ErrorIs(t, web.executeAllianceSelectionCommand("setAudienceDisplay", "unknown"), errInvalidAllianceSelectionValue)
	assert.ErrorIs(t, web.executeAllianceSelectionCommand("missing", nil), errInvalidAllianceSelectionCommand)
	assert.Equal(t, 45, web.allianceSelectionCommands.timeLimit())
}

func TestApiV1AllianceSelectionCommandIdempotencyAndValidation(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/alliance-selection/commands"
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "timer-command-1", `{"command":"setTimer","value":90}`)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.Equal(t, 90, web.allianceSelectionCommands.timeLimit())
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "timer-command-1", `{"command":"setTimer","value":90}`)
	require.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	conflict := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "timer-command-1", `{"command":"setTimer","value":100}`)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	invalid := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "timer-command-2", `{"command":"setTimer","value":-5}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	unknown := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "timer-command-3", `{"command":"unknown"}`)
	assert.Equal(t, http.StatusBadRequest, unknown.Code)
	display := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "display-command-1", `{"command":"setAudienceDisplay","value":"logo"}`)
	assert.Equal(t, http.StatusOK, display.Code, display.Body.String())
	assert.Equal(t, "logo", web.arena.AudienceDisplayMode)
}

func TestApiV1AllianceSelectionCommandRequiresCsrfAndAdmin(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/alliance-selection/commands"
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"command":"setTimer","value":90}`))
	request.Header.Set("Idempotency-Key", "security-1")
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	web.arena.EventSettings.AdminPassword = "required"
	request = httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"command":"setTimer","value":90}`))
	request.Header.Set("Idempotency-Key", "security-2")
	recorder = httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
