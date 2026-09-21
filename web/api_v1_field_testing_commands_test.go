// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/led"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldTestingCommandValidation(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.executeFieldTestingCommand("setLedMode", map[string]any{"RedMode": int(led.PurpleMode), "BlueMode": int(led.GreenMode)}))
	red, blue := web.arena.Leds.GetModes()
	assert.Equal(t, led.PurpleMode, red)
	assert.Equal(t, led.GreenMode, blue)
	assert.ErrorIs(t, web.executeFieldTestingCommand("playSound", "not-a-sound"), errInvalidFieldTestingValue)
	assert.ErrorIs(t, web.executeFieldTestingCommand("setPlcCoilOverride", map[string]any{"Index": -1, "Override": "on"}), errInvalidFieldTestingValue)
	web.arena.MatchState = field.TeleopPeriod
	assert.ErrorIs(t, web.executeFieldTestingCommand("setPlcCoilOverride", map[string]any{"Index": 7, "Override": "on"}), errFieldTestingCoilState)
	assert.ErrorIs(t, web.executeFieldTestingCommand("setLedMode", map[string]any{"RedMode": int(led.PurpleMode), "BlueMode": int(led.GreenMode)}), errFieldTestingLedState)
}

func TestApiV1FieldTestingCommandIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	path := "/api/v1/admin/field-testing/commands"
	body := `{"command":"setPlcCoilOverride","value":{"index":7,"override":"on"}}`
	first := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "coil-1", body)
	require.Equal(t, http.StatusOK, first.Code, first.Body.String())
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "coil-1", body)
	assert.Equal(t, http.StatusOK, replayed.Code)
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	conflict := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "coil-1", `{"command":"setPlcCoilOverride","value":{"index":7,"override":"off"}}`)
	assert.Equal(t, http.StatusConflict, conflict.Code)
	invalid := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "coil-2", `{"command":"setPlcCoilOverride","value":{"index":7000,"override":"on"}}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalid.Code)
	web.arena.MatchState = field.AutoPeriod
	disallowed := apiV1AdminIdempotentRequest(web, http.MethodPost, path, "coil-3", body)
	assert.Equal(t, http.StatusConflict, disallowed.Code)
}
