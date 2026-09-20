// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/mitchellh/mapstructure"
	"strconv"
)

var errInvalidRefereeCommand = errors.New("invalid referee command")
var errInvalidRefereeValue = errors.New("invalid referee command value")
var errRefereeScoreCommitted = errors.New("fouls already committed")

type refereeCommandArgs struct {
	Alliance string
	IsMajor  bool
	Index    int
	TeamId   int
	RuleId   int
	Card     string
}

type refereeCommandResult struct {
	Changed bool `json:"changed"`
	FoulId  int  `json:"foulId,omitempty"`
}

// executeRefereeScoreCommand is the single validation and mutation path for
// referee and scoring transports. The caller authenticates and audits.
func (web *Web) executeRefereeScoreCommand(command string, data any) (refereeCommandResult, error) {
	var result refereeCommandResult
	switch command {
	case "addFoul", "toggleFoulType", "updateFoulTeam", "updateFoulRule", "deleteFoul", "card":
	default:
		return result, errInvalidRefereeCommand
	}
	var args refereeCommandArgs
	if err := mapstructure.Decode(data, &args); err != nil || (args.Alliance != "red" && args.Alliance != "blue") {
		return result, errInvalidRefereeValue
	}
	if args.TeamId < 0 || args.RuleId < 0 {
		return result, errInvalidRefereeValue
	}
	if command == "card" && (args.Card != "" && args.Card != "yellow" && args.Card != "red" || args.TeamId <= 0) {
		return result, errInvalidRefereeValue
	}
	if command == "updateFoulRule" && args.RuleId != 0 && game.GetRuleById(args.RuleId) == nil {
		return result, errInvalidRefereeValue
	}

	web.refereeCommandMu.Lock()
	defer web.refereeCommandMu.Unlock()
	if web.arena.MatchState == field.TimeoutActive || web.arena.MatchState == field.PostTimeout {
		return result, errInvalidRefereeValue
	}
	var score *field.RealtimeScore
	if args.Alliance == "red" {
		score = web.arena.RedRealtimeScore
	} else {
		score = web.arena.BlueRealtimeScore
	}
	if score.FoulsCommitted {
		return result, errRefereeScoreCommitted
	}
	if command == "card" {
		if web.arena.CurrentMatch == nil {
			return result, errInvalidRefereeValue
		}
		cards := score.Cards
		if web.arena.CurrentMatch.Type == model.Playoff {
			var teams [3]int
			if args.Alliance == "red" {
				teams = [3]int{web.arena.CurrentMatch.Red1, web.arena.CurrentMatch.Red2, web.arena.CurrentMatch.Red3}
			} else {
				teams = [3]int{web.arena.CurrentMatch.Blue1, web.arena.CurrentMatch.Blue2, web.arena.CurrentMatch.Blue3}
			}
			for _, team := range teams {
				if team > 0 {
					cards[strconv.Itoa(team)] = args.Card
				}
			}
		} else {
			cards[strconv.Itoa(args.TeamId)] = args.Card
		}
		result.Changed = true
	} else if command == "addFoul" {
		foul := game.Foul{FoulId: web.arena.NextFoulId, IsMajor: args.IsMajor}
		web.arena.NextFoulId++
		score.CurrentScore.Fouls = append(score.CurrentScore.Fouls, foul)
		result.Changed, result.FoulId = true, foul.FoulId
	} else {
		fouls := &score.CurrentScore.Fouls
		if args.Index < 0 || args.Index >= len(*fouls) {
			return result, nil // Preserve the legacy no-op for stale row indices.
		}
		switch command {
		case "toggleFoulType":
			(*fouls)[args.Index].IsMajor = !(*fouls)[args.Index].IsMajor
			(*fouls)[args.Index].RuleId = 0
		case "updateFoulTeam":
			if (*fouls)[args.Index].TeamId == args.TeamId {
				(*fouls)[args.Index].TeamId = 0
			} else {
				(*fouls)[args.Index].TeamId = args.TeamId
			}
		case "updateFoulRule":
			(*fouls)[args.Index].RuleId = args.RuleId
		case "deleteFoul":
			*fouls = append((*fouls)[:args.Index], (*fouls)[args.Index+1:]...)
		}
		result.Changed = true
	}
	if result.Changed {
		web.arena.RealtimeScoreNotifier.Notify()
	}
	return result, nil
}
