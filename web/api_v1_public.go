// Copyright 2026 Team 254. All Rights Reserved.
//
// Public read-only resources for version 1 of the JSON API.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/partner"
	"github.com/Team254/cheesy-arena/playoff"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type apiV1Match struct {
	Id                  int               `json:"id"`
	Type                string            `json:"type"`
	TypeOrder           int               `json:"typeOrder"`
	Time                string            `json:"time"`
	LongName            string            `json:"longName"`
	ShortName           string            `json:"shortName"`
	NameDetail          string            `json:"nameDetail"`
	PlayoffMatchGroupId string            `json:"playoffMatchGroupId"`
	PlayoffRedAlliance  int               `json:"playoffRedAlliance"`
	PlayoffBlueAlliance int               `json:"playoffBlueAlliance"`
	Red                 [3]apiV1MatchTeam `json:"red"`
	Blue                [3]apiV1MatchTeam `json:"blue"`
	StartedAt           *string           `json:"startedAt"`
	ScoreCommittedAt    *string           `json:"scoreCommittedAt"`
	FieldReadyAt        *string           `json:"fieldReadyAt"`
	Status              string            `json:"status"`
	UseTiebreakCriteria bool              `json:"useTiebreakCriteria"`
	TbaMatchKey         apiV1TbaMatchKey  `json:"tbaMatchKey"`
}

type apiV1MatchTeam struct {
	TeamId      int  `json:"teamId"`
	IsSurrogate bool `json:"isSurrogate"`
}

type apiV1TbaMatchKey struct {
	CompLevel   string `json:"compLevel"`
	SetNumber   int    `json:"setNumber"`
	MatchNumber int    `json:"matchNumber"`
}

type apiV1MatchWithResult struct {
	Match  apiV1Match        `json:"match"`
	Result *apiV1MatchResult `json:"result"`
}

type apiV1MatchResult struct {
	Id          int               `json:"id"`
	PlayNumber  int               `json:"playNumber"`
	RedSummary  apiV1ScoreSummary `json:"redSummary"`
	BlueSummary apiV1ScoreSummary `json:"blueSummary"`
	RedCards    map[string]string `json:"redCards"`
	BlueCards   map[string]string `json:"blueCards"`
}

type apiV1ScoreSummary struct {
	AutoFuelPoints                int  `json:"autoFuelPoints"`
	AutoTowerPoints               int  `json:"autoTowerPoints"`
	TeleopFuelPoints              int  `json:"teleopFuelPoints"`
	TeleopTowerPoints             int  `json:"teleopTowerPoints"`
	NumFuel                       int  `json:"numFuel"`
	NumFuelPostMatch              int  `json:"numFuelPostMatch"`
	NumFuelGoal                   int  `json:"numFuelGoal"`
	MatchPoints                   int  `json:"matchPoints"`
	PostMatchPoints               int  `json:"postMatchPoints"`
	FoulPoints                    int  `json:"foulPoints"`
	Score                         int  `json:"score"`
	PlayoffDq                     bool `json:"playoffDq"`
	EnergizedBonusRankingPoint    bool `json:"energizedBonusRankingPoint"`
	SuperchargedBonusRankingPoint bool `json:"superchargedBonusRankingPoint"`
	TraversalBonusRankingPoint    bool `json:"traversalBonusRankingPoint"`
	BonusRankingPoints            int  `json:"bonusRankingPoints"`
	NumOpponentMajorFouls         int  `json:"numOpponentMajorFouls"`
}

type apiV1Ranking struct {
	TeamId            int    `json:"teamId"`
	Nickname          string `json:"nickname"`
	Rank              int    `json:"rank"`
	PreviousRank      int    `json:"previousRank"`
	RankingPoints     int    `json:"rankingPoints"`
	MatchPoints       int    `json:"matchPoints"`
	AutoFuelPoints    int    `json:"autoFuelPoints"`
	TowerPoints       int    `json:"towerPoints"`
	Wins              int    `json:"wins"`
	Losses            int    `json:"losses"`
	Ties              int    `json:"ties"`
	Disqualifications int    `json:"disqualifications"`
	Played            int    `json:"played"`
}

type apiV1Rankings struct {
	Rankings           []apiV1Ranking `json:"rankings"`
	HighestPlayedMatch string         `json:"highestPlayedMatch"`
}

type apiV1Alliance struct {
	Id      int    `json:"id"`
	TeamIds []int  `json:"teamIds"`
	Lineup  [3]int `json:"lineup"`
}

type apiV1SponsorSlide struct {
	Id             int    `json:"id"`
	Subtitle       string `json:"subtitle"`
	Line1          string `json:"line1"`
	Line2          string `json:"line2"`
	Image          string `json:"image"`
	DisplayTimeSec int    `json:"displayTimeSec"`
	DisplayOrder   int    `json:"displayOrder"`
}

type apiV1Team struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Nickname        string `json:"nickname"`
	City            string `json:"city"`
	StateProv       string `json:"stateProv"`
	Country         string `json:"country"`
	SchoolName      string `json:"schoolName"`
	RookieYear      int    `json:"rookieYear"`
	RobotName       string `json:"robotName"`
	Accomplishments string `json:"accomplishments"`
	YellowCard      bool   `json:"yellowCard"`
	HasConnected    bool   `json:"hasConnected"`
}

type apiV1Rule struct {
	Id             int    `json:"id"`
	RuleNumber     string `json:"ruleNumber"`
	IsMajor        bool   `json:"isMajor"`
	IsRankingPoint bool   `json:"isRankingPoint"`
	Description    string `json:"description"`
}

type apiV1Bracket struct {
	BracketType   string                `json:"bracketType"`
	ActiveMatchId *int                  `json:"activeMatchId"`
	Matchups      []apiV1BracketMatchup `json:"matchups"`
}

type apiV1BracketMatchup struct {
	Id                 string         `json:"id"`
	RedAllianceSource  string         `json:"redAllianceSource"`
	BlueAllianceSource string         `json:"blueAllianceSource"`
	RedAlliance        *apiV1Alliance `json:"redAlliance"`
	BlueAlliance       *apiV1Alliance `json:"blueAlliance"`
	IsActive           bool           `json:"isActive"`
	SeriesLeader       string         `json:"seriesLeader"`
	SeriesStatus       string         `json:"seriesStatus"`
	IsComplete         bool           `json:"isComplete"`
}

type apiV1MatchLogListItem struct {
	Id         int    `json:"id"`
	ShortName  string `json:"shortName"`
	Time       string `json:"time"`
	RedTeams   []int  `json:"redTeams"`
	BlueTeams  []int  `json:"blueTeams"`
	Status     string `json:"status"`
	IsComplete bool   `json:"isComplete"`
}

type apiV1MatchLogs struct {
	MatchesByType map[string][]apiV1MatchLogListItem `json:"matchesByType"`
}

type apiV1MatchLog struct {
	Filename  string             `json:"filename"`
	StartTime string             `json:"startTime"`
	TotalRows int                `json:"totalRows"`
	Rows      []apiV1MatchLogRow `json:"rows"`
}

type apiV1MatchLogDetail struct {
	Match           apiV1Match      `json:"match"`
	TeamId          int             `json:"teamId"`
	AllianceStation string          `json:"allianceStation"`
	Offset          int             `json:"offset"`
	Limit           int             `json:"limit"`
	SampleEvery     int             `json:"sampleEvery"`
	Logs            []apiV1MatchLog `json:"logs"`
}

type apiV1MatchLogRow struct {
	MatchTimeSec          float64 `json:"matchTimeSec"`
	PacketType            int     `json:"packetType"`
	TeamId                int     `json:"teamId"`
	AllianceStation       string  `json:"allianceStation"`
	DsLinked              bool    `json:"dsLinked"`
	RadioLinked           bool    `json:"radioLinked"`
	RioLinked             bool    `json:"rioLinked"`
	RobotLinked           bool    `json:"robotLinked"`
	Auto                  bool    `json:"auto"`
	Enabled               bool    `json:"enabled"`
	EmergencyStop         bool    `json:"emergencyStop"`
	AutonomousStop        bool    `json:"autonomousStop"`
	BatteryVoltage        float64 `json:"batteryVoltage"`
	MissedPacketCount     int     `json:"missedPacketCount"`
	DsRobotTripTimeMs     int     `json:"dsRobotTripTimeMs"`
	TxRate                float64 `json:"txRate"`
	RxRate                float64 `json:"rxRate"`
	SignalNoiseRatio      int     `json:"signalNoiseRatio"`
	EthernetConnected     bool    `json:"ethernetConnected"`
	DsReportedStatusValid bool    `json:"dsReportedStatusValid"`
	DsReportedAuto        bool    `json:"dsReportedAuto"`
	DsReportedTeleop      bool    `json:"dsReportedTeleop"`
	DsReportedDisabled    bool    `json:"dsReportedDisabled"`
	DsReportedEnabled     bool    `json:"dsReportedEnabled"`
}

func (web *Web) apiV1MatchesHandler(w http.ResponseWriter, r *http.Request) {
	matchType, err := model.MatchTypeFromString(r.PathValue("type"))
	if err != nil {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_match_type", "Match type is invalid.", map[string]string{"type": err.Error()})
		return
	}
	matches, err := web.arena.Database.GetMatchesByType(matchType, false)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load matches.", nil)
		return
	}
	response := make([]apiV1MatchWithResult, 0, len(matches))
	for i := range matches {
		item := apiV1MatchWithResult{Match: newApiV1Match(&matches[i])}
		result, resultErr := web.arena.Database.GetMatchResultForMatch(matches[i].Id)
		if resultErr != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load match results.", nil)
			return
		}
		if result != nil {
			item.Result = newApiV1MatchResult(result)
		}
		response = append(response, item)
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1RankingsHandler(w http.ResponseWriter, r *http.Request) {
	rankings, err := web.arena.Database.GetAllRankings()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load rankings.", nil)
		return
	}
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load teams.", nil)
		return
	}
	nicknames := make(map[int]string, len(teams))
	for _, team := range teams {
		nicknames[team.Id] = team.Nickname
	}
	response := apiV1Rankings{Rankings: make([]apiV1Ranking, 0, len(rankings))}
	for _, ranking := range rankings {
		response.Rankings = append(response.Rankings, newApiV1Ranking(ranking, nicknames[ranking.TeamId]))
	}
	matches, err := web.arena.Database.GetMatchesByType(model.Qualification, false)
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load matches.", nil)
		return
	}
	for _, match := range matches {
		if match.IsComplete() {
			response.HighestPlayedMatch = match.ShortName
		}
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1AlliancesHandler(w http.ResponseWriter, r *http.Request) {
	alliances, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load alliances.", nil)
		return
	}
	response := make([]apiV1Alliance, 0, len(alliances))
	for i := range alliances {
		response = append(response, newApiV1Alliance(&alliances[i]))
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1SponsorSlidesHandler(w http.ResponseWriter, r *http.Request) {
	slides, err := web.arena.Database.GetAllSponsorSlides()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load sponsor slides.", nil)
		return
	}
	response := make([]apiV1SponsorSlide, 0, len(slides))
	for _, slide := range slides {
		response = append(response, apiV1SponsorSlide{
			Id: slide.Id, Subtitle: slide.Subtitle, Line1: slide.Line1, Line2: slide.Line2, Image: slide.Image,
			DisplayTimeSec: slide.DisplayTimeSec, DisplayOrder: slide.DisplayOrder,
		})
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1TeamsHandler(w http.ResponseWriter, r *http.Request) {
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load teams.", nil)
		return
	}
	response := make([]apiV1Team, 0, len(teams))
	for _, team := range teams {
		response = append(response, apiV1Team{
			Id: team.Id, Name: team.Name, Nickname: team.Nickname, City: team.City, StateProv: team.StateProv,
			Country: team.Country, SchoolName: team.SchoolName, RookieYear: team.RookieYear, RobotName: team.RobotName,
			Accomplishments: team.Accomplishments, YellowCard: team.YellowCard, HasConnected: team.HasConnected,
		})
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1TeamAvatarHandler(w http.ResponseWriter, r *http.Request) {
	teamId, err := strconv.Atoi(r.PathValue("teamId"))
	if err != nil || teamId < 1 {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_team_id", "Team ID must be a positive integer.", nil)
		return
	}
	avatarPath := fmt.Sprintf("%s/%d.png", partner.AvatarsDir, teamId)
	if _, err = os.Stat(avatarPath); os.IsNotExist(err) {
		avatarPath = fmt.Sprintf("%s/0.png", partner.AvatarsDir)
	} else if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "avatar_error", "Unable to load the team avatar.", nil)
		return
	}
	http.ServeFile(w, r, avatarPath)
}

func (web *Web) apiV1RulesHandler(w http.ResponseWriter, r *http.Request) {
	rules := game.GetAllRules()
	ids := make([]int, 0, len(rules))
	for id := range rules {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	response := make([]apiV1Rule, 0, len(ids))
	for _, id := range ids {
		rule := rules[id]
		response = append(response, apiV1Rule{
			Id: rule.Id, RuleNumber: rule.RuleNumber, IsMajor: rule.IsMajor,
			IsRankingPoint: rule.IsRankingPoint, Description: rule.Description,
		})
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1BracketHandler(w http.ResponseWriter, r *http.Request) {
	alliances, err := web.arena.Database.GetAllAlliances()
	if err != nil {
		writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load alliances.", nil)
		return
	}
	allianceById := make(map[int]apiV1Alliance, len(alliances))
	for i := range alliances {
		allianceById[alliances[i].Id] = newApiV1Alliance(&alliances[i])
	}
	activeMatch := apiV1ActiveMatch(web, r.URL.Query().Get("activeMatch"))
	response := apiV1Bracket{BracketType: apiV1BracketType(web.arena.EventSettings), Matchups: []apiV1BracketMatchup{}}
	if activeMatch != nil {
		response.ActiveMatchId = &activeMatch.Id
	}
	if web.arena.PlayoffTournament != nil {
		for _, matchGroup := range web.arena.PlayoffTournament.MatchGroups() {
			matchup, ok := matchGroup.(*playoff.Matchup)
			if !ok {
				continue
			}
			leader, status := matchup.StatusText()
			item := apiV1BracketMatchup{
				Id: matchup.Id(), RedAllianceSource: matchup.RedAllianceSourceDisplayName(),
				BlueAllianceSource: matchup.BlueAllianceSourceDisplayName(), IsComplete: matchup.IsComplete(),
				SeriesLeader: leader, SeriesStatus: status,
			}
			if alliance, ok := allianceById[matchup.RedAllianceId]; ok {
				item.RedAlliance = &alliance
			}
			if alliance, ok := allianceById[matchup.BlueAllianceId]; ok {
				item.BlueAlliance = &alliance
			}
			if activeMatch != nil {
				item.IsActive = activeMatch.PlayoffMatchGroupId == matchup.Id()
			}
			response.Matchups = append(response.Matchups, item)
		}
		sort.Slice(response.Matchups, func(i, j int) bool { return response.Matchups[i].Id < response.Matchups[j].Id })
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1MatchLogsHandler(w http.ResponseWriter, r *http.Request) {
	response := apiV1MatchLogs{MatchesByType: make(map[string][]apiV1MatchLogListItem)}
	for _, matchType := range []model.MatchType{model.Practice, model.Qualification, model.Playoff} {
		matches, err := web.buildMatchLogsList(matchType)
		if err != nil {
			writeApiV1Error(w, r, http.StatusInternalServerError, "database_error", "Unable to load match logs.", nil)
			return
		}
		items := make([]apiV1MatchLogListItem, 0, len(matches))
		for _, match := range matches {
			items = append(items, apiV1MatchLogListItem{
				Id: match.Id, ShortName: match.ShortName, Time: match.Time, RedTeams: match.RedTeams,
				BlueTeams: match.BlueTeams, Status: apiV1ColorStatus(match.ColorClass), IsComplete: match.IsComplete,
			})
		}
		response.MatchesByType[apiV1MatchType(matchType)] = items
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func (web *Web) apiV1MatchLogHandler(w http.ResponseWriter, r *http.Request) {
	match, logs, _, err := web.getMatchLogFromRequest(r)
	if err != nil {
		writeApiV1Error(w, r, http.StatusNotFound, "match_log_not_found", "Unable to load the requested match log.", nil)
		return
	}
	if match == nil || logs == nil {
		writeApiV1Error(w, r, http.StatusNotFound, "match_log_not_found", "The requested match log does not exist.", nil)
		return
	}
	offset, limit, sampleEvery, parameterErr := apiV1LogParameters(r)
	if parameterErr != "" {
		writeApiV1Error(w, r, http.StatusBadRequest, "invalid_query_parameter", "Match log query parameters are invalid.", map[string]string{"query": parameterErr})
		return
	}
	response := apiV1MatchLogDetail{
		Match: newApiV1Match(match), TeamId: logs.TeamId, AllianceStation: logs.AllianceStation,
		Offset: offset, Limit: limit, SampleEvery: sampleEvery, Logs: make([]apiV1MatchLog, 0, len(logs.Logs)),
	}
	for _, logData := range logs.Logs {
		logResponse := apiV1MatchLog{Filename: logData.Filename, StartTime: logData.StartTime, TotalRows: len(logData.Rows), Rows: []apiV1MatchLogRow{}}
		selected := 0
		for index := offset; index < len(logData.Rows) && selected < limit; index += sampleEvery {
			logResponse.Rows = append(logResponse.Rows, newApiV1MatchLogRow(logData.Rows[index]))
			selected++
		}
		response.Logs = append(response.Logs, logResponse)
	}
	writeApiV1Data(w, r, http.StatusOK, response, nil)
}

func newApiV1Match(match *model.Match) apiV1Match {
	return apiV1Match{
		Id: match.Id, Type: apiV1MatchType(match.Type), TypeOrder: match.TypeOrder, Time: match.Time.Format(time.RFC3339),
		LongName: match.LongName, ShortName: match.ShortName, NameDetail: match.NameDetail,
		PlayoffMatchGroupId: match.PlayoffMatchGroupId, PlayoffRedAlliance: match.PlayoffRedAlliance,
		PlayoffBlueAlliance: match.PlayoffBlueAlliance,
		Red:                 [3]apiV1MatchTeam{{match.Red1, match.Red1IsSurrogate}, {match.Red2, match.Red2IsSurrogate}, {match.Red3, match.Red3IsSurrogate}},
		Blue:                [3]apiV1MatchTeam{{match.Blue1, match.Blue1IsSurrogate}, {match.Blue2, match.Blue2IsSurrogate}, {match.Blue3, match.Blue3IsSurrogate}},
		StartedAt:           apiV1OptionalTime(match.StartedAt), ScoreCommittedAt: apiV1OptionalTime(match.ScoreCommittedAt),
		FieldReadyAt: apiV1OptionalTime(match.FieldReadyAt), Status: apiV1MatchStatus(match.Status),
		UseTiebreakCriteria: match.UseTiebreakCriteria,
		TbaMatchKey:         apiV1TbaMatchKey{match.TbaMatchKey.CompLevel, match.TbaMatchKey.SetNumber, match.TbaMatchKey.MatchNumber},
	}
}

func newApiV1MatchResult(result *model.MatchResult) *apiV1MatchResult {
	redCards := result.RedCards
	if redCards == nil {
		redCards = map[string]string{}
	}
	blueCards := result.BlueCards
	if blueCards == nil {
		blueCards = map[string]string{}
	}
	return &apiV1MatchResult{
		Id: result.Id, PlayNumber: result.PlayNumber, RedSummary: newApiV1ScoreSummary(result.RedScoreSummary()),
		BlueSummary: newApiV1ScoreSummary(result.BlueScoreSummary()), RedCards: redCards, BlueCards: blueCards,
	}
}

func newApiV1ScoreSummary(summary *game.ScoreSummary) apiV1ScoreSummary {
	return apiV1ScoreSummary{
		AutoFuelPoints: summary.AutoFuelPoints, AutoTowerPoints: summary.AutoTowerPoints,
		TeleopFuelPoints: summary.TeleopFuelPoints, TeleopTowerPoints: summary.TeleopTowerPoints,
		NumFuel: summary.NumFuel, NumFuelPostMatch: summary.NumFuelPostMatch, NumFuelGoal: summary.NumFuelGoal,
		MatchPoints: summary.MatchPoints, PostMatchPoints: summary.PostMatchPoints, FoulPoints: summary.FoulPoints,
		Score: summary.Score, PlayoffDq: summary.PlayoffDq, EnergizedBonusRankingPoint: summary.EnergizedBonusRankingPoint,
		SuperchargedBonusRankingPoint: summary.SuperchargedBonusRankingPoint,
		TraversalBonusRankingPoint:    summary.TraversalBonusRankingPoint, BonusRankingPoints: summary.BonusRankingPoints,
		NumOpponentMajorFouls: summary.NumOpponentMajorFouls,
	}
}

func newApiV1Ranking(ranking game.Ranking, nickname string) apiV1Ranking {
	return apiV1Ranking{
		TeamId: ranking.TeamId, Nickname: nickname, Rank: ranking.Rank, PreviousRank: ranking.PreviousRank,
		RankingPoints: ranking.RankingPoints, MatchPoints: ranking.MatchPoints, AutoFuelPoints: ranking.AutoFuelPoints,
		TowerPoints: ranking.TowerPoints, Wins: ranking.Wins, Losses: ranking.Losses, Ties: ranking.Ties,
		Disqualifications: ranking.Disqualifications, Played: ranking.Played,
	}
}

func newApiV1Alliance(alliance *model.Alliance) apiV1Alliance {
	teamIds := append([]int{}, alliance.TeamIds...)
	return apiV1Alliance{Id: alliance.Id, TeamIds: teamIds, Lineup: alliance.Lineup}
}

func newApiV1MatchLogRow(row MatchLogRow) apiV1MatchLogRow {
	return apiV1MatchLogRow{
		MatchTimeSec: row.MatchTimeSec, PacketType: row.PacketType, TeamId: row.TeamId,
		AllianceStation: row.AllianceStation, DsLinked: row.DsLinked, RadioLinked: row.RadioLinked,
		RioLinked: row.RioLinked, RobotLinked: row.RobotLinked, Auto: row.Auto, Enabled: row.Enabled,
		EmergencyStop: row.EmergencyStop, AutonomousStop: row.AutonomousStop, BatteryVoltage: row.BatteryVoltage,
		MissedPacketCount: row.MissedPacketCount, DsRobotTripTimeMs: row.DsRobotTripTimeMs,
		TxRate: row.TxRate, RxRate: row.RxRate, SignalNoiseRatio: row.SignalNoiseRatio,
		EthernetConnected: row.EthernetConnected, DsReportedStatusValid: row.DsReportedStatusValid,
		DsReportedAuto: row.DsReportedAuto, DsReportedTeleop: row.DsReportedTeleop,
		DsReportedDisabled: row.DsReportedDisabled, DsReportedEnabled: row.DsReportedEnabled,
	}
}

func apiV1OptionalTime(value time.Time) *string {
	if value.IsZero() {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}

func apiV1MatchType(matchType model.MatchType) string {
	switch matchType {
	case model.Practice:
		return "practice"
	case model.Qualification:
		return "qualification"
	case model.Playoff:
		return "playoff"
	default:
		return "test"
	}
}

func apiV1MatchStatus(status game.MatchStatus) string {
	switch status {
	case game.MatchHidden:
		return "hidden"
	case game.RedWonMatch:
		return "red_won"
	case game.BlueWonMatch:
		return "blue_won"
	case game.TieMatch:
		return "tie"
	default:
		return "scheduled"
	}
}

func apiV1ColorStatus(color string) string {
	switch color {
	case "red":
		return "red_won"
	case "blue":
		return "blue_won"
	case "yellow":
		return "tie"
	default:
		return "scheduled"
	}
}

func apiV1ActiveMatch(web *Web, active string) *model.Match {
	if active == "current" {
		return web.arena.CurrentMatch
	}
	if active == "saved" {
		return web.arena.SavedMatch
	}
	return nil
}

func apiV1BracketType(settings *model.EventSettings) string {
	if settings.PlayoffType == model.DoubleEliminationPlayoff {
		if settings.NumPlayoffAlliances == 4 {
			return "double4"
		}
		return "double"
	}
	if settings.NumPlayoffAlliances > 8 {
		return "16"
	}
	if settings.NumPlayoffAlliances > 4 {
		return "8"
	}
	if settings.NumPlayoffAlliances > 2 {
		return "4"
	}
	return "2"
}

func apiV1LogParameters(r *http.Request) (int, int, int, string) {
	parse := func(name string, defaultValue int) (int, string) {
		value := r.URL.Query().Get(name)
		if value == "" {
			return defaultValue, ""
		}
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, name + " must be an integer"
		}
		return parsed, ""
	}
	offset, message := parse("offset", 0)
	if message != "" || offset < 0 {
		if message == "" {
			message = "offset must not be negative"
		}
		return 0, 0, 0, message
	}
	limit, message := parse("limit", 1000)
	if message != "" || limit < 1 || limit > 10000 {
		if message == "" {
			message = "limit must be between 1 and 10000"
		}
		return 0, 0, 0, message
	}
	sampleEvery, message := parse("sampleEvery", 1)
	if message != "" || sampleEvery < 1 || sampleEvery > 1000 {
		if message == "" {
			message = "sampleEvery must be between 1 and 1000"
		}
		return 0, 0, 0, message
	}
	return offset, limit, sampleEvery, ""
}
