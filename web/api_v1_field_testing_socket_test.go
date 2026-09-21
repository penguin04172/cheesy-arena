// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldTestingV1SocketAndRestShareIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	const csrf = "field-testing-socket-csrf"
	header := http.Header{}
	header.Set("Cookie", csrfTokenCookie+"="+csrf)
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/field-testing", header)
	require.NoError(t, err)
	defer conn.Close()
	readAllianceSelectionV1Message(t, conn, "ready")
	command := map[string]any{
		"requestId": "field-request-1", "idempotencyKey": "field-cross-1", "csrfToken": csrf,
		"command": "setPlcCoilOverride", "value": map[string]any{"index": 7, "override": "on"},
	}
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	result := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "field-request-1", result["requestId"])
	assert.Equal(t, false, result["replayed"])
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/field-testing/commands", "field-cross-1", `{"command":"setPlcCoilOverride","value":{"index":7,"override":"on"}}`)
	assert.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	command["csrfToken"] = "wrong"
	command["idempotencyKey"] = "field-cross-2"
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	rejected := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "invalid_csrf_token", rejected["code"])
}
