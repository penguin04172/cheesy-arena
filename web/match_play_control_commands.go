// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/model"
	"github.com/mitchellh/mapstructure"
)

var errInvalidMatchPlayControlCommand = errors.New("invalid match play control command")
var errInvalidMatchPlayControlValue = errors.New("invalid match play control value")
var errInvalidMatchPlayControlState = errors.New("command not allowed in current match state")

type matchPlayTimeoutInput struct {
	Description   string
	NextMatchName string
	DurationSec   float64
}

// executeMatchPlayControlCommand serializes non-scoring field controls shared
// by Match Play and Referee transports. Transport layers own auth and audit.
func (web *Web) executeMatchPlayControlCommand(command string, data any) error {
	web.matchPlayControlMu.Lock()
	defer web.matchPlayControlMu.Unlock()
	switch command {
	case "setAudienceDisplay":
		mode, ok := data.(string)
		if !ok || !validAudienceDisplayMode(mode) {
			return errInvalidMatchPlayControlValue
		}
		web.arena.SetAudienceDisplayMode(mode)
	case "setAllianceStationDisplay":
		mode, ok := data.(string)
		if !ok || !validAllianceStationDisplayMode(mode) {
			return errInvalidMatchPlayControlValue
		}
		web.arena.SetAllianceStationDisplayMode(mode)
	case "toggleBypass":
		station, ok := data.(string)
		if !ok {
			return fmt.Errorf("%w: Failed to parse '%s' message.", errInvalidMatchPlayControlValue, command)
		}
		if _, ok := web.arena.AllianceStations[station]; !ok {
			return fmt.Errorf("%w: Invalid alliance station", errInvalidMatchPlayControlValue)
		}
		return web.arena.ToggleBypass(station)
	case "signalVolunteers", "signalReset":
		if data != nil {
			return errInvalidMatchPlayControlValue
		}
		if web.arena.MatchState != field.PreMatch && web.arena.MatchState != field.PostMatch && web.arena.MatchState != field.TimeoutActive {
			return errInvalidMatchPlayControlState
		}
		if command == "signalVolunteers" {
			web.arena.SignalVolunteers()
		} else {
			web.arena.SignalReset()
		}
	case "startTimeout":
		if web.arena.MatchState != field.PreMatch {
			return errInvalidMatchPlayControlState
		}
		input := matchPlayTimeoutInput{}
		if seconds, ok := data.(float64); ok {
			input.DurationSec = seconds
		} else if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidMatchPlayControlValue
		}
		if input.DurationSec < 1 || input.DurationSec > 3600 || input.DurationSec != float64(int(input.DurationSec)) {
			return errInvalidMatchPlayControlValue
		}
		if input.Description == "" {
			input.Description = defaultTimeoutDescription
		}
		if len(input.Description) > 200 || len(input.NextMatchName) > 200 {
			return errInvalidMatchPlayControlValue
		}
		return web.arena.StartAdHocTimeout(input.Description, input.NextMatchName, int(input.DurationSec))
	case "setTimeoutDisplay":
		if data == nil {
			return errInvalidMatchPlayControlValue
		}
		input := matchPlayTimeoutInput{}
		if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidMatchPlayControlValue
		}
		if input.Description == "" {
			input.Description = defaultTimeoutDescription
		}
		if len(input.Description) > 200 || len(input.NextMatchName) > 200 {
			return errInvalidMatchPlayControlValue
		}
		web.arena.SetTimeoutDisplay(input.Description, input.NextMatchName)
	case "setTestMatchName":
		name, ok := data.(string)
		if !ok || len(name) > 200 {
			return errInvalidMatchPlayControlValue
		}
		if web.arena.CurrentMatch == nil || web.arena.CurrentMatch.Type != model.Test {
			return errInvalidMatchPlayControlState
		}
		web.arena.CurrentMatch.LongName = name
		web.arena.MatchLoadNotifier.Notify()
	default:
		return errInvalidMatchPlayControlCommand
	}
	return nil
}

func validAllianceStationDisplayMode(mode string) bool {
	switch mode {
	case "blank", "match", "logo", "timeout", "fieldReset", "signalCount":
		return true
	default:
		return false
	}
}
