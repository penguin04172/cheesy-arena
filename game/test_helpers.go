// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Helper methods for use in tests in this package and others.

package game

func TestScore1() *Score {
	fouls := []Foul{
		{true, 25, 13},
		{false, 1868, 14},
		{false, 1868, 14},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
		{true, 25, 15},
	}
	return &Score{
		LeaveStatuses:   [3]bool{true, true, false},
		CoralAlgae:      scoreElements1(),
		EndgameStatuses: [3]EndgameStatus{EndgameParked, EndgameNone, EndgameShallowCage},
		Fouls:           fouls,
		PlayoffDq:       false,
	}
}

func TestScore2() *Score {
	return &Score{
		LeaveStatuses:   [3]bool{false, true, false},
		CoralAlgae:      scoreElements2(),
		EndgameStatuses: [3]EndgameStatus{EndgameShallowCage, EndgameDeepCage, EndgameDeepCage},
		Fouls:           []Foul{},
		PlayoffDq:       false,
	}
}

func TestRanking1() *Ranking {
	return &Ranking{254, 1, 0, RankingFields{20, 625, 90, 554, 12, 0.254, 3, 2, 1, 0, 10}}
}

func TestRanking2() *Ranking {
	return &Ranking{1114, 2, 1, RankingFields{18, 700, 625, 90, 23, 0.1114, 1, 3, 2, 0, 10}}
}

func scoreElements1() CoralAlgae {
	return CoralAlgae{
		AutoScoring: [3][12]bool{
			{true, true, false, false, false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false, false, false, false, false},
			{false, false, false, false, false, false, false, false, false, false, false, false},
		},
		Nodes: [3][12]bool{
			{true, true, true, true, true, true, true, false, false, false, false, false},
			{false, false, false, false, false, false, false, true, true, true, true, true},
			{false, false, false, false, false, false, false, false, false, false, false, false},
		},
		AutoCoralTrough:      6,
		TotalCoralTrough:     3,
		AutoProcessorAlgae:   5,
		TeleopProcessorAlgae: 7,
		AutoNetAlgae:         4,
		TeleopNetAlgae:       6,
	}
}

func scoreElements2() CoralAlgae {
	return CoralAlgae{
		AutoScoring: [3][12]bool{
			{true, false, false, false, false, false, false, false, false, false, false, false},
			{true, false, false, false, false, false, false, false, false, false, false, false},
			{true, false, false, false, false, false, false, false, false, false, false, false},
		},
		Nodes: [3][12]bool{
			{true, true, true, true, true, false, false, false, false, false, false, false},
			{true, false, false, true, false, true, false, false, true, false, true, false},
			{true, false, false, true, false, false, true, false, false, true, true, false},
		},
		AutoCoralTrough:      6,
		TotalCoralTrough:     12,
		AutoProcessorAlgae:   10,
		TeleopProcessorAlgae: 14,
		AutoNetAlgae:         8,
		TeleopNetAlgae:       12,
	}
}
