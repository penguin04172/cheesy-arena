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

type apiV1MatchPlaySocketResult struct {
	RequestId string `json:"requestId"`
	Command   string `json:"command,omitempty"`
	Replayed  bool   `json:"replayed"`
	Code      string `json:"code,omitempty"`
}

func (web *Web) handleMatchPlayV1ControlCommands(ws *websocket.Websocket, r *http.Request) {
	defer ws.Close()
	ws.SetReadLimit(4096)
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			return
		}
		result := apiV1MatchPlaySocketResult{}
		if messageType != "command" {
			result.Code = "unknown_message_type"
		} else {
			body, marshalErr := json.Marshal(data)
			var input apiV1AllianceSelectionSocketCommand
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
					result.Replayed, err = web.executeMatchPlayControlIdempotent(input.IdempotencyKey, input.Command, input.Value)
					switch {
					case errors.Is(err, errAllianceSelectionKeyReused):
						result.Code = "idempotency_key_reused"
					case errors.Is(err, errInvalidMatchPlayControlCommand):
						result.Code = "unknown_command"
					case errors.Is(err, errInvalidMatchPlayControlState):
						result.Code = "invalid_match_state"
					case errors.Is(err, errInvalidMatchPlayControlValue):
						result.Code = "invalid_command_value"
					case err != nil:
						result.Code = "field_unavailable"
					}
				}
				outcome := "success"
				if result.Code != "" {
					outcome = "rejected"
				} else if result.Replayed {
					outcome = "replayed"
				}
				web.logApiV1Audit(r, "match-play."+input.Command, "field-control", outcome)
			}
		}
		if err := ws.WriteV1CommandResult(result); err != nil {
			return
		}
	}
}
