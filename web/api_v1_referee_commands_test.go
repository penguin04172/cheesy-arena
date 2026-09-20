// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefereeCommandServiceStateTransitions(t *testing.T) {
	web := setupTestWeb(t)
	first, err := web.executeRefereeScoreCommand("addFoul", map[string]any{"Alliance": "red", "IsMajor": true})
	require.NoError(t, err)
	assert.True(t, first.Changed)
	assert.Equal(t, 1, first.FoulId)
	second, err := web.executeRefereeScoreCommand("addFoul", map[string]any{"Alliance": "blue", "IsMajor": false})
	require.NoError(t, err)
	assert.Equal(t, 2, second.FoulId)
	_, err = web.executeRefereeScoreCommand("updateFoulTeam", map[string]any{"Alliance": "red", "Index": 0, "TeamId": 254})
	require.NoError(t, err)
	assert.Equal(t, 254, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].TeamId)
	_, err = web.executeRefereeScoreCommand("toggleFoulType", map[string]any{"Alliance": "red", "Index": 0})
	require.NoError(t, err)
	assert.False(t, web.arena.RedRealtimeScore.CurrentScore.Fouls[0].IsMajor)
	noChange, err := web.executeRefereeScoreCommand("deleteFoul", map[string]any{"Alliance": "red", "Index": 9})
	require.NoError(t, err)
	assert.False(t, noChange.Changed)
	web.arena.CurrentMatch.Type = model.Playoff
	web.arena.CurrentMatch.Blue1, web.arena.CurrentMatch.Blue2, web.arena.CurrentMatch.Blue3 = 1, 2, 3
	_, err = web.executeRefereeScoreCommand("card", map[string]any{"Alliance": "blue", "TeamId": 2, "Card": "yellow"})
	require.NoError(t, err)
	for _, team := range []string{"1", "2", "3"} {
		assert.Equal(t, "yellow", web.arena.BlueRealtimeScore.Cards[team])
	}
	web.arena.RedRealtimeScore.FoulsCommitted = true
	_, err = web.executeRefereeScoreCommand("addFoul", map[string]any{"Alliance": "red"})
	assert.ErrorIs(t, err, errRefereeScoreCommitted)
	web.arena.MatchState = field.TimeoutActive
	_, err = web.executeRefereeScoreCommand("addFoul", map[string]any{"Alliance": "blue"})
	assert.ErrorIs(t, err, errInvalidRefereeValue)
}

func TestApiV1RefereeCommandsIdempotencyAndValidation(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/referee/commands"
	body := `{"command":"addFoul","payload":{"alliance":"red","isMajor":true}}`
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "foul-1", body)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	assert.Contains(t, first.Body.String(), `"foulId":1`)
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "foul-1", body)
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.Len(t, web.arena.RedRealtimeScore.CurrentScore.Fouls, 1)
	conflict := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "foul-1", `{"command":"addFoul","payload":{"alliance":"blue"}}`)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	invalid := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "foul-2", `{"command":"card","payload":{"alliance":"red","teamId":254,"card":"purple"}}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	web.arena.RedRealtimeScore.FoulsCommitted = true
	committed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "foul-3", body)
	assert.Equal(t, http.StatusConflict, committed.Code)
}
