package game

type Score struct {
	LeaveStatuses   [3]bool
	CoralAlgae      CoralAlgae
	EndgameStatuses [3]EndgameStatus
	Fouls           []Foul
	PlayoffDq       bool
}

// Game-specific constants that cannot be changed by the user.
const (
	coralBonusCountThreshold = 5
	autoBonusPointThreshold  = 1
	autoBonusRobotThreshold  = 3
)

// Game-specific settings that can be changed by the user.
var CoralBonusLevelThresholdWithoutCoop = 4
var CoralBonusLevelThresholdWithCoop = 15
var BargeBonusPointThreshold = 14

// Represents the state of a robot at the end of the match.
type EndgameStatus int

const (
	EndgameNone EndgameStatus = iota
	EndgameParked
	EndgameShallowCage
	EndgameDeepCage
)

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentScore *Score) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the alliance was disqualified.
	if score.PlayoffDq {
		return summary
	}

	// Calculate autonomous period points.
	for _, status := range score.LeaveStatuses {
		if status {
			summary.LeavePoints += 2
		}
	}
	autoCoralPoints := score.CoralAlgae.AutoCoralPoints()
	autoAlgaePoints := score.CoralAlgae.AutoAlgaePoints()
	summary.AutoPoints = summary.LeavePoints + autoCoralPoints + autoAlgaePoints

	// Calculate Amp and Speaker points.
	summary.CoralPoints = score.CoralAlgae.AutoCoralPoints() + score.CoralAlgae.TeleopCoralPoints()
	summary.AlgaePoints = score.CoralAlgae.AutoAlgaePoints() + score.CoralAlgae.TeleopAlgaePoints()

	// Calculate endgame points.
	for _, status := range score.EndgameStatuses {
		switch status {
		case EndgameParked:
			summary.BargePoints += 2
		case EndgameShallowCage:
			summary.BargePoints += 6
		case EndgameDeepCage:
			summary.BargePoints += 12
		default:
		}
	}

	summary.MatchPoints = summary.LeavePoints + summary.CoralPoints + summary.AlgaePoints + summary.BargePoints

	// Calculate penalty points.
	for _, foul := range opponentScore.Fouls {
		summary.FoulPoints += foul.PointValue()
		// Store the number of tech fouls since it is used to break ties in playoffs.
		if foul.IsMajor {
			summary.NumOpponentMajorFouls++
		}

		rule := foul.Rule()
		if rule != nil {
			// Check for the opponent fouls that automatically trigger a ranking point.
			if rule.IsRankingPoint {
				if rule.RuleNumber == "G206" {
					summary.AutoBonusRankingPoint = false
					summary.BargeBonusRankingPoint = false
				} else if rule.RuleNumber == "G410" {
					summary.CoralBonusRankingPoint = true
				} else if rule.RuleNumber == "G428" {
					summary.BargeBonusRankingPoint = true
				}
			}
		}
	}

	summary.Score = summary.MatchPoints + summary.FoulPoints

	// Calculate bonus ranking points.
	coralNumEachLevel := score.CoralAlgae.NumScoredCoralEachRow()
	summary.CoralLevelMet = 0
	for _, numCoral := range coralNumEachLevel {
		if numCoral >= coralBonusCountThreshold {
			summary.CoralLevelMet++
		}
	}
	summary.CoralLevelGoal = CoralBonusLevelThresholdWithoutCoop
	if CoralBonusLevelThresholdWithCoop > 0 {
		// A MelodyBonusThresholdWithCoop of 0 disables the coopertition bonus.
		summary.CoopertitionCriteriaMet = score.CoralAlgae.IsCoopertitionThresholdAchieved()
		summary.CoopertitionBonus = summary.CoopertitionCriteriaMet && opponentScore.CoralAlgae.IsCoopertitionThresholdAchieved()
		if summary.CoopertitionBonus {
			summary.CoralLevelGoal = CoralBonusLevelThresholdWithCoop
		}
	}

	leaveCount := 0
	for _, status := range score.LeaveStatuses {
		if status {
			leaveCount++
		}
	}
	if score.CoralAlgae.AutoCoralPoints() > autoBonusPointThreshold && leaveCount >= autoBonusRobotThreshold {
		summary.AutoBonusRankingPoint = true
	}

	if summary.CoralLevelMet >= summary.CoralLevelGoal {
		summary.CoralBonusRankingPoint = true
	}
	if summary.BargePoints >= BargeBonusPointThreshold {
		summary.BargeBonusRankingPoint = true
	}

	if summary.AutoBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.CoralBonusRankingPoint {
		summary.BonusRankingPoints++
	}
	if summary.BargeBonusRankingPoint {
		summary.BonusRankingPoints++
	}

	return summary
}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.LeaveStatuses != other.LeaveStatuses ||
		score.CoralAlgae != other.CoralAlgae ||
		score.EndgameStatuses != other.EndgameStatuses ||
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
