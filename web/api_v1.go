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
	mux.Handle("GET /api/v1/event", web.apiV1Middleware(http.HandlerFunc(web.apiV1EventHandler)))
	mux.Handle("GET /api/v1/session", web.apiV1Middleware(http.HandlerFunc(web.apiV1SessionHandler)))
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
