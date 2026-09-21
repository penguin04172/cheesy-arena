// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/mitchellh/mapstructure"
)

var errInvalidScoringTowerCommand = errors.New("invalid scoring tower command")
var errInvalidScoringTowerValue = errors.New("invalid scoring tower value")
var errScoringTowerCommitted = errors.New("scoring panel already committed")

type scoringTowerInput struct {
	TeamPosition       int
	AutoTowerStatus    int
	EndgameTowerStatus int
}

// executeScoringTowerCommand is shared by the legacy scoring socket and API
// transports. A stale or invalid command cannot mutate a score.
func (web *Web) executeScoringTowerCommand(position, command string, data any) (bool, error) {
	if position != "red" && position != "blue" {
		return false, errInvalidScoringTowerValue
	}
	if command != "autoTower" && command != "endgame" {
		return false, errInvalidScoringTowerCommand
	}
	var input scoringTowerInput
	if err := mapstructure.Decode(data, &input); err != nil || input.TeamPosition < 1 || input.TeamPosition > 3 {
		return false, errInvalidScoringTowerValue
	}
	status := input.AutoTowerStatus
	if command == "endgame" {
		status = input.EndgameTowerStatus
	}
	if status < 0 || status > 3 {
		return false, errInvalidScoringTowerValue
	}
	web.refereeCommandMu.Lock()
	defer web.refereeCommandMu.Unlock()
	if web.arena.MatchState == field.TimeoutActive || web.arena.MatchState == field.PostTimeout {
		return false, errInvalidScoringTowerValue
	}
	if web.arena.ScoringPanelRegistry.GetNumScoreCommitted(position) > 0 {
		return false, errScoringTowerCommitted
	}
	var score *game.Score
	if position == "red" {
		score = &web.arena.RedRealtimeScore.CurrentScore
	} else {
		score = &web.arena.BlueRealtimeScore.CurrentScore
	}
	towerStatus := game.TowerStatus(status)
	if command == "autoTower" {
		if score.AutoTowerStatuses[input.TeamPosition-1] == towerStatus {
			web.arena.RealtimeScoreNotifier.Notify()
			return false, nil
		}
		score.AutoTowerStatuses[input.TeamPosition-1] = towerStatus
	} else {
		if score.EndgameTowerStatuses[input.TeamPosition-1] == towerStatus {
			web.arena.RealtimeScoreNotifier.Notify()
			return false, nil
		}
		score.EndgameTowerStatuses[input.TeamPosition-1] = towerStatus
	}
	web.arena.RealtimeScoreNotifier.Notify()
	return true, nil
}
