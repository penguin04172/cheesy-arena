// Copyright 2026 Team 254. All Rights Reserved.
//
// Structured audit logging for API mutations.

package web

import (
	"log"
	"net/http"
)

func (web *Web) logApiV1Audit(r *http.Request, action string, target string, result string) {
	actor := "anonymous"
	if web.arena.EventSettings.AdminPassword == "" {
		actor = "admin(auth-disabled)"
	} else if session := web.getUserSessionFromCookie(r); session != nil {
		actor = session.Username
	}
	log.Printf(
		"API audit requestId=%s actor=%q action=%q target=%q result=%q",
		apiV1RequestId(r), actor, action, target, result,
	)
}
