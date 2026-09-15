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

func TestApiV1AnnouncerDisplayScore(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{Id: 7, Type: model.Qualification, LongName: "Qual 17", Status: game.RedWonMatch, Red1: 254, Blue1: 1114}
	result := model.NewMatchResult()
	result.RedScore.Fouls = []game.Foul{{IsMajor: true, TeamId: 254, RuleId: 2}}
	result.RedCards["254"] = "yellow"
	web.arena.SavedMatch, web.arena.SavedMatchResult = &match, result
	web.arena.SavedRankings = game.Rankings{{TeamId: 254, Rank: 2, PreviousRank: 4}}

	legacy := web.getHttpResponse("/displays/announcer/score_posted")
	require.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "true", legacy.Header().Get("Deprecation"))
	assert.Contains(t, legacy.Header().Get("Link"), "/api/v1/displays/announcer/score")

	recorder := web.getHttpResponse("/api/v1/displays/announcer/score")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data apiV1AnnouncerScore `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, "red", response.Data.Winner)
	assert.Equal(t, "bg-danger", response.Data.WinnerClass)
	assert.Equal(t, "qualification", response.Data.MatchType)
	require.Len(t, response.Data.Red.Fouls, 1)
	assert.Equal(t, "G210", response.Data.Red.Fouls[0].RuleNumber)
	assert.NotEmpty(t, response.Data.Red.Fouls[0].RuleDescription)
	assert.Equal(t, []apiV1AnnouncerCard{{TeamId: 254, Card: "yellow"}}, response.Data.Red.Cards)
	assert.Equal(t, []apiV1AnnouncerRanking{{TeamId: 254, Rank: 2, PreviousRank: 4}}, response.Data.Red.Rankings)
}
