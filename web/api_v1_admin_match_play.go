// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
)

type apiV1MatchPlayListItem struct {
	Id            int    `json:"id"`
	ShortName     string `json:"shortName"`
	ScheduledAt   string `json:"scheduledAt"`
	DisplayTime   string `json:"displayTime"`
	Status        string `json:"status"`
	ColorClass    string `json:"colorClass"`
	CanShowResult bool   `json:"canShowResult"`
}

type apiV1MatchPlayMatches struct {
	CurrentMatchType string                              `json:"currentMatchType"`
	MatchesByType    map[string][]apiV1MatchPlayListItem `json:"matchesByType"`
}

func (web *Web) apiV1AdminMatchPlayMatchesHandler(w http.ResponseWriter, r *http.Request) {
	snapshot, err := web.buildMatchPlayMatchListSnapshot()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load match list.", nil)
		return
	}
	response := apiV1MatchPlayMatches{CurrentMatchType: apiV1MatchType(snapshot.CurrentMatchType), MatchesByType: make(map[string][]apiV1MatchPlayListItem, 3)}
	for _, matchType := range []model.MatchType{model.Practice, model.Qualification, model.Playoff} {
		items := make([]apiV1MatchPlayListItem, 0, len(snapshot.MatchesByType[matchType]))
		for _, item := range snapshot.MatchesByType[matchType] {
			scheduledAt := ""
			if !item.ScheduledAt.IsZero() {
				scheduledAt = item.ScheduledAt.Format(time.RFC3339)
			}
			items = append(items, apiV1MatchPlayListItem{Id: item.Id, ShortName: item.ShortName, ScheduledAt: scheduledAt, DisplayTime: item.Time, Status: apiV1MatchStatus(item.Status), ColorClass: item.ColorClass, CanShowResult: item.Status != game.MatchScheduled})
		}
		response.MatchesByType[apiV1MatchType(matchType)] = items
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}
