// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitCurrentMatchRetryReusesPostedResult(t *testing.T) {
	web := setupTestWeb(t)
	match := &model.Match{Type: model.Practice, ShortName: "P1", LongName: "Practice 1", StartedAt: time.Now().UTC()}
	require.NoError(t, web.arena.Database.CreateMatch(match))
	web.arena.CurrentMatch = match
	first := web.getCurrentMatchResult()
	require.NotEmpty(t, first.PostAttemptId)
	require.NoError(t, web.commitMatchScore(match, first, false))
	require.Positive(t, first.Id)
	assert.Equal(t, 1, first.PlayNumber)

	// Simulate losing process memory after the result row was persisted but
	// before the remainder of the post operation completed.
	reloaded, err := web.arena.Database.GetMatchById(match.Id)
	require.NoError(t, err)
	web.arena.CurrentMatch = reloaded
	second := web.getCurrentMatchResult()
	require.Equal(t, first.PostAttemptId, second.PostAttemptId)
	require.NoError(t, web.commitMatchScore(reloaded, second, false))
	assert.Equal(t, first.Id, second.Id)
	assert.Equal(t, first.PlayNumber, second.PlayNumber)
	stored, err := web.arena.Database.GetMatchResultByPostAttemptId(match.Id, first.PostAttemptId)
	require.NoError(t, err)
	assert.Equal(t, first.Id, stored.Id)
	assert.Equal(t, 1, stored.PlayNumber)

	// A deliberate replay has a new persisted start time and must still
	// create the next play number rather than reusing the old result.
	reloaded.StartedAt = reloaded.StartedAt.Add(time.Second)
	replay := web.getCurrentMatchResult()
	require.NotEqual(t, first.PostAttemptId, replay.PostAttemptId)
	require.NoError(t, web.commitMatchScore(reloaded, replay, false))
	assert.NotEqual(t, first.Id, replay.Id)
	assert.Equal(t, 2, replay.PlayNumber)
}
