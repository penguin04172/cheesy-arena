// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

type apiV1MatchPlayControlInput struct {
	Command string          `json:"command"`
	Value   json.RawMessage `json:"value,omitempty"`
}

type apiV1MatchPlayControlResult struct {
	Command  string `json:"command"`
	Replayed bool   `json:"replayed"`
}

func (web *Web) executeMatchPlayControlIdempotent(key, command string, value any) (bool, error) {
	canonical, err := json.Marshal(value)
	if err != nil {
		return false, errInvalidMatchPlayControlValue
	}
	operation := "match-play-control:" + command + ":" + string(canonical)
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	if previous, found := web.apiV1State.idempotency[key]; found {
		if previous.Operation != operation {
			return false, errAllianceSelectionKeyReused
		}
		return true, nil
	}
	if err := web.executeMatchPlayControlCommand(command, value); err != nil {
		return false, err
	}
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation}
	return false, nil
}

func (web *Web) apiV1MatchPlayControlCommandHandler(w http.ResponseWriter, r *http.Request) {
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	var input apiV1MatchPlayControlInput
	if err := decodeApiV1Json(w, r, &input); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid match play command body.", nil)
		return
	}
	var value any
	if len(input.Value) > 0 {
		if err := json.Unmarshal(input.Value, &value); err != nil {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid match play command value.", nil)
			return
		}
	}
	replayed, err := web.executeMatchPlayControlIdempotent(key, input.Command, value)
	if err != nil {
		switch {
		case errors.Is(err, errAllianceSelectionKeyReused):
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
		case errors.Is(err, errInvalidMatchPlayControlCommand):
			writeApiV1Error(w, r, http.StatusBadRequest, "unknown_command", "Unknown match play command.", nil)
		case errors.Is(err, errInvalidMatchPlayControlState):
			writeApiV1Error(w, r, http.StatusConflict, "invalid_match_state", "Command not allowed in the current match state.", nil)
		case errors.Is(err, errInvalidMatchPlayControlValue):
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_command_value", "Invalid match play command value.", nil)
		default:
			writeApiV1Error(w, r, http.StatusServiceUnavailable, "field_unavailable", "Field command could not be completed.", nil)
		}
		web.logApiV1Audit(r, "match-play."+input.Command, "field-control", "rejected")
		return
	}
	result := "success"
	if replayed {
		result = "replayed"
	}
	web.logApiV1Audit(r, "match-play."+input.Command, "field-control", result)
	writeApiV1Data(w, r, http.StatusOK, apiV1MatchPlayControlResult{Command: input.Command, Replayed: replayed}, nil)
}
