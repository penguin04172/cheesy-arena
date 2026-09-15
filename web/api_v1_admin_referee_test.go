// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminRefereeFoulsMatchesLegacyFragmentData(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.CurrentMatch.Id = 42
	web.arena.CurrentMatch.Red1, web.arena.CurrentMatch.Red2, web.arena.CurrentMatch.Red3 = 254, 1114, 2056
	web.arena.CurrentMatch.Blue1, web.arena.CurrentMatch.Blue2, web.arena.CurrentMatch.Blue3 = 1678, 1679, 1680
	web.arena.RedRealtimeScore.CurrentScore.Fouls = []game.Foul{{FoulId: 8, IsMajor: true, TeamId: 254, RuleId: 2}}
	web.arena.BlueRealtimeScore.CurrentScore.Fouls = []game.Foul{{FoulId: 9, TeamId: 1680, RuleId: 1}}

	legacy := web.getHttpResponse("/panels/referee/foul_list")
	require.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "true", legacy.Header().Get("Deprecation"))
	assert.Contains(t, legacy.Header().Get("Link"), "/api/v1/admin/referee/fouls")
	for _, value := range []string{"254", "1680", "G210", "G206"} {
		assert.Contains(t, legacy.Body.String(), value)
	}

	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/referee/fouls", "")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data apiV1RefereeFouls `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, 42, response.Data.MatchId)
	assert.Equal(t, []int{254, 1114, 2056}, response.Data.Red.TeamIds)
	assert.Equal(t, []int{1678, 1679, 1680}, response.Data.Blue.TeamIds)
	require.Len(t, response.Data.Red.Fouls, 1)
	assert.Equal(t, apiV1RefereeFoul{Index: 0, FoulId: 8, IsMajor: true, TeamId: 254, RuleId: 2}, response.Data.Red.Fouls[0])
	require.NotEmpty(t, response.Data.Rules)
	for index := 1; index < len(response.Data.Rules); index++ {
		assert.Less(t, response.Data.Rules[index-1].Id, response.Data.Rules[index].Id)
	}
}

func TestApiV1AdminRefereeFoulsReturnsEmptyArrays(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/referee/fouls", "")
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"fouls":[]`)
	assert.Contains(t, recorder.Body.String(), `"teamIds":[0,0,0]`)
}
