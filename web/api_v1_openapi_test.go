// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func TestApiV1OpenApiDocument(t *testing.T) {
	contents, err := os.ReadFile("../docs/openapi-v1.yaml")
	require.NoError(t, err)
	var document struct {
		OpenApi string                    `yaml:"openapi"`
		Paths   map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(contents, &document))
	assert.Equal(t, "3.1.0", document.OpenApi)
	for _, path := range []string{
		"/alliances", "/bracket", "/event", "/game/rules", "/match-logs",
		"/matches/{matchId}/stations/{stationId}/logs", "/matches/{type}", "/rankings",
		"/session", "/sponsor-slides", "/teams", "/teams/{teamId}/avatar",
	} {
		if assert.Contains(t, document.Paths, path) {
			assert.Contains(t, document.Paths[path], "get")
		}
	}
	for _, path := range []string{
		"/admin/awards", "/admin/awards/{id}",
		"/admin/lower-thirds", "/admin/lower-thirds/{id}", "/admin/lower-thirds/reorder",
		"/admin/scheduled-breaks", "/admin/scheduled-breaks/{id}",
		"/admin/sponsor-slides", "/admin/sponsor-slides/{id}", "/admin/sponsor-slides/reorder",
		"/admin/teams", "/admin/teams/{id}", "/admin/teams/refresh", "/admin/teams/wpa-keys",
		"/admin/jobs/{id}",
		"/admin/judging-schedule",
		"/admin/settings/{section}",
		"/admin/database/backups", "/admin/database/restore", "/admin/tournament-data/{type}",
	} {
		assert.Contains(t, document.Paths, path)
	}
}
