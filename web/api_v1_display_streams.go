// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/playoff"
	"github.com/Team254/cheesy-arena/websocket"
)

type apiV1MatchClock struct {
	State      string `json:"state"`
	ElapsedSec int    `json:"elapsedSec"`
}

type apiV1DisplayEventStatus struct {
	CycleTime        string `json:"cycleTime"`
	EarlyLateMessage string `json:"earlyLateMessage"`
}

type apiV1DisplayMatchTiming struct {
	AutoDurationSec            int `json:"autoDurationSec"`
	PauseDurationSec           int `json:"pauseDurationSec"`
	TransitionShiftDurationSec int `json:"transitionShiftDurationSec"`
	ShiftDurationSec           int `json:"shiftDurationSec"`
	EndgameDurationSec         int `json:"endgameDurationSec"`
	TimeoutDurationSec         int `json:"timeoutDurationSec"`
}

type apiV1DisplayRealtimeScore struct {
	Red  int `json:"red"`
	Blue int `json:"blue"`
}

type apiV1DisplayTeam struct {
	Id         int    `json:"id"`
	Nickname   string `json:"nickname"`
	YellowCard bool   `json:"yellowCard"`
}

type apiV1DisplaySeries struct {
	NumWinsToAdvance int `json:"numWinsToAdvance"`
	RedAllianceWins  int `json:"redAllianceWins"`
	BlueAllianceWins int `json:"blueAllianceWins"`
}

type apiV1DisplayMatch struct {
	Id                  int                          `json:"id"`
	Type                string                       `json:"type"`
	LongName            string                       `json:"longName"`
	NameDetail          string                       `json:"nameDetail"`
	PlayoffRedAlliance  int                          `json:"playoffRedAlliance"`
	PlayoffBlueAlliance int                          `json:"playoffBlueAlliance"`
	Red                 [3]int                       `json:"red"`
	Blue                [3]int                       `json:"blue"`
	Teams               map[string]*apiV1DisplayTeam `json:"teams"`
	Rankings            map[string]int               `json:"rankings"`
	RedOffFieldTeams    []apiV1DisplayTeam           `json:"redOffFieldTeams"`
	BlueOffFieldTeams   []apiV1DisplayTeam           `json:"blueOffFieldTeams"`
	Series              *apiV1DisplaySeries          `json:"series"`
	BreakDescription    string                       `json:"breakDescription"`
	BreakNextMatchName  string                       `json:"breakNextMatchName"`
}

type apiV1AudienceRealtimeAlliance struct {
	Summary            apiV1ScoreSummary `json:"summary"`
	ActiveRemainingSec int               `json:"activeRemainingSec"`
	ActiveDurationSec  int               `json:"activeDurationSec"`
}

type apiV1AudienceRealtimeScore struct {
	Red  apiV1AudienceRealtimeAlliance `json:"red"`
	Blue apiV1AudienceRealtimeAlliance `json:"blue"`
}

type apiV1AudiencePostedAlliance struct {
	Summary         apiV1ScoreSummary                 `json:"summary"`
	RankingPoints   int                               `json:"rankingPoints"`
	Cards           map[string]string                 `json:"cards"`
	Rankings        map[string]*apiV1AnnouncerRanking `json:"rankings"`
	OffFieldTeamIds []int                             `json:"offFieldTeamIds"`
	Won             bool                              `json:"won"`
	Wins            int                               `json:"wins"`
	Destination     string                            `json:"destination"`
}

type apiV1AudiencePostedScore struct {
	Match          apiV1DisplayMatch           `json:"match"`
	Red            apiV1AudiencePostedAlliance `json:"red"`
	Blue           apiV1AudiencePostedAlliance `json:"blue"`
	TiebreakReason string                      `json:"tiebreakReason"`
}

type apiV1AllianceSelectionTeam struct {
	Rank   int  `json:"rank"`
	TeamId int  `json:"teamId"`
	Picked bool `json:"picked"`
}

type apiV1AllianceSelectionAlliance struct {
	Id      int   `json:"id"`
	TeamIds []int `json:"teamIds"`
}

type apiV1AudienceAllianceSelection struct {
	Alliances        []apiV1AllianceSelectionAlliance `json:"alliances"`
	ShowTimer        bool                             `json:"showTimer"`
	TimeRemainingSec int                              `json:"timeRemainingSec"`
	RankedTeams      []apiV1AllianceSelectionTeam     `json:"rankedTeams"`
}

type apiV1AudienceLowerThird struct {
	TopText        string `json:"topText"`
	BottomText     string `json:"bottomText"`
	ShowLowerThird bool   `json:"showLowerThird"`
}

type apiV1AllianceStationStatus struct {
	Bypass      bool `json:"bypass"`
	DsLinked    bool `json:"dsLinked"`
	RobotLinked bool `json:"robotLinked"`
}

type apiV1AudienceBootstrap struct {
	StreamUrl         string                         `json:"streamUrl"`
	Match             apiV1DisplayMatch              `json:"match"`
	PostedScore       *apiV1AudiencePostedScore      `json:"postedScore"`
	RealtimeScore     apiV1AudienceRealtimeScore     `json:"realtimeScore"`
	MatchClock        apiV1MatchClock                `json:"matchClock"`
	Timing            apiV1DisplayMatchTiming        `json:"timing"`
	DisplayMode       string                         `json:"displayMode"`
	AllianceSelection apiV1AudienceAllianceSelection `json:"allianceSelection"`
	LowerThird        apiV1AudienceLowerThird        `json:"lowerThird"`
}

type apiV1AllianceStationBootstrap struct {
	StreamUrl     string                                `json:"streamUrl"`
	Match         apiV1DisplayMatch                     `json:"match"`
	RealtimeScore apiV1DisplayRealtimeScore             `json:"realtimeScore"`
	MatchClock    apiV1MatchClock                       `json:"matchClock"`
	Timing        apiV1DisplayMatchTiming               `json:"timing"`
	DisplayMode   string                                `json:"displayMode"`
	Stations      map[string]apiV1AllianceStationStatus `json:"stations"`
}

type apiV1WallBootstrap struct {
	StreamUrl     string                     `json:"streamUrl"`
	Match         apiV1DisplayMatch          `json:"match"`
	RealtimeScore apiV1AudienceRealtimeScore `json:"realtimeScore"`
	MatchClock    apiV1MatchClock            `json:"matchClock"`
	Timing        apiV1DisplayMatchTiming    `json:"timing"`
	DisplayMode   string                     `json:"displayMode"`
}

type apiV1UnpickedBootstrap struct {
	StreamUrl         string                         `json:"streamUrl"`
	DisplayMode       string                         `json:"displayMode"`
	AllianceSelection apiV1AudienceAllianceSelection `json:"allianceSelection"`
}

type apiV1AllianceSelectionControlBootstrap struct {
	StreamUrl         string                         `json:"streamUrl"`
	DisplayMode       string                         `json:"displayMode"`
	AllianceSelection apiV1AudienceAllianceSelection `json:"allianceSelection"`
}

type apiV1MatchPlayMatch struct {
	apiV1DisplayMatch
	AllowSubstitution bool `json:"allowSubstitution"`
	IsReplay          bool `json:"isReplay"`
}

type apiV1MatchPlayArenaStatus struct {
	State                 string                              `json:"state"`
	CanStartMatch         bool                                `json:"canStartMatch"`
	StartMatchConditions  []string                            `json:"startMatchConditions"`
	Stations              map[string]apiV1FieldMonitorStation `json:"stations"`
	AccessPointStatus     string                              `json:"accessPointStatus"`
	SwitchStatus          string                              `json:"switchStatus"`
	RedSccStatus          string                              `json:"redSccStatus"`
	BlueSccStatus         string                              `json:"blueSccStatus"`
	PlcIsHealthy          bool                                `json:"plcIsHealthy"`
	FieldEStop            bool                                `json:"fieldEStop"`
	IsFtaReady            bool                                `json:"isFtaReady"`
	PlcArmorBlockStatuses map[string]bool                     `json:"plcArmorBlockStatuses"`
}

type apiV1ScoringPositionStatus struct {
	Ready          bool `json:"ready"`
	NumPanels      int  `json:"numPanels"`
	NumPanelsReady int  `json:"numPanelsReady"`
}
type apiV1ScoringStatus struct {
	RefereeScoreReady bool                                  `json:"refereeScoreReady"`
	Positions         map[string]apiV1ScoringPositionStatus `json:"positions"`
}

type apiV1MatchPlayBootstrap struct {
	StreamUrl                  string                     `json:"streamUrl"`
	Match                      apiV1MatchPlayMatch        `json:"match"`
	ArenaStatus                apiV1MatchPlayArenaStatus  `json:"arenaStatus"`
	RealtimeScore              apiV1AudienceRealtimeScore `json:"realtimeScore"`
	PostedScore                *apiV1AudiencePostedScore  `json:"postedScore"`
	ScoringStatus              apiV1ScoringStatus         `json:"scoringStatus"`
	MatchClock                 apiV1MatchClock            `json:"matchClock"`
	Timing                     apiV1DisplayMatchTiming    `json:"timing"`
	Event                      apiV1DisplayEventStatus    `json:"event"`
	AudienceDisplayMode        string                     `json:"audienceDisplayMode"`
	AllianceStationDisplayMode string                     `json:"allianceStationDisplayMode"`
}

type apiV1ControlFoul struct {
	FoulId  int  `json:"foulId"`
	IsMajor bool `json:"isMajor"`
	TeamId  int  `json:"teamId"`
	RuleId  int  `json:"ruleId"`
}
type apiV1ControlScoreAlliance struct {
	AutoTowerStatuses    [3]int             `json:"autoTowerStatuses"`
	EndgameTowerStatuses [3]int             `json:"endgameTowerStatuses"`
	Fouls                []apiV1ControlFoul `json:"fouls"`
	Cards                map[string]string  `json:"cards"`
}
type apiV1ControlRealtimeScore struct {
	Red  apiV1ControlScoreAlliance `json:"red"`
	Blue apiV1ControlScoreAlliance `json:"blue"`
}
type apiV1ScoringPanelBootstrap struct {
	StreamUrl     string                    `json:"streamUrl"`
	Match         apiV1DisplayMatch         `json:"match"`
	MatchClock    apiV1MatchClock           `json:"matchClock"`
	RealtimeScore apiV1ControlRealtimeScore `json:"realtimeScore"`
}
type apiV1RefereePanelBootstrap struct {
	StreamUrl     string                                `json:"streamUrl"`
	Match         apiV1DisplayMatch                     `json:"match"`
	MatchClock    apiV1MatchClock                       `json:"matchClock"`
	RealtimeScore apiV1ControlRealtimeScore             `json:"realtimeScore"`
	ScoringStatus apiV1ScoringStatus                    `json:"scoringStatus"`
	Stations      map[string]apiV1AllianceStationStatus `json:"stations"`
}

type apiV1FieldMonitorTeam struct {
	Id       int     `json:"id"`
	FtaNotes *string `json:"ftaNotes,omitempty"`
}

type apiV1FieldMonitorConnection struct {
	WrongStation              string  `json:"wrongStation"`
	DsLinked                  bool    `json:"dsLinked"`
	RadioLinked               bool    `json:"radioLinked"`
	RioLinked                 bool    `json:"rioLinked"`
	RobotLinked               bool    `json:"robotLinked"`
	BatteryVoltage            float64 `json:"batteryVoltage"`
	DsRobotTripTimeMs         int     `json:"dsRobotTripTimeMs"`
	MissedPacketCount         int     `json:"missedPacketCount"`
	SecondsSinceLastRobotLink float64 `json:"secondsSinceLastRobotLink"`
}

type apiV1FieldMonitorWifi struct {
	TeamId            int     `json:"teamId"`
	RadioLinked       bool    `json:"radioLinked"`
	MBits             float64 `json:"mBits"`
	ConnectionQuality int     `json:"connectionQuality"`
}

type apiV1FieldMonitorStation struct {
	Team       *apiV1FieldMonitorTeam       `json:"team"`
	Connection *apiV1FieldMonitorConnection `json:"connection"`
	Wifi       apiV1FieldMonitorWifi        `json:"wifi"`
	Ethernet   bool                         `json:"ethernet"`
	AStop      bool                         `json:"aStop"`
	EStop      bool                         `json:"eStop"`
	Bypass     bool                         `json:"bypass"`
}

type apiV1FieldMonitorStatus struct {
	MatchId           int                                 `json:"matchId"`
	AccessPointStatus string                              `json:"accessPointStatus"`
	SwitchStatus      string                              `json:"switchStatus"`
	Stations          map[string]apiV1FieldMonitorStation `json:"stations"`
}

type apiV1FieldMonitorBootstrap struct {
	StreamUrl     string                     `json:"streamUrl"`
	ArenaStatus   apiV1FieldMonitorStatus    `json:"arenaStatus"`
	Event         apiV1DisplayEventStatus    `json:"event"`
	Match         apiV1DisplayMatch          `json:"match"`
	RealtimeScore apiV1AudienceRealtimeScore `json:"realtimeScore"`
	MatchClock    apiV1MatchClock            `json:"matchClock"`
	Timing        apiV1DisplayMatchTiming    `json:"timing"`
}

type apiV1QueueingBootstrap struct {
	StreamUrl  string                  `json:"streamUrl"`
	Matches    []apiV1QueueingMatch    `json:"matches"`
	MatchClock apiV1MatchClock         `json:"matchClock"`
	Timing     apiV1DisplayMatchTiming `json:"timing"`
	Event      apiV1DisplayEventStatus `json:"event"`
}

type apiV1AnnouncerBootstrap struct {
	StreamUrl           string                    `json:"streamUrl"`
	Match               apiV1AnnouncerMatch       `json:"match"`
	PostedScore         *apiV1AnnouncerScore      `json:"postedScore"`
	RealtimeScore       apiV1DisplayRealtimeScore `json:"realtimeScore"`
	MatchClock          apiV1MatchClock           `json:"matchClock"`
	Timing              apiV1DisplayMatchTiming   `json:"timing"`
	Event               apiV1DisplayEventStatus   `json:"event"`
	AudienceDisplayMode string                    `json:"audienceDisplayMode"`
}

type apiV1DisplayState struct {
	mu              sync.RWMutex
	queueing        apiV1QueueingBootstrap
	announcer       apiV1AnnouncerBootstrap
	audience        apiV1AudienceBootstrap
	allianceStation apiV1AllianceStationBootstrap
	wall            apiV1WallBootstrap
	unpicked        apiV1UnpickedBootstrap
	fieldMonitor    apiV1FieldMonitorBootstrap
	fieldMonitorFta apiV1FieldMonitorBootstrap
	matchPlay       apiV1MatchPlayBootstrap
	scoringPanel    apiV1ScoringPanelBootstrap
	refereePanel    apiV1RefereePanelBootstrap

	queueingMatches       *websocket.Notifier
	announcerMatch        *websocket.Notifier
	postedScore           *websocket.Notifier
	realtimeScore         *websocket.Notifier
	matchClock            *websocket.Notifier
	timing                *websocket.Notifier
	eventStatus           *websocket.Notifier
	audienceMode          *websocket.Notifier
	reload                *websocket.Notifier
	displayMatch          *websocket.Notifier
	audienceRealtime      *websocket.Notifier
	audiencePosted        *websocket.Notifier
	allianceSelection     *websocket.Notifier
	lowerThird            *websocket.Notifier
	playSound             *websocket.Notifier
	allianceStationMode   *websocket.Notifier
	stationStatuses       *websocket.Notifier
	fieldMonitorStatus    *websocket.Notifier
	fieldMonitorFtaStatus *websocket.Notifier
	matchPlayMatch        *websocket.Notifier
	matchPlayArenaStatus  *websocket.Notifier
	scoringStatus         *websocket.Notifier
	controlRealtimeScore  *websocket.Notifier
}

func (web *Web) initializeApiV1DisplayState() {
	state := &apiV1DisplayState{}
	web.apiV1Displays = state
	state.queueing.StreamUrl = "/api/v1/streams/displays/queueing"
	state.announcer.StreamUrl = "/api/v1/streams/displays/announcer"
	state.audience.StreamUrl = "/api/v1/streams/displays/audience"
	state.allianceStation.StreamUrl = "/api/v1/streams/displays/alliance-station"
	state.wall.StreamUrl = "/api/v1/streams/displays/wall"
	state.unpicked.StreamUrl = "/api/v1/streams/displays/unpicked"
	state.fieldMonitor.StreamUrl = "/api/v1/streams/displays/field-monitor"
	state.fieldMonitorFta.StreamUrl = "/api/v1/streams/displays/field-monitor?fta=true"
	state.matchPlay.StreamUrl = "/api/v1/streams/admin/match-play"
	state.scoringPanel.StreamUrl = "/api/v1/streams/admin/scoring"
	state.refereePanel.StreamUrl = "/api/v1/streams/admin/referee"
	web.refreshApiV1DisplayMatches()
	web.refreshApiV1PostedScore()
	web.refreshApiV1RealtimeScore()
	web.refreshApiV1MatchClock()
	web.refreshApiV1DisplayTiming()
	web.refreshApiV1EventStatus()
	web.refreshApiV1AudienceMode()
	web.refreshApiV1AudienceState()
	web.refreshApiV1AllianceStationState()
	web.refreshApiV1FieldMonitorStatus()
	web.refreshApiV1MatchPlayState()
	web.refreshApiV1ControlRealtimeScore()

	state.queueingMatches = websocket.NewNotifier("matches", func() any { return web.apiV1QueueingMatchesSnapshot() })
	state.announcerMatch = websocket.NewNotifier("match", func() any { return web.apiV1AnnouncerMatchSnapshot() })
	state.postedScore = websocket.NewNotifier("postedScore", func() any { return web.apiV1PostedScoreSnapshot() })
	state.realtimeScore = websocket.NewNotifier("realtimeScore", func() any { return web.apiV1RealtimeScoreSnapshot() })
	state.matchClock = websocket.NewNotifier("matchClock", func() any { return web.apiV1MatchClockSnapshot() })
	state.timing = websocket.NewNotifier("timing", func() any { return web.apiV1TimingSnapshot() })
	state.eventStatus = websocket.NewNotifier("eventStatus", func() any { return web.apiV1EventStatusSnapshot() })
	state.audienceMode = websocket.NewNotifier("audienceDisplayMode", func() any { return web.apiV1AudienceModeSnapshot() })
	state.reload = websocket.NewNotifier("reload", nil)
	state.displayMatch = websocket.NewNotifier("match", func() any { return web.apiV1DisplayMatchSnapshot() })
	state.audienceRealtime = websocket.NewNotifier("realtimeScore", func() any { return web.apiV1AudienceRealtimeSnapshot() })
	state.audiencePosted = websocket.NewNotifier("postedScore", func() any { return web.apiV1AudiencePostedSnapshot() })
	state.allianceSelection = websocket.NewNotifier("allianceSelection", func() any { return web.apiV1AllianceSelectionSnapshot() })
	state.lowerThird = websocket.NewNotifier("lowerThird", func() any { return web.apiV1LowerThirdSnapshot() })
	state.playSound = websocket.NewNotifier("playSound", nil)
	state.allianceStationMode = websocket.NewNotifier("allianceStationDisplayMode", func() any { return web.apiV1AllianceStationModeSnapshot() })
	state.stationStatuses = websocket.NewNotifier("stationStatuses", func() any { return web.apiV1StationStatusesSnapshot() })
	state.fieldMonitorStatus = websocket.NewNotifier("arenaStatus", func() any { return web.apiV1FieldMonitorStatusSnapshot(false) })
	state.fieldMonitorFtaStatus = websocket.NewNotifier("arenaStatus", func() any { return web.apiV1FieldMonitorStatusSnapshot(true) })
	state.matchPlayMatch = websocket.NewNotifier("match", func() any { return web.apiV1MatchPlayMatchSnapshot() })
	state.matchPlayArenaStatus = websocket.NewNotifier("arenaStatus", func() any { return web.apiV1MatchPlayArenaStatusSnapshot() })
	state.scoringStatus = websocket.NewNotifier("scoringStatus", func() any { return web.apiV1ScoringStatusSnapshot() })
	state.controlRealtimeScore = websocket.NewNotifier("realtimeScore", func() any { return web.apiV1ControlRealtimeScoreSnapshot() })

	web.arena.MatchLoadNotifier.Observe(func(any) {
		web.refreshApiV1DisplayMatches()
		web.refreshApiV1DisplayMatch()
		state.queueingMatches.Notify()
		state.announcerMatch.Notify()
		state.displayMatch.Notify()
		web.refreshApiV1MatchPlayMatch()
		state.matchPlayMatch.Notify()
	})
	web.arena.ScorePostedNotifier.Observe(func(any) {
		web.refreshApiV1PostedScore()
		web.refreshApiV1AudiencePostedScore()
		state.postedScore.Notify()
		state.audiencePosted.Notify()
	})
	web.arena.RealtimeScoreNotifier.Observe(func(any) {
		web.refreshApiV1RealtimeScore()
		web.refreshApiV1AudienceRealtimeScore()
		state.realtimeScore.Notify()
		state.audienceRealtime.Notify()
		web.refreshApiV1ControlRealtimeScore()
		state.controlRealtimeScore.Notify()
	})
	web.arena.MatchTimeNotifier.Observe(func(any) { web.refreshApiV1MatchClock(); state.matchClock.Notify() })
	web.arena.MatchTimingNotifier.Observe(func(any) { web.refreshApiV1DisplayTiming(); state.timing.Notify() })
	web.arena.EventStatusNotifier.Observe(func(any) { web.refreshApiV1EventStatus(); state.eventStatus.Notify() })
	web.arena.AudienceDisplayModeNotifier.Observe(func(any) { web.refreshApiV1AudienceMode(); state.audienceMode.Notify() })
	web.arena.AllianceSelectionNotifier.Observe(func(any) { web.refreshApiV1AllianceSelection(); state.allianceSelection.Notify() })
	web.arena.LowerThirdNotifier.Observe(func(any) { web.refreshApiV1LowerThird(); state.lowerThird.Notify() })
	web.arena.PlaySoundNotifier.Observe(func(value any) { state.playSound.NotifyWithMessage(value) })
	web.arena.AllianceStationDisplayModeNotifier.Observe(func(any) { web.refreshApiV1AllianceStationMode(); state.allianceStationMode.Notify() })
	web.arena.ArenaStatusNotifier.Observe(func(any) {
		web.refreshApiV1StationStatuses()
		web.refreshApiV1FieldMonitorStatus()
		state.stationStatuses.Notify()
		state.fieldMonitorStatus.Notify()
		state.fieldMonitorFtaStatus.Notify()
		web.refreshApiV1MatchPlayArenaStatus()
		state.matchPlayArenaStatus.Notify()
	})
	web.arena.ScoringStatusNotifier.Observe(func(any) { web.refreshApiV1ScoringStatus(); state.scoringStatus.Notify() })
	web.arena.ReloadDisplaysNotifier.Observe(func(value any) { state.reload.NotifyWithMessage(value) })
}

func (web *Web) refreshApiV1DisplayMatches() {
	queueing, err := web.buildApiV1QueueingMatches()
	if err != nil {
		log.Printf("Unable to refresh API v1 queueing snapshot: %v", err)
		queueing = []apiV1QueueingMatch{}
	}
	announcer, err := web.buildApiV1AnnouncerMatch()
	if err != nil {
		log.Printf("Unable to refresh API v1 announcer snapshot: %v", err)
	}
	web.apiV1Displays.mu.Lock()
	defer web.apiV1Displays.mu.Unlock()
	web.apiV1Displays.queueing.Matches = queueing
	web.apiV1Displays.announcer.Match = announcer
}

func (web *Web) refreshApiV1PostedScore() {
	score, _ := web.buildApiV1AnnouncerScore()
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.announcer.PostedScore = score
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1RealtimeScore() {
	red, blue := web.arena.RedScoreSummary(), web.arena.BlueScoreSummary()
	value := apiV1DisplayRealtimeScore{Red: red.Score - red.PostMatchPoints, Blue: blue.Score - blue.PostMatchPoints}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.announcer.RealtimeScore = value
	web.apiV1Displays.allianceStation.RealtimeScore = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1MatchClock() {
	value := apiV1MatchClock{State: apiV1MatchState(web.arena.MatchState), ElapsedSec: int(web.arena.MatchTimeSec())}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.MatchClock = value
	web.apiV1Displays.announcer.MatchClock = value
	web.apiV1Displays.audience.MatchClock = value
	web.apiV1Displays.allianceStation.MatchClock = value
	web.apiV1Displays.wall.MatchClock = value
	web.apiV1Displays.fieldMonitor.MatchClock = value
	web.apiV1Displays.fieldMonitorFta.MatchClock = value
	web.apiV1Displays.matchPlay.MatchClock = value
	web.apiV1Displays.scoringPanel.MatchClock = value
	web.apiV1Displays.refereePanel.MatchClock = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1DisplayTiming() {
	value := apiV1DisplayMatchTiming{game.MatchTiming.AutoDurationSec, game.MatchTiming.PauseDurationSec, game.MatchTiming.TransitionShiftDurationSec, game.MatchTiming.ShiftDurationSec, game.MatchTiming.EndgameDurationSec, game.MatchTiming.TimeoutDurationSec}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.Timing = value
	web.apiV1Displays.announcer.Timing = value
	web.apiV1Displays.audience.Timing = value
	web.apiV1Displays.allianceStation.Timing = value
	web.apiV1Displays.wall.Timing = value
	web.apiV1Displays.fieldMonitor.Timing = value
	web.apiV1Displays.fieldMonitorFta.Timing = value
	web.apiV1Displays.matchPlay.Timing = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1EventStatus() {
	value := apiV1DisplayEventStatus{web.arena.EventStatus.CycleTime, web.arena.EventStatus.EarlyLateMessage}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.Event = value
	web.apiV1Displays.announcer.Event = value
	web.apiV1Displays.fieldMonitor.Event = value
	web.apiV1Displays.fieldMonitorFta.Event = value
	web.apiV1Displays.matchPlay.Event = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1AudienceMode() {
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.announcer.AudienceDisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.audience.DisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.wall.DisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.unpicked.DisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.matchPlay.AudienceDisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1AudienceState() {
	web.refreshApiV1DisplayMatch()
	web.refreshApiV1AudienceRealtimeScore()
	web.refreshApiV1AudiencePostedScore()
	web.refreshApiV1AllianceSelection()
	web.refreshApiV1LowerThird()
}

func (web *Web) refreshApiV1AllianceStationState() {
	web.refreshApiV1AllianceStationMode()
	web.refreshApiV1StationStatuses()
}

func (web *Web) refreshApiV1DisplayMatch() {
	value := web.buildApiV1DisplayMatch(web.arena.CurrentMatch)
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.audience.Match = value
	web.apiV1Displays.allianceStation.Match = value
	web.apiV1Displays.wall.Match = value
	web.apiV1Displays.fieldMonitor.Match = value
	web.apiV1Displays.fieldMonitorFta.Match = value
	web.apiV1Displays.scoringPanel.Match = value
	web.apiV1Displays.refereePanel.Match = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) buildApiV1DisplayMatch(match *model.Match) apiV1DisplayMatch {
	value := apiV1DisplayMatch{Teams: make(map[string]*apiV1DisplayTeam), Rankings: make(map[string]int), RedOffFieldTeams: []apiV1DisplayTeam{}, BlueOffFieldTeams: []apiV1DisplayTeam{}}
	if match == nil {
		return value
	}
	value.Id, value.Type, value.LongName, value.NameDetail = match.Id, apiV1MatchType(match.Type), match.LongName, match.NameDetail
	value.PlayoffRedAlliance, value.PlayoffBlueAlliance = match.PlayoffRedAlliance, match.PlayoffBlueAlliance
	value.Red, value.Blue = [3]int{match.Red1, match.Red2, match.Red3}, [3]int{match.Blue1, match.Blue2, match.Blue3}
	value.BreakDescription, value.BreakNextMatchName = web.arena.DisplayBreakDetails()
	for _, station := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		team := web.arena.AllianceStations[station].Team
		if team == nil {
			value.Teams[station] = nil
			continue
		}
		value.Teams[station] = &apiV1DisplayTeam{Id: team.Id, Nickname: team.Nickname, YellowCard: team.YellowCard}
		if ranking, err := web.arena.Database.GetRankingForTeam(team.Id); err == nil && ranking != nil {
			value.Rankings[strconv.Itoa(team.Id)] = ranking.Rank
		}
	}
	if match.Type == model.Playoff {
		if group := web.arena.PlayoffTournament.MatchGroups()[match.PlayoffMatchGroupId]; group != nil {
			if series, ok := group.(*playoff.Matchup); ok {
				value.Series = &apiV1DisplaySeries{series.NumWinsToAdvance, series.RedAllianceWins, series.BlueAllianceWins}
			}
		}
		redIds, blueIds, err := web.arena.Database.GetOffFieldTeamIds(match)
		if err == nil {
			value.RedOffFieldTeams = web.apiV1DisplayTeams(redIds)
			value.BlueOffFieldTeams = web.apiV1DisplayTeams(blueIds)
		}
	}
	return value
}

func (web *Web) apiV1DisplayTeams(ids []int) []apiV1DisplayTeam {
	items := make([]apiV1DisplayTeam, 0, len(ids))
	for _, id := range ids {
		team, err := web.arena.Database.GetTeamById(id)
		if err != nil || team == nil {
			continue
		}
		items = append(items, apiV1DisplayTeam{Id: team.Id, Nickname: team.Nickname, YellowCard: team.YellowCard})
		if ranking, err := web.arena.Database.GetRankingForTeam(id); err == nil && ranking != nil {
			// The ranking is used by Alliance Station for off-field context only through the match-level map.
			_ = ranking
		}
	}
	return items
}

func (web *Web) refreshApiV1AudienceRealtimeScore() {
	red, blue := web.arena.RedScoreSummary(), web.arena.BlueScoreSummary()
	value := apiV1AudienceRealtimeScore{
		Red:  apiV1AudienceRealtimeAlliance{newApiV1ScoreSummary(red), web.arena.RedRealtimeScore.ActiveRemainingSec, web.arena.RedRealtimeScore.ActiveDurationSec},
		Blue: apiV1AudienceRealtimeAlliance{newApiV1ScoreSummary(blue), web.arena.BlueRealtimeScore.ActiveRemainingSec, web.arena.BlueRealtimeScore.ActiveDurationSec},
	}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.audience.RealtimeScore = value
	web.apiV1Displays.wall.RealtimeScore = value
	web.apiV1Displays.fieldMonitor.RealtimeScore = value
	web.apiV1Displays.fieldMonitorFta.RealtimeScore = value
	web.apiV1Displays.matchPlay.RealtimeScore = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1AudiencePostedScore() {
	value := web.buildApiV1AudiencePostedScore()
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.audience.PostedScore = value
	web.apiV1Displays.matchPlay.PostedScore = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) buildApiV1AudiencePostedScore() *apiV1AudiencePostedScore {
	if web.arena.SavedMatch == nil || web.arena.SavedMatchResult == nil {
		return nil
	}
	match, result := web.arena.SavedMatch, web.arena.SavedMatchResult
	redSummary, blueSummary := result.RedScoreSummary(), result.BlueScoreSummary()
	_, tiebreakReason := game.DetermineMatchStatus(redSummary, blueSummary, match.UseTiebreakCriteria)
	redRp, blueRp := redSummary.BonusRankingPoints, blueSummary.BonusRankingPoints
	redWon, blueWon := match.Status == game.RedWonMatch, match.Status == game.BlueWonMatch
	switch match.Status {
	case game.RedWonMatch:
		redRp += game.GetWinRankingPoints()
	case game.BlueWonMatch:
		blueRp += game.GetWinRankingPoints()
	case game.TieMatch:
		redRp++
		blueRp++
	}
	redOff, blueOff := []int{}, []int{}
	redWins, blueWins, redDestination, blueDestination := 0, 0, "", ""
	if match.Type == model.Playoff {
		redOff, blueOff, _ = web.arena.Database.GetOffFieldTeamIds(match)
		if group := web.arena.PlayoffTournament.MatchGroups()[match.PlayoffMatchGroupId]; group != nil {
			if series, ok := group.(*playoff.Matchup); ok {
				redWins, blueWins, redDestination, blueDestination = series.RedAllianceWins, series.BlueAllianceWins, series.RedAllianceDestination(), series.BlueAllianceDestination()
			}
		}
	}
	return &apiV1AudiencePostedScore{
		Match: web.buildApiV1DisplayMatch(match), TiebreakReason: tiebreakReason,
		Red:  apiV1AudiencePostedAlliance{newApiV1ScoreSummary(redSummary), redRp, copyStringMap(result.RedCards), web.apiV1AudienceRankings(match.Red1, match.Red2, match.Red3), redOff, redWon, redWins, redDestination},
		Blue: apiV1AudiencePostedAlliance{newApiV1ScoreSummary(blueSummary), blueRp, copyStringMap(result.BlueCards), web.apiV1AudienceRankings(match.Blue1, match.Blue2, match.Blue3), blueOff, blueWon, blueWins, blueDestination},
	}
}

func copyStringMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func (web *Web) apiV1AudienceRankings(teamIds ...int) map[string]*apiV1AnnouncerRanking {
	result := make(map[string]*apiV1AnnouncerRanking, len(teamIds))
	for _, teamId := range teamIds {
		key := strconv.Itoa(teamId)
		result[key] = nil
		for _, ranking := range web.arena.SavedRankings {
			if ranking.TeamId == teamId {
				item := apiV1AnnouncerRanking{ranking.TeamId, ranking.Rank, ranking.PreviousRank}
				result[key] = &item
				break
			}
		}
	}
	return result
}

func (web *Web) refreshApiV1AllianceSelection() {
	value := apiV1AudienceAllianceSelection{Alliances: make([]apiV1AllianceSelectionAlliance, 0, len(web.arena.AllianceSelectionAlliances)), RankedTeams: make([]apiV1AllianceSelectionTeam, 0, len(web.arena.AllianceSelectionRankedTeams)), ShowTimer: web.arena.AllianceSelectionShowTimer, TimeRemainingSec: web.arena.AllianceSelectionTimeRemainingSec}
	for _, alliance := range web.arena.AllianceSelectionAlliances {
		value.Alliances = append(value.Alliances, apiV1AllianceSelectionAlliance{alliance.Id, append([]int(nil), alliance.TeamIds...)})
	}
	for _, team := range web.arena.AllianceSelectionRankedTeams {
		value.RankedTeams = append(value.RankedTeams, apiV1AllianceSelectionTeam{team.Rank, team.TeamId, team.Picked})
	}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.audience.AllianceSelection = value
	web.apiV1Displays.unpicked.AllianceSelection = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1LowerThird() {
	value := apiV1AudienceLowerThird{ShowLowerThird: web.arena.ShowLowerThird}
	if web.arena.LowerThird != nil {
		value.TopText, value.BottomText = web.arena.LowerThird.TopText, web.arena.LowerThird.BottomText
	}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.audience.LowerThird = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1AllianceStationMode() {
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.allianceStation.DisplayMode = web.arena.AllianceStationDisplayMode
	web.apiV1Displays.matchPlay.AllianceStationDisplayMode = web.arena.AllianceStationDisplayMode
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1StationStatuses() {
	value := make(map[string]apiV1AllianceStationStatus, 6)
	for _, station := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		status := web.arena.AllianceStations[station]
		item := apiV1AllianceStationStatus{Bypass: status.Bypass}
		if status.DsConn != nil {
			item.DsLinked, item.RobotLinked = status.DsConn.DsLinked, status.DsConn.RobotLinked
		}
		value[station] = item
	}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.allianceStation.Stations = value
	web.apiV1Displays.refereePanel.Stations = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1FieldMonitorStatus() {
	publicStatus := web.buildApiV1FieldMonitorStatus(false)
	ftaStatus := web.buildApiV1FieldMonitorStatus(true)
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.fieldMonitor.ArenaStatus = publicStatus
	web.apiV1Displays.fieldMonitorFta.ArenaStatus = ftaStatus
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1MatchPlayState() {
	web.refreshApiV1MatchPlayMatch()
	web.refreshApiV1MatchPlayArenaStatus()
	web.refreshApiV1ScoringStatus()
}

func (web *Web) refreshApiV1MatchPlayMatch() {
	match := web.arena.CurrentMatch
	value := apiV1MatchPlayMatch{apiV1DisplayMatch: web.buildApiV1DisplayMatch(match)}
	if match != nil {
		value.AllowSubstitution = match.ShouldAllowSubstitution() && !(web.arena.EventSettings.NexusEnabled && match.ShouldAllowNexusSubstitution())
		if result, err := web.arena.Database.GetMatchResultForMatch(match.Id); err == nil {
			value.IsReplay = result != nil
		}
	}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.matchPlay.Match = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1MatchPlayArenaStatus() {
	control := web.arena.MatchPlayControlStatusSnapshot()
	stations := web.buildApiV1FieldMonitorStatus(false).Stations
	value := apiV1MatchPlayArenaStatus{State: apiV1MatchState(control.MatchState), CanStartMatch: control.CanStartMatch,
		StartMatchConditions: control.StartMatchConditions, Stations: stations, AccessPointStatus: control.AccessPointStatus,
		SwitchStatus: control.SwitchStatus, RedSccStatus: control.RedSCCStatus, BlueSccStatus: control.BlueSCCStatus,
		PlcIsHealthy: control.PlcIsHealthy, FieldEStop: control.FieldEStop, IsFtaReady: control.IsFtaReady,
		PlcArmorBlockStatuses: control.PlcArmorBlockStatuses}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.matchPlay.ArenaStatus = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1ScoringStatus() {
	positions := make(map[string]apiV1ScoringPositionStatus, 2)
	for _, position := range []string{"red", "blue"} {
		numPanels := web.arena.ScoringPanelRegistry.GetNumPanels(position)
		numReady := web.arena.ScoringPanelRegistry.GetNumScoreCommitted(position)
		positions[position] = apiV1ScoringPositionStatus{Ready: numPanels > 0 && numReady >= numPanels, NumPanels: numPanels, NumPanelsReady: numReady}
	}
	value := apiV1ScoringStatus{RefereeScoreReady: web.arena.RedRealtimeScore.FoulsCommitted && web.arena.BlueRealtimeScore.FoulsCommitted, Positions: positions}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.matchPlay.ScoringStatus = value
	web.apiV1Displays.refereePanel.ScoringStatus = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) refreshApiV1ControlRealtimeScore() {
	build := func(source *field.RealtimeScore) apiV1ControlScoreAlliance {
		value := apiV1ControlScoreAlliance{Fouls: make([]apiV1ControlFoul, 0, len(source.CurrentScore.Fouls)), Cards: copyStringMap(source.Cards)}
		for index := 0; index < 3; index++ {
			value.AutoTowerStatuses[index] = int(source.CurrentScore.AutoTowerStatuses[index])
			value.EndgameTowerStatuses[index] = int(source.CurrentScore.EndgameTowerStatuses[index])
		}
		for _, foul := range source.CurrentScore.Fouls {
			value.Fouls = append(value.Fouls, apiV1ControlFoul{foul.FoulId, foul.IsMajor, foul.TeamId, foul.RuleId})
		}
		return value
	}
	value := apiV1ControlRealtimeScore{Red: build(web.arena.RedRealtimeScore), Blue: build(web.arena.BlueRealtimeScore)}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.scoringPanel.RealtimeScore = value
	web.apiV1Displays.refereePanel.RealtimeScore = value
	web.apiV1Displays.mu.Unlock()
}

func (web *Web) buildApiV1FieldMonitorStatus(includeFtaNotes bool) apiV1FieldMonitorStatus {
	accessPointStatus, switchStatus := web.arena.FieldMonitorInfrastructureStatus()
	value := apiV1FieldMonitorStatus{AccessPointStatus: accessPointStatus, SwitchStatus: switchStatus, Stations: make(map[string]apiV1FieldMonitorStation, 6)}
	if web.arena.CurrentMatch != nil {
		value.MatchId = web.arena.CurrentMatch.Id
	}
	for _, station := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		source := web.arena.AllianceStations[station]
		item := apiV1FieldMonitorStation{Ethernet: source.Ethernet, AStop: source.AStop, EStop: source.EStop, Bypass: source.Bypass,
			Wifi: apiV1FieldMonitorWifi{TeamId: source.WifiStatus.TeamId, RadioLinked: source.WifiStatus.RadioLinked, MBits: source.WifiStatus.MBits, ConnectionQuality: source.WifiStatus.ConnectionQuality}}
		if source.Team != nil {
			item.Team = &apiV1FieldMonitorTeam{Id: source.Team.Id}
			if includeFtaNotes {
				notes := source.Team.FtaNotes
				item.Team.FtaNotes = &notes
			}
		}
		if source.DsConn != nil {
			connection := source.DsConn
			item.Connection = &apiV1FieldMonitorConnection{WrongStation: connection.WrongStation, DsLinked: connection.DsLinked, RadioLinked: connection.RadioLinked, RioLinked: connection.RioLinked, RobotLinked: connection.RobotLinked, BatteryVoltage: connection.BatteryVoltage, DsRobotTripTimeMs: connection.DsRobotTripTimeMs, MissedPacketCount: connection.MissedPacketCount, SecondsSinceLastRobotLink: connection.SecondsSinceLastRobotLink}
		}
		value.Stations[station] = item
	}
	return value
}

func (web *Web) apiV1QueueingMatchesSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.queueing.Matches
}
func (web *Web) apiV1AnnouncerMatchSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.announcer.Match
}
func (web *Web) apiV1PostedScoreSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.announcer.PostedScore
}
func (web *Web) apiV1RealtimeScoreSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.announcer.RealtimeScore
}
func (web *Web) apiV1MatchClockSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.queueing.MatchClock
}
func (web *Web) apiV1TimingSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.queueing.Timing
}
func (web *Web) apiV1EventStatusSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.queueing.Event
}
func (web *Web) apiV1AudienceModeSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.announcer.AudienceDisplayMode
}
func (web *Web) apiV1DisplayMatchSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.audience.Match
}
func (web *Web) apiV1AudienceRealtimeSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.audience.RealtimeScore
}
func (web *Web) apiV1AudiencePostedSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.audience.PostedScore
}
func (web *Web) apiV1AllianceSelectionSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.audience.AllianceSelection
}
func (web *Web) apiV1LowerThirdSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.audience.LowerThird
}
func (web *Web) apiV1AllianceStationModeSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.allianceStation.DisplayMode
}
func (web *Web) apiV1StationStatusesSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.allianceStation.Stations
}
func (web *Web) apiV1FieldMonitorStatusSnapshot(isFta bool) any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	if isFta {
		return web.apiV1Displays.fieldMonitorFta.ArenaStatus
	}
	return web.apiV1Displays.fieldMonitor.ArenaStatus
}
func (web *Web) apiV1MatchPlayMatchSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.matchPlay.Match
}
func (web *Web) apiV1MatchPlayArenaStatusSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.matchPlay.ArenaStatus
}
func (web *Web) apiV1ScoringStatusSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.matchPlay.ScoringStatus
}
func (web *Web) apiV1ControlRealtimeScoreSnapshot() any {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	return web.apiV1Displays.scoringPanel.RealtimeScore
}

func (web *Web) apiV1QueueingBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.queueing, nil)
}
func (web *Web) apiV1AnnouncerBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.announcer, nil)
}
func (web *Web) apiV1AudienceBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.audience, nil)
}
func (web *Web) apiV1AllianceStationBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.allianceStation, nil)
}
func (web *Web) apiV1WallBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.wall, nil)
}
func (web *Web) apiV1UnpickedBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.unpicked, nil)
}
func (web *Web) apiV1FieldMonitorBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	isFta := r.URL.Query().Get("fta") == "true"
	if isFta {
		web.apiV1RequireAdmin(http.HandlerFunc(web.apiV1FieldMonitorFtaBootstrapHandler)).ServeHTTP(w, r)
		return
	}
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.fieldMonitor, nil)
}

func (web *Web) apiV1AllianceSelectionControlBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, apiV1AllianceSelectionControlBootstrap{
		StreamUrl: "/api/v1/streams/admin/alliance-selection", DisplayMode: web.apiV1Displays.unpicked.DisplayMode,
		AllianceSelection: web.apiV1Displays.unpicked.AllianceSelection,
	}, nil)
}
func (web *Web) apiV1MatchPlayBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.matchPlay, nil)
}
func (web *Web) apiV1ScoringPanelBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	position := r.PathValue("position")
	if _, ok := positionParameters[position]; !ok {
		writeApiV1Error(w, r, http.StatusNotFound, "not_found", "Scoring position not found.", nil)
		return
	}
	web.apiV1Displays.mu.RLock()
	value := web.apiV1Displays.scoringPanel
	web.apiV1Displays.mu.RUnlock()
	value.StreamUrl = "/api/v1/streams/admin/scoring/" + position
	writeApiV1Data(w, r, http.StatusOK, value, nil)
}
func (web *Web) apiV1RefereePanelBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.refereePanel, nil)
}
func (web *Web) apiV1FieldMonitorFtaBootstrapHandler(w http.ResponseWriter, r *http.Request) {
	web.apiV1Displays.mu.RLock()
	defer web.apiV1Displays.mu.RUnlock()
	writeApiV1Data(w, r, http.StatusOK, web.apiV1Displays.fieldMonitorFta, nil)
}

func (web *Web) apiV1QueueingStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/queueing/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.queueingMatches, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.eventStatus, web.apiV1Displays.reload)
}
func (web *Web) apiV1AnnouncerStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/announcer/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.announcerMatch, web.apiV1Displays.postedScore, web.apiV1Displays.realtimeScore, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.eventStatus, web.apiV1Displays.audienceMode, web.apiV1Displays.reload)
}

func (web *Web) apiV1AudienceStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/audience/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.displayMatch, web.apiV1Displays.audienceRealtime, web.apiV1Displays.audiencePosted, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.audienceMode, web.apiV1Displays.allianceSelection, web.apiV1Displays.lowerThird, web.apiV1Displays.playSound, web.apiV1Displays.reload)
}

func (web *Web) apiV1AllianceStationStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/alliance_station/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.displayMatch, web.apiV1Displays.realtimeScore, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.allianceStationMode, web.apiV1Displays.stationStatuses, web.apiV1Displays.reload)
}

func (web *Web) apiV1WallStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/wall/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.displayMatch, web.apiV1Displays.audienceRealtime, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.audienceMode, web.apiV1Displays.reload)
}

func (web *Web) apiV1UnpickedStreamHandler(w http.ResponseWriter, r *http.Request) {
	display, err := web.registerDisplayForPath(r, "/displays/unpicked/websocket")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(display.Notifier, web.apiV1Displays.allianceSelection, web.apiV1Displays.audienceMode, web.apiV1Displays.reload)
}

func (web *Web) apiV1FieldMonitorStreamHandler(w http.ResponseWriter, r *http.Request) {
	isFta := r.URL.Query().Get("fta") == "true"
	if isFta && !web.userIsAdmin(w, r) {
		return
	}
	legacyPath := "/displays/field_monitor/websocket"
	if r.URL.Query().Get("fms") == "true" {
		legacyPath = "/displays/fms_field_monitor/websocket"
	}
	display, err := web.registerDisplayForPath(r, legacyPath)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer web.arena.MarkDisplayDisconnected(display.DisplayConfiguration.Id)
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	statusNotifier := web.apiV1Displays.fieldMonitorStatus
	if isFta {
		statusNotifier = web.apiV1Displays.fieldMonitorFtaStatus
	}
	ws.HandleNotifiersV1(display.Notifier, statusNotifier, web.apiV1Displays.eventStatus, web.apiV1Displays.displayMatch, web.apiV1Displays.audienceRealtime, web.apiV1Displays.timing, web.apiV1Displays.matchClock, web.apiV1Displays.reload)
}

func (web *Web) apiV1AllianceSelectionControlStreamHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.allianceSelection, web.apiV1Displays.audienceMode)
}
func (web *Web) apiV1MatchPlayStreamHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.matchPlayMatch, web.apiV1Displays.matchPlayArenaStatus,
		web.apiV1Displays.audienceMode, web.apiV1Displays.allianceStationMode, web.apiV1Displays.eventStatus,
		web.apiV1Displays.audienceRealtime, web.apiV1Displays.audiencePosted, web.apiV1Displays.scoringStatus,
		web.apiV1Displays.timing, web.apiV1Displays.matchClock)
}
func (web *Web) apiV1ScoringPanelStreamHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	if _, ok := positionParameters[r.PathValue("position")]; !ok {
		http.NotFound(w, r)
		return
	}
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.displayMatch, web.apiV1Displays.matchClock, web.apiV1Displays.controlRealtimeScore)
}
func (web *Web) apiV1RefereePanelStreamHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.displayMatch, web.apiV1Displays.matchClock, web.apiV1Displays.controlRealtimeScore, web.apiV1Displays.scoringStatus, web.apiV1Displays.stationStatuses)
}

func apiV1MatchState(state field.MatchState) string {
	switch state {
	case field.PreMatch:
		return "pre_match"
	case field.StartMatch:
		return "start_match"
	case field.AutoPeriod:
		return "auto"
	case field.PausePeriod:
		return "pause"
	case field.TeleopPeriod:
		return "teleop"
	case field.PostMatch:
		return "post_match"
	case field.TimeoutActive:
		return "timeout_active"
	case field.PostTimeout:
		return "post_timeout"
	default:
		return "unknown"
	}
}
