// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readAllianceSelectionV1Message(t *testing.T, conn *gorillawebsocket.Conn, want string) map[string]any {
	t.Helper()
	for i := 0; i < 8; i++ {
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(3*time.Second)))
		var message map[string]any
		require.NoError(t, conn.ReadJSON(&message))
		if message["type"] == want {
			return message
		}
	}
	t.Fatalf("did not receive %s", want)
	return nil
}

func TestAllianceSelectionV1SocketAndRestShareIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	const csrf = "socket-csrf-token"
	header := http.Header{}
	header.Set("Cookie", csrfTokenCookie+"="+csrf)
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/alliance-selection", header)
	require.NoError(t, err)
	defer conn.Close()
	readAllianceSelectionV1Message(t, conn, "ready")
	command := map[string]any{"requestId": "request-1", "idempotencyKey": "cross-transport-1", "csrfToken": csrf, "command": "setTimer", "value": 75}
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	result := readAllianceSelectionV1Message(t, conn, "commandResult")
	assert.Equal(t, "request-1", result["data"].(map[string]any)["requestId"])
	assert.Equal(t, false, result["data"].(map[string]any)["replayed"])
	assert.Equal(t, 75, web.allianceSelectionCommands.timeLimit())
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/alliance-selection/commands", "cross-transport-1", `{"command":"setTimer","value":75}`)
	assert.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)

	command["csrfToken"] = "wrong"
	command["idempotencyKey"] = "cross-transport-2"
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	rejected := readAllianceSelectionV1Message(t, conn, "commandResult")
	assert.Equal(t, "invalid_csrf_token", rejected["data"].(map[string]any)["code"])
}
