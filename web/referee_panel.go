// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for the referee interface.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
)

// Renders the referee interface for assigning fouls.
func (web *Web) refereePanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/referee_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Renders a partial template for when the foul list is updated.
func (web *Web) refereePanelFoulListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Link", "</api/v1/admin/referee/fouls>; rel=\"successor-version\"")
	template, err := web.parseFiles("templates/referee_panel_foul_list.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		Match     *model.Match
		RedFouls  []game.Foul
		BlueFouls []game.Foul
		Rules     map[int]*game.Rule
	}{
		web.arena.CurrentMatch,
		web.arena.RedRealtimeScore.CurrentScore.Fouls,
		web.arena.BlueRealtimeScore.CurrentScore.Fouls,
		game.GetAllRules(),
	}
	err = template.ExecuteTemplate(w, "referee_panel_foul_list", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the refereee interface client to send control commands and receive status updates.
func (web *Web) refereePanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer closeWebsocket(ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ScoringStatusNotifier,
		web.arena.ReloadDisplaysNotifier,
		web.arena.ArenaStatusNotifier,
	)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "addFoul", "toggleFoulType", "updateFoulTeam", "updateFoulRule", "deleteFoul", "card":
			if _, err := web.executeRefereeScoreCommand(messageType, data); err != nil {
				writeWebsocketError(ws, err.Error())
			}
		case "toggleBypass":
			if err := web.executeMatchPlayControlCommand(messageType, data); err != nil {
				writeWebsocketError(ws, err.Error())
			}
		case "signalVolunteers", "signalReset":
			if err := web.executeMatchPlayControlCommand(messageType, data); err != nil {
				writeWebsocketError(ws, err.Error())
			}
		case "commitAndPost":
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the fouls until the match is over.
				continue
			}
			err = web.executeMatchPlayLifecycleCommand(messageType, data, true)
			if err != nil {
				writeWebsocketError(ws, err.Error())
				continue
			}
		default:
			writeWebsocketError(ws, fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}
