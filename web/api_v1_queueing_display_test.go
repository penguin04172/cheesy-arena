// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1QueueingDisplayMatchesMatchesLegacyFragmentData(t *testing.T) {
	web := setupTestWeb(t)
	now := time.Now().UTC().Truncate(time.Second)
	completed := model.Match{Type: model.Qualification, TypeOrder: 1, ShortName: "Q1", Time: now, Status: game.RedWonMatch}
	current := model.Match{Type: model.Qualification, TypeOrder: 2, ShortName: "Q2", Time: now.Add(time.Minute), Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6, Status: game.MatchScheduled}
	next := model.Match{Type: model.Qualification, TypeOrder: 3, ShortName: "Q3", Time: now.Add(2 * time.Minute), Red1: 7, Blue1: 8, Status: game.MatchScheduled}
	for _, match := range []*model.Match{&completed, &current, &next} {
		require.NoError(t, web.arena.Database.CreateMatch(match))
	}
	web.arena.CurrentMatch = &current

	legacy := web.getHttpResponse("/displays/queueing/match_load")
	require.Equal(t, http.StatusOK, legacy.Code)
	assert.Equal(t, "true", legacy.Header().Get("Deprecation"))
	assert.Contains(t, legacy.Header().Get("Link"), "/api/v1/displays/queueing/matches")
	assert.NotContains(t, legacy.Body.String(), "Q1")
	assert.Contains(t, legacy.Body.String(), "Q2")
	assert.Contains(t, legacy.Body.String(), "Q3")

	recorder := web.getHttpResponse("/api/v1/displays/queueing/matches")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data []apiV1QueueingMatch `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 2)
	assert.Equal(t, "on_field", response.Data[0].Position)
	assert.Equal(t, "On Field", response.Data[0].PositionLabel)
	assert.Equal(t, "Q2", response.Data[0].ShortName)
	assert.Equal(t, now.Add(time.Minute).Format(time.RFC3339), response.Data[0].ScheduledAt)
	assert.Equal(t, []int{1, 2, 3}, response.Data[0].Red.TeamIds)
	assert.Empty(t, response.Data[0].Red.OffFieldTeamIds)
	assert.Equal(t, "on_deck", response.Data[1].Position)
	assert.Equal(t, []int{8}, response.Data[1].Blue.TeamIds)
}

func TestApiV1QueueingDisplayMatchesStopsAtScheduleGap(t *testing.T) {
	web := setupTestWeb(t)
	now := time.Now().UTC()
	first := model.Match{Type: model.Practice, TypeOrder: 1, ShortName: "P1", Time: now, Status: game.MatchScheduled}
	afterGap := model.Match{Type: model.Practice, TypeOrder: 2, ShortName: "P2", Time: now.Add((field.MaxMatchGapMin + 1) * time.Minute), Status: game.MatchScheduled}
	require.NoError(t, web.arena.Database.CreateMatch(&first))
	require.NoError(t, web.arena.Database.CreateMatch(&afterGap))
	web.arena.CurrentMatch = &first

	recorder := web.getHttpResponse("/api/v1/displays/queueing/matches")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data []apiV1QueueingMatch `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, "P1", response.Data[0].ShortName)
}

func TestApiV1QueueingDisplayMatchesIncludesPlayoffOffFieldTeams(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1, TeamIds: []int{1, 2, 3, 9}}))
	require.NoError(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 2, TeamIds: []int{4, 5, 6, 10}}))
	match := model.Match{
		Type: model.Playoff, TypeOrder: 1, ShortName: "F1", Status: game.MatchScheduled,
		PlayoffRedAlliance: 1, PlayoffBlueAlliance: 2,
		Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6,
	}
	require.NoError(t, web.arena.Database.CreateMatch(&match))
	web.arena.CurrentMatch = &match

	recorder := web.getHttpResponse("/api/v1/displays/queueing/matches")
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data []apiV1QueueingMatch `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, []int{9}, response.Data[0].Red.OffFieldTeamIds)
	assert.Equal(t, []int{10}, response.Data[0].Blue.OffFieldTeamIds)
	assert.Equal(t, 1, response.Data[0].Red.PlayoffAllianceId)
	assert.Empty(t, response.Data[0].ScheduledAt)
}

func TestApiV1QueueingDisplayMatchesReturnsEmptyArray(t *testing.T) {
	web := setupTestWeb(t)
	recorder := web.getHttpResponse("/api/v1/displays/queueing/matches")
	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"data":[]`)
}
