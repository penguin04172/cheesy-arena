// Copyright 2026 Team 254. All Rights Reserved.
//
// Administrative team APIs and asynchronous import/refresh jobs.

package web

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Team254/cheesy-arena/model"
	"github.com/dchest/uniuri"
)

type apiV1AdminTeam struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	Nickname         string `json:"nickname"`
	City             string `json:"city"`
	StateProv        string `json:"stateProv"`
	Country          string `json:"country"`
	SchoolName       string `json:"schoolName"`
	RookieYear       int    `json:"rookieYear"`
	RobotName        string `json:"robotName"`
	Accomplishments  string `json:"accomplishments"`
	WpaKeyConfigured bool   `json:"wpaKeyConfigured"`
	YellowCard       bool   `json:"yellowCard"`
	HasConnected     bool   `json:"hasConnected"`
	FtaNotes         string `json:"ftaNotes"`
}

type apiV1AdminTeamList struct {
	Teams                  []apiV1AdminTeam `json:"teams"`
	CanModifyTeamList      bool             `json:"canModifyTeamList"`
	NetworkSecurityEnabled bool             `json:"networkSecurityEnabled"`
}

type apiV1AdminTeamImportInput struct {
	TeamIds              []int `json:"teamIds"`
	DownloadOfficialData bool  `json:"downloadOfficialData"`
}

type apiV1SecretInput struct {
	Action string `json:"action"`
	Value  string `json:"value"`
}

type apiV1AdminTeamUpdateInput struct {
	Name            string            `json:"name"`
	Nickname        string            `json:"nickname"`
	City            string            `json:"city"`
	StateProv       string            `json:"stateProv"`
	Country         string            `json:"country"`
	SchoolName      string            `json:"schoolName"`
	RookieYear      int               `json:"rookieYear"`
	RobotName       string            `json:"robotName"`
	Accomplishments string            `json:"accomplishments"`
	WpaKey          *apiV1SecretInput `json:"wpaKey"`
	HasConnected    bool              `json:"hasConnected"`
	FtaNotes        string            `json:"ftaNotes"`
}

type apiV1AdminTeamWpaKeysInput struct {
	ReplaceAll bool `json:"replaceAll"`
}

func (web *Web) apiV1AdminTeamsHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load teams.", nil)
		return
	}
	response := make([]apiV1AdminTeam, 0, len(teams))
	for _, team := range teams {
		response = append(response, newApiV1AdminTeam(team))
	}
	writeApiV1Data(w, r, http.StatusOK, apiV1AdminTeamList{
		Teams: response, CanModifyTeamList: web.canModifyTeamList(),
		NetworkSecurityEnabled: web.arena.EventSettings.NetworkSecurityEnabled,
	}, nil)
}

func (web *Web) apiV1AdminTeamHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	team, err := web.arena.Database.GetTeamById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load team.", nil)
		return
	}
	if team == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "team_not_found", "Team not found.", nil)
		return
	}
	writeApiV1Data(w, r, http.StatusOK, newApiV1AdminTeam(*team), nil)
}

func (web *Web) apiV1AdminTeamImportHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminTeamImportInput
	if !decodeApiV1Input(w, r, &input) || !validateApiV1TeamIds(w, r, input.TeamIds) {
		return
	}
	if !web.canModifyTeamList() {
		writeApiV1Error(w, r, http.StatusConflict, "team_list_locked", "Team list cannot be changed after qualification matches are generated.", nil)
		return
	}
	for _, id := range input.TeamIds {
		team, err := web.arena.Database.GetTeamById(id)
		if err != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to validate teams.", nil)
			return
		}
		if team != nil {
			writeApiV1Error(w, r, http.StatusConflict, "team_exists", fmt.Sprintf("Team %d already exists.", id), nil)
			return
		}
	}
	job := web.apiV1State.jobs.create("team_import")
	web.logApiV1Audit(r, "team.import", job.Id, "accepted")
	go web.runApiV1TeamImport(job.Id, input)
	writeApiV1Data(w, r, http.StatusAccepted, job, nil)
}

func (web *Web) apiV1AdminTeamRefreshHandler(w http.ResponseWriter, r *http.Request) {
	job := web.apiV1State.jobs.create("team_refresh")
	web.logApiV1Audit(r, "team.refresh", job.Id, "accepted")
	go web.runApiV1TeamRefresh(job.Id)
	writeApiV1Data(w, r, http.StatusAccepted, job, nil)
}

func (web *Web) apiV1AdminJobHandler(w http.ResponseWriter, r *http.Request) {
	job, ok := web.apiV1State.jobs.get(r.PathValue("id"))
	if !ok {
		writeApiV1Error(w, r, http.StatusNotFound, "job_not_found", "Job not found.", nil)
		return
	}
	writeApiV1Data(w, r, http.StatusOK, job, nil)
}

func (web *Web) apiV1AdminTeamUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	web.apiV1State.teamMutationMu.Lock()
	defer web.apiV1State.teamMutationMu.Unlock()
	team, err := web.arena.Database.GetTeamById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load team.", nil)
		return
	}
	if team == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "team_not_found", "Team not found.", nil)
		return
	}
	var input apiV1AdminTeamUpdateInput
	if !decodeApiV1Input(w, r, &input) || !web.validateApiV1TeamUpdate(w, r, input) {
		return
	}
	team.Name, team.Nickname, team.City = input.Name, input.Nickname, input.City
	team.StateProv, team.Country, team.SchoolName = input.StateProv, input.Country, input.SchoolName
	team.RookieYear, team.RobotName, team.Accomplishments = input.RookieYear, input.RobotName, input.Accomplishments
	team.HasConnected, team.FtaNotes = input.HasConnected, input.FtaNotes
	if input.WpaKey != nil {
		switch input.WpaKey.Action {
		case "replace":
			team.WpaKey = input.WpaKey.Value
		case "clear":
			team.WpaKey = ""
		}
	}
	if err = web.arena.Database.UpdateTeam(team); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to update team.", nil)
		return
	}
	web.logApiV1Audit(r, "team.update", strconv.Itoa(id), "success")
	writeApiV1Data(w, r, http.StatusOK, newApiV1AdminTeam(*team), nil)
}

func (web *Web) apiV1AdminTeamDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	web.apiV1State.teamMutationMu.Lock()
	defer web.apiV1State.teamMutationMu.Unlock()
	if !web.canModifyTeamList() {
		writeApiV1Error(w, r, http.StatusConflict, "team_list_locked", "Team list cannot be changed after qualification matches are generated.", nil)
		return
	}
	team, err := web.arena.Database.GetTeamById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load team.", nil)
		return
	}
	if team == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "team_not_found", "Team not found.", nil)
		return
	}
	if err = web.arena.Database.DeleteTeam(id); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to delete team.", nil)
		return
	}
	web.logApiV1Audit(r, "team.delete", strconv.Itoa(id), "success")
	w.WriteHeader(http.StatusNoContent)
}

func (web *Web) apiV1AdminTeamWpaKeysHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminTeamWpaKeysInput
	if !decodeApiV1Input(w, r, &input) {
		return
	}
	if !web.arena.EventSettings.NetworkSecurityEnabled {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "network_security_disabled", "Network security is disabled.", nil)
		return
	}
	web.apiV1State.teamMutationMu.Lock()
	defer web.apiV1State.teamMutationMu.Unlock()
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load teams.", nil)
		return
	}
	originals := append([]model.Team(nil), teams...)
	updated := 0
	for index := range teams {
		if teams[index].WpaKey != "" && !input.ReplaceAll {
			continue
		}
		teams[index].WpaKey = uniuri.NewLen(wpaKeyLength)
		if err = web.arena.Database.UpdateTeam(&teams[index]); err != nil {
			for rollback := 0; rollback < index; rollback++ {
				_ = web.arena.Database.UpdateTeam(&originals[rollback])
			}
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to generate WPA keys.", nil)
			return
		}
		updated++
	}
	web.logApiV1Audit(r, "team.wpa_keys", "all", "success")
	writeApiV1Data(w, r, http.StatusOK, map[string]int{"updated": updated}, nil)
}

func (web *Web) runApiV1TeamImport(jobId string, input apiV1AdminTeamImportInput) {
	web.apiV1State.jobs.start(jobId)
	web.apiV1State.teamMutationMu.Lock()
	defer web.apiV1State.teamMutationMu.Unlock()
	if !web.canModifyTeamList() {
		web.failApiV1Job(jobId, "Team list is locked.", fmt.Errorf("qualification matches exist"))
		return
	}
	prepared := make([]model.Team, len(input.TeamIds))
	for index, id := range input.TeamIds {
		prepared[index].Id = id
		if input.DownloadOfficialData {
			if err := web.populateOfficialTeamInfo(&prepared[index]); err != nil {
				web.failApiV1Job(jobId, fmt.Sprintf("Unable to download official data for team %d.", id), err)
				return
			}
		}
		web.apiV1State.jobs.progress(jobId, (index+1)*50/len(prepared))
	}
	created := make([]int, 0, len(prepared))
	for index := range prepared {
		if existing, err := web.arena.Database.GetTeamById(prepared[index].Id); err != nil || existing != nil {
			web.rollbackApiV1Teams(created)
			if err != nil {
				web.failApiV1Job(jobId, "Unable to validate teams.", err)
			} else {
				web.failApiV1Job(jobId, fmt.Sprintf("Team %d already exists.", prepared[index].Id), fmt.Errorf("team already exists"))
			}
			return
		}
		if err := web.arena.Database.CreateTeam(&prepared[index]); err != nil {
			web.rollbackApiV1Teams(created)
			web.failApiV1Job(jobId, fmt.Sprintf("Unable to create team %d.", prepared[index].Id), err)
			return
		}
		created = append(created, prepared[index].Id)
		web.apiV1State.jobs.progress(jobId, 50+(index+1)*50/len(prepared))
	}
	web.apiV1State.jobs.finish(jobId, nil)
}

func (web *Web) runApiV1TeamRefresh(jobId string) {
	web.apiV1State.jobs.start(jobId)
	web.apiV1State.teamMutationMu.Lock()
	defer web.apiV1State.teamMutationMu.Unlock()
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		web.failApiV1Job(jobId, "Unable to load teams.", err)
		return
	}
	if len(teams) == 0 {
		web.apiV1State.jobs.finish(jobId, nil)
		return
	}
	originals := append([]model.Team(nil), teams...)
	for index := range teams {
		if err = web.populateOfficialTeamInfo(&teams[index]); err != nil {
			web.failApiV1Job(jobId, fmt.Sprintf("Unable to download official data for team %d.", teams[index].Id), err)
			return
		}
		web.apiV1State.jobs.progress(jobId, (index+1)*50/len(teams))
	}
	for index := range teams {
		if err = web.arena.Database.UpdateTeam(&teams[index]); err != nil {
			for rollback := 0; rollback < index; rollback++ {
				_ = web.arena.Database.UpdateTeam(&originals[rollback])
			}
			web.failApiV1Job(jobId, fmt.Sprintf("Unable to update team %d.", teams[index].Id), err)
			return
		}
		web.apiV1State.jobs.progress(jobId, 50+(index+1)*50/len(teams))
	}
	web.apiV1State.jobs.finish(jobId, nil)
}

func (web *Web) rollbackApiV1Teams(ids []int) {
	for _, id := range ids {
		_ = web.arena.Database.DeleteTeam(id)
	}
}

func validateApiV1TeamIds(w http.ResponseWriter, r *http.Request, ids []int) bool {
	if len(ids) == 0 || len(ids) > 1000 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Between 1 and 1000 team IDs are required.", map[string]string{"teamIds": "must contain between 1 and 1000 IDs"})
		return false
	}
	seen := make(map[int]bool, len(ids))
	for _, id := range ids {
		if id < 1 || seen[id] {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Team IDs must be positive and unique.", map[string]string{"teamIds": "must contain positive unique IDs"})
			return false
		}
		seen[id] = true
	}
	return true
}

func (web *Web) validateApiV1TeamUpdate(w http.ResponseWriter, r *http.Request, input apiV1AdminTeamUpdateInput) bool {
	if input.RookieYear < 0 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Rookie year is invalid.", map[string]string{"rookieYear": "must not be negative"})
		return false
	}
	if input.WpaKey == nil || input.WpaKey.Action == "keep" {
		return true
	}
	if !web.arena.EventSettings.NetworkSecurityEnabled {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "network_security_disabled", "Network security is disabled.", map[string]string{"wpaKey": "cannot be changed while network security is disabled"})
		return false
	}
	if input.WpaKey.Action == "clear" {
		return true
	}
	if input.WpaKey.Action != "replace" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "WPA key action is invalid.", map[string]string{"wpaKey.action": "must be keep, replace, or clear"})
		return false
	}
	if len(input.WpaKey.Value) < 8 || len(input.WpaKey.Value) > 63 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "WPA key is invalid.", map[string]string{"wpaKey.value": "must be between 8 and 63 characters"})
		return false
	}
	return true
}

func newApiV1AdminTeam(team model.Team) apiV1AdminTeam {
	return apiV1AdminTeam{
		Id: team.Id, Name: team.Name, Nickname: team.Nickname, City: team.City, StateProv: team.StateProv,
		Country: team.Country, SchoolName: team.SchoolName, RookieYear: team.RookieYear, RobotName: team.RobotName,
		Accomplishments: team.Accomplishments, WpaKeyConfigured: team.WpaKey != "", YellowCard: team.YellowCard,
		HasConnected: team.HasConnected, FtaNotes: team.FtaNotes,
	}
}
