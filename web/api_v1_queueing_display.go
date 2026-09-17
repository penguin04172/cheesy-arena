// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"fmt"
	"net/http"
	"time"
)

type apiV1QueueingAlliance struct {
	TeamIds           []int `json:"teamIds"`
	OffFieldTeamIds   []int `json:"offFieldTeamIds"`
	PlayoffAllianceId int   `json:"playoffAllianceId"`
}

type apiV1QueueingMatch struct {
	Position      string                `json:"position"`
	PositionLabel string                `json:"positionLabel"`
	Id            int                   `json:"id"`
	ShortName     string                `json:"shortName"`
	ScheduledAt   string                `json:"scheduledAt"`
	DisplayTime   string                `json:"displayTime"`
	Red           apiV1QueueingAlliance `json:"red"`
	Blue          apiV1QueueingAlliance `json:"blue"`
}

func (web *Web) apiV1QueueingDisplayMatchesHandler(w http.ResponseWriter, r *http.Request) {
	items, err := web.buildApiV1QueueingMatches()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load queueing matches.", nil)
		return
	}
	writeApiV1Data(w, r, http.StatusOK, items, nil)
}

func (web *Web) buildApiV1QueueingMatches() ([]apiV1QueueingMatch, error) {
	snapshot, err := web.buildQueueingDisplayMatchList()
	if err != nil {
		return nil, err
	}
	items := make([]apiV1QueueingMatch, 0, len(snapshot.Matches))
	for index, match := range snapshot.Matches {
		position, label := apiV1QueueingPosition(index)
		scheduledAt := ""
		displayTime := ""
		if !match.Time.IsZero() {
			scheduledAt = match.Time.Format(time.RFC3339)
			displayTime = match.Time.Local().Format("3:04 PM")
		}
		items = append(items, apiV1QueueingMatch{
			Position: position, PositionLabel: label, Id: match.Id, ShortName: match.ShortName,
			ScheduledAt: scheduledAt, DisplayTime: displayTime,
			Red:  apiV1QueueingAlliance{TeamIds: nonzeroTeamIds(match.Red1, match.Red2, match.Red3), OffFieldTeamIds: snapshot.RedOffFieldTeams[index], PlayoffAllianceId: match.PlayoffRedAlliance},
			Blue: apiV1QueueingAlliance{TeamIds: nonzeroTeamIds(match.Blue1, match.Blue2, match.Blue3), OffFieldTeamIds: snapshot.BlueOffFieldTeams[index], PlayoffAllianceId: match.PlayoffBlueAlliance},
		})
	}
	return items, nil
}

func apiV1QueueingPosition(index int) (string, string) {
	switch index {
	case 0:
		return "on_field", "On Field"
	case 1:
		return "on_deck", "On Deck"
	default:
		return fmt.Sprintf("up_in_%d", index), fmt.Sprintf("Up In %d", index)
	}
}

func nonzeroTeamIds(ids ...int) []int {
	result := make([]int, 0, len(ids))
	for _, id := range ids {
		if id != 0 {
			result = append(result, id)
		}
	}
	return result
}
