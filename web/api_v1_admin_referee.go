// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"
	"sort"

	"github.com/Team254/cheesy-arena/game"
)

type apiV1RefereeFoul struct {
	Index   int  `json:"index"`
	FoulId  int  `json:"foulId"`
	IsMajor bool `json:"isMajor"`
	TeamId  int  `json:"teamId"`
	RuleId  int  `json:"ruleId"`
}

type apiV1RefereeAllianceFouls struct {
	TeamIds []int              `json:"teamIds"`
	Fouls   []apiV1RefereeFoul `json:"fouls"`
}

type apiV1RefereeRule struct {
	Id             int    `json:"id"`
	RuleNumber     string `json:"ruleNumber"`
	IsMajor        bool   `json:"isMajor"`
	IsRankingPoint bool   `json:"isRankingPoint"`
	Description    string `json:"description"`
}

type apiV1RefereeFouls struct {
	MatchId int                       `json:"matchId"`
	Red     apiV1RefereeAllianceFouls `json:"red"`
	Blue    apiV1RefereeAllianceFouls `json:"blue"`
	Rules   []apiV1RefereeRule        `json:"rules"`
}

func (web *Web) apiV1AdminRefereeFoulsHandler(w http.ResponseWriter, r *http.Request) {
	match := web.arena.CurrentMatch
	response := apiV1RefereeFouls{
		MatchId: match.Id,
		Red: apiV1RefereeAllianceFouls{
			TeamIds: []int{match.Red1, match.Red2, match.Red3},
			Fouls:   apiV1RefereeFoulItems(web.arena.RedRealtimeScore.CurrentScore.Fouls),
		},
		Blue: apiV1RefereeAllianceFouls{
			TeamIds: []int{match.Blue1, match.Blue2, match.Blue3},
			Fouls:   apiV1RefereeFoulItems(web.arena.BlueRealtimeScore.CurrentScore.Fouls),
		},
		Rules: apiV1RefereeRuleItems(game.GetAllRules()),
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func apiV1RefereeFoulItems(fouls []game.Foul) []apiV1RefereeFoul {
	items := make([]apiV1RefereeFoul, 0, len(fouls))
	for index, foul := range fouls {
		items = append(items, apiV1RefereeFoul{Index: index, FoulId: foul.FoulId, IsMajor: foul.IsMajor, TeamId: foul.TeamId, RuleId: foul.RuleId})
	}
	return items
}

func apiV1RefereeRuleItems(rules map[int]*game.Rule) []apiV1RefereeRule {
	ids := make([]int, 0, len(rules))
	for id := range rules {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	items := make([]apiV1RefereeRule, 0, len(ids))
	for _, id := range ids {
		rule := rules[id]
		if rule != nil {
			items = append(items, apiV1RefereeRule{Id: rule.Id, RuleNumber: rule.RuleNumber, IsMajor: rule.IsMajor, IsRankingPoint: rule.IsRankingPoint, Description: rule.Description})
		}
	}
	return items
}
