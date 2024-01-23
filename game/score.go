// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	LeaveStatuses                     [3]bool
	AutoNoteAmp                       int
	AutoNoteSpeaker                   int
	TeleopNoteAmp                     int
	TeleopNoteSpeaker                 int
	TeleopNoteAmplifiedSpeaker        int
	AccumulateNote                    int
	TrapStatuses                      [3]bool
	EndgameStatuses                   [3]EndgameStatus
	EndgameHarmony                    bool
	Coopertition                      bool
	CoopertitionActive                bool
	Amplification                     bool
	AmplifiedNoteCount                int
	AmplificationRemainingDurationSec float64
	AmplificationStartedTimeSec       float64
	Fouls                             []Foul
	PlayoffDq                         bool
}

var CoopertitionActiveDurationSec = 45
var AmplificationDurationSec = 13
var AmplificationNoteThreshold = 4
var MelodyBonusThresholdWithoutCoop = 18
var MelodyBonusThresholdWithCoop = 15
var EnsembleBonusPointThreshold = 10
var EnsembleBonusOnstageRobotThreshold = 2

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameParked
	EndgameOnstage
	EndgameOnstageWithSpotlit
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, leave := range score.LeaveStatuses {
		if leave {
			summary.LeavePoints += 2
		}
	}

	autoNotePoints := score.AutoNoteAmp*2 + score.AutoNoteSpeaker*5
	summary.AutoPoints = summary.LeavePoints + autoNotePoints

	// Calculate teleoperated period points.

	teleopNotePoints := score.TeleopNoteAmp + score.TeleopNoteSpeaker*2 + score.TeleopNoteAmplifiedSpeaker*5
	for i := 0; i < 3; i++ {
		switch score.EndgameStatuses[i] {
		case EndgameParked:
			summary.StagePoints += 1
		case EndgameOnstage:
			summary.StagePoints += 3
			summary.NumOnstages += 1
		case EndgameOnstageWithSpotlit:
			summary.StagePoints += 4
			summary.NumOnstages += 1
		}

		if score.TrapStatuses[i] {
			summary.TrapPoints += 5
			summary.StagePoints += 5
		}
	}

	if score.EndgameHarmony {
		summary.StagePoints += 2
	}

	summary.NotePoints = autoNotePoints + teleopNotePoints
	summary.EndgamePoints = summary.StagePoints
	summary.MatchPoints = summary.LeavePoints + summary.NotePoints + summary.StagePoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of tech fouls since it is used to break ties in playoffs.
		if foul.IsTechnical {
			summary.NumOpponentTechFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				summary.EnsembleBonusRankingPoint = true
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	summary.CoopertitionBonus = score.Coopertition && opponentScore.Coopertition
	summary.NumNotes = score.AutoNoteAmp + score.AutoNoteSpeaker + score.TeleopNoteAmp + score.TeleopNoteSpeaker + score.TeleopNoteAmplifiedSpeaker
	summary.NumNotesGoal = MelodyBonusThresholdWithoutCoop
	// A SustainabilityBonusLinkThresholdWithCoop of 0 disables the coopertition bonus.
	if MelodyBonusThresholdWithCoop > 0 && summary.CoopertitionBonus {
		summary.NumNotesGoal = MelodyBonusThresholdWithCoop
	}
	if summary.NumNotes >= summary.NumNotesGoal {
		summary.MelodyBonusRankingPoint = true
	}

	summary.NumOnstagesGoal = EnsembleBonusOnstageRobotThreshold
	summary.EnsembleBonusRankingPoint = summary.EndgamePoints >= EnsembleBonusPointThreshold && summary.NumOnstages >= summary.NumOnstagesGoal

	if summary.MelodyBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.EnsembleBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.LeaveStatuses != other.LeaveStatuses ||
		score.AutoNoteAmp != other.AutoNoteAmp ||
		score.AutoNoteSpeaker != other.AutoNoteSpeaker ||
		score.TeleopNoteAmp != other.TeleopNoteAmp ||
		score.TeleopNoteSpeaker != other.TeleopNoteSpeaker ||
		score.TeleopNoteAmplifiedSpeaker != other.TeleopNoteAmplifiedSpeaker ||
		score.TrapStatuses != other.TrapStatuses ||
		score.EndgameStatuses != other.EndgameStatuses ||
		score.EndgameHarmony != other.EndgameHarmony ||
		score.Coopertition != other.Coopertition ||
		score.PlayoffDq != other.PlayoffDq ||
		len(score.Fouls) != len(other.Fouls) {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
