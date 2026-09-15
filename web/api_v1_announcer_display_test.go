// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AnnouncerDisplayMatchMatchesLegacyFragmentData(t *testing.T) {
	web := setupTestWeb(t)
	for _, team := range []*model.Team{
		{Id: 254, Nickname: "The Cheesy Poofs", SchoolName: "Bellarmine", City: "San Jose", StateProv: "CA", Country: "USA", RookieYear: 1999, RobotName: "Robot", Accomplishments: "Awards"},
		{Id: 1114, Nickname: "Simbotics"},
		{Id: 2056, Nickname: "OP Robotics"},
		{Id: 9999, Nickname: "Backup"},
	} {
		require.NoError(t, web.arena.Database.CreateTeam(team))
	}
	require.NoError(t, web.arena.Database.CreateRanking(&game.Ranking{TeamId: 254, Rank: 3}))
	require.NoError(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1, TeamIds: []int{254, 1114, 9999}}))
	match := model.Match{Type: model.Playoff, LongName: "Final 1", Red1: 254, Red2: 1114, Blue3: 2056, PlayoffRedAlliance: 1}
	require.NoError(t, web.arena.LoadMatch(&match))

	legacy := web.getHttpResponse("/displays/announcer/match_load")
	require.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "true", legacy.Header().Get("Deprecation"))
	assert.Contains(t, legacy.Header().Get("Link"), "/api/v1/displays/announcer/match")
	for _, value := range []string{"254", "The Cheesy Poofs", "9999", "not on field"} {
		assert.Contains(t, legacy.Body.String(), value)
	}

	recorder := web.getHttpResponse("/api/v1/displays/announcer/match")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data apiV1AnnouncerMatch `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "playoff", response.Data.Type)
	assert.Equal(t, "Final 1", response.Data.LongName)
	require.Len(t, response.Data.Red.Teams, 4)
	assert.Equal(t, 254, response.Data.Red.Teams[0].Id)
	require.NotNil(t, response.Data.Red.Teams[0].Rank)
	assert.Equal(t, 3, *response.Data.Red.Teams[0].Rank)
	assert.Nil(t, response.Data.Red.Teams[2])
	assert.Equal(t, 9999, response.Data.Red.Teams[3].Id)
	assert.True(t, response.Data.Red.Teams[3].IsOffField)
	require.Len(t, response.Data.Blue.Teams, 3)
	assert.Nil(t, response.Data.Blue.Teams[0])
	assert.Equal(t, 2056, response.Data.Blue.Teams[2].Id)
}
