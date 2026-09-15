// Copyright 2026 Team 254. All Rights Reserved.
//
// In-memory status tracking for bounded background API jobs.

package web

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type apiV1State struct {
	jobs           *apiV1JobManager
	teamMutationMu sync.Mutex
	judgingMu      sync.Mutex
	settingsMu     sync.Mutex
}

type apiV1JobManager struct {
	mu   sync.RWMutex
	jobs map[string]*apiV1Job
}

type apiV1Job struct {
	Id          string
	Type        string
	Status      string
	Progress    int
	Error       string
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type apiV1JobDto struct {
	Id          string  `json:"id"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	Progress    int     `json:"progress"`
	Error       *string `json:"error"`
	CreatedAt   string  `json:"createdAt"`
	StartedAt   *string `json:"startedAt"`
	CompletedAt *string `json:"completedAt"`
}

func newApiV1State() *apiV1State {
	return &apiV1State{jobs: &apiV1JobManager{jobs: make(map[string]*apiV1Job)}}
}

func (manager *apiV1JobManager) create(jobType string) apiV1JobDto {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	for id, existing := range manager.jobs {
		if existing.CompletedAt != nil && existing.CompletedAt.Before(cutoff) {
			delete(manager.jobs, id)
		}
	}
	job := &apiV1Job{Id: uuid.NewString(), Type: jobType, Status: "queued", CreatedAt: time.Now().UTC()}
	manager.jobs[job.Id] = job
	return newApiV1JobDto(job)
}

func (web *Web) failApiV1Job(jobId string, publicMessage string, cause error) {
	log.Printf("API v1 job failed jobId=%s: %v", jobId, cause)
	web.apiV1State.jobs.finish(jobId, errors.New(publicMessage))
}

func (manager *apiV1JobManager) get(id string) (apiV1JobDto, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	job, ok := manager.jobs[id]
	if !ok {
		return apiV1JobDto{}, false
	}
	return newApiV1JobDto(job), true
}

func (manager *apiV1JobManager) start(id string) {
	manager.update(id, func(job *apiV1Job) {
		now := time.Now().UTC()
		job.Status, job.StartedAt = "running", &now
	})
}

func (manager *apiV1JobManager) progress(id string, progress int) {
	manager.update(id, func(job *apiV1Job) { job.Progress = progress })
}

func (manager *apiV1JobManager) finish(id string, err error) {
	manager.update(id, func(job *apiV1Job) {
		now := time.Now().UTC()
		job.CompletedAt = &now
		if err == nil {
			job.Status, job.Progress = "succeeded", 100
		} else {
			job.Status, job.Error = "failed", err.Error()
		}
	})
}

func (manager *apiV1JobManager) update(id string, update func(*apiV1Job)) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if job := manager.jobs[id]; job != nil {
		update(job)
	}
}

func newApiV1JobDto(job *apiV1Job) apiV1JobDto {
	dto := apiV1JobDto{
		Id: job.Id, Type: job.Type, Status: job.Status, Progress: job.Progress,
		CreatedAt: job.CreatedAt.Format(time.RFC3339Nano),
	}
	if job.Error != "" {
		dto.Error = &job.Error
	}
	if job.StartedAt != nil {
		value := job.StartedAt.Format(time.RFC3339Nano)
		dto.StartedAt = &value
	}
	if job.CompletedAt != nil {
		value := job.CompletedAt.Format(time.RFC3339Nano)
		dto.CompletedAt = &value
	}
	return dto
}
