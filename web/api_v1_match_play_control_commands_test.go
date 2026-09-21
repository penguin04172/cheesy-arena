// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPlayControlCommandStateTransitions(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.executeMatchPlayControlCommand("setAudienceDisplay", "logo"))
	assert.Equal(t, "logo", web.arena.AudienceDisplayMode)
	require.NoError(t, web.executeMatchPlayControlCommand("setAllianceStationDisplay", "match"))
	assert.Equal(t, "match", web.arena.AllianceStationDisplayMode)
	require.NoError(t, web.executeMatchPlayControlCommand("toggleBypass", "R3"))
	assert.True(t, web.arena.AllianceStations["R3"].Bypass)
	require.NoError(t, web.executeMatchPlayControlCommand("signalVolunteers", nil))
	assert.True(t, web.arena.FieldVolunteers)
	require.NoError(t, web.executeMatchPlayControlCommand("signalReset", nil))
	assert.True(t, web.arena.FieldReset)
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("setAudienceDisplay", "invalid"), errInvalidMatchPlayControlValue)
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("setAllianceStationDisplay", "invalid"), errInvalidMatchPlayControlValue)
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("toggleBypass", "R4"), errInvalidMatchPlayControlValue)
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("startTimeout", map[string]any{"durationSec": 1.5}), errInvalidMatchPlayControlValue)
	web.arena.MatchState = field.AutoPeriod
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("signalReset", nil), errInvalidMatchPlayControlState)
	assert.ErrorIs(t, web.executeMatchPlayControlCommand("startTimeout", float64(60)), errInvalidMatchPlayControlState)
}

func TestApiV1MatchPlayControlIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/match-play/commands"
	body := `{"command":"toggleBypass","value":"R3"}`
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "bypass-1", body)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.True(t, web.arena.AllianceStations["R3"].Bypass)
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "bypass-1", body)
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.True(t, web.arena.AllianceStations["R3"].Bypass)
	conflict := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "bypass-1", `{"command":"toggleBypass","value":"B3"}`)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	invalid := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "bypass-2", `{"command":"setAllianceStationDisplay","value":"unsafe"}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	unknown := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "bypass-3", `{"command":"startMatch"}`)
	assert.Equal(t, http.StatusBadRequest, unknown.Code)
}
