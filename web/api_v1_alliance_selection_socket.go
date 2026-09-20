// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Team254/cheesy-arena/websocket"
)

type apiV1AllianceSelectionSocketCommand struct {
	RequestId      string `json:"requestId"`
	IdempotencyKey string `json:"idempotencyKey"`
	CsrfToken      string `json:"csrfToken"`
	Command        string `json:"command"`
	Value          any    `json:"value"`
}

type apiV1AllianceSelectionSocketResult struct {
	RequestId string `json:"requestId"`
	Command   string `json:"command,omitempty"`
	Replayed  bool   `json:"replayed"`
	Code      string `json:"code,omitempty"`
}

func (web *Web) handleAllianceSelectionV1Commands(ws *websocket.Websocket, r *http.Request) {
	defer ws.Close()
	ws.SetReadLimit(4096)
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			return
		}
		result := apiV1AllianceSelectionSocketResult{}
		if messageType != "command" {
			result.Code = "unknown_message_type"
		} else {
			body, marshalErr := json.Marshal(data)
			var input apiV1AllianceSelectionSocketCommand
			if marshalErr != nil || json.Unmarshal(body, &input) != nil {
				result.Code = "invalid_command"
			} else {
				result.RequestId = input.RequestId
				result.Command = input.Command
				cookie, cookieErr := r.Cookie(csrfTokenCookie)
				if cookieErr != nil || cookie.Value == "" || input.CsrfToken == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(input.CsrfToken)) != 1 {
					result.Code = "invalid_csrf_token"
				} else if strings.TrimSpace(input.IdempotencyKey) == "" || len(input.IdempotencyKey) > 128 {
					result.Code = "invalid_idempotency_key"
				} else {
					result.Replayed, err = web.executeAllianceSelectionIdempotent(input.IdempotencyKey, input.Command, input.Value)
					switch {
					case errors.Is(err, errAllianceSelectionKeyReused):
						result.Code = "idempotency_key_reused"
					case errors.Is(err, errInvalidAllianceSelectionCommand):
						result.Code = "unknown_command"
					case err != nil:
						result.Code = "invalid_command_value"
					}
					auditResult := "success"
					if result.Code != "" {
						auditResult = "rejected"
					} else if result.Replayed {
						auditResult = "replayed"
					}
					web.logApiV1Audit(r, "alliance-selection."+input.Command, "timer/display", auditResult)
				}
			}
		}
		if err := ws.WriteV1CommandResult(result); err != nil {
			log.Printf("Alliance selection v1 command response failed: %v", err)
			return
		}
	}
}
