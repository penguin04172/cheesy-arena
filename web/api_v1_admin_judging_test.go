// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminJudgingScheduleGenerateGetAndClear(t *testing.T) {
	web := setupTestWeb(t)
	createApiV1JudgingPrerequisites(t, web)
	body := `{
		"numJudges":3,"durationMinutes":10,"previousSpacingMinutes":15,"nextSpacingMinutes":15
	}`
	generated := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/judging-schedule", body)
	require.Equal(t, http.StatusCreated, generated.Code, generated.Body.String())
	var response struct {
		Data apiV1JudgingSchedule `json:"data"`
	}
	require.NoError(t, json.Unmarshal(generated.Body.Bytes(), &response))
	assert.Equal(t, 3, response.Data.Params.NumJudges)
	assert.Len(t, response.Data.Slots, 6)
	for index := 1; index < len(response.Data.Slots); index++ {
		assert.LessOrEqual(t, response.Data.Slots[index-1].JudgeNumber, response.Data.Slots[index].JudgeNumber)
	}

	listed := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/judging-schedule", "")
	assert.Equal(t, http.StatusOK, listed.Code)
	assert.Contains(t, listed.Body.String(), `"numJudges":3`)

	conflict := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/judging-schedule", body)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	assert.Contains(t, conflict.Body.String(), "judging_schedule_exists")

	cleared := apiV1AdminJsonRequest(web, http.MethodDelete, "/api/v1/admin/judging-schedule", "")
	assert.Equal(t, http.StatusNoContent, cleared.Code)
	slots, err := web.arena.Database.GetAllJudgingSlots()
	require.NoError(t, err)
	assert.Empty(t, slots)
}

func TestApiV1AdminJudgingScheduleValidation(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/judging-schedule", `{
		"numJudges":0,"durationMinutes":0,"previousSpacingMinutes":0,"nextSpacingMinutes":0
	}`)
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "numJudges")
	assert.Contains(t, recorder.Body.String(), "durationMinutes")
}

func TestApiV1AdminJudgingScheduleRequiresPrerequisites(t *testing.T) {
	web := setupTestWeb(t)
	body := `{
		"numJudges":3,"durationMinutes":10,"previousSpacingMinutes":15,"nextSpacingMinutes":15
	}`
	noTeams := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/judging-schedule", body)
	assert.Equal(t, http.StatusConflict, noTeams.Code)
	assert.Contains(t, noTeams.Body.String(), "teams_required")

	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 1}))
	noMatches := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/judging-schedule", body)
	assert.Equal(t, http.StatusConflict, noMatches.Code)
	assert.Contains(t, noMatches.Body.String(), "qualification_schedule_required")
}

func createApiV1JudgingPrerequisites(t *testing.T, web *Web) {
	t.Helper()
	for id := 1; id <= 6; id++ {
		require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: id}))
	}
	start := time.Now().UTC().Add(time.Hour)
	require.NoError(t, web.arena.Database.CreateMatch(&model.Match{
		Type: model.Qualification, TypeOrder: 1, Time: start,
		Red1: 1, Red2: 2, Red3: 3, Blue1: 4, Blue2: 5, Blue3: 6,
	}))
	require.NoError(t, web.arena.Database.CreateMatch(&model.Match{
		Type: model.Qualification, TypeOrder: 2, Time: start.Add(time.Hour),
		Red1: 6, Red2: 5, Red3: 4, Blue1: 3, Blue2: 2, Blue3: 1,
	}))
}
