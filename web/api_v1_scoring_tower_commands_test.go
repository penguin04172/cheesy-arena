// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoringTowerCommandService(t *testing.T) {
	web := setupTestWeb(t)
	changed, err := web.executeScoringTowerCommand("red", "autoTower", map[string]any{"TeamPosition": 2, "AutoTowerStatus": 3})
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, game.TowerLevel3, web.arena.RedRealtimeScore.CurrentScore.AutoTowerStatuses[1])
	changed, err = web.executeScoringTowerCommand("red", "autoTower", map[string]any{"TeamPosition": 2, "AutoTowerStatus": 3})
	require.NoError(t, err)
	assert.False(t, changed)
	for _, input := range []map[string]any{
		{"TeamPosition": 0, "AutoTowerStatus": 1},
		{"TeamPosition": 4, "AutoTowerStatus": 1},
		{"TeamPosition": 1, "AutoTowerStatus": 4},
	} {
		_, err := web.executeScoringTowerCommand("red", "autoTower", input)
		assert.ErrorIs(t, err, errInvalidScoringTowerValue)
	}
	web.arena.MatchState = field.TimeoutActive
	_, err = web.executeScoringTowerCommand("red", "endgame", map[string]any{"TeamPosition": 1, "EndgameTowerStatus": 1})
	assert.ErrorIs(t, err, errInvalidScoringTowerValue)
}

func TestApiV1ScoringTowerCommandsIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/scoring/red/commands"
	body := `{"command":"autoTower","payload":{"teamPosition":1,"autoTowerStatus":2}}`
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "tower-1", body)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.Contains(t, first.Body.String(), `"changed":true`)
	assert.Equal(t, game.TowerLevel2, web.arena.RedRealtimeScore.CurrentScore.AutoTowerStatuses[0])
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "tower-1", body)
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	conflict := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "tower-1", `{"command":"autoTower","payload":{"teamPosition":1,"autoTowerStatus":3}}`)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	invalid := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "tower-2", `{"command":"endgame","payload":{"teamPosition":1,"endgameTowerStatus":4}}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	missing := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/scoring/green/commands", "tower-3", body)
	assert.Equal(t, http.StatusNotFound, missing.Code)
}
