// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminPublishingJobAndReplay(t *testing.T) {
	web := setupTestWeb(t)
	tbaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer tbaServer.Close()
	web.arena.EventSettings.TbaPublishingEnabled = true
	web.arena.TbaClient.BaseUrl = tbaServer.URL
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/publishing/teams", "publish-teams-1", "")
	require.Equal(t, http.StatusAccepted, first.Code, first.Body.String())
	var response struct {
		Data apiV1PublishingResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &response))
	job := waitForApiV1Job(t, web, response.Data.Job.Id)
	assert.Equal(t, "succeeded", job.Status)

	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/publishing/teams", "publish-teams-1", "")
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.Contains(t, replayed.Body.String(), `"status":"succeeded"`)
}

func TestApiV1AdminPublishingDisabled(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/publishing/awards", "publish-awards-1", "")
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "publishing_disabled")
}

func TestApiV1AdminPublishingFailureIsSanitized(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.TbaPublishingEnabled = true
	web.arena.TbaClient.BaseUrl = "http://127.0.0.1:1/private-path"
	recorder := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/publishing/teams", "publish-fail-1", "")
	require.Equal(t, http.StatusAccepted, recorder.Code)
	var response struct {
		Data apiV1PublishingResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	job := waitForApiV1Job(t, web, response.Data.Job.Id)
	require.Equal(t, "failed", job.Status)
	require.NotNil(t, job.Error)
	assert.Equal(t, "Unable to publish teams to TBA.", *job.Error)
}

func TestApiV1AdminPublishingUnknownResource(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/publishing/unknown", "publish-unknown", "")
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
