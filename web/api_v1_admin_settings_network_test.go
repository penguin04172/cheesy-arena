// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminNetworkSettingsMaskAndUpdate(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.ApPassword = "ap-secret"
	web.arena.EventSettings.SwitchPassword = "switch-secret"
	web.arena.EventSettings.SCCPassword = "scc-secret"
	require.NoError(t, web.arena.Database.UpdateEventSettings(web.arena.EventSettings))
	read := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/settings/network", "")
	assert.Equal(t, http.StatusOK, read.Code)
	assert.Contains(t, read.Body.String(), `"apPasswordConfigured":true`)
	assert.NotContains(t, read.Body.String(), "ap-secret")
	assert.NotContains(t, read.Body.String(), "switch-secret")
	assert.NotContains(t, read.Body.String(), "scc-secret")

	updated := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/network", `{
		"networkSecurityEnabled":true,"apAddress":"10.0.100.2","apChannel":44,
		"apPassword":{"action":"keep","value":""},
		"switchPassword":{"action":"replace","value":"new-switch"},
		"sccPassword":{"action":"clear","value":""}
	}`)
	require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
	assert.Equal(t, "ap-secret", web.arena.EventSettings.ApPassword)
	assert.Equal(t, "new-switch", web.arena.EventSettings.SwitchPassword)
	assert.Empty(t, web.arena.EventSettings.SCCPassword)
	assert.Equal(t, 44, web.arena.EventSettings.ApChannel)
	assert.NotContains(t, updated.Body.String(), "new-switch")
}

func TestApiV1AdminNetworkSettingsValidation(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/network", `{"apChannel":0}`)
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Equal(t, 36, web.arena.EventSettings.ApChannel)
}
