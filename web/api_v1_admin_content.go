// Copyright 2026 Team 254. All Rights Reserved.
//
// Admin CRUD APIs for awards, scheduled breaks, sponsor slides, and lower thirds.

package web

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
	"net/http"
	"strconv"
	"time"
)

type apiV1AdminAward struct {
	Id         int    `json:"id"`
	AwardName  string `json:"awardName"`
	TeamId     int    `json:"teamId"`
	PersonName string `json:"personName"`
}

type apiV1AdminScheduledBreak struct {
	Id              int    `json:"id"`
	MatchType       string `json:"matchType"`
	TypeOrderBefore int    `json:"typeOrderBefore"`
	Time            string `json:"time"`
	DurationSec     int    `json:"durationSec"`
	Description     string `json:"description"`
}

type apiV1AdminLowerThird struct {
	Id           int    `json:"id"`
	TopText      string `json:"topText"`
	BottomText   string `json:"bottomText"`
	DisplayOrder int    `json:"displayOrder"`
	AwardId      int    `json:"awardId"`
}

type apiV1AdminAwardInput struct {
	AwardName  string `json:"awardName"`
	TeamId     int    `json:"teamId"`
	PersonName string `json:"personName"`
}

type apiV1AdminScheduledBreakInput struct {
	Description string `json:"description"`
}

type apiV1AdminSponsorSlideInput struct {
	Subtitle       string `json:"subtitle"`
	Line1          string `json:"line1"`
	Line2          string `json:"line2"`
	Image          string `json:"image"`
	DisplayTimeSec int    `json:"displayTimeSec"`
}

type apiV1AdminLowerThirdInput struct {
	TopText    string `json:"topText"`
	BottomText string `json:"bottomText"`
}

type apiV1AdminReorderInput struct {
	Ids []int `json:"ids"`
}

func (web *Web) apiV1AdminAwardsHandler(w http.ResponseWriter, r *http.Request) {
	awards, err := web.arena.Database.GetAllAwards()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load awards.", nil)
		return
	}
	response := make([]apiV1AdminAward, 0, len(awards))
	for _, award := range awards {
		response = append(response, newApiV1AdminAward(award))
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1AdminAwardCreateHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminAwardInput
	if !decodeApiV1Input(w, r, &input) || !web.validateApiV1Award(w, r, input) {
		return
	}
	award := model.Award{Type: model.JudgedAward, AwardName: input.AwardName, TeamId: input.TeamId, PersonName: input.PersonName}
	if err := tournament.CreateOrUpdateAward(web.arena.Database, &award, true); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to create award.", nil)
		return
	}
	web.logApiV1Audit(r, "award.create", strconv.Itoa(award.Id), "success")
	writeApiV1Data(w, r, http.StatusCreated, newApiV1AdminAward(award), nil)
}

func (web *Web) apiV1AdminAwardUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	award, err := web.arena.Database.GetAwardById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load award.", nil)
		return
	}
	if award == nil || award.Type != model.JudgedAward {
		writeApiV1Error(w, r, http.StatusNotFound, "award_not_found", "Award not found.", nil)
		return
	}
	var input apiV1AdminAwardInput
	if !decodeApiV1Input(w, r, &input) || !web.validateApiV1Award(w, r, input) {
		return
	}
	award.AwardName, award.TeamId, award.PersonName = input.AwardName, input.TeamId, input.PersonName
	if err = tournament.CreateOrUpdateAward(web.arena.Database, award, true); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to update award.", nil)
		return
	}
	web.logApiV1Audit(r, "award.update", strconv.Itoa(id), "success")
	writeApiV1Data(w, r, http.StatusOK, newApiV1AdminAward(*award), nil)
}

func (web *Web) apiV1AdminAwardDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	award, err := web.arena.Database.GetAwardById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load award.", nil)
		return
	}
	if award == nil || award.Type != model.JudgedAward {
		writeApiV1Error(w, r, http.StatusNotFound, "award_not_found", "Award not found.", nil)
		return
	}
	if err = tournament.DeleteAward(web.arena.Database, id); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to delete award.", nil)
		return
	}
	web.logApiV1Audit(r, "award.delete", strconv.Itoa(id), "success")
	w.WriteHeader(http.StatusNoContent)
}

func (web *Web) apiV1AdminScheduledBreaksHandler(w http.ResponseWriter, r *http.Request) {
	matchType := model.Playoff
	if matchTypeValue := r.URL.Query().Get("matchType"); matchTypeValue != "" {
		var err error
		matchType, err = model.MatchTypeFromString(matchTypeValue)
		if err != nil || matchType == model.Test {
			writeApiV1Error(w, r, http.StatusBadRequest, "invalid_match_type", "Match type is invalid.", nil)
			return
		}
	}
	breaks, err := web.arena.Database.GetScheduledBreaksByMatchType(matchType)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load scheduled breaks.", nil)
		return
	}
	response := make([]apiV1AdminScheduledBreak, 0, len(breaks))
	for _, scheduledBreak := range breaks {
		response = append(response, apiV1AdminScheduledBreak{
			Id: scheduledBreak.Id, MatchType: apiV1MatchType(scheduledBreak.MatchType),
			TypeOrderBefore: scheduledBreak.TypeOrderBefore, Time: scheduledBreak.Time.Format(time.RFC3339),
			DurationSec: scheduledBreak.DurationSec, Description: scheduledBreak.Description,
		})
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1AdminScheduledBreakUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	scheduledBreak, err := web.arena.Database.GetScheduledBreakById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load scheduled break.", nil)
		return
	}
	if scheduledBreak == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "scheduled_break_not_found", "Scheduled break not found.", nil)
		return
	}
	var input apiV1AdminScheduledBreakInput
	if !decodeApiV1Input(w, r, &input) {
		return
	}
	scheduledBreak.Description = input.Description
	if err = web.arena.Database.UpdateScheduledBreak(scheduledBreak); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to update scheduled break.", nil)
		return
	}
	web.logApiV1Audit(r, "scheduled_break.update", strconv.Itoa(id), "success")
	writeApiV1Data(w, r, http.StatusOK, apiV1AdminScheduledBreak{
		Id: scheduledBreak.Id, MatchType: apiV1MatchType(scheduledBreak.MatchType),
		TypeOrderBefore: scheduledBreak.TypeOrderBefore, Time: scheduledBreak.Time.Format(time.RFC3339),
		DurationSec: scheduledBreak.DurationSec, Description: scheduledBreak.Description,
	}, nil)
}

func (web *Web) apiV1AdminSponsorSlidesHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1SponsorSlidesHandler(w, r)
}

func (web *Web) apiV1AdminSponsorSlideCreateHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminSponsorSlideInput
	if !decodeApiV1Input(w, r, &input) || !validateApiV1SponsorSlide(w, r, input) {
		return
	}
	slide := model.SponsorSlide{
		Subtitle: input.Subtitle, Line1: input.Line1, Line2: input.Line2, Image: input.Image,
		DisplayTimeSec: input.DisplayTimeSec, DisplayOrder: web.arena.Database.GetNextSponsorSlideDisplayOrder(),
	}
	if err := web.arena.Database.CreateSponsorSlide(&slide); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to create sponsor slide.", nil)
		return
	}
	web.logApiV1Audit(r, "sponsor_slide.create", strconv.Itoa(slide.Id), "success")
	writeApiV1Data(w, r, http.StatusCreated, newApiV1SponsorSlide(slide), nil)
}

func (web *Web) apiV1AdminSponsorSlideUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	slide, err := web.arena.Database.GetSponsorSlideById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load sponsor slide.", nil)
		return
	}
	if slide == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "sponsor_slide_not_found", "Sponsor slide not found.", nil)
		return
	}
	var input apiV1AdminSponsorSlideInput
	if !decodeApiV1Input(w, r, &input) || !validateApiV1SponsorSlide(w, r, input) {
		return
	}
	slide.Subtitle, slide.Line1, slide.Line2, slide.Image = input.Subtitle, input.Line1, input.Line2, input.Image
	slide.DisplayTimeSec = input.DisplayTimeSec
	if err = web.arena.Database.UpdateSponsorSlide(slide); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to update sponsor slide.", nil)
		return
	}
	web.logApiV1Audit(r, "sponsor_slide.update", strconv.Itoa(id), "success")
	writeApiV1Data(w, r, http.StatusOK, newApiV1SponsorSlide(*slide), nil)
}

func (web *Web) apiV1AdminSponsorSlideDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	slide, err := web.arena.Database.GetSponsorSlideById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load sponsor slide.", nil)
		return
	}
	if slide == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "sponsor_slide_not_found", "Sponsor slide not found.", nil)
		return
	}
	if err = web.arena.Database.DeleteSponsorSlide(id); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to delete sponsor slide.", nil)
		return
	}
	web.logApiV1Audit(r, "sponsor_slide.delete", strconv.Itoa(id), "success")
	w.WriteHeader(http.StatusNoContent)
}

func (web *Web) apiV1AdminSponsorSlideReorderHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminReorderInput
	if !decodeApiV1Input(w, r, &input) {
		return
	}
	slides, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load sponsor slides.", nil)
		return
	}
	if message := validateApiV1Order(input.Ids, sponsorSlideIds(slides)); message != "" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_order", message, nil)
		return
	}
	if err = web.arena.Database.ReorderSponsorSlides(input.Ids); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to reorder sponsor slides.", nil)
		return
	}
	web.logApiV1Audit(r, "sponsor_slide.reorder", "all", "success")
	web.apiV1SponsorSlidesHandler(w, r)
}

func (web *Web) apiV1AdminLowerThirdsHandler(w http.ResponseWriter, r *http.Request) {
	lowerThirds, err := web.arena.Database.GetAllLowerThirds()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load lower thirds.", nil)
		return
	}
	response := make([]apiV1AdminLowerThird, 0, len(lowerThirds))
	for _, lowerThird := range lowerThirds {
		response = append(response, newApiV1AdminLowerThird(lowerThird))
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1AdminLowerThirdCreateHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminLowerThirdInput
	if !decodeApiV1Input(w, r, &input) || !validateApiV1LowerThird(w, r, input) {
		return
	}
	lowerThird := model.LowerThird{
		TopText: input.TopText, BottomText: input.BottomText,
		DisplayOrder: web.arena.Database.GetNextLowerThirdDisplayOrder(),
	}
	if err := web.arena.Database.CreateLowerThird(&lowerThird); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to create lower third.", nil)
		return
	}
	web.logApiV1Audit(r, "lower_third.create", strconv.Itoa(lowerThird.Id), "success")
	writeApiV1Data(w, r, http.StatusCreated, newApiV1AdminLowerThird(lowerThird), nil)
}

func (web *Web) apiV1AdminLowerThirdUpdateHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	lowerThird, err := web.arena.Database.GetLowerThirdById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load lower third.", nil)
		return
	}
	if lowerThird == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "lower_third_not_found", "Lower third not found.", nil)
		return
	}
	var input apiV1AdminLowerThirdInput
	if !decodeApiV1Input(w, r, &input) || !validateApiV1LowerThird(w, r, input) {
		return
	}
	lowerThird.TopText, lowerThird.BottomText = input.TopText, input.BottomText
	if err = web.arena.Database.UpdateLowerThird(lowerThird); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to update lower third.", nil)
		return
	}
	web.logApiV1Audit(r, "lower_third.update", strconv.Itoa(id), "success")
	writeApiV1Data(w, r, http.StatusOK, newApiV1AdminLowerThird(*lowerThird), nil)
}

func (web *Web) apiV1AdminLowerThirdDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := readApiV1PathId(w, r)
	if !ok {
		return
	}
	lowerThird, err := web.arena.Database.GetLowerThirdById(id)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load lower third.", nil)
		return
	}
	if lowerThird == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "lower_third_not_found", "Lower third not found.", nil)
		return
	}
	if err = web.arena.Database.DeleteLowerThird(id); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to delete lower third.", nil)
		return
	}
	web.logApiV1Audit(r, "lower_third.delete", strconv.Itoa(id), "success")
	w.WriteHeader(http.StatusNoContent)
}

func (web *Web) apiV1AdminLowerThirdReorderHandler(w http.ResponseWriter, r *http.Request) {
	var input apiV1AdminReorderInput
	if !decodeApiV1Input(w, r, &input) {
		return
	}
	lowerThirds, err := web.arena.Database.GetAllLowerThirds()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load lower thirds.", nil)
		return
	}
	if message := validateApiV1Order(input.Ids, lowerThirdIds(lowerThirds)); message != "" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "invalid_order", message, nil)
		return
	}
	if err = web.arena.Database.ReorderLowerThirds(input.Ids); err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to reorder lower thirds.", nil)
		return
	}
	web.logApiV1Audit(r, "lower_third.reorder", "all", "success")
	web.apiV1AdminLowerThirdsHandler(w, r)
}

func decodeApiV1Input(w http.ResponseWriter, r *http.Request, destination any) bool {
	if err := decodeApiV1Json(w, r, destination); err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_json", "Request body is invalid.", map[string]string{"body": err.Error()})
		return false
	}
	return true
}

func readApiV1PathId(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := apiV1PathId(r)
	if err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_id", err.Error(), nil)
		return 0, false
	}
	return id, true
}

func (web *Web) validateApiV1Award(w http.ResponseWriter, r *http.Request, input apiV1AdminAwardInput) bool {
	if input.AwardName == "" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Award name is required.", map[string]string{"awardName": "must not be empty"})
		return false
	}
	if input.TeamId < 0 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Team ID is invalid.", map[string]string{"teamId": "must not be negative"})
		return false
	}
	if input.TeamId > 0 {
		team, err := web.arena.Database.GetTeamById(input.TeamId)
		if err != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to validate team.", nil)
			return false
		}
		if team == nil {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Team is not present at this event.", map[string]string{"teamId": "team does not exist"})
			return false
		}
	}
	return true
}

func validateApiV1SponsorSlide(w http.ResponseWriter, r *http.Request, input apiV1AdminSponsorSlideInput) bool {
	if input.DisplayTimeSec < 1 || input.DisplayTimeSec > 3600 {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Display time is invalid.", map[string]string{"displayTimeSec": "must be between 1 and 3600"})
		return false
	}
	return true
}

func validateApiV1LowerThird(w http.ResponseWriter, r *http.Request, input apiV1AdminLowerThirdInput) bool {
	if input.TopText == "" && input.BottomText == "" {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Lower third text is required.", map[string]string{"topText": "topText or bottomText must not be empty"})
		return false
	}
	return true
}

func validateApiV1Order(requested []int, existing []int) string {
	if len(requested) != len(existing) {
		return "Order must contain every resource ID exactly once."
	}
	expected := make(map[int]bool, len(existing))
	for _, id := range existing {
		expected[id] = true
	}
	seen := make(map[int]bool, len(requested))
	for _, id := range requested {
		if !expected[id] || seen[id] {
			return "Order must contain every resource ID exactly once."
		}
		seen[id] = true
	}
	return ""
}

func sponsorSlideIds(slides []model.SponsorSlide) []int {
	ids := make([]int, 0, len(slides))
	for _, slide := range slides {
		ids = append(ids, slide.Id)
	}
	return ids
}

func lowerThirdIds(lowerThirds []model.LowerThird) []int {
	ids := make([]int, 0, len(lowerThirds))
	for _, lowerThird := range lowerThirds {
		ids = append(ids, lowerThird.Id)
	}
	return ids
}

func newApiV1AdminAward(award model.Award) apiV1AdminAward {
	return apiV1AdminAward{Id: award.Id, AwardName: award.AwardName, TeamId: award.TeamId, PersonName: award.PersonName}
}

func newApiV1SponsorSlide(slide model.SponsorSlide) apiV1SponsorSlide {
	return apiV1SponsorSlide{
		Id: slide.Id, Subtitle: slide.Subtitle, Line1: slide.Line1, Line2: slide.Line2, Image: slide.Image,
		DisplayTimeSec: slide.DisplayTimeSec, DisplayOrder: slide.DisplayOrder,
	}
}

func newApiV1AdminLowerThird(lowerThird model.LowerThird) apiV1AdminLowerThird {
	return apiV1AdminLowerThird{
		Id: lowerThird.Id, TopText: lowerThird.TopText, BottomText: lowerThird.BottomText,
		DisplayOrder: lowerThird.DisplayOrder, AwardId: lowerThird.AwardId,
	}
}
