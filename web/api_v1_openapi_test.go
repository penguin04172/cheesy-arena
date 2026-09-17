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
		"/displays/queueing/matches",
		"/displays/announcer/match",
		"/displays/announcer/score",
		"/displays/queueing/bootstrap", "/displays/announcer/bootstrap",
		"/displays/audience/bootstrap", "/displays/alliance-station/bootstrap",
		"/displays/wall/bootstrap", "/displays/unpicked/bootstrap",
		"/displays/field-monitor/bootstrap",
		"/streams/displays/queueing", "/streams/displays/announcer",
		"/streams/displays/audience", "/streams/displays/alliance-station",
		"/streams/displays/wall", "/streams/displays/unpicked",
		"/streams/displays/field-monitor",
		"/streams/admin/alliance-selection",
		"/streams/admin/match-play",
		"/streams/admin/scoring/{position}", "/streams/admin/referee",
		"/streams/admin/field-testing",
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
		"/admin/publishing/{resource}",
		"/admin/match-play/matches",
		"/admin/referee/fouls",
		"/admin/alliance-selection/bootstrap",
		"/admin/match-play/bootstrap",
		"/admin/scoring/{position}/bootstrap", "/admin/referee/bootstrap",
		"/admin/field-testing/bootstrap",
	} {
		assert.Contains(t, document.Paths, path)
	}
}
