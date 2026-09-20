// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type apiV1AllianceSelectionCommandInput struct {
	Command string          `json:"command"`
	Value   json.RawMessage `json:"value,omitempty"`
}

type apiV1AllianceSelectionCommandResult struct {
	Command  string `json:"command"`
	Replayed bool   `json:"replayed"`
}

var errAllianceSelectionKeyReused = errors.New("idempotency key reused")

func (web *Web) executeAllianceSelectionIdempotent(key, command string, value any) (bool, error) {
	if strings.TrimSpace(key) == "" || len(key) > 128 {
		return false, errInvalidAllianceSelectionValue
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return false, errInvalidAllianceSelectionValue
	}
	operation := "alliance-selection:" + command + ":" + string(canonical)
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	if previous, found := web.apiV1State.idempotency[key]; found {
		if previous.Operation != operation {
			return false, errAllianceSelectionKeyReused
		}
		return true, nil
	}
	if err := web.executeAllianceSelectionCommand(command, value); err != nil {
		return false, err
	}
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation}
	return false, nil
}

// Commands require admin session, CSRF token and an idempotency key. The key
// remains reserved for the exact command and value until this process exits.
func (web *Web) apiV1AllianceSelectionCommandHandler(w http.ResponseWriter, r *http.Request) {
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	var input apiV1AllianceSelectionCommandInput
	if err := decodeApiV1Json(w, r, &input); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid command body.", nil)
		return
	}
	var value any
	if len(input.Value) > 0 {
		if err := json.Unmarshal(input.Value, &value); err != nil {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid command value.", nil)
			return
		}
	}
	replayed, err := web.executeAllianceSelectionIdempotent(key, input.Command, value)
	if err != nil {
		if errors.Is(err, errInvalidAllianceSelectionCommand) {
			writeApiV1Error(w, r, http.StatusBadRequest, "unknown_command", "Unknown alliance selection command.", nil)
		} else if errors.Is(err, errAllianceSelectionKeyReused) {
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
		} else {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_command_value", "Invalid alliance selection command value.", nil)
		}
		web.logApiV1Audit(r, "alliance-selection."+input.Command, "timer/display", "rejected")
		return
	}
	result := apiV1AllianceSelectionCommandResult{Command: input.Command, Replayed: replayed}
	auditResult := "success"
	if replayed {
		auditResult = "replayed"
	}
	web.logApiV1Audit(r, "alliance-selection."+input.Command, "timer/display", auditResult)
	writeApiV1Data(w, r, http.StatusOK, result, nil)
}
