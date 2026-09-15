// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"fmt"
	"net/http"

	"github.com/Team254/cheesy-arena/partner"
)

type apiV1PublishingResult struct {
	Job      apiV1JobDto `json:"job"`
	Replayed bool        `json:"replayed"`
}

func (web *Web) apiV1AdminPublishingHandler(w http.ResponseWriter, r *http.Request) {
	resource := r.PathValue("resource")
	if !validApiV1PublishingResource(resource) {
		writeApiV1Error(w, r, http.StatusNotFound, "publishing_resource_not_found", "Publishing resource not found.", nil)
		return
	}
	key, ok := readApiV1IdempotencyKey(w, r)
	if !ok {
		return
	}
	web.apiV1State.publishingMu.Lock()
	defer web.apiV1State.publishingMu.Unlock()
	operation := "publishing." + resource
	if data, found, conflict := web.getApiV1Idempotency(key, operation); found {
		if conflict {
			writeApiV1Error(w, r, http.StatusConflict, "idempotency_key_reused", "Idempotency key was already used for another operation.", nil)
			return
		}
		result := data.(apiV1PublishingResult)
		result.Replayed = true
		if current, exists := web.apiV1State.jobs.get(result.Job.Id); exists {
			result.Job = current
		}
		writeApiV1Data(w, r, http.StatusOK, result, nil)
		return
	}
	if !web.arena.EventSettings.TbaPublishingEnabled {
		writeApiV1Error(w, r, http.StatusUnprocessableEntity, "publishing_disabled", "TBA publishing is not enabled.", nil)
		return
	}
	job := web.apiV1State.jobs.create("publish_" + resource)
	result := apiV1PublishingResult{Job: job}
	web.storeApiV1Idempotency(key, operation, result)
	web.logApiV1Audit(r, "publishing."+resource, job.Id, "accepted")
	go web.runApiV1Publishing(job.Id, resource, web.arena.TbaClient)
	writeApiV1Data(w, r, http.StatusAccepted, result, nil)
}

func (web *Web) runApiV1Publishing(jobId, resource string, client *partner.TbaClient) {
	web.apiV1State.jobs.start(jobId)
	web.apiV1State.destructiveMu.Lock()
	defer web.apiV1State.destructiveMu.Unlock()
	var err error
	switch resource {
	case "teams":
		err = client.PublishTeams(web.arena.Database)
	case "matches":
		err = client.DeletePublishedMatches()
		if err == nil {
			err = client.PublishMatches(web.arena.Database)
		}
	case "rankings":
		err = client.PublishRankings(web.arena.Database)
	case "alliances":
		err = client.PublishAlliances(web.arena.Database)
	case "awards":
		err = client.PublishAwards(web.arena.Database)
	default:
		err = fmt.Errorf("unsupported publishing resource %q", resource)
	}
	if err != nil {
		web.failApiV1Job(jobId, "Unable to publish "+resource+" to TBA.", err)
		return
	}
	web.apiV1State.jobs.finish(jobId, nil)
}

func validApiV1PublishingResource(resource string) bool {
	switch resource {
	case "teams", "matches", "rankings", "alliances", "awards":
		return true
	default:
		return false
	}
}
