// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiV1DisplayClientsUseBootstrapAndV1Streams(t *testing.T) {
	for _, test := range []struct {
		path             string
		bootstrapPath    string
		streamPath       string
		legacyStreamPath string
	}{
		{"../static/js/queueing_display.js", "/api/v1/displays/queueing/bootstrap", "/api/v1/streams/displays/queueing", "/displays/queueing/websocket"},
		{"../static/js/announcer_display.js", "/api/v1/displays/announcer/bootstrap", "/api/v1/streams/displays/announcer", "/displays/announcer/websocket"},
	} {
		contents, err := os.ReadFile(test.path)
		require.NoError(t, err)
		source := string(contents)
		assert.Contains(t, source, test.bootstrapPath)
		assert.Contains(t, source, test.streamPath)
		assert.Contains(t, source, "new CheesyWebsocketV1")
		assert.NotContains(t, source, test.legacyStreamPath)
	}
}
