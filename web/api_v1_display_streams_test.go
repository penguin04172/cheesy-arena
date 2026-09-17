// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"net/http"
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

func TestApiV1QueueingStreamBootstrapReadyAndUpdate(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/api/v1/streams/displays/queueing", nil)
	require.NoError(t, err)
	defer conn.Close()

	for _, expectedType := range []string{"matches", "matchClock", "timing", "eventStatus"} {
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
