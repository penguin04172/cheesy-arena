// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchPlayV1SocketAndRestShareControlIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	const csrf = "match-play-socket-csrf"
	header := http.Header{}
	header.Set("Cookie", csrfTokenCookie+"="+csrf)
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/match-play", header)
	require.NoError(t, err)
	defer conn.Close()
	readAllianceSelectionV1Message(t, conn, "ready")
	command := map[string]any{"requestId": "match-request-1", "idempotencyKey": "match-control-1", "csrfToken": csrf, "command": "toggleBypass", "value": "R3"}
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	result := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "match-request-1", result["requestId"])
	assert.Equal(t, false, result["replayed"])
	assert.True(t, web.arena.AllianceStations["R3"].Bypass)
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/match-play/commands", "match-control-1", `{"command":"toggleBypass","value":"R3"}`)
	assert.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.True(t, web.arena.AllianceStations["R3"].Bypass)

	command["csrfToken"] = "wrong"
	command["idempotencyKey"] = "match-control-2"
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	rejected := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "invalid_csrf_token", rejected["code"])
}
