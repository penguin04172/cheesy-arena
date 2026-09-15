// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentReorderValidationLeavesOriginalOrder(t *testing.T) {
	database := setupTestDb(t)
	first := SponsorSlide{DisplayOrder: 1}
	second := SponsorSlide{DisplayOrder: 2}
	require.NoError(t, database.CreateSponsorSlide(&first))
	require.NoError(t, database.CreateSponsorSlide(&second))
	require.Error(t, database.ReorderSponsorSlides([]int{second.Id, second.Id}))
	slides, err := database.GetAllSponsorSlides()
	require.NoError(t, err)
	require.Len(t, slides, 2)
	assert.Equal(t, first.Id, slides[0].Id)
	assert.Equal(t, second.Id, slides[1].Id)

	require.NoError(t, database.ReorderSponsorSlides([]int{second.Id, first.Id}))
	slides, err = database.GetAllSponsorSlides()
	require.NoError(t, err)
	assert.Equal(t, second.Id, slides[0].Id)
}

func TestAwardAndLowerThirdsTransaction(t *testing.T) {
	database := setupTestDb(t)
	award := Award{Type: JudgedAward, AwardName: "Impact Award", TeamId: 254}
	lowerThirds := []LowerThird{{TopText: "Impact Award"}, {TopText: "Impact Award", BottomText: "Team 254"}}
	require.NoError(t, database.SaveAwardWithLowerThirds(&award, lowerThirds))
	assert.Positive(t, award.Id)
	storedLowerThirds, err := database.GetLowerThirdsByAwardId(award.Id)
	require.NoError(t, err)
	assert.Len(t, storedLowerThirds, 2)
	require.NoError(t, database.DeleteAwardWithLowerThirds(award.Id))
	storedAward, err := database.GetAwardById(award.Id)
	require.NoError(t, err)
	assert.Nil(t, storedAward)
	storedLowerThirds, err = database.GetLowerThirdsByAwardId(award.Id)
	require.NoError(t, err)
	assert.Empty(t, storedLowerThirds)
}
