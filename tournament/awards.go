// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for managing awards and their associated lower thirds.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
)

// Creates or updates the given award, depending on whether or not it already exists.
func CreateOrUpdateAward(database *model.Database, award *model.Award, createIntroLowerThird bool) error {
	// Validate the award data.
	if award.AwardName == "" {
		return fmt.Errorf("Award name cannot be blank.")
	}
	var team *model.Team
	if award.TeamId > 0 {
		var err error
		team, err = database.GetTeamById(award.TeamId)
		if err != nil {
			return err
		}
		if team == nil {
			return fmt.Errorf("Team %d is not present at this event.", award.TeamId)
		}
	}

	// Create or update the award and associated lower thirds atomically.
	awardIntroLowerThird := model.LowerThird{TopText: award.AwardName, AwardId: award.Id}
	awardWinnerLowerThird := model.LowerThird{
		TopText: award.AwardName, BottomText: award.PersonName, AwardId: award.Id,
	}
	if team != nil {
		if award.PersonName == "" {
			awardWinnerLowerThird.BottomText = fmt.Sprintf("Team %d, %s", team.Id, team.Nickname)
		} else {
			awardWinnerLowerThird.BottomText = fmt.Sprintf(
				"%s &ndash; Team %d, %s", award.PersonName, team.Id, team.Nickname,
			)
		}
	}
	if awardWinnerLowerThird.BottomText == "" {
		awardWinnerLowerThird.BottomText = "(No awardee assigned yet)"
	}
	desiredLowerThirds := make([]model.LowerThird, 0, 2)
	if createIntroLowerThird {
		desiredLowerThirds = append(desiredLowerThirds, awardIntroLowerThird)
	}
	desiredLowerThirds = append(desiredLowerThirds, awardWinnerLowerThird)
	return database.SaveAwardWithLowerThirds(award, desiredLowerThirds)
}

// Deletes the given award and any associated lower thirds.
func DeleteAward(database *model.Database, awardId int) error {
	return database.DeleteAwardWithLowerThirds(awardId)
}

// Generates awards and lower thirds for the tournament winners and finalists.
func CreateOrUpdateWinnerAndFinalistAwards(database *model.Database, winnerAllianceId, finalistAllianceId int) error {
	var winnerAlliance, finalistAlliance *model.Alliance
	var err error
	if winnerAlliance, err = database.GetAllianceById(winnerAllianceId); err != nil {
		return err
	}
	if finalistAlliance, err = database.GetAllianceById(finalistAllianceId); err != nil {
		return err
	}
	if winnerAlliance == nil || finalistAlliance == nil {
		return fmt.Errorf("Winner and/or finalist alliances do not exist.")
	}
	if len(winnerAlliance.TeamIds) == 0 || len(finalistAlliance.TeamIds) == 0 {
		return fmt.Errorf("Winner and/or finalist alliances do not contain teams.")
	}

	// Clear out any awards that may exist if the final match was scored more than once.
	winnerAwards, err := database.GetAwardsByType(model.WinnerAward)
	if err != nil {
		return err
	}
	finalistAwards, err := database.GetAwardsByType(model.FinalistAward)
	if err != nil {
		return err
	}
	for _, award := range append(winnerAwards, finalistAwards...) {
		if err = DeleteAward(database, award.Id); err != nil {
			return err
		}
	}

	// Create the finalist awards first since they're usually presented first.
	finalistAward := model.Award{
		AwardName: "Finalist",
		Type:      model.FinalistAward,
		TeamId:    finalistAlliance.TeamIds[0],
	}
	if err = CreateOrUpdateAward(database, &finalistAward, true); err != nil {
		return err
	}
	for _, allianceTeamId := range finalistAlliance.TeamIds[1:] {
		finalistAward.Id = 0
		finalistAward.TeamId = allianceTeamId
		if err = CreateOrUpdateAward(database, &finalistAward, false); err != nil {
			return err
		}
	}

	// Create the winner awards.
	winnerAward := model.Award{
		AwardName: "Winner",
		Type:      model.WinnerAward,
		TeamId:    winnerAlliance.TeamIds[0],
	}
	if err = CreateOrUpdateAward(database, &winnerAward, true); err != nil {
		return err
	}
	for _, allianceTeamId := range winnerAlliance.TeamIds[1:] {
		winnerAward.Id = 0
		winnerAward.TeamId = allianceTeamId
		if err = CreateOrUpdateAward(database, &winnerAward, false); err != nil {
			return err
		}
	}

	return nil
}
