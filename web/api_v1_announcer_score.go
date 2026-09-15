// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
)

type apiV1AnnouncerFoul struct {
	IsMajor         bool   `json:"isMajor"`
	TeamId          int    `json:"teamId"`
	RuleNumber      string `json:"ruleNumber"`
	RuleDescription string `json:"ruleDescription"`
	IsRankingPoint  bool   `json:"isRankingPoint"`
}
type apiV1AnnouncerCard struct {
	TeamId int    `json:"teamId"`
	Card   string `json:"card"`
}
type apiV1AnnouncerRanking struct {
	TeamId       int `json:"teamId"`
	Rank         int `json:"rank"`
	PreviousRank int `json:"previousRank"`
}
type apiV1AnnouncerScoreAlliance struct {
	Summary       apiV1ScoreSummary       `json:"summary"`
	RankingPoints int                     `json:"rankingPoints"`
	Fouls         []apiV1AnnouncerFoul    `json:"fouls"`
	Cards         []apiV1AnnouncerCard    `json:"cards"`
	Rankings      []apiV1AnnouncerRanking `json:"rankings"`
}
type apiV1AnnouncerScore struct {
	MatchId     int                         `json:"matchId"`
	MatchName   string                      `json:"matchName"`
	MatchType   string                      `json:"matchType"`
	Winner      string                      `json:"winner"`
	WinnerClass string                      `json:"winnerClass"`
	Red         apiV1AnnouncerScoreAlliance `json:"red"`
	Blue        apiV1AnnouncerScoreAlliance `json:"blue"`
}

func (web *Web) apiV1AnnouncerDisplayScoreHandler(w http.ResponseWriter, r *http.Request) {
	if web.arena.SavedMatch == nil || web.arena.SavedMatchResult == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "score_not_available", "No saved match score is available.", nil)
		return
	}
	match, result := web.arena.SavedMatch, web.arena.SavedMatchResult
	redSummary, blueSummary := result.RedScoreSummary(), result.BlueScoreSummary()
	redRp, blueRp := redSummary.BonusRankingPoints, blueSummary.BonusRankingPoints
	winner, winnerClass := "tie", "bg-tie"
	switch match.Status {
	case game.RedWonMatch:
		winner, winnerClass, redRp = "red", "bg-danger", redRp+game.GetWinRankingPoints()
	case game.BlueWonMatch:
		winner, winnerClass, blueRp = "blue", "bg-primary", blueRp+game.GetWinRankingPoints()
	case game.TieMatch:
		redRp, blueRp = redRp+1, blueRp+1
	}
	response := apiV1AnnouncerScore{MatchId: match.Id, MatchName: match.LongName, MatchType: apiV1MatchType(match.Type), Winner: winner, WinnerClass: winnerClass,
		Red:  web.newApiV1AnnouncerScoreAlliance(redSummary, redRp, result.RedScore.Fouls, result.RedCards, []int{match.Red1, match.Red2, match.Red3}),
		Blue: web.newApiV1AnnouncerScoreAlliance(blueSummary, blueRp, result.BlueScore.Fouls, result.BlueCards, []int{match.Blue1, match.Blue2, match.Blue3})}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) newApiV1AnnouncerScoreAlliance(summary *game.ScoreSummary, rankingPoints int, fouls []game.Foul, cards map[string]string, teamIds []int) apiV1AnnouncerScoreAlliance {
	foulItems := make([]apiV1AnnouncerFoul, 0, len(fouls))
	for _, foul := range fouls {
		item := apiV1AnnouncerFoul{IsMajor: foul.IsMajor, TeamId: foul.TeamId}
		if rule := foul.Rule(); rule != nil {
			item.RuleNumber, item.RuleDescription, item.IsRankingPoint = rule.RuleNumber, rule.Description, rule.IsRankingPoint
		}
		foulItems = append(foulItems, item)
	}
	cardItems := make([]apiV1AnnouncerCard, 0, len(cards))
	rankingItems := make([]apiV1AnnouncerRanking, 0, len(teamIds))
	for _, teamId := range teamIds {
		if teamId == 0 {
			continue
		}
		if card := cards[strconv.Itoa(teamId)]; card != "" {
			cardItems = append(cardItems, apiV1AnnouncerCard{teamId, card})
		}
		for _, ranking := range web.arena.SavedRankings {
			if ranking.TeamId == teamId {
				rankingItems = append(rankingItems, apiV1AnnouncerRanking{teamId, ranking.Rank, ranking.PreviousRank})
				break
			}
		}
	}
	return apiV1AnnouncerScoreAlliance{newApiV1ScoreSummary(summary), rankingPoints, foulItems, cardItems, rankingItems}
}
