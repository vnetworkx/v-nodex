# API

## Health
- `GET /healthz`
- `GET /readyz`

## State
- `GET /v1/state`
- `GET /v1/events?limit=&offset=`
- `GET /v1/peers`
- `GET /v1/snapshots`

## Writes
- `POST /v1/events/submit`
- `POST /v1/events/ingest`
- `POST /v1/peers`

## Replay
- `GET /v1/replay`
- `POST /v1/replay`
