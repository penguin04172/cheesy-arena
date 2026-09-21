// Copyright 2026 Team 254. All Rights Reserved.
//
// Shared routing and response helpers for version 1 of the JSON API.

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"strconv"
)

const maxApiV1RequestBodyBytes = 1024 * 1024

type apiV1ContextKey string

const apiV1RequestIdKey apiV1ContextKey = "requestId"

type apiV1Meta struct {
	RequestId string `json:"requestId"`
	Version   *int   `json:"version,omitempty"`
}

type apiV1Response struct {
	Data any       `json:"data"`
	Meta apiV1Meta `json:"meta"`
}

type apiV1ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type apiV1ErrorResponse struct {
	Error apiV1ErrorDetail `json:"error"`
	Meta  apiV1Meta        `json:"meta"`
}

type apiV1ResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (writer *apiV1ResponseWriter) WriteHeader(statusCode int) {
	writer.wroteHeader = true
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *apiV1ResponseWriter) Write(bytes []byte) (int, error) {
	writer.wroteHeader = true
	return writer.ResponseWriter.Write(bytes)
}

func (web *Web) registerApiV1Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/v1/alliances", web.apiV1Middleware(http.HandlerFunc(web.apiV1AlliancesHandler)))
	mux.Handle("GET /api/v1/bracket", web.apiV1Middleware(http.HandlerFunc(web.apiV1BracketHandler)))
	mux.Handle("GET /api/v1/event", web.apiV1Middleware(http.HandlerFunc(web.apiV1EventHandler)))
	mux.Handle("GET /api/v1/game/rules", web.apiV1Middleware(http.HandlerFunc(web.apiV1RulesHandler)))
	mux.Handle("GET /api/v1/match-logs", web.apiV1Middleware(http.HandlerFunc(web.apiV1MatchLogsHandler)))
	mux.Handle("GET /api/v1/matches/{matchId}/stations/{stationId}/logs", web.apiV1Middleware(http.HandlerFunc(web.apiV1MatchLogHandler)))
	mux.Handle("GET /api/v1/matches/{type}", web.apiV1Middleware(http.HandlerFunc(web.apiV1MatchesHandler)))
	mux.Handle("GET /api/v1/rankings", web.apiV1Middleware(http.HandlerFunc(web.apiV1RankingsHandler)))
	mux.Handle("GET /api/v1/session", web.apiV1Middleware(http.HandlerFunc(web.apiV1SessionHandler)))
	mux.Handle("GET /api/v1/sponsor-slides", web.apiV1Middleware(http.HandlerFunc(web.apiV1SponsorSlidesHandler)))
	mux.Handle("GET /api/v1/teams", web.apiV1Middleware(http.HandlerFunc(web.apiV1TeamsHandler)))
	mux.Handle("GET /api/v1/teams/{teamId}/avatar", web.apiV1Middleware(http.HandlerFunc(web.apiV1TeamAvatarHandler)))
	mux.Handle("GET /api/v1/admin/awards", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminAwardsHandler)))
	mux.Handle("POST /api/v1/admin/awards", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminAwardCreateHandler)))
	mux.Handle("PATCH /api/v1/admin/awards/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminAwardUpdateHandler)))
	mux.Handle("DELETE /api/v1/admin/awards/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminAwardDeleteHandler)))
	mux.Handle("GET /api/v1/admin/lower-thirds", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminLowerThirdsHandler)))
	mux.Handle("POST /api/v1/admin/lower-thirds", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminLowerThirdCreateHandler)))
	mux.Handle("PATCH /api/v1/admin/lower-thirds/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminLowerThirdUpdateHandler)))
	mux.Handle("DELETE /api/v1/admin/lower-thirds/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminLowerThirdDeleteHandler)))
	mux.Handle("POST /api/v1/admin/lower-thirds/reorder", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminLowerThirdReorderHandler)))
	mux.Handle("GET /api/v1/admin/scheduled-breaks", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminScheduledBreaksHandler)))
	mux.Handle("PATCH /api/v1/admin/scheduled-breaks/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminScheduledBreakUpdateHandler)))
	mux.Handle("GET /api/v1/admin/sponsor-slides", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminSponsorSlidesHandler)))
	mux.Handle("POST /api/v1/admin/sponsor-slides", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminSponsorSlideCreateHandler)))
	mux.Handle("PATCH /api/v1/admin/sponsor-slides/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminSponsorSlideUpdateHandler)))
	mux.Handle("DELETE /api/v1/admin/sponsor-slides/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminSponsorSlideDeleteHandler)))
	mux.Handle("POST /api/v1/admin/sponsor-slides/reorder", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminSponsorSlideReorderHandler)))
	mux.Handle("GET /api/v1/admin/teams", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminTeamsHandler)))
	mux.Handle("POST /api/v1/admin/teams", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTeamImportHandler)))
	mux.Handle("GET /api/v1/admin/teams/{id}", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminTeamHandler)))
	mux.Handle("PATCH /api/v1/admin/teams/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTeamUpdateHandler)))
	mux.Handle("DELETE /api/v1/admin/teams/{id}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTeamDeleteHandler)))
	mux.Handle("POST /api/v1/admin/teams/refresh", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTeamRefreshHandler)))
	mux.Handle("POST /api/v1/admin/teams/wpa-keys", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTeamWpaKeysHandler)))
	mux.Handle("GET /api/v1/admin/jobs/{id}", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminJobHandler)))
	mux.Handle("GET /api/v1/admin/judging-schedule", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminJudgingScheduleHandler)))
	mux.Handle("POST /api/v1/admin/judging-schedule", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminJudgingScheduleGenerateHandler)))
	mux.Handle("DELETE /api/v1/admin/judging-schedule", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminJudgingScheduleClearHandler)))
	mux.Handle("GET /api/v1/admin/settings/{section}", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminSettingsHandler)))
	mux.Handle("PATCH /api/v1/admin/settings/{section}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminSettingsUpdateHandler)))
	mux.Handle("POST /api/v1/admin/database/backups", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminDatabaseBackupHandler)))
	mux.Handle("POST /api/v1/admin/database/restore", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminDatabaseRestoreHandler)))
	mux.Handle("DELETE /api/v1/admin/tournament-data/{type}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminTournamentDataClearHandler)))
	mux.Handle("POST /api/v1/admin/publishing/{resource}", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AdminPublishingHandler)))
	mux.Handle("GET /api/v1/admin/match-play/matches", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminMatchPlayMatchesHandler)))
	mux.Handle("GET /api/v1/admin/referee/fouls", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AdminRefereeFoulsHandler)))
	mux.Handle("POST /api/v1/admin/referee/commands", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1RefereeCommandHandler)))
	mux.Handle("GET /api/v1/admin/alliance-selection/bootstrap", web.apiV1AdminRead(http.HandlerFunc(web.apiV1AllianceSelectionControlBootstrapHandler)))
	mux.Handle("POST /api/v1/admin/alliance-selection/commands", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1AllianceSelectionCommandHandler)))
	mux.Handle("GET /api/v1/admin/match-play/bootstrap", web.apiV1AdminRead(http.HandlerFunc(web.apiV1MatchPlayBootstrapHandler)))
	mux.Handle("POST /api/v1/admin/match-play/commands", web.apiV1AdminMutation(http.HandlerFunc(web.apiV1MatchPlayControlCommandHandler)))
	mux.Handle("GET /api/v1/admin/scoring/{position}/bootstrap", web.apiV1AdminRead(http.HandlerFunc(web.apiV1ScoringPanelBootstrapHandler)))
	mux.Handle("GET /api/v1/admin/referee/bootstrap", web.apiV1AdminRead(http.HandlerFunc(web.apiV1RefereePanelBootstrapHandler)))
	mux.Handle("GET /api/v1/admin/field-testing/bootstrap", web.apiV1AdminRead(http.HandlerFunc(web.apiV1FieldTestingBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/queueing/matches", web.apiV1Middleware(http.HandlerFunc(web.apiV1QueueingDisplayMatchesHandler)))
	mux.Handle("GET /api/v1/displays/announcer/match", web.apiV1Middleware(http.HandlerFunc(web.apiV1AnnouncerDisplayMatchHandler)))
	mux.Handle("GET /api/v1/displays/announcer/score", web.apiV1Middleware(http.HandlerFunc(web.apiV1AnnouncerDisplayScoreHandler)))
	mux.Handle("GET /api/v1/displays/queueing/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1QueueingBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/announcer/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1AnnouncerBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/audience/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1AudienceBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/alliance-station/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1AllianceStationBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/wall/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1WallBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/unpicked/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1UnpickedBootstrapHandler)))
	mux.Handle("GET /api/v1/displays/field-monitor/bootstrap", web.apiV1Middleware(http.HandlerFunc(web.apiV1FieldMonitorBootstrapHandler)))
	mux.HandleFunc("GET /api/v1/streams/displays/queueing", web.apiV1QueueingStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/announcer", web.apiV1AnnouncerStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/audience", web.apiV1AudienceStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/alliance-station", web.apiV1AllianceStationStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/wall", web.apiV1WallStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/unpicked", web.apiV1UnpickedStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/displays/field-monitor", web.apiV1FieldMonitorStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/admin/alliance-selection", web.apiV1AllianceSelectionControlStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/admin/match-play", web.apiV1MatchPlayStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/admin/scoring/{position}", web.apiV1ScoringPanelStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/admin/referee", web.apiV1RefereePanelStreamHandler)
	mux.HandleFunc("GET /api/v1/streams/admin/field-testing", web.apiV1FieldTestingStreamHandler)
}

func (web *Web) apiV1AdminRead(next http.Handler) http.Handler {
	return web.apiV1Middleware(web.apiV1RequireAdmin(next))
}

func (web *Web) apiV1AdminMutation(next http.Handler) http.Handler {
	return web.apiV1Middleware(web.apiV1RequireAdmin(web.apiV1RequireCsrf(next)))
}

func (web *Web) apiV1Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := uuid.NewString()
		w.Header().Set("X-Request-ID", requestId)
		writer := &apiV1ResponseWriter{ResponseWriter: w}
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("API v1 panic requestId=%s method=%s path=%s: %v", requestId, r.Method, r.URL.Path, recovered)
				if !writer.wroteHeader {
					writeApiV1Error(writer, r, http.StatusInternalServerError, "internal_error", "Internal server error.", nil)
				}
			}
		}()

		ctx := context.WithValue(r.Context(), apiV1RequestIdKey, requestId)
		next.ServeHTTP(writer, r.WithContext(ctx))
	})
}

func writeApiV1Data(w http.ResponseWriter, r *http.Request, statusCode int, data any, version *int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(apiV1Response{
		Data: data,
		Meta: apiV1Meta{RequestId: apiV1RequestId(r), Version: version},
	}); err != nil {
		log.Printf("API v1 response encoding error requestId=%s: %v", apiV1RequestId(r), err)
	}
}

func writeApiV1Error(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	code string,
	message string,
	fields map[string]string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(apiV1ErrorResponse{
		Error: apiV1ErrorDetail{Code: code, Message: message, Fields: fields},
		Meta:  apiV1Meta{RequestId: apiV1RequestId(r)},
	}); err != nil {
		log.Printf("API v1 error response encoding error requestId=%s: %v", apiV1RequestId(r), err)
	}
}

func decodeApiV1Json(w http.ResponseWriter, r *http.Request, destination any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxApiV1RequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("request body must contain a single JSON value")
		}
		return fmt.Errorf("request body must contain a single JSON value: %w", err)
	}
	return nil
}

func apiV1RequestId(r *http.Request) string {
	requestId, _ := r.Context().Value(apiV1RequestIdKey).(string)
	return requestId
}

func apiV1PathId(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, fmt.Errorf("ID must be a positive integer")
	}
	return id, nil
}
