// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScoringV1SocketAndRestShareTowerIdempotency(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	const csrf = "scoring-socket-csrf"
	header := http.Header{}
	header.Set("Cookie", csrfTokenCookie+"="+csrf)
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/scoring/red", header)
	require.NoError(t, err)
	defer conn.Close()
	readAllianceSelectionV1Message(t, conn, "ready")
	command := map[string]any{
		"requestId": "scoring-request-1", "idempotencyKey": "scoring-cross-1", "csrfToken": csrf,
		"command": "autoTower", "payload": map[string]any{"teamPosition": 1, "autoTowerStatus": 2},
	}
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	result := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "scoring-request-1", result["requestId"])
	assert.Equal(t, true, result["changed"])
	assert.Equal(t, game.TowerLevel2, web.arena.RedRealtimeScore.CurrentScore.AutoTowerStatuses[0])
	replayed := apiV1AdminIdempotentRequest(web, http.MethodPost, "/api/v1/admin/scoring/red/commands", "scoring-cross-1", `{"command":"autoTower","payload":{"teamPosition":1,"autoTowerStatus":2}}`)
	assert.Equal(t, http.StatusOK, replayed.Code, replayed.Body.String())
	assert.Contains(t, replayed.Body.String(), `"replayed":true`)
	command["csrfToken"] = "wrong"
	command["idempotencyKey"] = "scoring-cross-2"
	require.NoError(t, conn.WriteJSON(map[string]any{"type": "command", "data": command}))
	rejected := readAllianceSelectionV1Message(t, conn, "commandResult")["data"].(map[string]any)
	assert.Equal(t, "invalid_csrf_token", rejected["code"])
}
