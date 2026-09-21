// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/mitchellh/mapstructure"
)

var errInvalidMatchLifecycleCommand = errors.New("invalid match lifecycle command")
var errInvalidMatchLifecycleValue = errors.New("invalid match lifecycle value")

type matchLifecycleMatchInput struct {
	MatchId int
}

type matchLifecycleSubstitutionInput struct {
	Red1  int
	Red2  int
	Red3  int
	Blue1 int
	Blue2 int
	Blue3 int
}

type matchLifecycleStartInput struct {
	MuteMatchSounds bool
}

// executeMatchPlayLifecycleCommand centralizes the existing high-impact
// transitions. They remain legacy-only until durable post/retry semantics are
// established for API transports.
func (web *Web) executeMatchPlayLifecycleCommand(command string, data any, fromReferee bool) error {
	web.matchPlayLifecycleMu.Lock()
	defer web.matchPlayLifecycleMu.Unlock()
	switch command {
	case "loadMatch":
		var input matchLifecycleMatchInput
		if err := mapstructure.Decode(data, &input); err != nil || input.MatchId < 0 {
			return errInvalidMatchLifecycleValue
		}
		var match *model.Match
		if input.MatchId != 0 {
			var err error
			match, err = web.arena.Database.GetMatchById(input.MatchId)
			if err != nil {
				return err
			}
			if match == nil {
				return fmt.Errorf("invalid match ID %d", input.MatchId)
			}
		}
		if err := web.arena.ResetMatch(); err != nil {
			return err
		}
		if match == nil {
			return web.arena.LoadTestMatch()
		}
		return web.arena.LoadMatch(match)
	case "showResult":
		var input matchLifecycleMatchInput
		if err := mapstructure.Decode(data, &input); err != nil || input.MatchId < 0 {
			return errInvalidMatchLifecycleValue
		}
		if input.MatchId == 0 {
			web.arena.SavedMatch = &model.Match{}
			web.arena.SavedMatchResult = model.NewMatchResult()
			web.arena.ScorePostedNotifier.Notify()
			return nil
		}
		match, err := web.arena.Database.GetMatchById(input.MatchId)
		if err != nil {
			return err
		}
		if match == nil {
			return fmt.Errorf("invalid match ID %d", input.MatchId)
		}
		result, err := web.arena.Database.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		if result == nil {
			return fmt.Errorf("No result found for match ID %d.", input.MatchId)
		}
		rankings := game.Rankings{}
		if match.ShouldUpdateRankings() {
			rankings, err = web.arena.Database.GetAllRankings()
			if err != nil {
				return err
			}
		}
		web.arena.SavedRankings = rankings
		web.arena.SavedMatch = match
		web.arena.SavedMatchResult = result
		web.arena.ScorePostedNotifier.Notify()
	case "substituteTeams":
		var input matchLifecycleSubstitutionInput
		if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidMatchLifecycleValue
		}
		return web.arena.SubstituteTeams(input.Red1, input.Red2, input.Red3, input.Blue1, input.Blue2, input.Blue3)
	case "startMatch":
		var input matchLifecycleStartInput
		if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidMatchLifecycleValue
		}
		previousMute := web.arena.MuteMatchSounds
		web.arena.MuteMatchSounds = input.MuteMatchSounds
		if err := web.arena.StartMatch(); err != nil {
			web.arena.MuteMatchSounds = previousMute
			return err
		}
	case "abortMatch":
		return web.arena.AbortMatch()
	case "discardResults":
		if err := web.arena.ResetMatch(); err != nil {
			return err
		}
		return web.arena.LoadNextMatch(false)
	case "commitAndPost":
		if fromReferee {
			if web.arena.MatchState != field.PostMatch {
				return fmt.Errorf("cannot commit match while it is in progress")
			}
			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.ScoringStatusNotifier.Notify()
		}
		return web.commitPostAndLoadNextMatch()
	default:
		return errInvalidMatchLifecycleCommand
	}
	return nil
}
