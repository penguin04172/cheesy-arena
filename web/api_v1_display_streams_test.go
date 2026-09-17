// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	cawebsocket "github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1DisplayBootstrapsUseProjectedDtos(t *testing.T) {
	web := setupTestWeb(t)
	team := model.Team{Id: 254, Nickname: "Poofs", WpaKey: "never-expose", FtaNotes: "private-notes"}
	require.NoError(t, web.arena.Database.CreateTeam(&team))
	match := model.Match{Type: model.Practice, TypeOrder: 1, LongName: "Practice 1", ShortName: "P1", Red1: 254, Status: game.MatchScheduled, Time: time.Now().UTC()}
	require.NoError(t, web.arena.Database.CreateMatch(&match))
	require.NoError(t, web.arena.LoadMatch(&match))
	web.arena.EventStatus.CycleTime = "6:00"
	web.arena.EventStatus.EarlyLateMessage = "On time"
	web.arena.EventStatusNotifier.Notify()

	queueing := web.getHttpResponse("/api/v1/displays/queueing/bootstrap")
	require.Equal(t, http.StatusOK, queueing.Code)
	var queueingResponse struct {
		Data apiV1QueueingBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(queueing.Body.Bytes(), &queueingResponse))
	assert.Equal(t, "/api/v1/streams/displays/queueing", queueingResponse.Data.StreamUrl)
	require.Len(t, queueingResponse.Data.Matches, 1)
	assert.Equal(t, "P1", queueingResponse.Data.Matches[0].ShortName)
	assert.Equal(t, "6:00", queueingResponse.Data.Event.CycleTime)

	announcer := web.getHttpResponse("/api/v1/displays/announcer/bootstrap")
	require.Equal(t, http.StatusOK, announcer.Code)
	assert.NotContains(t, announcer.Body.String(), "never-expose")
	assert.NotContains(t, announcer.Body.String(), "private-notes")
	var announcerResponse struct {
		Data apiV1AnnouncerBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(announcer.Body.Bytes(), &announcerResponse))
	assert.Equal(t, "/api/v1/streams/displays/announcer", announcerResponse.Data.StreamUrl)
	assert.Equal(t, "Practice 1", announcerResponse.Data.Match.LongName)
	require.NotNil(t, announcerResponse.Data.Match.Red.Teams[0])
	assert.Equal(t, "Poofs", announcerResponse.Data.Match.Red.Teams[0].Nickname)
}

func TestApiV1AudienceAndAllianceStationBootstrapsUseSafeDtos(t *testing.T) {
	web := setupTestWeb(t)
	team := model.Team{Id: 254, Nickname: "Poofs", WpaKey: "audience-secret", FtaNotes: "station-secret", YellowCard: true}
	require.NoError(t, web.arena.Database.CreateTeam(&team))
	match := model.Match{Type: model.Qualification, TypeOrder: 1, LongName: "Qualification 1", Red1: 254, Status: game.MatchScheduled, Time: time.Now().UTC()}
	require.NoError(t, web.arena.Database.CreateMatch(&match))
	require.NoError(t, web.arena.LoadMatch(&match))

	for _, path := range []string{"/api/v1/displays/audience/bootstrap", "/api/v1/displays/alliance-station/bootstrap"} {
		response := web.getHttpResponse(path)
		require.Equal(t, http.StatusOK, response.Code)
		assert.NotContains(t, response.Body.String(), "audience-secret")
		assert.NotContains(t, response.Body.String(), "station-secret")
	}
	var audienceResponse struct {
		Data apiV1AudienceBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(web.getHttpResponse("/api/v1/displays/audience/bootstrap").Body.Bytes(), &audienceResponse))
	assert.Equal(t, "/api/v1/streams/displays/audience", audienceResponse.Data.StreamUrl)
	require.NotNil(t, audienceResponse.Data.Match.Teams["R1"])
	assert.True(t, audienceResponse.Data.Match.Teams["R1"].YellowCard)
	assert.NotNil(t, audienceResponse.Data.AllianceSelection.Alliances)

	var stationResponse struct {
		Data apiV1AllianceStationBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(web.getHttpResponse("/api/v1/displays/alliance-station/bootstrap").Body.Bytes(), &stationResponse))
	assert.Equal(t, "/api/v1/streams/displays/alliance-station", stationResponse.Data.StreamUrl)
	assert.Contains(t, stationResponse.Data.Stations, "R1")
}

func TestApiV1AudienceAndAllianceStationStreamContracts(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	for _, test := range []struct {
		path     string
		expected []string
	}{
		{"/api/v1/streams/displays/audience?displayId=3", []string{"displayConfiguration", "match", "realtimeScore", "postedScore", "timing", "matchClock", "audienceDisplayMode", "allianceSelection", "lowerThird"}},
		{"/api/v1/streams/displays/alliance-station?displayId=4", []string{"displayConfiguration", "match", "realtimeScore", "timing", "matchClock", "allianceStationDisplayMode", "stationStatuses"}},
		{"/api/v1/streams/displays/wall?displayId=5", []string{"displayConfiguration", "match", "realtimeScore", "timing", "matchClock", "audienceDisplayMode"}},
		{"/api/v1/streams/displays/unpicked?displayId=6", []string{"displayConfiguration", "allianceSelection", "audienceDisplayMode"}},
	} {
		conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+test.path, nil)
		require.NoError(t, err)
		for _, expectedType := range test.expected {
			var message cawebsocket.V1Message
			require.NoError(t, conn.ReadJSON(&message))
			assert.Equal(t, expectedType, message.Type)
			assert.True(t, message.Meta.Bootstrap)
		}
		var ready cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&ready))
		assert.Equal(t, "ready", ready.Type)
		require.NoError(t, conn.Close())
	}
}

func TestAudienceAndAllianceStationClientsUseV1Streams(t *testing.T) {
	for _, path := range []string{"../static/js/audience_display.js", "../static/js/alliance_station_display.js", "../static/js/wall_display.js", "../static/js/unpicked_display.js"} {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		source := string(contents)
		assert.Contains(t, source, "new CheesyWebsocketV1")
		assert.Contains(t, source, "/api/v1/displays/")
		assert.NotContains(t, source, "new CheesyWebsocket(\"")
	}
}

func TestApiV1WallAndUnpickedBootstraps(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.AudienceDisplayMode = "allianceSelection"
	web.arena.AllianceSelectionRankedTeams = []model.AllianceSelectionRankedTeam{{Rank: 1, TeamId: 254}}
	web.arena.AudienceDisplayModeNotifier.Notify()
	web.arena.AllianceSelectionNotifier.Notify()

	var wall struct {
		Data apiV1WallBootstrap `json:"data"`
	}
	response := web.getHttpResponse("/api/v1/displays/wall/bootstrap")
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &wall))
	assert.Equal(t, "/api/v1/streams/displays/wall", wall.Data.StreamUrl)
	assert.Equal(t, "allianceSelection", wall.Data.DisplayMode)

	var unpicked struct {
		Data apiV1UnpickedBootstrap `json:"data"`
	}
	response = web.getHttpResponse("/api/v1/displays/unpicked/bootstrap")
	require.Equal(t, http.StatusOK, response.Code)
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &unpicked))
	assert.Equal(t, "/api/v1/streams/displays/unpicked", unpicked.Data.StreamUrl)
	require.Len(t, unpicked.Data.AllianceSelection.RankedTeams, 1)
	assert.Equal(t, 254, unpicked.Data.AllianceSelection.RankedTeams[0].TeamId)
}

func TestApiV1FieldMonitorBootstrapRedactsFtaNotes(t *testing.T) {
	web := setupTestWeb(t)
	team := model.Team{Id: 254, Nickname: "Poofs", WpaKey: "wifi-secret", FtaNotes: "FTA-only note"}
	require.NoError(t, web.arena.Database.CreateTeam(&team))
	require.NoError(t, web.arena.SubstituteTeams(0, 0, 0, 254, 0, 0))
	web.arena.ArenaStatusNotifier.Notify()

	publicResponse := web.getHttpResponse("/api/v1/displays/field-monitor/bootstrap?fta=false")
	require.Equal(t, http.StatusOK, publicResponse.Code)
	assert.NotContains(t, publicResponse.Body.String(), "FTA-only note")
	assert.NotContains(t, publicResponse.Body.String(), "wifi-secret")

	ftaResponse := web.getHttpResponse("/api/v1/displays/field-monitor/bootstrap?fta=true")
	require.Equal(t, http.StatusOK, ftaResponse.Code)
	assert.Contains(t, ftaResponse.Body.String(), "FTA-only note")
	assert.NotContains(t, ftaResponse.Body.String(), "wifi-secret")

	web.arena.EventSettings.AdminPassword = "password"
	request := httptest.NewRequest(http.MethodGet, "/api/v1/displays/field-monitor/bootstrap?fta=true", nil)
	unauthorized := httptest.NewRecorder()
	web.newHandler().ServeHTTP(unauthorized, request)
	assert.Equal(t, http.StatusUnauthorized, unauthorized.Code)
	assert.Contains(t, unauthorized.Body.String(), "authentication_required")
}

func TestApiV1FieldMonitorStreamContract(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/displays/field-monitor?displayId=7&fta=false", nil)
	require.NoError(t, err)
	defer conn.Close()
	for _, expectedType := range []string{"displayConfiguration", "arenaStatus", "eventStatus", "match", "realtimeScore", "timing", "matchClock"} {
		var message cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&message))
		assert.Equal(t, expectedType, message.Type)
		assert.True(t, message.Meta.Bootstrap)
	}
	var ready cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&ready))
	assert.Equal(t, "ready", ready.Type)
}

func TestFieldMonitorClientSplitsV1StateAndLegacyCommands(t *testing.T) {
	contents, err := os.ReadFile("../static/js/field_monitor_display.js")
	require.NoError(t, err)
	source := string(contents)
	assert.Contains(t, source, "new CheesyWebsocketV1")
	assert.Contains(t, source, "/api/v1/displays/field-monitor/bootstrap")
	assert.Contains(t, source, "commandWebsocket = new CheesyWebsocket")
	assert.Contains(t, source, "commandWebsocket.send(\"updateTeamNotes\"")
}

func TestApiV1AllianceSelectionControlBootstrapAndStream(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.AllianceSelectionTimeRemainingSec = 42
	web.arena.AudienceDisplayMode = "allianceSelection"
	web.arena.AllianceSelectionNotifier.Notify()
	web.arena.AudienceDisplayModeNotifier.Notify()
	response := web.getHttpResponse("/api/v1/admin/alliance-selection/bootstrap")
	require.Equal(t, http.StatusOK, response.Code)
	var body struct {
		Data apiV1AllianceSelectionControlBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, 42, body.Data.AllianceSelection.TimeRemainingSec)
	assert.Equal(t, "allianceSelection", body.Data.DisplayMode)

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/alliance-selection", nil)
	require.NoError(t, err)
	defer conn.Close()
	for _, expected := range []string{"allianceSelection", "audienceDisplayMode"} {
		var message cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&message))
		assert.Equal(t, expected, message.Type)
	}
	var ready cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&ready))
	assert.Equal(t, "ready", ready.Type)
}

func TestAllianceSelectionClientSplitsStateAndCommands(t *testing.T) {
	contents, err := os.ReadFile("../static/js/alliance_selection.js")
	require.NoError(t, err)
	source := string(contents)
	assert.Contains(t, source, "new CheesyWebsocketV1")
	assert.Contains(t, source, "new CheesyWebsocket(\"/alliance_selection/websocket\", {})")
	assert.Contains(t, source, "/api/v1/admin/alliance-selection/bootstrap")
}

func TestApiV1MatchPlayBootstrapAndStream(t *testing.T) {
	web := setupTestWeb(t)
	team := model.Team{Id: 254, Nickname: "Poofs", WpaKey: "match-play-secret", FtaNotes: "private"}
	require.NoError(t, web.arena.Database.CreateTeam(&team))
	require.NoError(t, web.arena.SubstituteTeams(254, 0, 0, 0, 0, 0))
	web.arena.MatchLoadNotifier.Notify()
	web.arena.ArenaStatusNotifier.Notify()
	response := web.getHttpResponse("/api/v1/admin/match-play/bootstrap")
	require.Equal(t, http.StatusOK, response.Code)
	assert.NotContains(t, response.Body.String(), "match-play-secret")
	assert.NotContains(t, response.Body.String(), "private")
	var body struct {
		Data apiV1MatchPlayBootstrap `json:"data"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "/api/v1/streams/admin/match-play", body.Data.StreamUrl)
	require.NotNil(t, body.Data.Match.Teams["R1"])

	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/admin/match-play", nil)
	require.NoError(t, err)
	defer conn.Close()
	for _, expected := range []string{"match", "arenaStatus", "audienceDisplayMode", "allianceStationDisplayMode", "eventStatus", "realtimeScore", "postedScore", "scoringStatus", "timing", "matchClock"} {
		var message cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&message))
		assert.Equal(t, expected, message.Type)
	}
	var ready cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&ready))
	assert.Equal(t, "ready", ready.Type)
}

func TestMatchPlayClientSplitsStateAndCommands(t *testing.T) {
	contents, err := os.ReadFile("../static/js/match_play.js")
	require.NoError(t, err)
	source := string(contents)
	assert.Contains(t, source, "new CheesyWebsocketV1")
	assert.Contains(t, source, "new CheesyWebsocket(\"/match_play/websocket\", {})")
	assert.Contains(t, source, "/api/v1/admin/match-play/bootstrap")
}

func TestApiV1QueueingStreamBootstrapReadyAndUpdate(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/displays/queueing?displayId=1", nil)
	require.NoError(t, err)
	defer conn.Close()

	for _, expectedType := range []string{"displayConfiguration", "matches", "timing", "matchClock", "eventStatus"} {
		var message cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&message))
		assert.Equal(t, expectedType, message.Type)
		assert.True(t, message.Meta.Bootstrap)
		assert.Equal(t, cawebsocket.V1ProtocolVersion, message.Meta.Version)
	}
	var ready cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&ready))
	assert.Equal(t, "ready", ready.Type)

	web.arena.EventStatus.EarlyLateMessage = "Event is on schedule"
	web.arena.EventStatusNotifier.Notify()
	var update cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&update))
	assert.Equal(t, "eventStatus", update.Type)
	assert.False(t, update.Meta.Bootstrap)
	assert.Equal(t, uint64(1), update.Meta.Sequence)
}

func TestApiV1AnnouncerStreamBootstrapContract(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/displays/announcer?displayId=2", nil)
	require.NoError(t, err)
	defer conn.Close()

	for _, expectedType := range []string{
		"displayConfiguration", "match", "postedScore", "realtimeScore", "timing", "matchClock",
		"eventStatus", "audienceDisplayMode",
	} {
		var message cawebsocket.V1Message
		require.NoError(t, conn.ReadJSON(&message))
		assert.Equal(t, expectedType, message.Type)
		assert.True(t, message.Meta.Bootstrap)
	}
	var ready cawebsocket.V1Message
	require.NoError(t, conn.ReadJSON(&ready))
	assert.Equal(t, "ready", ready.Type)
	assert.Equal(t, 1, web.arena.Displays["2"].ConnectionCount)
}
