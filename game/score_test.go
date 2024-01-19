// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScoreSummary(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redSummary := redScore.Summarize(blueScore)
	assert.Equal(t, 6, redSummary.LeavePoints)
	assert.Equal(t, 36, redSummary.AutoPoints)
	assert.Equal(t, 52, redSummary.NotePoints)
	assert.Equal(t, 18, redSummary.StagePoints)
	assert.Equal(t, 12, redSummary.EndgamePoints)
	assert.Equal(t, 78, redSummary.MatchPoints)
	assert.Equal(t, 0, redSummary.FoulPoints)
	assert.Equal(t, 78, redSummary.Score)
	assert.Equal(t, false, redSummary.CoopertitionBonus)
	assert.Equal(t, 0, redSummary.NumNotes)
	assert.Equal(t, 6, redSummary.NumNotesGoal)
	assert.Equal(t, false, redSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, redSummary.EnsembleBonusRankingPoint)
	assert.Equal(t, 0, redSummary.BonusRankingPoints)
	assert.Equal(t, 0, redSummary.NumOpponentTechFouls)

	blueSummary := blueScore.Summarize(redScore)
	assert.Equal(t, 3, blueSummary.LeavePoints)
	assert.Equal(t, 43, blueSummary.AutoPoints)
	assert.Equal(t, 154, blueSummary.NotePoints)
	assert.Equal(t, 30, blueSummary.StagePoints)
	assert.Equal(t, 18, blueSummary.EndgamePoints)
	assert.Equal(t, 187, blueSummary.MatchPoints)
	assert.Equal(t, 29, blueSummary.FoulPoints)
	assert.Equal(t, 216, blueSummary.Score)
	assert.Equal(t, false, blueSummary.CoopertitionBonus)
	assert.Equal(t, 9, blueSummary.NumNotes)
	assert.Equal(t, 6, blueSummary.NumNotes)
	assert.Equal(t, true, blueSummary.MelodyBonusRankingPoint)
	assert.Equal(t, true, blueSummary.EnsembleBonusRankingPoint)
	assert.Equal(t, 2, blueSummary.BonusRankingPoints)
	assert.Equal(t, 2, blueSummary.NumOpponentTechFouls)

	// Test that unsetting the team and rule ID don't invalidate the foul.
	redScore.Fouls[0].TeamId = 0
	redScore.Fouls[0].RuleId = 0
	assert.Equal(t, 29, blueScore.Summarize(redScore).FoulPoints)

	// Test playoff disqualification.
	redScore.PlayoffDq = true
	assert.Equal(t, 0, redScore.Summarize(blueScore).Score)
	assert.NotEqual(t, 0, blueScore.Summarize(blueScore).Score)
	blueScore.PlayoffDq = true
	assert.Equal(t, 0, blueScore.Summarize(redScore).Score)
}

func TestScoreMelodyBonusRankingPoint(t *testing.T) {
	redScore := TestScore1()
	blueScore := TestScore2()

	redScoreSummary := redScore.Summarize(blueScore)
	blueScoreSummary := blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumNotes)
	assert.Equal(t, 6, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 9, blueScoreSummary.NumNotes)
	assert.Equal(t, 6, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Reduce blue links to 8 and verify that the bonus is still awarded.
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumNotes)
	assert.Equal(t, 6, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 8, blueScoreSummary.NumNotes)
	assert.Equal(t, 6, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Increase non-coopertition threshold to 9.
	MelodyBonusThresholdWithoutCoop = 22
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumNotes)
	assert.Equal(t, 9, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 8, blueScoreSummary.NumNotes)
	assert.Equal(t, 9, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, false, blueScoreSummary.MelodyBonusRankingPoint)

	// Reduce blue links to 6 and verify that the sustainability bonus is not awarded.
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, false, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumNotes)
	assert.Equal(t, 9, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, false, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumNotes)
	assert.Equal(t, 9, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, false, blueScoreSummary.MelodyBonusRankingPoint)

	// Make red fulfill the coopertition bonus requirement.
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 0, redScoreSummary.NumNotes)
	assert.Equal(t, 5, redScoreSummary.NumNotesGoal)
	assert.Equal(t, false, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumNotes)
	assert.Equal(t, 5, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)

	// Reduce coopertition threshold to 1 and make red fulfill the sustainability bonus requirement.
	MelodyBonusThresholdWithCoop = 12
	redScoreSummary = redScore.Summarize(blueScore)
	blueScoreSummary = blueScore.Summarize(redScore)
	assert.Equal(t, true, redScoreSummary.CoopertitionBonus)
	assert.Equal(t, 1, redScoreSummary.NumNotes)
	assert.Equal(t, 1, redScoreSummary.NumNotesGoal)
	assert.Equal(t, true, redScoreSummary.MelodyBonusRankingPoint)
	assert.Equal(t, true, blueScoreSummary.CoopertitionBonus)
	assert.Equal(t, 6, blueScoreSummary.NumNotes)
	assert.Equal(t, 1, blueScoreSummary.NumNotesGoal)
	assert.Equal(t, true, blueScoreSummary.MelodyBonusRankingPoint)
}

func TestScoreEnsembleBonusRankingPoint(t *testing.T) {
	var score Score

	score.EndgameStatuses = [3]EndgameStatus{EndgameNone, EndgameNone, EndgameNone}
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	score.EndgameStatuses = [3]EndgameStatus{EndgameOnstage, EndgameNone, EndgameOnstage}
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	score.EndgameHarmony = true
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	score.EndgameHarmony = false
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	EnsembleBonusPointThreshold = 10
	score.EndgameHarmony = true
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	score.EndgameHarmony = true
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	EnsembleBonusPointThreshold = 20
	score.EndgameStatuses = [3]EndgameStatus{EndgameOnstageWithSpotlit, EndgameOnstageWithSpotlit, EndgameOnstageWithSpotlit}
	score.EndgameHarmony = true
	assert.Equal(t, true, score.Summarize(&Score{}).EnsembleBonusRankingPoint)

	EnsembleBonusPointThreshold = 43
	assert.Equal(t, false, score.Summarize(&Score{}).EnsembleBonusRankingPoint)
}

func TestScoreEquals(t *testing.T) {
	score1 := TestScore1()
	score2 := TestScore1()
	assert.True(t, score1.Equals(score2))
	assert.True(t, score2.Equals(score1))

	score3 := TestScore2()
	assert.False(t, score1.Equals(score3))
	assert.False(t, score3.Equals(score1))

	score2 = TestScore1()
	score2.LeaveStatuses[0] = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameStatuses[1] = EndgameParked
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.EndgameHarmony = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls = []Foul{}
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].IsTechnical = false
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].TeamId += 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.Fouls[0].RuleId = 1
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))

	score2 = TestScore1()
	score2.PlayoffDq = !score2.PlayoffDq
	assert.False(t, score1.Equals(score2))
	assert.False(t, score2.Equals(score1))
}
