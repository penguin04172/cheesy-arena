// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/led"
	"github.com/mitchellh/mapstructure"
)

var errInvalidFieldTestingCommand = errors.New("invalid field testing command")
var errInvalidFieldTestingValue = errors.New("invalid field testing command value")
var errFieldTestingCoilState = errors.New(fieldTestingOverrideDisabledMessage)
var errFieldTestingLedState = errors.New(fieldTestingLedModeDisabledMessage)

type fieldTestingCoilInput struct {
	Index    int
	Override string
}

type fieldTestingLedInput struct {
	RedMode  led.Mode
	BlueMode led.Mode
}

func (web *Web) executeFieldTestingCommand(command string, data any) error {
	web.fieldTestingCommandMu.Lock()
	defer web.fieldTestingCommandMu.Unlock()
	switch command {
	case "playSound":
		sound, ok := data.(string)
		if !ok {
			return fmt.Errorf("%w: Failed to parse '%s' message.", errInvalidFieldTestingValue, command)
		}
		valid := false
		for _, matchSound := range game.UniqueMatchSounds() {
			if matchSound.Name == sound {
				valid = true
				break
			}
		}
		if !valid {
			return errInvalidFieldTestingValue
		}
		web.arena.PlaySoundNotifier.NotifyWithMessage(sound)
	case "setPlcCoilOverride":
		var input fieldTestingCoilInput
		if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidFieldTestingValue
		}
		if !fieldTestingOverridesAllowed(web.arena.MatchState) {
			return errFieldTestingCoilState
		}
		if input.Index < 0 || input.Index >= len(web.arena.Plc.GetCoilNames()) {
			return errInvalidFieldTestingValue
		}
		switch input.Override {
		case "auto":
			web.arena.Plc.ClearCoilOverride(input.Index)
		case "on":
			web.arena.Plc.SetCoilOverride(input.Index, true)
		case "off":
			web.arena.Plc.SetCoilOverride(input.Index, false)
		default:
			return fmt.Errorf("Invalid coil override state '%s'.", input.Override)
		}
		web.arena.Plc.IoChangeNotifier().Notify()
	case "setLedMode":
		var input fieldTestingLedInput
		if err := mapstructure.Decode(data, &input); err != nil {
			return errInvalidFieldTestingValue
		}
		if !fieldTestingOverridesAllowed(web.arena.MatchState) {
			return errFieldTestingLedState
		}
		if _, ok := led.ModeNames[input.RedMode]; !ok {
			return fmt.Errorf("Invalid LED mode '%d'.", input.RedMode)
		}
		if _, ok := led.ModeNames[input.BlueMode]; !ok {
			return fmt.Errorf("Invalid LED mode '%d'.", input.BlueMode)
		}
		web.arena.Leds.SetMode(input.RedMode, input.BlueMode)
	default:
		return errInvalidFieldTestingCommand
	}
	return nil
}
