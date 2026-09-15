// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"strings"

	"github.com/Team254/cheesy-arena/model"
)

type apiV1DatabaseBackupResult struct {
	Created  bool `json:"created"`
	Replayed bool `json:"replayed"`
}

type apiV1TournamentDataClearInput struct {
	EventId      int    `json:"eventId"`
	Confirmation string `json:"confirmation"`
}

type apiV1TournamentDataClearResult struct {
	MatchType       string `json:"matchType"`
	Matches         int    `json:"matches"`
	MatchResults    int    `json:"matchResults"`
	ScheduledBreaks int    `json:"scheduledBreaks"`
	Rankings        int    `json:"rankings"`
	Alliances       int    `json:"alliances"`
	Replayed        bool   `json:"replayed"`
}

func (web *Web) apiV1AdminDatabaseBackupHandler(w http.ResponseWriter, r *http.Request) {
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	web.apiV1State.destructiveMu.Lock()
	defer web.apiV1State.destructiveMu.Unlock()
	operation := "database.backup"
	if data, found, conflict := web.getApiV1Idempotency(key, operation); found {
		if conflict {
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
			return
		}
		result := data.(apiV1DatabaseBackupResult)
		result.Replayed = true
		writeApiV1Data(w, r, http.StatusOK, result, nil)
		return
	}
	if err := web.arena.Database.Backup(web.arena.EventSettings.Name, "api_manual"); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "backup_failed", "Unable to create database backup.", nil)
		return
	}
	result := apiV1DatabaseBackupResult{Created: true}
	web.storeApiV1Idempotency(key, operation, result)
	web.logApiV1Audit(r, "database.backup", key, "success")
	writeApiV1Data(w, r, http.StatusCreated, result, nil)
}

func (web *Web) apiV1AdminTournamentDataClearHandler(w http.ResponseWriter, r *http.Request) {
	matchType, err := model.MatchTypeFromString(r.PathValue("type"))
	if err != nil || matchType == model.Test {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_match_type", "Match type must be practice, qualification, or playoff.", nil)
		return
	}
	var input apiV1TournamentDataClearInput
	if !decodeApiV1Input(w, r, &input) {
		return
	}
	typeName := strings.ToLower(matchType.String())
	if input.EventId != web.arena.EventSettings.Id || input.Confirmation != "clear:"+typeName {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "confirmation_mismatch", "Event ID and confirmation token must match the requested clear operation.", map[string]string{"confirmation": "must equal clear:" + typeName})
		return
	}
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	web.apiV1State.destructiveMu.Lock()
	defer web.apiV1State.destructiveMu.Unlock()
	operation := "tournament_data.clear." + typeName
	if data, found, conflict := web.getApiV1Idempotency(key, operation); found {
		if conflict {
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
			return
		}
		result := data.(apiV1TournamentDataClearResult)
		result.Replayed = true
		writeApiV1Data(w, r, http.StatusOK, result, nil)
		return
	}
	if err = web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_api_clear_"+typeName); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "backup_failed", "Unable to create the required pre-clear backup.", nil)
		return
	}
	cleared, err := web.arena.Database.ClearTournamentData(matchType)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to clear tournament data.", nil)
		return
	}
	if matchType == model.Playoff {
		web.arena.AllianceSelectionAlliances = []model.Alliance{}
		web.arena.AllianceSelectionRankedTeams = []model.AllianceSelectionRankedTeam{}
	}
	result := apiV1TournamentDataClearResult{typeName, cleared.Matches, cleared.MatchResults, cleared.ScheduledBreaks, cleared.Rankings, cleared.Alliances, false}
	web.storeApiV1Idempotency(key, operation, result)
	web.logApiV1Audit(r, "tournament_data.clear", typeName, "success")
	writeApiV1Data(w, r, http.StatusOK, result, nil)
}

func readApiV1IdempotencyKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" || len(key) > 128 {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_idempotency_key", "Idempotency-Key header is required and must be at most 128 characters.", nil)
		return "", false
	}
	return key, true
}

func (web *Web) getApiV1Idempotency(key, operation string) (any, bool, bool) {
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	record, found := web.apiV1State.idempotency[key]
	return record.Data, found, found && record.Operation != operation
}

func (web *Web) storeApiV1Idempotency(key, operation string, data any) {
	web.apiV1State.idempotencyMu.Lock()
	defer web.apiV1State.idempotencyMu.Unlock()
	web.apiV1State.idempotency[key] = apiV1IdempotencyRecord{Operation: operation, Data: data}
}
