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

func TestApiV1AdminTeamImportJob(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/teams", `{
		"teamIds":[254,1114],"downloadOfficialData":false
	}`)
	require.Equal(t, http.StatusAccepted, recorder.Code)
	var response struct {
		Data apiV1JobDto `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	job := waitForApiV1Job(t, web, response.Data.Id)
	assert.Equal(t, "succeeded", job.Status)
	assert.Equal(t, 100, job.Progress)

	jobResponse := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/jobs/"+job.Id, "")
	assert.Equal(t, http.StatusOK, jobResponse.Code)
	teams, err := web.arena.Database.GetAllTeams()
	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, 254, teams[0].Id)
}

func TestApiV1AdminTeamImportValidation(t *testing.T) {
	web := setupTestWeb(t)
	duplicate := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/teams", `{
		"teamIds":[254,254],"downloadOfficialData":false
	}`)
	assert.Equal(t, http.StatusUnprocessableEntity, duplicate.Code)

	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 254}))
	existing := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/teams", `{
		"teamIds":[254],"downloadOfficialData":false
	}`)
	assert.Equal(t, http.StatusConflict, existing.Code)
}

func TestApiV1AdminTeamReadMasksWpaKey(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{
		Id: 254, Nickname: "Cheesy Poofs", WpaKey: "supersecret", FtaNotes: "Inspect radio",
	}))
	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/teams", "")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"wpaKeyConfigured":true`)
	assert.NotContains(t, recorder.Body.String(), "supersecret")
	assert.Contains(t, recorder.Body.String(), "Inspect radio")
}

func TestApiV1AdminTeamUpdateSecretSemantics(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.NetworkSecurityEnabled = true
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 254, WpaKey: "original1"}))
	path := "/api/v1/admin/teams/254"
	keep := apiV1AdminJsonRequest(web, http.MethodPatch, path, `{
		"name":"Team 254","nickname":"Poofs","city":"San Jose","stateProv":"CA","country":"USA",
		"schoolName":"Bellarmine","rookieYear":1999,"robotName":"Robot","accomplishments":"Wins",
		"wpaKey":{"action":"keep","value":""},"hasConnected":true,"ftaNotes":"Ready"
	}`)
	assert.Equal(t, http.StatusOK, keep.Code)
	team, err := web.arena.Database.GetTeamById(254)
	require.NoError(t, err)
	assert.Equal(t, "original1", team.WpaKey)
	assert.Equal(t, "Ready", team.FtaNotes)

	replace := apiV1AdminJsonRequest(web, http.MethodPatch, path, `{
		"name":"Team 254","nickname":"Poofs","city":"San Jose","stateProv":"CA","country":"USA",
		"schoolName":"Bellarmine","rookieYear":1999,"robotName":"Robot","accomplishments":"Wins",
		"wpaKey":{"action":"replace","value":"newsecret"},"hasConnected":true,"ftaNotes":"Ready"
	}`)
	assert.Equal(t, http.StatusOK, replace.Code)
	team, err = web.arena.Database.GetTeamById(254)
	require.NoError(t, err)
	assert.Equal(t, "newsecret", team.WpaKey)
	assert.NotContains(t, replace.Body.String(), "newsecret")
}

func TestApiV1AdminTeamDeleteConflict(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 254}))
	require.NoError(t, web.arena.Database.CreateMatch(&model.Match{Type: model.Qualification}))
	recorder := apiV1AdminJsonRequest(web, http.MethodDelete, "/api/v1/admin/teams/254", "")
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "team_list_locked")
}

func TestApiV1AdminTeamGenerateWpaKeys(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.NetworkSecurityEnabled = true
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 254, WpaKey: "existing"}))
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 1114}))
	recorder := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/teams/wpa-keys", `{"replaceAll":false}`)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"updated":1`)
	first, err := web.arena.Database.GetTeamById(254)
	require.NoError(t, err)
	second, err := web.arena.Database.GetTeamById(1114)
	require.NoError(t, err)
	assert.Equal(t, "existing", first.WpaKey)
	assert.Len(t, second.WpaKey, wpaKeyLength)
}

func TestApiV1AdminJobNotFound(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/jobs/missing", "")
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func waitForApiV1Job(t *testing.T, web *Web, id string) apiV1JobDto {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, ok := web.apiV1State.jobs.get(id)
		require.True(t, ok)
		if job.Status == "succeeded" || job.Status == "failed" {
			return job
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("job %s did not complete", id)
	return apiV1JobDto{}
}
