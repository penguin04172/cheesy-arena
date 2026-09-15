// Copyright 2026 Team 254. All Rights Reserved.
//
// Sectioned administrative settings APIs. Secrets are write-only.

package web

import (
	"errors"
	"net/http"

	"github.com/Team254/cheesy-arena/model"
)

var errApiV1ResponseWritten = errors.New("API response already written")

type apiV1EventSettings struct {
	Name                       string `json:"name"`
	PlayoffType                string `json:"playoffType"`
	NumPlayoffAlliances        int    `json:"numPlayoffAlliances"`
	SelectionRound2Order       string `json:"selectionRound2Order"`
	SelectionRound3Order       string `json:"selectionRound3Order"`
	SelectionShowUnpickedTeams bool   `json:"selectionShowUnpickedTeams"`
	AutoAudienceDisplayEnabled bool   `json:"autoAudienceDisplayEnabled"`
	AdminPasswordConfigured    bool   `json:"adminPasswordConfigured"`
}

type apiV1EventSettingsInput struct {
	Name                       *string           `json:"name"`
	PlayoffType                *string           `json:"playoffType"`
	NumPlayoffAlliances        *int              `json:"numPlayoffAlliances"`
	SelectionRound2Order       *string           `json:"selectionRound2Order"`
	SelectionRound3Order       *string           `json:"selectionRound3Order"`
	SelectionShowUnpickedTeams *bool             `json:"selectionShowUnpickedTeams"`
	AutoAudienceDisplayEnabled *bool             `json:"autoAudienceDisplayEnabled"`
	AdminPassword              *apiV1SecretInput `json:"adminPassword"`
}

type apiV1IntegrationSettings struct {
	TbaDownloadEnabled          bool   `json:"tbaDownloadEnabled"`
	TbaPublishingEnabled        bool   `json:"tbaPublishingEnabled"`
	TbaEventCode                string `json:"tbaEventCode"`
	TbaSecretIdConfigured       bool   `json:"tbaSecretIdConfigured"`
	TbaSecretConfigured         bool   `json:"tbaSecretConfigured"`
	NexusEnabled                bool   `json:"nexusEnabled"`
	NexusAutoQueueEnabled       bool   `json:"nexusAutoQueueEnabled"`
	NexusAutoQueueKeyConfigured bool   `json:"nexusAutoQueueKeyConfigured"`
}

type apiV1IntegrationSettingsInput struct {
	TbaDownloadEnabled    *bool             `json:"tbaDownloadEnabled"`
	TbaPublishingEnabled  *bool             `json:"tbaPublishingEnabled"`
	TbaEventCode          *string           `json:"tbaEventCode"`
	TbaSecretId           *apiV1SecretInput `json:"tbaSecretId"`
	TbaSecret             *apiV1SecretInput `json:"tbaSecret"`
	NexusEnabled          *bool             `json:"nexusEnabled"`
	NexusAutoQueueEnabled *bool             `json:"nexusAutoQueueEnabled"`
	NexusAutoQueueKey     *apiV1SecretInput `json:"nexusAutoQueueKey"`
}

type apiV1GameSettings struct {
	AutoDurationSec            int `json:"autoDurationSec"`
	PauseDurationSec           int `json:"pauseDurationSec"`
	TransitionShiftDurationSec int `json:"transitionShiftDurationSec"`
	ShiftDurationSec           int `json:"shiftDurationSec"`
	EndgameDurationSec         int `json:"endgameDurationSec"`
	EnergizedBonusThreshold    int `json:"energizedBonusThreshold"`
	SuperchargedBonusThreshold int `json:"superchargedBonusThreshold"`
	TraversalBonusThreshold    int `json:"traversalBonusThreshold"`
}

type apiV1GameSettingsInput struct {
	AutoDurationSec            *int `json:"autoDurationSec"`
	PauseDurationSec           *int `json:"pauseDurationSec"`
	TransitionShiftDurationSec *int `json:"transitionShiftDurationSec"`
	ShiftDurationSec           *int `json:"shiftDurationSec"`
	EndgameDurationSec         *int `json:"endgameDurationSec"`
	EnergizedBonusThreshold    *int `json:"energizedBonusThreshold"`
	SuperchargedBonusThreshold *int `json:"superchargedBonusThreshold"`
	TraversalBonusThreshold    *int `json:"traversalBonusThreshold"`
}

func (web *Web) apiV1AdminSettingsHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1State.settingsMu.Lock()
	defer web.apiV1State.settingsMu.Unlock()
	web.writeApiV1SettingsSection(w, r, r.PathValue("section"))
}

func (web *Web) apiV1AdminSettingsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !settingsSaveAllowed(web.arena.MatchState) {
		writeApiV1Error(w, r, http.StatusConflict, "settings_locked", "Settings cannot be changed while a match is in progress or uncommitted.", nil)
		return
	}
	web.apiV1State.settingsMu.Lock()
	defer web.apiV1State.settingsMu.Unlock()
	original := *web.arena.EventSettings
	candidate := original
	section := r.PathValue("section")
	var err error
	switch section {
	case "event":
		err = web.applyApiV1EventSettings(w, r, &candidate)
	case "integrations":
		err = web.applyApiV1IntegrationSettings(w, r, &candidate)
	case "game":
		err = web.applyApiV1GameSettings(w, r, &candidate)
	case "network", "hardware":
		writeApiV1Error(w, r, http.StatusNotImplemented, "section_not_implemented", "This settings section is not implemented yet.", nil)
		return
	default:
		writeApiV1Error(w, r, http.StatusNotFound, "settings_section_not_found", "Settings section not found.", nil)
		return
	}
	if err != nil {
		return
	}
	if err = web.arena.Database.UpdateEventSettings(&candidate); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to save settings.", nil)
		return
	}
	if err = web.arena.LoadSettings(); err != nil {
		_ = web.arena.Database.UpdateEventSettings(&original)
		_ = web.arena.LoadSettings()
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "settings_apply_failed", "Settings could not be applied; the previous settings were restored.", nil)
		return
	}
	if candidate.AdminPassword != original.AdminPassword {
		if err = web.arena.Database.TruncateUserSessions(); err != nil {
			_ = web.arena.Database.UpdateEventSettings(&original)
			_ = web.arena.LoadSettings()
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to invalidate existing sessions; settings were restored.", nil)
			return
		}
	}
	web.logApiV1Audit(r, "settings.update", section, "success")
	web.writeApiV1SettingsSection(w, r, section)
}

func (web *Web) applyApiV1EventSettings(w http.ResponseWriter, r *http.Request, settings *model.EventSettings) error {
	var input apiV1EventSettingsInput
	if !decodeApiV1Input(w, r, &input) {
		return errApiV1ResponseWritten
	}
	if input.Name != nil {
		settings.Name = *input.Name
	}
	if input.PlayoffType != nil {
		switch *input.PlayoffType {
		case "double_elimination":
			settings.PlayoffType = model.DoubleEliminationPlayoff
		case "single_elimination":
			settings.PlayoffType = model.SingleEliminationPlayoff
		default:
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Playoff type is invalid.", map[string]string{"playoffType": "must be double_elimination or single_elimination"})
			return errApiV1ResponseWritten
		}
	}
	if input.NumPlayoffAlliances != nil {
		settings.NumPlayoffAlliances = *input.NumPlayoffAlliances
	}
	if input.SelectionRound2Order != nil {
		settings.SelectionRound2Order = *input.SelectionRound2Order
	}
	if input.SelectionRound3Order != nil {
		settings.SelectionRound3Order = *input.SelectionRound3Order
	}
	if input.SelectionShowUnpickedTeams != nil {
		settings.SelectionShowUnpickedTeams = *input.SelectionShowUnpickedTeams
	}
	if input.AutoAudienceDisplayEnabled != nil {
		settings.AutoAudienceDisplayEnabled = *input.AutoAudienceDisplayEnabled
	}
	if input.AdminPassword != nil {
		if !applyApiV1Secret(w, r, "adminPassword", input.AdminPassword, &settings.AdminPassword) {
			return errApiV1ResponseWritten
		}
	}
	if settings.Name == "" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Event name is required.", map[string]string{"name": "must not be empty"})
		return errApiV1ResponseWritten
	}
	validAlliances := settings.PlayoffType == model.SingleEliminationPlayoff && settings.NumPlayoffAlliances >= 2 && settings.NumPlayoffAlliances <= 16 || settings.PlayoffType == model.DoubleEliminationPlayoff && (settings.NumPlayoffAlliances == 4 || settings.NumPlayoffAlliances == 8)
	if !validAlliances {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Number of playoff alliances is invalid for this playoff type.", map[string]string{"numPlayoffAlliances": "single: 2-16; double: 4 or 8"})
		return errApiV1ResponseWritten
	}
	current := web.arena.EventSettings
	if current.PlayoffType != settings.PlayoffType || current.NumPlayoffAlliances != settings.NumPlayoffAlliances {
		alliances, err := web.arena.Database.GetAllAlliances()
		if err != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to validate alliances.", nil)
			return err
		}
		if len(alliances) > 0 {
			writeApiV1Error(w, r, http.StatusConflict, "alliance_selection_finalized", "Playoff type or size cannot change after alliance selection is finalized.", nil)
			return errApiV1ResponseWritten
		}
	}
	return nil
}

func (web *Web) applyApiV1IntegrationSettings(w http.ResponseWriter, r *http.Request, settings *model.EventSettings) error {
	var input apiV1IntegrationSettingsInput
	if !decodeApiV1Input(w, r, &input) {
		return errApiV1ResponseWritten
	}
	if input.TbaDownloadEnabled != nil {
		settings.TbaDownloadEnabled = *input.TbaDownloadEnabled
	}
	if input.TbaPublishingEnabled != nil {
		settings.TbaPublishingEnabled = *input.TbaPublishingEnabled
	}
	if input.TbaEventCode != nil {
		settings.TbaEventCode = *input.TbaEventCode
	}
	if input.NexusEnabled != nil {
		settings.NexusEnabled = *input.NexusEnabled
	}
	if input.NexusAutoQueueEnabled != nil {
		settings.NexusAutoQueueEnabled = *input.NexusAutoQueueEnabled
	}
	for _, secret := range []struct {
		name  string
		input *apiV1SecretInput
		value *string
	}{
		{"tbaSecretId", input.TbaSecretId, &settings.TbaSecretId}, {"tbaSecret", input.TbaSecret, &settings.TbaSecret}, {"nexusAutoQueueKey", input.NexusAutoQueueKey, &settings.NexusAutoQueueKey},
	} {
		if secret.input != nil && !applyApiV1Secret(w, r, secret.name, secret.input, secret.value) {
			return errApiV1ResponseWritten
		}
	}
	return nil
}

func (web *Web) applyApiV1GameSettings(w http.ResponseWriter, r *http.Request, settings *model.EventSettings) error {
	var input apiV1GameSettingsInput
	if !decodeApiV1Input(w, r, &input) {
		return errApiV1ResponseWritten
	}
	values := []struct {
		input  *int
		output *int
		name   string
	}{
		{input.AutoDurationSec, &settings.AutoDurationSec, "autoDurationSec"}, {input.PauseDurationSec, &settings.PauseDurationSec, "pauseDurationSec"},
		{input.TransitionShiftDurationSec, &settings.TransitionShiftDurationSec, "transitionShiftDurationSec"}, {input.ShiftDurationSec, &settings.ShiftDurationSec, "shiftDurationSec"},
		{input.EndgameDurationSec, &settings.EndgameDurationSec, "endgameDurationSec"}, {input.EnergizedBonusThreshold, &settings.EnergizedBonusThreshold, "energizedBonusThreshold"},
		{input.SuperchargedBonusThreshold, &settings.SuperchargedBonusThreshold, "superchargedBonusThreshold"}, {input.TraversalBonusThreshold, &settings.TraversalBonusThreshold, "traversalBonusThreshold"},
	}
	for _, value := range values {
		if value.input != nil {
			if *value.input < 0 || *value.input > 86400 {
				writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Game setting is out of range.", map[string]string{value.name: "must be between 0 and 86400"})
				return errApiV1ResponseWritten
			}
			*value.output = *value.input
		}
	}
	return nil
}

func applyApiV1Secret(w http.ResponseWriter, r *http.Request, name string, input *apiV1SecretInput, value *string) bool {
	switch input.Action {
	case "keep":
		return true
	case "clear":
		*value = ""
		return true
	case "replace":
		if input.Value == "" {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Replacement secret must not be empty.", map[string]string{name + ".value": "must not be empty"})
			return false
		}
		*value = input.Value
		return true
	default:
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Secret action is invalid.", map[string]string{name + ".action": "must be keep, replace, or clear"})
		return false
	}
}

func (web *Web) writeApiV1SettingsSection(w http.ResponseWriter, r *http.Request, section string) {
	settings := web.arena.EventSettings
	var data any
	switch section {
	case "event":
		data = apiV1EventSettings{
			Name: settings.Name, PlayoffType: apiV1PlayoffType(settings.PlayoffType),
			NumPlayoffAlliances: settings.NumPlayoffAlliances, SelectionRound2Order: settings.SelectionRound2Order,
			SelectionRound3Order: settings.SelectionRound3Order, SelectionShowUnpickedTeams: settings.SelectionShowUnpickedTeams,
			AutoAudienceDisplayEnabled: settings.AutoAudienceDisplayEnabled, AdminPasswordConfigured: settings.AdminPassword != "",
		}
	case "integrations":
		data = apiV1IntegrationSettings{
			TbaDownloadEnabled: settings.TbaDownloadEnabled, TbaPublishingEnabled: settings.TbaPublishingEnabled,
			TbaEventCode: settings.TbaEventCode, TbaSecretIdConfigured: settings.TbaSecretId != "",
			TbaSecretConfigured: settings.TbaSecret != "", NexusEnabled: settings.NexusEnabled,
			NexusAutoQueueEnabled: settings.NexusAutoQueueEnabled, NexusAutoQueueKeyConfigured: settings.NexusAutoQueueKey != "",
		}
	case "game":
		data = apiV1GameSettings{
			AutoDurationSec: settings.AutoDurationSec, PauseDurationSec: settings.PauseDurationSec,
			TransitionShiftDurationSec: settings.TransitionShiftDurationSec, ShiftDurationSec: settings.ShiftDurationSec,
			EndgameDurationSec: settings.EndgameDurationSec, EnergizedBonusThreshold: settings.EnergizedBonusThreshold,
			SuperchargedBonusThreshold: settings.SuperchargedBonusThreshold, TraversalBonusThreshold: settings.TraversalBonusThreshold,
		}
	case "network", "hardware":
		writeApiV1Error(w, r, http.StatusNotImplemented, "section_not_implemented", "This settings section is not implemented yet.", nil)
		return
	default:
		writeApiV1Error(w, r, http.StatusNotFound, "settings_section_not_found", "Settings section not found.", nil)
		return
	}
	writeApiV1Data(w, r, http.StatusOK, data, nil)
}
