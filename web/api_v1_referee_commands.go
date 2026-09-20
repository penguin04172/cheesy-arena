// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

type apiV1RefereeCommandInput struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}

type apiV1RefereeCommandResult struct {
	Command  string `json:"command"`
	Changed  bool   `json:"changed"`
	FoulId   int    `json:"foulId,omitempty"`
	Replayed bool   `json:"replayed"`
}

func (web *Web) executeRefereeIdempotent(key, command string, payload any) (apiV1RefereeCommandResult, error) {
	result := apiV1RefereeCommandResult{Command: command}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return result, errInvalidRefereeValue
	}
	operation := "referee:" + command + ":" + string(canonical)
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	if previous, found := web.apiV1State.idempotency[key]; found {
		if previous.Operation != operation {
			return result, errAllianceSelectionKeyReused
		}
		result = previous.Data.(apiV1RefereeCommandResult)
		result.Replayed = true
		return result, nil
	}
	commandResult, err := web.executeRefereeScoreCommand(command, payload)
	if err != nil {
		return result, err
	}
	result.Changed, result.FoulId = commandResult.Changed, commandResult.FoulId
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation, Data: result}
	return result, nil
}

func (web *Web) apiV1RefereeCommandHandler(w http.ResponseWriter, r *http.Request) {
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	var input apiV1RefereeCommandInput
	if err := decodeApiV1Json(w, r, &input); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid referee command body.", nil)
		return
	}
	var payload any
	if len(input.Payload) > 0 {
		if err := json.Unmarshal(input.Payload, &payload); err != nil {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid referee command payload.", nil)
			return
		}
	}
	result, err := web.executeRefereeIdempotent(key, input.Command, payload)
	if err != nil {
		switch {
		case errors.Is(err, errAllianceSelectionKeyReused):
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
		case errors.Is(err, errInvalidRefereeCommand):
			writeApiV1Error(w, r, http.StatusBadRequest, "unknown_command", "Unknown referee command.", nil)
		case errors.Is(err, errRefereeScoreCommitted):
			writeApiV1Error(w, r, http.StatusConflict, "score_committed", "Fouls are already committed.", nil)
		default:
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_command_value", "Invalid referee command value or match state.", nil)
		}
		web.logApiV1Audit(r, "referee."+input.Command, "score", "rejected")
		return
	}
	auditResult := "success"
	if result.Replayed {
		auditResult = "replayed"
	}
	web.logApiV1Audit(r, "referee."+input.Command, "score", auditResult)
	writeApiV1Data(w, r, http.StatusOK, result, nil)
}
