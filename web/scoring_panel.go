// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

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
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		Alliance     string
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), alliance}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}
		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else {
			args := struct {
				Target int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			switch command {
			case "leaveStatus":
				if args.Target >= 1 && args.Target <= 3 {
					score.LeaveStatuses[args.Target-1] = !score.LeaveStatuses[args.Target-1]
					scoreChanged = true
				}
			case "trapStatus":
				if args.Target >= 1 && args.Target <= 3 {
					score.TrapStatuses[args.Target-1] = !score.TrapStatuses[args.Target-1]
					scoreChanged = true
				}
			case "plus":
				if args.Target >= 1 && args.Target <= 5 {
					switch args.Target {
					case 1:
						incrementGoal(&score.AutoNoteAmp)
						if !score.Amplification {
							incrementGoal(&score.AccumulateNote)
						}
					case 2:
						incrementGoal(&score.TeleopNoteAmp)
						if !score.Amplification {
							incrementGoal(&score.AccumulateNote)
						}
					case 3:
						incrementGoal(&score.AutoNoteSpeaker)
					case 4:
						incrementGoal(&score.TeleopNoteSpeaker)
					case 5:
						if score.Amplification {
							incrementGoal(&score.TeleopNoteAmplifiedSpeaker)
							decrementGoal(&score.AmplificationRemainingNote)
							if score.AmplificationRemainingNote == 0 {
								score.Amplification = false
								score.AmplificationRemainingDurationSec = 0
								score.AccumulateNote = 0
							}
						}
					}
					scoreChanged = true
				}
			case "minus":
				if args.Target >= 1 && args.Target <= 5 {
					switch args.Target {
					case 1:
						decrementGoal(&score.AutoNoteAmp)
						if !score.Amplification {
							decrementGoal(&score.AccumulateNote)
						}
					case 2:
						decrementGoal(&score.TeleopNoteAmp)
						if !score.Amplification {
							decrementGoal(&score.AccumulateNote)
						}
					case 3:
						decrementGoal(&score.AutoNoteSpeaker)
					case 4:
						decrementGoal(&score.TeleopNoteSpeaker)
					case 5:
						decrementGoal(&score.TeleopNoteAmplifiedSpeaker)
						if score.TeleopNoteAmplifiedSpeaker > 0 {
							incrementGoal(&score.AmplificationRemainingNote)
						}
					}
					scoreChanged = true
				}
			case "endgameStatus":
				if args.Target >= 1 && args.Target <= 3 {
					score.EndgameStatuses[args.Target-1]++
					if score.EndgameStatuses[args.Target-1] > 3 {
						score.EndgameStatuses[args.Target-1] = 0
					}
					scoreChanged = true
				}
			case "coopertition":
				if score.AccumulateNote >= 1 {
					score.CoopertitionActive = false
					score.Coopertition = true
					decrementGoal(&score.AccumulateNote)
					scoreChanged = true
				}
			case "amplification":
				if score.AccumulateNote >= 2 {
					score.Amplification = true
					score.AmplificationStartedTimeSec = web.arena.MatchTimeSec()
					score.AmplificationRemainingDurationSec = float64(game.AmplificationDurationSec)
					score.AmplificationRemainingNote = game.AmplificationNoteThreshold
					score.AccumulateNote = 0
					scoreChanged = true
				}
			case "endgameHarmony":
				score.EndgameHarmony = !score.EndgameHarmony
				scoreChanged = true
			}
			if scoreChanged {
				web.arena.RealtimeScoreNotifier.Notify()
			}
		}
	}
}

// Increments the cargo count for the given goal.
func incrementGoal(goal *int) bool {
	// Use just the first hub quadrant for manual scoring.
	*goal++
	return true
}

// Decrements the cargo for the given goal.
func decrementGoal(goal *int) bool {
	// Use just the first hub quadrant for manual scoring.
	if *goal > 0 {
		*goal--
		return true
	}
	return false
}
