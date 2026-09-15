# API v1 Migration Baseline

Recorded on 2026-09-15 from commit `e86332e402caecc5e7855b88f24f8235101eb0ac` on branch `api`.

## Environment

- Go: `go1.27.1 windows/amd64`
- Full page routes reviewed: 35
- Existing web test files: 33
- Existing API style: unversioned public JSON endpoints plus page-specific websockets and HTML fragments

## Baseline test result

`go test ./...` was run before API implementation changes.

Passing packages:

- game
- led
- model
- network
- partner
- playoff
- plc
- tournament
- websocket

Known baseline failures:

- `field/TestPlcEStopAStop`: multiple expected PLC E-stop/A-stop state assertions differ.
- `web/TestAllianceStationDisplayWebsocket`: expected `realtimeScore`, received `arenaStatus` because bootstrap message order is timing-dependent.
- `web/TestAnnouncerDisplayWebsocket`: bootstrap message presence/order assertions fail for the same notifier ordering behavior.

The API migration must not add new failures. The known failures are tracked separately and should be made deterministic without changing arena behavior before they are used as release gates.

## Phase 0 exit record

- API conventions documented.
- OpenAPI 3.1 draft created for the first two foundation endpoints.
- Existing route and test coverage inventory recorded.
- No runtime behavior changed.
