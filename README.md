# relay-engine

Go service for Relay's performance-critical decisioning: disposition scoring,
Rescue geo/TTL + decay pricing, and next-owner matching (incl. demand-weighting
and Pair Rescue). Called by `relay-api` over HTTP. See `relay-dev/docs/plan.md`
§5 (contracts) and §7 (specs).

## Routes (T0 = health live; rest are stubs mirroring §5)

| Method | Path | Tier | Task |
|---|---|---|---|
| GET | `/health` | T0 | engine-skeleton |
| POST | `/disposition/score` | T1 | engine-disposition (+ T2 demand-weight) |
| POST | `/match/rescue` | T1 | engine-rescue-ttl (+ decay pricing) |
| POST | `/match/wishlist` | T1 | engine-match-vector |
| POST | `/match/pair-rescue` | T2 | engine-pair-rescue |

## Run locally

```bash
go mod tidy        # fetch chi, write go.sum
go run .           # listens on :8002
```

```bash
curl localhost:8002/health
```

## Docker

```bash
docker build -t relay-engine .
docker run -p 8002:8002 relay-engine
```

## Layout

```
main.go                       # chi router + route wiring
internal/config/config.go     # env-driven config
internal/handlers/handlers.go # health + endpoint stubs
```
