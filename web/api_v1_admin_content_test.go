// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestApiV1AdminAwardCrud(t *testing.T) {
	web := setupTestWeb(t)
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 254}))
	require.NoError(t, web.arena.Database.CreateTeam(&model.Team{Id: 1114}))
	created := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/awards", `{
		"awardName":"Engineering Inspiration","teamId":254,"personName":""
	}`)
	assert.Equal(t, http.StatusCreated, created.Code)
	var createResponse struct {
		Data apiV1AdminAward `json:"data"`
	}
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &createResponse))
	assert.Positive(t, createResponse.Data.Id)

	path := "/api/v1/admin/awards/" + strconv.Itoa(createResponse.Data.Id)
	updated := apiV1AdminJsonRequest(web, http.MethodPatch, path, `{
		"awardName":"Excellence in Engineering","teamId":1114,"personName":""
	}`)
	assert.Equal(t, http.StatusOK, updated.Code)
	award, err := web.arena.Database.GetAwardById(createResponse.Data.Id)
	require.NoError(t, err)
	require.NotNil(t, award)
	assert.Equal(t, "Excellence in Engineering", award.AwardName)

	list := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/awards", "")
	assert.Equal(t, http.StatusOK, list.Code)
	assert.Contains(t, list.Body.String(), "Excellence in Engineering")

	deleted := apiV1AdminJsonRequest(web, http.MethodDelete, path, "")
	assert.Equal(t, http.StatusNoContent, deleted.Code)
	award, err = web.arena.Database.GetAwardById(createResponse.Data.Id)
	require.NoError(t, err)
	assert.Nil(t, award)
}

func TestApiV1AdminAwardValidation(t *testing.T) {
	web := setupTestWeb(t)
	recorder := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/awards", `{
		"awardName":"","teamId":-1,"personName":""
	}`)
	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "validation_failed")
}

func TestApiV1AdminScheduledBreakUpdate(t *testing.T) {
	web := setupTestWeb(t)
	scheduledBreak := model.ScheduledBreak{
		MatchType: model.Playoff, TypeOrderBefore: 2, Time: time.Now(), DurationSec: 600, Description: "Old",
	}
	require.NoError(t, web.arena.Database.CreateScheduledBreak(&scheduledBreak))
	path := "/api/v1/admin/scheduled-breaks/" + strconv.Itoa(scheduledBreak.Id)
	recorder := apiV1AdminJsonRequest(web, http.MethodPatch, path, `{"description":"Field Reset"}`)
	assert.Equal(t, http.StatusOK, recorder.Code)
	updated, err := web.arena.Database.GetScheduledBreakById(scheduledBreak.Id)
	require.NoError(t, err)
	assert.Equal(t, "Field Reset", updated.Description)

	list := apiV1AdminJsonRequest(web, http.MethodGet, "/api/v1/admin/scheduled-breaks", "")
	assert.Equal(t, http.StatusOK, list.Code)
	assert.Contains(t, list.Body.String(), "Field Reset")
}

func TestApiV1AdminSponsorSlideCrudAndReorder(t *testing.T) {
	web := setupTestWeb(t)
	first := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/sponsor-slides", `{
		"subtitle":"First","line1":"One","line2":"","image":"","displayTimeSec":10
	}`)
	second := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/sponsor-slides", `{
		"subtitle":"Second","line1":"Two","line2":"","image":"","displayTimeSec":15
	}`)
	assert.Equal(t, http.StatusCreated, first.Code)
	assert.Equal(t, http.StatusCreated, second.Code)
	var firstResponse, secondResponse struct {
		Data apiV1SponsorSlide `json:"data"`
	}
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstResponse))
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondResponse))

	reorderBody := `{"ids":[` + strconv.Itoa(secondResponse.Data.Id) + `,` + strconv.Itoa(firstResponse.Data.Id) + `]}`
	reordered := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/sponsor-slides/reorder", reorderBody)
	assert.Equal(t, http.StatusOK, reordered.Code)
	slides, err := web.arena.Database.GetAllSponsorSlides()
	require.NoError(t, err)
	require.Len(t, slides, 2)
	assert.Equal(t, secondResponse.Data.Id, slides[0].Id)

	invalidOrder := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/sponsor-slides/reorder", `{"ids":[1,1]}`)
	assert.Equal(t, http.StatusUnprocessableEntity, invalidOrder.Code)

	deletePath := "/api/v1/admin/sponsor-slides/" + strconv.Itoa(firstResponse.Data.Id)
	deleted := apiV1AdminJsonRequest(web, http.MethodDelete, deletePath, "")
	assert.Equal(t, http.StatusNoContent, deleted.Code)
}

func TestApiV1AdminLowerThirdCrudAndReorder(t *testing.T) {
	web := setupTestWeb(t)
	first := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/lower-thirds", `{
		"topText":"First","bottomText":"One"
	}`)
	second := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/lower-thirds", `{
		"topText":"Second","bottomText":"Two"
	}`)
	assert.Equal(t, http.StatusCreated, first.Code)
	assert.Equal(t, http.StatusCreated, second.Code)
	var firstResponse, secondResponse struct {
		Data apiV1AdminLowerThird `json:"data"`
	}
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstResponse))
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondResponse))

	reorderBody := `{"ids":[` + strconv.Itoa(secondResponse.Data.Id) + `,` + strconv.Itoa(firstResponse.Data.Id) + `]}`
	reordered := apiV1AdminJsonRequest(web, http.MethodPost, "/api/v1/admin/lower-thirds/reorder", reorderBody)
	assert.Equal(t, http.StatusOK, reordered.Code)
	lowerThirds, err := web.arena.Database.GetAllLowerThirds()
	require.NoError(t, err)
	require.Len(t, lowerThirds, 2)
	assert.Equal(t, secondResponse.Data.Id, lowerThirds[0].Id)

	updatePath := "/api/v1/admin/lower-thirds/" + strconv.Itoa(firstResponse.Data.Id)
	updated := apiV1AdminJsonRequest(web, http.MethodPatch, updatePath, `{"topText":"Updated","bottomText":""}`)
	assert.Equal(t, http.StatusOK, updated.Code)
	deleted := apiV1AdminJsonRequest(web, http.MethodDelete, updatePath, "")
	assert.Equal(t, http.StatusNoContent, deleted.Code)
}

func TestApiV1AdminMutationRequiresCsrf(t *testing.T) {
	web := setupTestWeb(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/awards", strings.NewReader(`{
		"awardName":"Test","teamId":0,"personName":""
	}`))
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "invalid_csrf_token")
}

func apiV1AdminJsonRequest(web *Web, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-CSRF-Token", "test-csrf-token")
	request.AddCookie(&http.Cookie{Name: csrfTokenCookie, Value: "test-csrf-token"})
	recorder := httptest.NewRecorder()
	web.newHandler().ServeHTTP(recorder, request)
	return recorder
}
