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

type apiV1FieldTestingSocketResult struct {
	RequestId string `json:"requestId"`
	Command   string `json:"command,omitempty"`
	Replayed  bool   `json:"replayed"`
	Code      string `json:"code,omitempty"`
}

func (web *Web) handleFieldTestingV1Commands(ws *websocket.Websocket, r *http.Request) {
	defer ws.Close()
	ws.SetReadLimit(4096)
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			return
		}
		result := apiV1FieldTestingSocketResult{}
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
					result.Replayed, err = web.executeFieldTestingIdempotent(input.IdempotencyKey, input.Command, input.Value)
					switch {
					case errors.Is(err, errAllianceSelectionKeyReused):
						result.Code = "idempotency_key_reused"
					case errors.Is(err, errInvalidFieldTestingCommand):
						result.Code = "unknown_command"
					case errors.Is(err, errFieldTestingCoilState), errors.Is(err, errFieldTestingLedState):
						result.Code = "invalid_match_state"
					case err != nil:
						result.Code = "invalid_command_value"
					}
				}
				outcome := "success"
				if result.Code != "" {
					outcome = "rejected"
				} else if result.Replayed {
					outcome = "replayed"
				}
				web.logApiV1Audit(r, "field-testing."+input.Command, "field-hardware", outcome)
			}
		}
		if err := ws.WriteV1CommandResult(result); err != nil {
			return
		}
	}
}
