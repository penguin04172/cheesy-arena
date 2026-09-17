// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
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
	for _, path := range []string{"../static/js/audience_display.js", "../static/js/alliance_station_display.js"} {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		source := string(contents)
		assert.Contains(t, source, "new CheesyWebsocketV1")
		assert.Contains(t, source, "/api/v1/displays/")
		assert.NotContains(t, source, "new CheesyWebsocket(\"")
	}
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
