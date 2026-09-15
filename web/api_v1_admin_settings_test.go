// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminEventSettingsPartialUpdate(t *testing.T) {
	web := setupTestWeb(t)
	originalAlliances := web.arena.EventSettings.NumPlayoffAlliances
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/event", `{
		"name":"API Event","selectionShowUnpickedTeams":false
	}`)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, "API Event", web.arena.EventSettings.Name)
	assert.False(t, web.arena.EventSettings.SelectionShowUnpickedTeams)
	assert.Equal(t, originalAlliances, web.arena.EventSettings.NumPlayoffAlliances)
	assert.Contains(t, recorder.Body.String(), `"adminPasswordConfigured":false`)
}

func TestApiV1AdminEventSettingsPlayoffConflict(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateAlliance(&model.Alliance{Id: 1}))
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/event", `{
		"playoffType":"single_elimination","numPlayoffAlliances":16
	}`)
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "alliance_selection_finalized")
	assert.Equal(t, model.DoubleEliminationPlayoff, web.arena.EventSettings.PlayoffType)
}

func TestApiV1AdminIntegrationSettingsMaskAndSecretSemantics(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.TbaSecretId = "existing-id"
	web.arena.EventSettings.TbaSecret = "existing-secret"
	web.arena.EventSettings.NexusAutoQueueKey = "existing-nexus"
	require.NoError(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))

	read := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/settings/integrations", "")
	assert.Equal(t, http.StatusOK, read.Code)
	assert.Contains(t, read.Body.String(), `"tbaSecretConfigured":true`)
	assert.NotContains(t, read.Body.String(), "existing-secret")
	assert.NotContains(t, read.Body.String(), "existing-nexus")

	updated := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/integrations", `{
		"tbaPublishingEnabled":true,
		"tbaSecretId":{"action":"keep","value":""},
		"tbaSecret":{"action":"replace","value":"replacement"},
		"nexusAutoQueueKey":{"action":"clear","value":""}
	}`)
	require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
	assert.Equal(t, "existing-id", web.arena.EventSettings.TbaSecretId)
	assert.Equal(t, "replacement", web.arena.EventSettings.TbaSecret)
	assert.Empty(t, web.arena.EventSettings.NexusAutoQueueKey)
	assert.NotContains(t, updated.Body.String(), "replacement")
}

func TestApiV1AdminGameSettingsPartialUpdate(t *testing.T) {
	web := setupTestWeb(t)
	originalAuto := web.arena.EventSettings.AutoDurationSec
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/game", `{
		"shiftDurationSec":42,"endgameDurationSec":30
	}`)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, originalAuto, web.arena.EventSettings.AutoDurationSec)
	assert.Equal(t, 42, web.arena.EventSettings.ShiftDurationSec)
	assert.Equal(t, 30, web.arena.EventSettings.EndgameDurationSec)
}

func TestApiV1AdminSettingsLockedDuringMatch(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.MatchState = field.AutoPeriod
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/event", `{"name":"Blocked"}`)
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "settings_locked")
	assert.NotEqual(t, "Blocked", web.arena.EventSettings.Name)
}

func TestApiV1AdminSettingsSectionValidation(t *testing.T) {
	web := setupTestWeb(t)
	missing := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/settings/unknown", "")
	assert.Equal(t, http.StatusNotFound, missing.Code)
	invalid := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/game", `{"autoDurationSec":-1}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
}
