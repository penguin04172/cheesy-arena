// Copyright 2026 Team 254. All Rights Reserved.
//
// Public event configuration for version 1 of the JSON API.

package web

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"net/http"
)

type apiV1MatchTiming struct {
	AutoDurationSec            int `json:"autoDurationSec"`
	PauseDurationSec           int `json:"pauseDurationSec"`
	TransitionShiftDurationSec int `json:"transitionShiftDurationSec"`
	ShiftDurationSec           int `json:"shiftDurationSec"`
	EndgameDurationSec         int `json:"endgameDurationSec"`
}

type apiV1PublicEvent struct {
	Id                         int              `json:"id"`
	Name                       string           `json:"name"`
	PlayoffType                string           `json:"playoffType"`
	NumPlayoffAlliances        int              `json:"numPlayoffAlliances"`
	SelectionShowUnpickedTeams bool             `json:"selectionShowUnpickedTeams"`
	NetworkSecurityEnabled     bool             `json:"networkSecurityEnabled"`
	MatchTiming                apiV1MatchTiming `json:"matchTiming"`
}

func (web *Web) apiV1EventHandler(w http.ResponseWriter, r *http.Request) {
	settings := web.arena.EventSettings
	response := apiV1PublicEvent{
		Id:                         settings.Id,
		Name:                       settings.Name,
		PlayoffType:                apiV1PlayoffType(settings.PlayoffType),
		NumPlayoffAlliances:        settings.NumPlayoffAlliances,
		SelectionShowUnpickedTeams: settings.SelectionShowUnpickedTeams,
		NetworkSecurityEnabled:     settings.NetworkSecurityEnabled,
		MatchTiming: apiV1MatchTiming{
			AutoDurationSec:            game.MatchTiming.AutoDurationSec,
			PauseDurationSec:           game.MatchTiming.PauseDurationSec,
			TransitionShiftDurationSec: game.MatchTiming.TransitionShiftDurationSec,
			ShiftDurationSec:           game.MatchTiming.ShiftDurationSec,
			EndgameDurationSec:         game.MatchTiming.EndgameDurationSec,
		},
	}
	w.Header().Set("Cache-Control", "no-cache")
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func apiV1PlayoffType(playoffType model.PlayoffType) string {
	if playoffType == model.SingleEliminationPlayoff {
		return "single_elimination"
	}
	return "double_elimination"
}
