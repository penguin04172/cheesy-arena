// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefereeV1SocketAndRestShareIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	const csrf = "referee-socket-csrf"
	header := http.Header{}
	header.Set("Cookie", csrfTokenCookie+"="+csrf)
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/referee", header)
	require.NoError(t, err)
	defer conn.Close()
	readAllianceSelectionV1Message(t, conn, "ready")
	command := map[string]any{
		"requestId": "referee-request-1", "idempotencyKey": "referee-cross-1", "csrfToken": csrf,
		"command": "addFoul", "payload": map[string]any{"alliance": "red", "isMajor": true},
	}
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	result := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "referee-request-1", result["requestId"])
	assert.Equal(t, float64(1), result["foulId"])
	assert.Equal(t, false, result["replayed"])
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/referee/commands", "referee-cross-1", `{"command":"addFoul","payload":{"alliance":"red","isMajor":true}}`)
	assert.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	assert.Len(t, web.arena.RedRealtimeScore.CurrentScore.Fouls, 1)

	command["csrfToken"] = "wrong"
	command["idempotencyKey"] = "referee-cross-2"
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	rejected := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "invalid_csrf_token", rejected["code"])
}
