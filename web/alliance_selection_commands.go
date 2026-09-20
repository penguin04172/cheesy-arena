// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"errors"
	"github.com/Team254/cheesy-arena/field"
	"sync"
	"time"
)

var errInvalidAllianceSelectionCommand = errors.New("invalid alliance selection command")
var errInvalidAllianceSelectionValue = errors.New("invalid alliance selection command value")

// All timer transitions, including ticks, are serialized across command transports.
type allianceSelectionCommandState struct {
	mu                  sync.Mutex
	timeLimitSec        int
	currentTimeLimitSec int
	stopTicking         chan struct{}
}

func newAllianceSelectionCommandState() *allianceSelectionCommandState {
	return &allianceSelectionCommandState{timeLimitSec: 45}
}

func (state *allianceSelectionCommandState) timeLimit() int {
	state.mu.Lock()
	defer state.mu.Unlock()
	return state.timeLimitSec
}

func (state *allianceSelectionCommandState) stopLocked() {
	if state.stopTicking != nil {
		close(state.stopTicking)
		state.stopTicking = nil
	}
}

func (state *allianceSelectionCommandState) startLocked(arena *field.Arena) {
	state.stopLocked()
	if arena.AllianceSelectionTimeRemainingSec <= 0 {
		return
	}
	done := make(chan struct{})
	state.stopTicking = done
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				state.mu.Lock()
				if state.stopTicking != done {
					state.mu.Unlock()
					return
				}
				arena.AllianceSelectionTimeRemainingSec--
				remaining := arena.AllianceSelectionTimeRemainingSec
				arena.AllianceSelectionNotifier.Notify()
				if remaining <= 0 {
					state.stopLocked()
				}
				playSound := state.currentTimeLimitSec != allianceSelectionBreakDurationSec
				state.mu.Unlock()
				if playSound && remaining == 5 {
					arena.PlaySound("pick_clock")
				} else if playSound && remaining == 0 {
					arena.PlaySound("pick_clock_expired")
				}
				if remaining <= 0 {
					return
				}
			}
		}
	}()
}

// executeAllianceSelectionCommand is shared by the legacy socket and JSON API.
// Authentication and auditing are transport responsibilities.
func (web *Web) executeAllianceSelectionCommand(command string, value any) error {
	state := web.allianceSelectionCommands
	state.mu.Lock()
	defer state.mu.Unlock()
	switch command {
	case "setTimer":
		seconds, ok := value.(float64)
		if !ok || seconds < 0 || seconds > 3600 || seconds != float64(int(seconds)) {
			return errInvalidAllianceSelectionValue
		}
		state.timeLimitSec = int(seconds)
	case "startTimer":
		if value != nil {
			return errInvalidAllianceSelectionValue
		}
		if state.stopTicking != nil {
			return nil
		}
		if web.arena.AllianceSelectionTimeRemainingSec <= 0 {
			web.arena.AllianceSelectionTimeRemainingSec = state.timeLimitSec
			state.currentTimeLimitSec = state.timeLimitSec
		}
		web.arena.AllianceSelectionShowTimer = true
		web.arena.AllianceSelectionNotifier.Notify()
		state.startLocked(web.arena)
	case "stopTimer":
		if value != nil {
			return errInvalidAllianceSelectionValue
		}
		state.stopLocked()
		web.arena.AllianceSelectionNotifier.Notify()
	case "restartTimer":
		if value != nil {
			return errInvalidAllianceSelectionValue
		}
		state.stopLocked()
		web.arena.AllianceSelectionShowTimer = true
		web.arena.AllianceSelectionTimeRemainingSec = state.timeLimitSec
		state.currentTimeLimitSec = state.timeLimitSec
		web.arena.AllianceSelectionNotifier.Notify()
	case "hideTimer":
		if value != nil {
			return errInvalidAllianceSelectionValue
		}
		state.stopLocked()
		web.arena.AllianceSelectionShowTimer = false
		web.arena.AllianceSelectionTimeRemainingSec = 0
		web.arena.AllianceSelectionNotifier.Notify()
	case "setAudienceDisplay":
		mode, ok := value.(string)
		if !ok || !validAudienceDisplayMode(mode) {
			return errInvalidAllianceSelectionValue
		}
		web.arena.SetAudienceDisplayMode(mode)
	default:
		return errInvalidAllianceSelectionCommand
	}
	return nil
}

func validAudienceDisplayMode(mode string) bool {
	switch mode {
	case "blank", "intro", "match", "score", "bracket", "logo", "logoLuma", "sponsor", "allianceSelection", "timeout":
		return true
	default:
		return false
	}
}
