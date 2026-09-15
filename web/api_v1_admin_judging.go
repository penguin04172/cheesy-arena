// Copyright 2026 Team 254. All Rights Reserved.
//
// Administrative judging schedule snapshot and commands.

package web

import (
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
)

type apiV1JudgingScheduleParams struct {
	NumJudges              int `json:"numJudges"`
	DurationMinutes        int `json:"durationMinutes"`
	PreviousSpacingMinutes int `json:"previousSpacingMinutes"`
	NextSpacingMinutes     int `json:"nextSpacingMinutes"`
}

type apiV1JudgingSlot struct {
	Id                  int    `json:"id"`
	Time                string `json:"time"`
	TeamId              int    `json:"teamId"`
	PreviousMatchNumber int    `json:"previousMatchNumber"`
	PreviousMatchTime   string `json:"previousMatchTime"`
	NextMatchNumber     int    `json:"nextMatchNumber"`
	NextMatchTime       string `json:"nextMatchTime"`
	JudgeNumber         int    `json:"judgeNumber"`
}

type apiV1JudgingSchedule struct {
	Params apiV1JudgingScheduleParams `json:"params"`
	Slots  []apiV1JudgingSlot         `json:"slots"`
}

func (web *Web) apiV1AdminJudgingScheduleHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1State.judgingMu.Lock()
	defer web.apiV1State.judgingMu.Unlock()
	web.writeApiV1JudgingSchedule(w, r, http.StatusOK)
}

func (web *Web) apiV1AdminJudgingScheduleGenerateHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1JudgingScheduleParams
	if !decodeApiV1Input(w, r, &input) || !validateApiV1JudgingScheduleParams(w, r, input) {
		return
	}
	web.apiV1State.judgingMu.Lock()
	defer web.apiV1State.judgingMu.Unlock()
	slots, err := web.arena.Database.GetAllJudgingSlots()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load judging schedule.", nil)
		return
	}
	if len(slots) > 0 {
		writeApiV1Error(w, r, http.StatusConflict, "judging_schedule_exists", "Judging schedule already exists. Clear it before generating a new one.", nil)
		return
	}
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load teams.", nil)
		return
	}
	if len(teams) == 0 {
		writeApiV1Error(w, r, http.StatusConflict, "teams_required", "Teams must be imported before generating a judging schedule.", nil)
		return
	}
	matches, err := web.arena.Database.GetMatchesByType(model.Qualification, true)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load qualification matches.", nil)
		return
	}
	if len(matches) < 2 {
		writeApiV1Error(w, r, http.StatusConflict, "qualification_schedule_required", "At least two qualification matches are required.", nil)
		return
	}
	params := tournament.JudgingScheduleParams{
		NumJudges: input.NumJudges, DurationMinutes: input.DurationMinutes,
		PreviousSpacingMinutes: input.PreviousSpacingMinutes, NextSpacingMinutes: input.NextSpacingMinutes,
	}
	if err = tournament.BuildJudgingSchedule(web.arena.Database, params); err != nil {
		if rollbackErr := web.arena.Database.TruncateJudgingSlots(); rollbackErr != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "rollback_failed", "Unable to generate or roll back judging schedule.", nil)
			return
		}
		log.Printf("API v1 judging schedule generation failed requestId=%s: %v", apiV1RequestId(r), err)
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "schedule_generation_failed", "Unable to generate a judging schedule with these parameters.", nil)
		return
	}
	setJudgingScheduleParams(params)
	web.logApiV1Audit(r, "judging_schedule.generate", "all", "success")
	web.writeApiV1JudgingSchedule(w, r, http.StatusCreated)
}

func (web *Web) apiV1AdminJudgingScheduleClearHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1State.judgingMu.Lock()
	defer web.apiV1State.judgingMu.Unlock()
	if err := web.arena.Database.TruncateJudgingSlots(); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to clear judging schedule.", nil)
		return
	}
	web.logApiV1Audit(r, "judging_schedule.clear", "all", "success")
	w.WriteHeader(http.StatusNoContent)
}

func (web *Web) writeApiV1JudgingSchedule(w http.ResponseWriter, r *http.Request, status int) {
	slots, err := web.arena.Database.GetAllJudgingSlots()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load judging schedule.", nil)
		return
	}
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].JudgeNumber != slots[j].JudgeNumber {
			return slots[i].JudgeNumber < slots[j].JudgeNumber
		}
		return slots[i].Time.Before(slots[j].Time)
	})
	response := make([]apiV1JudgingSlot, 0, len(slots))
	for _, slot := range slots {
		response = append(response, apiV1JudgingSlot{
			Id: slot.Id, Time: apiV1Time(slot.Time), TeamId: slot.TeamId,
			PreviousMatchNumber: slot.PreviousMatchNumber, PreviousMatchTime: apiV1Time(slot.PreviousMatchTime),
			NextMatchNumber: slot.NextMatchNumber, NextMatchTime: apiV1Time(slot.NextMatchTime),
			JudgeNumber: slot.JudgeNumber,
		})
	}
	params := getJudgingScheduleParams()
	writeApiV1Data(w, r, status, apiV1JudgingSchedule{
		Params: apiV1JudgingScheduleParams{
			NumJudges: params.NumJudges, DurationMinutes: params.DurationMinutes,
			PreviousSpacingMinutes: params.PreviousSpacingMinutes, NextSpacingMinutes: params.NextSpacingMinutes,
		},
		Slots: response,
	}, nil)
}

func validateApiV1JudgingScheduleParams(w http.ResponseWriter, r *http.Request, input apiV1JudgingScheduleParams) bool {
	fields := make(map[string]string)
	if input.NumJudges < 1 || input.NumJudges > 100 {
		fields["numJudges"] = "must be between 1 and 100"
	}
	if input.DurationMinutes < 1 || input.DurationMinutes > 1440 {
		fields["durationMinutes"] = "must be between 1 and 1440"
	}
	if input.PreviousSpacingMinutes < 1 || input.PreviousSpacingMinutes > 1440 {
		fields["previousSpacingMinutes"] = "must be between 1 and 1440"
	}
	if input.NextSpacingMinutes < 1 || input.NextSpacingMinutes > 1440 {
		fields["nextSpacingMinutes"] = "must be between 1 and 1440"
	}
	if len(fields) > 0 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Judging schedule parameters are invalid.", fields)
		return false
	}
	return true
}

func apiV1Time(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
