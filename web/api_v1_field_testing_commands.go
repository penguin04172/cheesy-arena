// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

type apiV1FieldTestingCommandInput struct {
	Command string          `json:"command"`
	Value   json.RawMessage `json:"value,omitempty"`
}

type apiV1FieldTestingCommandResult struct {
	Command  string `json:"command"`
	Replayed bool   `json:"replayed"`
}

func (web *Web) executeFieldTestingIdempotent(key, command string, value any) (bool, error) {
	canonical, err := json.Marshal(value)
	if err != nil {
		return false, errInvalidFieldTestingValue
	}
	operation := "field-testing:" + command + ":" + string(canonical)
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	if previous, found := web.apiV1State.idempotency[key]; found {
		if previous.Operation != operation {
			return false, errAllianceSelectionKeyReused
		}
		return true, nil
	}
	if err := web.executeFieldTestingCommand(command, value); err != nil {
		return false, err
	}
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation}
	return false, nil
}

func (web *Web) apiV1FieldTestingCommandHandler(w http.ResponseWriter, r *http.Request) {
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	var input apiV1FieldTestingCommandInput
	if err := decodeApiV1Json(w, r, &input); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid field testing command body.", nil)
		return
	}
	var value any
	if len(input.Value) > 0 {
		if err := json.Unmarshal(input.Value, &value); err != nil {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid field testing command value.", nil)
			return
		}
	}
	replayed, err := web.executeFieldTestingIdempotent(key, input.Command, value)
	if err != nil {
		switch {
		case errors.Is(err, errAllianceSelectionKeyReused):
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
		case errors.Is(err, errInvalidFieldTestingCommand):
			writeApiV1Error(w, r, http.StatusBadRequest, "unknown_command", "Unknown field testing command.", nil)
		case errors.Is(err, errFieldTestingCoilState), errors.Is(err, errFieldTestingLedState):
			writeApiV1Error(w, r, http.StatusConflict, "invalid_match_state", "Command not allowed during a match.", nil)
		default:
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_command_value", "Invalid field testing command value.", nil)
		}
		web.logApiV1Audit(r, "field-testing."+input.Command, "field-hardware", "rejected")
		return
	}
	outcome := "success"
	if replayed {
		outcome = "replayed"
	}
	web.logApiV1Audit(r, "field-testing."+input.Command, "field-hardware", outcome)
	writeApiV1Data(w, r, http.StatusOK, apiV1FieldTestingCommandResult{Command: input.Command, Replayed: replayed}, nil)
}
