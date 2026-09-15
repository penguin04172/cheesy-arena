// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestApiV1Matches(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{
		Type: model.Qualification, TypeOrder: 1, Time: time.Unix(60, 0), ShortName: "Q1",
		Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6, Blue1IsSurrogate: true,
	}
	require.NoError(t, web.arena.Database.CreateMatch(&match))
	result := model.BuildTestMatchResult(match.Id, 1)
	require.NoError(t, web.arena.Database.CreateMatchResult(result))

	recorder := web.getHttpResponse("/api/v1/matches/qualification")
	assert.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data []apiV1MatchWithResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, "qualification", response.Data[0].Match.Type)
	assert.Equal(t, "Q1", response.Data[0].Match.ShortName)
	assert.Equal(t, 4, response.Data[0].Match.Blue[0].TeamId)
	assert.True(t, response.Data[0].Match.Blue[0].IsSurrogate)
	require.NotNil(t, response.Data[0].Result)
	assert.Equal(t, result.PlayNumber, response.Data[0].Result.PlayNumber)

	invalid := web.getHttpResponse("/api/v1/matches/invalid")
	assert.Equal(t, http.StatusBadRequest, invalid.Code)
	assert.Contains(t, invalid.Body.String(), "invalid_match_type")
}

func TestApiV1Rankings(t *testing.T) {
	web := setupTestWeb(t)
	ranking := game.TestRanking1()
	require.NoError(t, web.arena.Database.CreateRanking(ranking))
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: ranking.TeamId, Nickname: "Nickname"}))
	require.NoError(t, web.arena.Database.CreateMatch(&model.Match{
		Type: model.Qualification, TypeOrder: 1, ShortName: "Q1", Status: game.RedWonMatch,
	}))

	recorder := web.getHttpResponse("/api/v1/rankings")
	assert.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data apiV1Rankings `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data.Rankings, 1)
	assert.Equal(t, "Nickname", response.Data.Rankings[0].Nickname)
	assert.Equal(t, "Q1", response.Data.HighestPlayedMatch)
}

func TestApiV1AlliancesAndSponsorSlides(t *testing.T) {
	web := setupTestWeb(t)
	model.BuildTestAlliances(web.arena.Database)
	require.NoError(t, web.arena.Database.CreateSponsorSlide(&model.SponsorSlide{
		Subtitle: "Sponsor", DisplayTimeSec: 10, DisplayOrder: 1,
	}))

	alliancesRecorder := web.getHttpResponse("/api/v1/alliances")
	assert.Equal(t, http.StatusOK, alliancesRecorder.Code)
	var alliancesResponse struct {
		Data []apiV1Alliance `json:"data"`
	}
	require.NoError(t, json.Unmarshal(alliancesRecorder.Body.Bytes(), &alliancesResponse))
	require.Len(t, alliancesResponse.Data, 2)
	assert.Equal(t, 254, alliancesResponse.Data[0].TeamIds[0])

	slidesRecorder := web.getHttpResponse("/api/v1/sponsor-slides")
	assert.Equal(t, http.StatusOK, slidesRecorder.Code)
	var slidesResponse struct {
		Data []apiV1SponsorSlide `json:"data"`
	}
	require.NoError(t, json.Unmarshal(slidesRecorder.Body.Bytes(), &slidesResponse))
	require.Len(t, slidesResponse.Data, 1)
	assert.Equal(t, "Sponsor", slidesResponse.Data[0].Subtitle)
}

func TestApiV1TeamsRedactsPrivateFields(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{
		Id: 254, Name: "Team", Nickname: "Cheesy Poofs", WpaKey: "private-wpa-key", FtaNotes: "private-notes",
	}))

	recorder := web.getHttpResponse("/api/v1/teams")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "private-wpa-key")
	assert.NotContains(t, recorder.Body.String(), "private-notes")
	var response struct {
		Data []apiV1Team `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, 254, response.Data[0].Id)

	invalidAvatar := web.getHttpResponse("/api/v1/teams/not-a-number/avatar")
	assert.Equal(t, http.StatusBadRequest, invalidAvatar.Code)
	assert.Contains(t, invalidAvatar.Body.String(), "invalid_team_id")
}

func TestApiV1RulesAndBracket(t *testing.T) {
	web := setupTestWeb(t)

	rulesRecorder := web.getHttpResponse("/api/v1/game/rules")
	assert.Equal(t, http.StatusOK, rulesRecorder.Code)
	var rulesResponse struct {
		Data []apiV1Rule `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rulesRecorder.Body.Bytes(), &rulesResponse))
	require.NotEmpty(t, rulesResponse.Data)
	assert.Equal(t, 1, rulesResponse.Data[0].Id)

	web.arena.EventSettings.NumPlayoffAlliances = 4
	tournament.CreateTestAlliances(web.arena.Database, 4)
	web.arena.CreatePlayoffTournament()
	bracketRecorder := web.getHttpResponse("/api/v1/bracket")
	assert.Equal(t, http.StatusOK, bracketRecorder.Code)
	var bracketResponse struct {
		Data apiV1Bracket `json:"data"`
	}
	require.NoError(t, json.Unmarshal(bracketRecorder.Body.Bytes(), &bracketResponse))
	assert.Equal(t, "double4", bracketResponse.Data.BracketType)
	assert.NotEmpty(t, bracketResponse.Data.Matchups)
}

func TestApiV1MatchLogs(t *testing.T) {
	web := setupTestWeb(t)
	match := model.Match{Type: model.Qualification, TypeOrder: 1, ShortName: "Q9876", Red1: 9876}
	require.NoError(t, web.arena.Database.CreateMatch(&match))

	listRecorder := web.getHttpResponse("/api/v1/match-logs")
	assert.Equal(t, http.StatusOK, listRecorder.Code)
	var listResponse struct {
		Data apiV1MatchLogs `json:"data"`
	}
	require.NoError(t, json.Unmarshal(listRecorder.Body.Bytes(), &listResponse))
	require.Len(t, listResponse.Data.MatchesByType["qualification"], 1)

	detailRecorder := web.getHttpResponse("/api/v1/matches/" + stringInt(match.Id) + "/stations/R1/logs?limit=20&sampleEvery=2")
	assert.Equal(t, http.StatusOK, detailRecorder.Code)
	var detailResponse struct {
		Data apiV1MatchLogDetail `json:"data"`
	}
	require.NoError(t, json.Unmarshal(detailRecorder.Body.Bytes(), &detailResponse))
	assert.Equal(t, 9876, detailResponse.Data.TeamId)
	assert.Equal(t, 20, detailResponse.Data.Limit)
	assert.Equal(t, 2, detailResponse.Data.SampleEvery)
	assert.Empty(t, detailResponse.Data.Logs)

	invalidRecorder := web.getHttpResponse("/api/v1/matches/" + stringInt(match.Id) + "/stations/R1/logs?limit=10001")
	assert.Equal(t, http.StatusBadRequest, invalidRecorder.Code)
}

func stringInt(value int) string {
	return strconv.Itoa(value)
}
