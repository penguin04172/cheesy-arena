// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPlayLifecycleCommandStateValidation(t *testing.T) {
	web := setupTestWeb(t)
	current := web.arena.CurrentMatch
	err := web.executeMatchPlayLifecycleCommand("loadMatch", map[string]any{"MatchId": 999999}, false)
	require.Error(t, err)
	assert.Same(t, current, web.arena.CurrentMatch, "invalid match ID must not reset the field")
	assert.Equal(t, field.PreMatch, web.arena.MatchState)
	assert.ErrorIs(t, web.executeMatchPlayLifecycleCommand("loadMatch", map[string]any{"MatchId": -1}, false), errInvalidMatchLifecycleValue)
	assert.ErrorIs(t, web.executeMatchPlayLifecycleCommand("showResult", map[string]any{"MatchId": -1}, false), errInvalidMatchLifecycleValue)
	assert.ErrorIs(t, web.executeMatchPlayLifecycleCommand("unknown", nil, false), errInvalidMatchLifecycleCommand)
	assert.Error(t, web.executeMatchPlayLifecycleCommand("abortMatch", nil, false))
	assert.Error(t, web.executeMatchPlayLifecycleCommand("commitAndPost", nil, false))
	assert.False(t, web.arena.RedRealtimeScore.FoulsCommitted)
	assert.Error(t, web.executeMatchPlayLifecycleCommand("commitAndPost", nil, true))
	assert.False(t, web.arena.RedRealtimeScore.FoulsCommitted, "referee must not commit fouls before post-match")
}

func TestMatchPlayLifecycleFailedStartRestoresMutePreference(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.MuteMatchSounds = false
	err := web.executeMatchPlayLifecycleCommand("startMatch", map[string]any{"MuteMatchSounds": true}, false)
	require.Error(t, err)
	assert.Equal(t, field.PreMatch, web.arena.MatchState)
	assert.False(t, web.arena.MuteMatchSounds)
}
