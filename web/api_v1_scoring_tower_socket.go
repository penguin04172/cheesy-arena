// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Team254/cheesy-arena/websocket"
)

type apiV1ScoringSocketResult struct {
	RequestId string `json:"requestId"`
	Command   string `json:"command,omitempty"`
	Changed   bool   `json:"changed"`
	Replayed  bool   `json:"replayed"`
	Code      string `json:"code,omitempty"`
}

func (web *Web) handleScoringV1TowerCommands(ws *websocket.Websocket, r *http.Request, position string) {
	defer ws.Close()
	ws.SetReadLimit(4096)
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			return
		}
		result := apiV1ScoringSocketResult{}
		if messageType != "command" {
			result.Code = "unknown_message_type"
		} else {
			body, marshalErr := json.Marshal(data)
			var input apiV1RefereeSocketCommand
			if marshalErr != nil || json.Unmarshal(body, &input) != nil {
				result.Code = "invalid_command"
			} else {
				result.RequestId, result.Command = input.RequestId, input.Command
				cookie, cookieErr := r.Cookie(csrfTokenCookie)
				switch {
				case cookieErr != nil || cookie.Value == "" || input.CsrfToken == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(input.CsrfToken)) != 1:
					result.Code = "invalid_csrf_token"
				case strings.TrimSpace(input.IdempotencyKey) == "" || len(input.IdempotencyKey) > 128:
					result.Code = "invalid_idempotency_key"
				default:
					commandResult, commandErr := web.executeScoringTowerIdempotent(input.IdempotencyKey, position, input.Command, input.Payload)
					result.Changed, result.Replayed = commandResult.Changed, commandResult.Replayed
					switch {
					case errors.Is(commandErr, errAllianceSelectionKeyReused):
						result.Code = "idempotency_key_reused"
					case errors.Is(commandErr, errInvalidScoringTowerCommand):
						result.Code = "unknown_command"
					case errors.Is(commandErr, errScoringTowerCommitted):
						result.Code = "score_committed"
					case commandErr != nil:
						result.Code = "invalid_command_value"
					}
				}
				outcome := "success"
				if result.Code != "" {
					outcome = "rejected"
				} else if result.Replayed {
					outcome = "replayed"
				}
				web.logApiV1Audit(r, "scoring."+input.Command, position, outcome)
			}
		}
		if err := ws.WriteV1CommandResult(result); err != nil {
			return
		}
	}
}
