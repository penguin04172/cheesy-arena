// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/model"
)

type apiV1AnnouncerTeam struct {
	Id              int    `json:"id"`
	Nickname        string `json:"nickname"`
	SchoolName      string `json:"schoolName"`
	City            string `json:"city"`
	StateProv       string `json:"stateProv"`
	Country         string `json:"country"`
	RookieYear      int    `json:"rookieYear"`
	RobotName       string `json:"robotName"`
	Accomplishments string `json:"accomplishments"`
	Rank            *int   `json:"rank"`
	IsOffField      bool   `json:"isOffField"`
}

type apiV1AnnouncerAlliance struct {
	PlayoffAllianceId int                   `json:"playoffAllianceId"`
	Teams             []*apiV1AnnouncerTeam `json:"teams"`
}

type apiV1AnnouncerMatch struct {
	Id       int                    `json:"id"`
	Type     string                 `json:"type"`
	LongName string                 `json:"longName"`
	Red      apiV1AnnouncerAlliance `json:"red"`
	Blue     apiV1AnnouncerAlliance `json:"blue"`
}

func (web *Web) apiV1AnnouncerDisplayMatchHandler(w http.ResponseWriter, r *http.Request) {
	response, err := web.buildApiV1AnnouncerMatch()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load announcer match data.", nil)
		return
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) buildApiV1AnnouncerMatch() (apiV1AnnouncerMatch, error) {
	match := web.arena.CurrentMatch
	red, err := web.buildApiV1AnnouncerAlliance(match.PlayoffRedAlliance, "R1", "R2", "R3")
	if err != nil {
		return apiV1AnnouncerMatch{}, err
	}
	blue, err := web.buildApiV1AnnouncerAlliance(match.PlayoffBlueAlliance, "B1", "B2", "B3")
	if err != nil {
		return apiV1AnnouncerMatch{}, err
	}
	return apiV1AnnouncerMatch{Id: match.Id, Type: apiV1MatchType(match.Type), LongName: match.LongName, Red: red, Blue: blue}, nil
}

func (web *Web) buildApiV1AnnouncerAlliance(allianceId int, stations ...string) (apiV1AnnouncerAlliance, error) {
	teams := make([]*apiV1AnnouncerTeam, 0, len(stations)+1)
	fieldTeamIds := make(map[int]bool)
	for _, station := range stations {
		team := web.arena.AllianceStations[station].Team
		item, err := web.newApiV1AnnouncerTeam(team, false)
		if err != nil {
			return apiV1AnnouncerAlliance{}, err
		}
		teams = append(teams, item)
		if team != nil {
			fieldTeamIds[team.Id] = true
		}
	}
	if allianceId != 0 {
		alliance, err := web.arena.Database.GetAllianceById(allianceId)
		if err != nil {
			return apiV1AnnouncerAlliance{}, err
		}
		if alliance != nil {
			for _, teamId := range alliance.TeamIds {
				if fieldTeamIds[teamId] {
					continue
				}
				team, err := web.arena.Database.GetTeamById(teamId)
				if err != nil {
					return apiV1AnnouncerAlliance{}, err
				}
				item, err := web.newApiV1AnnouncerTeam(team, true)
				if err != nil {
					return apiV1AnnouncerAlliance{}, err
				}
				teams = append(teams, item)
			}
		}
	}
	return apiV1AnnouncerAlliance{PlayoffAllianceId: allianceId, Teams: teams}, nil
}

func (web *Web) newApiV1AnnouncerTeam(team *model.Team, isOffField bool) (*apiV1AnnouncerTeam, error) {
	if team == nil {
		return nil, nil
	}
	var rank *int
	ranking, err := web.arena.Database.GetRankingForTeam(team.Id)
	if err != nil {
		return nil, err
	}
	if ranking != nil {
		value := ranking.Rank
		rank = &value
	}
	return &apiV1AnnouncerTeam{
		Id: team.Id, Nickname: team.Nickname, SchoolName: team.SchoolName, City: team.City,
		StateProv: team.StateProv, Country: team.Country, RookieYear: team.RookieYear,
		RobotName: team.RobotName, Accomplishments: team.Accomplishments, Rank: rank, IsOffField: isOffField,
	}, nil
}
