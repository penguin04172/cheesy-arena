# API v1 Conventions

This document defines the contract that all Cheesy Arena `/api/v1` endpoints must follow.

## Compatibility

- Existing HTML, `/api/*`, and websocket routes remain available during the migration.
- Version 1 uses explicit DTOs. Database and arena structs must not be serialized directly.
- Additive response fields are backward-compatible. Removing or changing a field type requires a new API version.

## JSON

Successful responses use a `data` member and optional `meta` member.

```json
{
  "data": {},
  "meta": {
    "requestId": "request-id",
    "version": 1
  }
}
```

Errors use a stable machine-readable code.

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Request validation failed.",
    "fields": {
      "field": "reason"
    }
  },
  "meta": {
    "requestId": "request-id"
  }
}
```

- JSON request bodies reject unknown fields and trailing JSON values.
- Empty lists are encoded as `[]` and empty maps as `{}`.
- Times use RFC 3339 with an explicit offset.
- Durations are integer seconds and use a `Sec` suffix.
- Enums use stable lower-case strings rather than Go integer values.

## HTTP status codes

| Status | Meaning |
|---|---|
| 200 | Successful read or update |
| 201 | Resource created |
| 204 | Successful operation with no response body |
| 400 | Malformed request |
| 401 | No valid session |
| 403 | Insufficient permission or invalid CSRF token |
| 404 | Resource not found |
| 409 | State, version, or idempotency conflict |
| 422 | Business validation failed |
| 429 | Rate limit exceeded |
| 500 | Unexpected server error |
| 503 | Required hardware or partner service unavailable |

## Authentication and sensitive data

- Public endpoints never return admin passwords, integration secrets, network passwords, WPA keys, or FTA notes.
- Admin endpoints use the existing server-side session and same-origin cookie authentication.
- State-changing requests require a CSRF token.
- Admin CORS must never use a wildcard origin with credentials.
- Secret settings expose configured state, not the stored value. Updates distinguish keep, replace, and clear.

## Concurrency and commands

- Mutable shared resources return a version and require an expected version on update.
- Version conflicts return HTTP 409 and the current version.
- High-risk commands accept an `Idempotency-Key` header.
- Websocket commands include a command ID and receive an accepted or rejected acknowledgement with the same ID.

## Websocket recovery

- REST bootstrap endpoints are the source of truth for recovery.
- Version 1 websocket events include a per-stream sequence and snapshot version.
- A client that reconnects or detects a sequence gap reloads its bootstrap endpoint.
- Slow consumers are disconnected instead of silently losing state.
