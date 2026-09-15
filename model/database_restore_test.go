// Copyright 2026 Team 254. All Rights Reserved.

package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseRestoreFrom(t *testing.T) {
	destination := setupTestDb(t)
	require.NoError(t, destination.CreateTeam(&Team{Id: 1}))
	sourcePath := filepath.Join(t.TempDir(), "source.db")
	source, err := OpenDatabase(sourcePath)
	require.NoError(t, err)
	settings, err := source.GetEventSettings()
	require.NoError(t, err)
	settings.Name = "Restored Event"
	require.NoError(t, source.UpdateEventSettings(settings))
	require.NoError(t, source.CreateTeam(&Team{Id: 254, Nickname: "Cheesy Poofs"}))
	require.NoError(t, source.Close())

	require.NoError(t, destination.RestoreFrom(sourcePath))
	restoredSettings, err := destination.GetEventSettings()
	require.NoError(t, err)
	assert.Equal(t, "Restored Event", restoredSettings.Name)
	restoredTeam, err := destination.GetTeamById(254)
	require.NoError(t, err)
	require.NotNil(t, restoredTeam)
	assert.Equal(t, "Cheesy Poofs", restoredTeam.Nickname)
	oldTeam, err := destination.GetTeamById(1)
	require.NoError(t, err)
	assert.Nil(t, oldTeam)
}

func TestDatabaseRestoreRejectsNonBoltFile(t *testing.T) {
	destination := setupTestDb(t)
	invalidPath := filepath.Join(t.TempDir(), "invalid.db")
	require.NoError(t, os.WriteFile(invalidPath, []byte("not a bolt database"), 0600))
	assert.Error(t, destination.RestoreFrom(invalidPath))
}
