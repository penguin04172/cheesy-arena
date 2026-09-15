// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"net/http"

	"github.com/Team254/cheesy-arena/model"
)

type apiV1TeamSignSettings struct {
	Red1      int `json:"red1"`
	Red2      int `json:"red2"`
	Red3      int `json:"red3"`
	RedTimer  int `json:"redTimer"`
	Blue1     int `json:"blue1"`
	Blue2     int `json:"blue2"`
	Blue3     int `json:"blue3"`
	BlueTimer int `json:"blueTimer"`
}

type apiV1TeamSignSettingsInput struct {
	Red1      *int `json:"red1"`
	Red2      *int `json:"red2"`
	Red3      *int `json:"red3"`
	RedTimer  *int `json:"redTimer"`
	Blue1     *int `json:"blue1"`
	Blue2     *int `json:"blue2"`
	Blue3     *int `json:"blue3"`
	BlueTimer *int `json:"blueTimer"`
}

type apiV1CompanionButton struct {
	Page   int `json:"page"`
	Row    int `json:"row"`
	Column int `json:"column"`
}

type apiV1CompanionButtonInput struct {
	Page   *int `json:"page"`
	Row    *int `json:"row"`
	Column *int `json:"column"`
}

type apiV1CompanionSettings struct {
	Address           string               `json:"address"`
	Port              int                  `json:"port"`
	MatchPreview      apiV1CompanionButton `json:"matchPreview"`
	SetAudience       apiV1CompanionButton `json:"setAudience"`
	MatchStart        apiV1CompanionButton `json:"matchStart"`
	TeleopStart       apiV1CompanionButton `json:"teleopStart"`
	EndgameStart      apiV1CompanionButton `json:"endgameStart"`
	MatchEnd          apiV1CompanionButton `json:"matchEnd"`
	PostResult        apiV1CompanionButton `json:"postResult"`
	AllianceSelection apiV1CompanionButton `json:"allianceSelection"`
	MatchAbort        apiV1CompanionButton `json:"matchAbort"`
}

type apiV1CompanionSettingsInput struct {
	Address           *string                    `json:"address"`
	Port              *int                       `json:"port"`
	MatchPreview      *apiV1CompanionButtonInput `json:"matchPreview"`
	SetAudience       *apiV1CompanionButtonInput `json:"setAudience"`
	MatchStart        *apiV1CompanionButtonInput `json:"matchStart"`
	TeleopStart       *apiV1CompanionButtonInput `json:"teleopStart"`
	EndgameStart      *apiV1CompanionButtonInput `json:"endgameStart"`
	MatchEnd          *apiV1CompanionButtonInput `json:"matchEnd"`
	PostResult        *apiV1CompanionButtonInput `json:"postResult"`
	AllianceSelection *apiV1CompanionButtonInput `json:"allianceSelection"`
	MatchAbort        *apiV1CompanionButtonInput `json:"matchAbort"`
}

type apiV1HardwareSettings struct {
	PlcAddress           string                 `json:"plcAddress"`
	LedControllerAddress string                 `json:"ledControllerAddress"`
	LedUniverseMode      string                 `json:"ledUniverseMode"`
	TeamSigns            apiV1TeamSignSettings  `json:"teamSigns"`
	UseLiteUdpPort       bool                   `json:"useLiteUdpPort"`
	BlackmagicAddresses  string                 `json:"blackmagicAddresses"`
	Companion            apiV1CompanionSettings `json:"companion"`
}

type apiV1HardwareSettingsInput struct {
	PlcAddress           *string                      `json:"plcAddress"`
	LedControllerAddress *string                      `json:"ledControllerAddress"`
	LedUniverseMode      *string                      `json:"ledUniverseMode"`
	TeamSigns            *apiV1TeamSignSettingsInput  `json:"teamSigns"`
	UseLiteUdpPort       *bool                        `json:"useLiteUdpPort"`
	BlackmagicAddresses  *string                      `json:"blackmagicAddresses"`
	Companion            *apiV1CompanionSettingsInput `json:"companion"`
}

func (web *Web) applyApiV1HardwareSettings(w http.ResponseWriter, r *http.Request, settings *model.EventSettings) error {
	var input apiV1HardwareSettingsInput
	if !decodeApiV1Input(w, r, &input) {
		return errApiV1ResponseWritten
	}
	if input.PlcAddress != nil {
		settings.PlcAddress = *input.PlcAddress
	}
	if input.LedControllerAddress != nil {
		settings.LedControllerAddress = *input.LedControllerAddress
	}
	if input.LedUniverseMode != nil {
		if *input.LedUniverseMode != "single" && *input.LedUniverseMode != "two" {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "LED universe mode is invalid.", map[string]string{"ledUniverseMode": "must be single or two"})
			return errApiV1ResponseWritten
		}
		settings.LedUniverseMode = *input.LedUniverseMode
	}
	if input.UseLiteUdpPort != nil {
		settings.UseLiteUdpPort = *input.UseLiteUdpPort
	}
	if input.BlackmagicAddresses != nil {
		settings.BlackmagicAddresses = *input.BlackmagicAddresses
	}
	if input.TeamSigns != nil && !applyApiV1TeamSigns(w, r, input.TeamSigns, settings) {
		return errApiV1ResponseWritten
	}
	if input.Companion != nil && !applyApiV1Companion(w, r, input.Companion, settings) {
		return errApiV1ResponseWritten
	}
	return nil
}

func applyApiV1TeamSigns(w http.ResponseWriter, r *http.Request, input *apiV1TeamSignSettingsInput, settings *model.EventSettings) bool {
	values := []struct {
		input  *int
		output *int
		name   string
	}{
		{input.Red1, &settings.TeamSignRed1Id, "teamSigns.red1"}, {input.Red2, &settings.TeamSignRed2Id, "teamSigns.red2"},
		{input.Red3, &settings.TeamSignRed3Id, "teamSigns.red3"}, {input.RedTimer, &settings.TeamSignRedTimerId, "teamSigns.redTimer"},
		{input.Blue1, &settings.TeamSignBlue1Id, "teamSigns.blue1"}, {input.Blue2, &settings.TeamSignBlue2Id, "teamSigns.blue2"},
		{input.Blue3, &settings.TeamSignBlue3Id, "teamSigns.blue3"}, {input.BlueTimer, &settings.TeamSignBlueTimerId, "teamSigns.blueTimer"},
	}
	for _, value := range values {
		if value.input != nil {
			if *value.input < 0 {
				writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Team sign ID is invalid.", map[string]string{value.name: "must not be negative"})
				return false
			}
			*value.output = *value.input
		}
	}
	return true
}

func applyApiV1Companion(w http.ResponseWriter, r *http.Request, input *apiV1CompanionSettingsInput, settings *model.EventSettings) bool {
	if input.Address != nil {
		settings.CompanionAddress = *input.Address
	}
	if input.Port != nil {
		if *input.Port < 0 || *input.Port > 65535 {
			writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Companion port is invalid.", map[string]string{"companion.port": "must be between 0 and 65535"})
			return false
		}
		settings.CompanionPort = *input.Port
	}
	buttons := []struct {
		input             *apiV1CompanionButtonInput
		page, row, column *int
		name              string
	}{
		{input.MatchPreview, &settings.CompanionMatchPreviewPage, &settings.CompanionMatchPreviewRow, &settings.CompanionMatchPreviewColumn, "matchPreview"},
		{input.SetAudience, &settings.CompanionSetAudiencePage, &settings.CompanionSetAudienceRow, &settings.CompanionSetAudienceColumn, "setAudience"},
		{input.MatchStart, &settings.CompanionMatchStartPage, &settings.CompanionMatchStartRow, &settings.CompanionMatchStartColumn, "matchStart"},
		{input.TeleopStart, &settings.CompanionTeleopStartPage, &settings.CompanionTeleopStartRow, &settings.CompanionTeleopStartColumn, "teleopStart"},
		{input.EndgameStart, &settings.CompanionEndgameStartPage, &settings.CompanionEndgameStartRow, &settings.CompanionEndgameStartColumn, "endgameStart"},
		{input.MatchEnd, &settings.CompanionMatchEndPage, &settings.CompanionMatchEndRow, &settings.CompanionMatchEndColumn, "matchEnd"},
		{input.PostResult, &settings.CompanionPostResultPage, &settings.CompanionPostResultRow, &settings.CompanionPostResultColumn, "postResult"},
		{input.AllianceSelection, &settings.CompanionAllianceSelectionPage, &settings.CompanionAllianceSelectionRow, &settings.CompanionAllianceSelectionColumn, "allianceSelection"},
		{input.MatchAbort, &settings.CompanionMatchAbortPage, &settings.CompanionMatchAbortRow, &settings.CompanionMatchAbortColumn, "matchAbort"},
	}
	for _, button := range buttons {
		if button.input == nil {
			continue
		}
		for _, value := range []struct {
			input  *int
			output *int
			suffix string
		}{{button.input.Page, button.page, "page"}, {button.input.Row, button.row, "row"}, {button.input.Column, button.column, "column"}} {
			if value.input != nil {
				if *value.input < 0 {
					writeApiV1Error(w, r, http.StatusUnprocessableEntity, "validation_failed", "Companion button coordinate is invalid.", map[string]string{"companion." + button.name + "." + value.suffix: "must not be negative"})
					return false
				}
				*value.output = *value.input
			}
		}
	}
	return true
}

func newApiV1HardwareSettings(s *model.EventSettings) apiV1HardwareSettings {
	button := func(page, row, column int) apiV1CompanionButton {
		return apiV1CompanionButton{Page: page, Row: row, Column: column}
	}
	return apiV1HardwareSettings{
		PlcAddress: s.PlcAddress, LedControllerAddress: s.LedControllerAddress, LedUniverseMode: s.LedUniverseMode,
		TeamSigns:      apiV1TeamSignSettings{Red1: s.TeamSignRed1Id, Red2: s.TeamSignRed2Id, Red3: s.TeamSignRed3Id, RedTimer: s.TeamSignRedTimerId, Blue1: s.TeamSignBlue1Id, Blue2: s.TeamSignBlue2Id, Blue3: s.TeamSignBlue3Id, BlueTimer: s.TeamSignBlueTimerId},
		UseLiteUdpPort: s.UseLiteUdpPort, BlackmagicAddresses: s.BlackmagicAddresses,
		Companion: apiV1CompanionSettings{Address: s.CompanionAddress, Port: s.CompanionPort,
			MatchPreview: button(s.CompanionMatchPreviewPage, s.CompanionMatchPreviewRow, s.CompanionMatchPreviewColumn), SetAudience: button(s.CompanionSetAudiencePage, s.CompanionSetAudienceRow, s.CompanionSetAudienceColumn),
			MatchStart: button(s.CompanionMatchStartPage, s.CompanionMatchStartRow, s.CompanionMatchStartColumn), TeleopStart: button(s.CompanionTeleopStartPage, s.CompanionTeleopStartRow, s.CompanionTeleopStartColumn),
			EndgameStart: button(s.CompanionEndgameStartPage, s.CompanionEndgameStartRow, s.CompanionEndgameStartColumn), MatchEnd: button(s.CompanionMatchEndPage, s.CompanionMatchEndRow, s.CompanionMatchEndColumn),
			PostResult: button(s.CompanionPostResultPage, s.CompanionPostResultRow, s.CompanionPostResultColumn), AllianceSelection: button(s.CompanionAllianceSelectionPage, s.CompanionAllianceSelectionRow, s.CompanionAllianceSelectionColumn),
			MatchAbort: button(s.CompanionMatchAbortPage, s.CompanionMatchAbortRow, s.CompanionMatchAbortColumn)},
	}
}
