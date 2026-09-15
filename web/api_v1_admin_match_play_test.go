// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminMatchPlayMatchesMatchesLegacyFragmentData(t *testing.T) {
	web := setupTestWeb(t)
	scheduled := model.Match{Type: model.Practice, ShortName: "P2", Time: time.Now().UTC(), Status: game.MatchScheduled}
	completed := model.Match{Type: model.Practice, ShortName: "P1", Time: time.Now().UTC().Add(-time.Hour), Status: game.RedWonMatch}
	qualification := model.Match{Type: model.Qualification, ShortName: "Q1", Time: time.Now().UTC(), Status: game.BlueWonMatch}
	require.NoError(t, web.arena.Database.CreateMatch(&completed))
	require.NoError(t, web.arena.Database.CreateMatch(&scheduled))
	require.NoError(t, web.arena.Database.CreateMatch(&qualification))

	legacy := web.getHttpResponse("/match_play/match_load")
	require.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "true", legacy.Header().Get("Deprecation"))
	assert.Contains(t, legacy.Header().Get("Link"), "/api/v1/admin/match-play/matches")
	for _, name := range []string{"P1", "P2", "Q1"} {
		assert.Contains(t, legacy.Body.String(), name)
	}

	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/match-play/matches", "")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data apiV1MatchPlayMatches `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "practice", response.Data.CurrentMatchType)
	require.Len(t, response.Data.MatchesByType["practice"], 2)
	assert.Equal(t, "P2", response.Data.MatchesByType["practice"][0].ShortName)
	assert.False(t, response.Data.MatchesByType["practice"][0].CanShowResult)
	assert.Equal(t, "P1", response.Data.MatchesByType["practice"][1].ShortName)
	assert.True(t, response.Data.MatchesByType["practice"][1].CanShowResult)
	assert.Equal(t, "red", response.Data.MatchesByType["practice"][1].ColorClass)
	assert.Len(t, response.Data.MatchesByType["qualification"], 1)
	assert.Empty(t, response.Data.MatchesByType["playoff"])
}
