// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

type apiV1ScoringTowerInput struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}

type apiV1ScoringTowerResult struct {
	Command  string `json:"command"`
	Changed  bool   `json:"changed"`
	Replayed bool   `json:"replayed"`
}

func (web *Web) executeScoringTowerIdempotent(key, position, command string, payload any) (apiV1ScoringTowerResult, error) {
	result := apiV1ScoringTowerResult{Command: command}
	canonical, err := json.Marshal(payload)
	if err != nil {
		return result, errInvalidScoringTowerValue
	}
	operation := "scoring:" + position + ":" + command + ":" + string(canonical)
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	if previous, found := web.apiV1State.idempotency[key]; found {
		if previous.Operation != operation {
			return result, errAllianceSelectionKeyReused
		}
		result = previous.Data.(apiV1ScoringTowerResult)
		result.Replayed = true
		return result, nil
	}
	result.Changed, err = web.executeScoringTowerCommand(position, command, payload)
	if err != nil {
		return result, err
	}
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation, Data: result}
	return result, nil
}

func (web *Web) apiV1ScoringTowerCommandHandler(w http.ResponseWriter, r *http.Request) {
	position := r.PathValue("position")
	if _, ok := positionParameters[position]; !ok {
		writeApiV1Error(w, r, http.StatusNotFound, "scoring_position_not_found", "Scoring position not found.", nil)
		return
	}
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	var input apiV1ScoringTowerInput
	if err := decodeApiV1Json(w, r, &input); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid scoring command body.", nil)
		return
	}
	var payload any
	if len(input.Payload) > 0 {
		if err := json.Unmarshal(input.Payload, &payload); err != nil {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Invalid scoring command payload.", nil)
			return
		}
	}
	result, err := web.executeScoringTowerIdempotent(key, position, input.Command, payload)
	if err != nil {
		switch {
		case errors.Is(err, errAllianceSelectionKeyReused):
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
		case errors.Is(err, errInvalidScoringTowerCommand):
			writeApiV1Error(w, r, http.StatusBadRequest, "unknown_command", "Unknown scoring command.", nil)
		case errors.Is(err, errScoringTowerCommitted):
			writeApiV1Error(w, r, http.StatusConflict, "score_committed", "Scoring panel already committed.", nil)
		default:
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_command_value", "Invalid scoring command value or match state.", nil)
		}
		web.logApiV1Audit(r, "scoring."+input.Command, position, "rejected")
		return
	}
	outcome := "success"
	if result.Replayed {
		outcome = "replayed"
	}
	web.logApiV1Audit(r, "scoring."+input.Command, position, outcome)
	writeApiV1Data(w, r, http.StatusOK, result, nil)
}
