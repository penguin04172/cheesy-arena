// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"log"
	"net/http"
	"sync"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
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
	mu        sync.RWMutex
	queueing  apiV1QueueingBootstrap
	announcer apiV1AnnouncerBootstrap

	queueingMatches *websocket.Notifier
	announcerMatch  *websocket.Notifier
	postedScore     *websocket.Notifier
	realtimeScore   *websocket.Notifier
	matchClock      *websocket.Notifier
	timing          *websocket.Notifier
	eventStatus     *websocket.Notifier
	audienceMode    *websocket.Notifier
	reload          *websocket.Notifier
}

func (web *Web) initializeApiV1DisplayState() {
	state := &apiV1DisplayState{}
	web.apiV1Displays = state
	state.queueing.StreamUrl = "/api/v1/streams/displays/queueing"
	state.announcer.StreamUrl = "/api/v1/streams/displays/announcer"
	web.refreshApiV1DisplayMatches()
	web.refreshApiV1PostedScore()
	web.refreshApiV1RealtimeScore()
	web.refreshApiV1MatchClock()
	web.refreshApiV1DisplayTiming()
	web.refreshApiV1EventStatus()
	web.refreshApiV1AudienceMode()

	state.queueingMatches = websocket.NewNotifier("matches", func() any { return web.apiV1QueueingMatchesSnapshot() })
	state.announcerMatch = websocket.NewNotifier("match", func() any { return web.apiV1AnnouncerMatchSnapshot() })
	state.postedScore = websocket.NewNotifier("postedScore", func() any { return web.apiV1PostedScoreSnapshot() })
	state.realtimeScore = websocket.NewNotifier("realtimeScore", func() any { return web.apiV1RealtimeScoreSnapshot() })
	state.matchClock = websocket.NewNotifier("matchClock", func() any { return web.apiV1MatchClockSnapshot() })
	state.timing = websocket.NewNotifier("timing", func() any { return web.apiV1TimingSnapshot() })
	state.eventStatus = websocket.NewNotifier("eventStatus", func() any { return web.apiV1EventStatusSnapshot() })
	state.audienceMode = websocket.NewNotifier("audienceDisplayMode", func() any { return web.apiV1AudienceModeSnapshot() })
	state.reload = websocket.NewNotifier("reload", nil)

	web.arena.MatchLoadNotifier.Observe(func(any) {
		web.refreshApiV1DisplayMatches()
		state.queueingMatches.Notify()
		state.announcerMatch.Notify()
	})
	web.arena.ScorePostedNotifier.Observe(func(any) { web.refreshApiV1PostedScore(); state.postedScore.Notify() })
	web.arena.RealtimeScoreNotifier.Observe(func(any) { web.refreshApiV1RealtimeScore(); state.realtimeScore.Notify() })
	web.arena.MatchTimeNotifier.Observe(func(any) { web.refreshApiV1MatchClock(); state.matchClock.Notify() })
	web.arena.MatchTimingNotifier.Observe(func(any) { web.refreshApiV1DisplayTiming(); state.timing.Notify() })
	web.arena.EventStatusNotifier.Observe(func(any) { web.refreshApiV1EventStatus(); state.eventStatus.Notify() })
	web.arena.AudienceDisplayModeNotifier.Observe(func(any) { web.refreshApiV1AudienceMode(); state.audienceMode.Notify() })
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
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1MatchClock() {
	value := apiV1MatchClock{State: apiV1MatchState(web.arena.MatchState), ElapsedSec: int(web.arena.MatchTimeSec())}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.MatchClock = value
	web.apiV1Displays.announcer.MatchClock = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1DisplayTiming() {
	value := apiV1DisplayMatchTiming{game.MatchTiming.AutoDurationSec, game.MatchTiming.PauseDurationSec, game.MatchTiming.TransitionShiftDurationSec, game.MatchTiming.ShiftDurationSec, game.MatchTiming.EndgameDurationSec, game.MatchTiming.TimeoutDurationSec}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.Timing = value
	web.apiV1Displays.announcer.Timing = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1EventStatus() {
	value := apiV1DisplayEventStatus{web.arena.EventStatus.CycleTime, web.arena.EventStatus.EarlyLateMessage}
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.queueing.Event = value
	web.apiV1Displays.announcer.Event = value
	web.apiV1Displays.mu.Unlock()
}
func (web *Web) refreshApiV1AudienceMode() {
	web.apiV1Displays.mu.Lock()
	web.apiV1Displays.announcer.AudienceDisplayMode = web.arena.AudienceDisplayMode
	web.apiV1Displays.mu.Unlock()
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

func (web *Web) apiV1QueueingStreamHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.queueingMatches, web.apiV1Displays.matchClock, web.apiV1Displays.timing, web.apiV1Displays.eventStatus, web.apiV1Displays.reload)
}
func (web *Web) apiV1AnnouncerStreamHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		return
	}
	defer ws.Close()
	ws.HandleNotifiersV1(web.apiV1Displays.announcerMatch, web.apiV1Displays.postedScore, web.apiV1Displays.realtimeScore, web.apiV1Displays.matchClock, web.apiV1Displays.timing, web.apiV1Displays.eventStatus, web.apiV1Displays.audienceMode, web.apiV1Displays.reload)
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
