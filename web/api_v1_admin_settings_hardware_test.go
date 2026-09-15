// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1AdminHardwareSettingsPartialUpdate(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/hardware", `{
		"ledUniverseMode":"two","teamSigns":{"red1":51,"blueTimer":54},
		"useLiteUdpPort":true,
		"companion":{"address":"10.0.0.8","port":51234,"matchStart":{"page":2,"row":1,"column":3}}
	}`)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, "two", web.arena.EventSettings.LedUniverseMode)
	assert.Equal(t, 51, web.arena.EventSettings.TeamSignRed1Id)
	assert.Equal(t, 54, web.arena.EventSettings.TeamSignBlueTimerId)
	assert.True(t, web.arena.EventSettings.UseLiteUdpPort)
	assert.Equal(t, "10.0.0.8", web.arena.EventSettings.CompanionAddress)
	assert.Equal(t, 51234, web.arena.EventSettings.CompanionPort)
	assert.Equal(t, 2, web.arena.EventSettings.CompanionMatchStartPage)
	assert.Equal(t, 1, web.arena.EventSettings.CompanionMatchStartRow)
	assert.Equal(t, 3, web.arena.EventSettings.CompanionMatchStartColumn)
	assert.Contains(t, recorder.Body.String(), `"matchStart":{"page":2,"row":1,"column":3}`)
}

func TestApiV1AdminHardwareSettingsValidation(t *testing.T) {
	web := setupTestWeb(t)
	invalidMode := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/hardware", `{"ledUniverseMode":"three"}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalidMode.Code)
	assert.Equal(t, "single", web.arena.EventSettings.LedUniverseMode)

	invalidPort := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/hardware", `{"companion":{"port":70000}}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalidPort.Code)
	assert.Zero(t, web.arena.EventSettings.CompanionPort)

	invalidCoordinate := apiV1AdminJsonRequest(web, http.MethodPatch, "/api/v1/admin/settings/hardware", `{"companion":{"matchEnd":{"row":-1}}}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalidCoordinate.Code)
}

func TestApiV1AdminHardwareSettingsGet(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.BlackmagicAddresses = "10.0.0.20,10.0.0.21"
	recorder := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/settings/hardware", "")
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "10.0.0.20,10.0.0.21")
	assert.Contains(t, recorder.Body.String(), `"teamSigns"`)
	assert.Contains(t, recorder.Body.String(), `"companion"`)
}
