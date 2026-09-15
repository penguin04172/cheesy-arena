// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearTournamentDataIsStageScoped(t *testing.T) {
	database := setupTestDb(t)
	practice := Match{Type: Practice}
	qualification := Match{Type: Qualification}
	playoff := Match{Type: Playoff}
	require.NoError(t, database.CreateMatch(&practice))
	require.NoError(t, database.CreateMatch(&qualification))
	require.NoError(t, database.CreateMatch(&playoff))
	require.NoError(t, database.CreateMatchResult(&MatchResult{MatchId: qualification.Id}))
	require.NoError(t, database.CreateMatchResult(&MatchResult{MatchId: playoff.Id}))
	require.NoError(t, database.CreateScheduledBreak(&ScheduledBreak{MatchType: Qualification}))
	require.NoError(t, database.CreateScheduledBreak(&ScheduledBreak{MatchType: Playoff}))
	require.NoError(t, database.CreateRanking(&game.Ranking{TeamId: 254}))
	require.NoError(t, database.CreateAlliance(&Alliance{Id: 1}))

	result, err := database.ClearTournamentData(Qualification)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Matches)
	assert.Equal(t, 1, result.MatchResults)
	assert.Equal(t, 1, result.ScheduledBreaks)
	assert.Equal(t, 1, result.Rankings)
	assert.Zero(t, result.Alliances)
	qualificationMatches, _ := database.GetMatchesByType(Qualification, true)
	playoffMatches, _ := database.GetMatchesByType(Playoff, true)
	alliances, _ := database.GetAllAlliances()
	assert.Empty(t, qualificationMatches)
	assert.Len(t, playoffMatches, 1)
	assert.Len(t, alliances, 1)
}
