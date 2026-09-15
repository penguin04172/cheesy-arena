// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Team254/cheesy-arena/model"
)

const maxApiV1DatabaseUploadBytes = 256 * 1024 * 1024

type apiV1DatabaseBackupResult struct {
	Created  bool `json:"created"`
	Replayed bool `json:"replayed"`
}

type apiV1DatabaseRestoreResult struct {
	Restored bool `json:"restored"`
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

func (web *Web) apiV1AdminDatabaseRestoreHandler(w http.ResponseWriter, r *http.Request) {
	if !settingsSaveAllowed(web.arena.MatchState) {
		writeApiV1Error(w, r, http.StatusConflict, "settings_locked", "Database cannot be restored while a match is in progress or uncommitted.", nil)
		return
	}
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	web.apiV1State.destructiveMu.Lock()
	defer web.apiV1State.destructiveMu.Unlock()
	operation := "database.restore"
	if data, found, conflict := web.getApiV1Idempotency(key, operation); found {
		if conflict {
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
			return
		}
		result := data.(apiV1DatabaseRestoreResult)
		result.Replayed = true
		writeApiV1Data(w, r, http.StatusOK, result, nil)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxApiV1DatabaseUploadBytes)
	if err := r.ParseMultipartForm(8 * 1024 * 1024); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeApiV1Error(w, r, http.StatusRequestEntityTooLarge, "upload_too_large", "Database backup exceeds the 256 MiB limit.", nil)
		} else {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_multipart", "Multipart upload is invalid.", nil)
		}
		return
	}
	defer r.MultipartForm.RemoveAll()
	if r.FormValue("confirmation") != "restore" || r.FormValue("eventId") != strconv.Itoa(web.arena.EventSettings.Id) {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "confirmation_mismatch", "Current event ID and confirmation value restore are required.", nil)
		return
	}
	file, _, err := r.FormFile("databaseFile")
	if err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "database_file_required", "A database backup file is required.", nil)
		return
	}
	defer file.Close()
	databaseDir := filepath.Dir(web.arena.Database.Path)
	uploaded, err := os.CreateTemp(databaseDir, "api-restore-upload-*.db")
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "temporary_file_failed", "Unable to stage database backup.", nil)
		return
	}
	uploadedPath := uploaded.Name()
	defer os.Remove(uploadedPath)
	if _, err = io.Copy(uploaded, file); err != nil {
		uploaded.Close()
		writeApiV1Error(w, r, http.StatusInternalServerError, "upload_failed", "Unable to stage database backup.", nil)
		return
	}
	if err = uploaded.Close(); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "upload_failed", "Unable to stage database backup.", nil)
		return
	}
	if err = web.arena.Database.ValidateRestoreFile(uploadedPath); err != nil {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_database_backup", "Uploaded file is not a valid Cheesy Arena database backup.", nil)
		return
	}
	rollback, err := os.CreateTemp(databaseDir, "api-restore-rollback-*.db")
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "temporary_file_failed", "Unable to create rollback snapshot.", nil)
		return
	}
	rollbackPath := rollback.Name()
	defer os.Remove(rollbackPath)
	if err = web.arena.Database.WriteBackup(rollback); err != nil {
		rollback.Close()
		writeApiV1Error(w, r, http.StatusInternalServerError, "backup_failed", "Unable to create rollback snapshot.", nil)
		return
	}
	if err = rollback.Close(); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "backup_failed", "Unable to create rollback snapshot.", nil)
		return
	}
	if err = web.arena.Database.Backup(web.arena.EventSettings.Name, "pre_api_restore"); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "backup_failed", "Unable to create the required pre-restore backup.", nil)
		return
	}
	if err = web.arena.Database.RestoreFrom(uploadedPath); err != nil {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_database_backup", "Uploaded file is not a valid Cheesy Arena database backup.", nil)
		return
	}
	if err = web.arena.LoadSettings(); err != nil {
		rollbackErr := web.arena.Database.RestoreFrom(rollbackPath)
		loadErr := web.arena.LoadSettings()
		if rollbackErr != nil || loadErr != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "rollback_failed", "Restore failed and automatic rollback could not be completed.", nil)
		} else {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "settings_apply_failed", "Restored settings could not be applied; the previous database was restored.", nil)
		}
		return
	}
	if err = web.arena.Database.TruncateUserSessions(); err != nil {
		rollbackErr := web.arena.Database.RestoreFrom(rollbackPath)
		loadErr := web.arena.LoadSettings()
		if rollbackErr != nil || loadErr != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "rollback_failed", "Session invalidation failed and automatic rollback could not be completed.", nil)
		} else {
			writeApiV1Error(w, r, http.StatusInternalServerError, "session_invalidation_failed", "Existing sessions could not be invalidated; the previous database was restored.", nil)
		}
		return
	}
	result := apiV1DatabaseRestoreResult{Restored: true}
	web.storeApiV1Idempotency(key, operation, result)
	web.logApiV1Audit(r, "database.restore", key, "success")
	writeApiV1Data(w, r, http.StatusOK, result, nil)
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
