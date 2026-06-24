# Auction Project

> This project is part of the [FullCycle](https://fullcycle.com.br/) learning program (Pós-Graduação).

## Table of Contents

- [What the System Does](#what-the-system-does)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [API Documentation](#api-documentation)
- [Running Tests](#running-tests)
- [Service Management](#service-management)
- [Requirements Coverage (FullCycle)](#requirements-coverage-fullcycle)
- [License](#license)

## What the System Does

A REST API for online auctions. Users open auctions for products and place bids on the ones that are still
open. The feature added by this challenge is **automatic auction closing**: an auction created with status
`Active` is automatically moved to `Completed` once its configured duration elapses — no manual action and
no external scheduler required.

Closing is handled by two complementary mechanisms:

- **Scheduled close** — when an auction is created, a goroutine is scheduled to close it after
  `AUCTION_INTERVAL`.
- **Background monitor** — a single goroutine periodically sweeps the database and closes any `Active`
  auction whose time has already elapsed. This is a safety net for auctions whose scheduled close was lost
  (for example, after a process restart), since the database — not in-memory state — is the source of truth.

Both paths update the auction with a filter on `status = Active`, which makes the operation idempotent and
safe under concurrency.

## Technology Stack

- **Go** 1.26.4
- **Gin** — HTTP web framework
- **MongoDB** — persistence (official `mongo-driver`)
- **Uber Zap** — structured logging
- **go-playground/validator v10** — request validation
- **Docker / Docker Compose** — local environment
- **Testcontainers for Go** — ephemeral MongoDB for integration tests

## Project Structure

```
cmd/auction/                entrypoint (main.go), manual dependency wiring
internal/
  apperr/                   application error types
  config/                   env parsing
  entity/                   domain entities (auction, bid, user)
  usecase/                  application use cases (auction, bid, user)
  infra/
    api/web/
      controller/           Gin controllers (auction, bid, user)
      httperr/              HTTP error responses
      validation/           request validation
    database/
      mongodb/              MongoDB connection factory
      auction/              repository + background auction closer
      bid/                  repository
      user/                 repository
  observability/logger/     Zap logger setup
docs/CHALLENGE.md           original challenge brief (pt-BR)
```

The automatic-close logic lives in `internal/infra/database/auction/create.go`.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- (optional, for running tests/builds locally) Go 1.26+

## Quick Start

Create your local `.env` from the template, then bring up the API and MongoDB with Docker Compose:

```bash
cp .env.example .env
docker compose up --build
```

The API will be available at `http://localhost:8080`.

> The real `.env` is git-ignored; `.env.example` is the versioned template. The `cp` step is required
> because Docker Compose reads `.env` via `env_file`.

### Environment variables

Configured in `.env` at the project root (loaded by the app and by both containers):

| Variable | Example | Description |
|----------|---------|-------------|
| `AUCTION_INTERVAL` | `20s` | Auction duration. After this, an auction is automatically closed. |
| `AUCTION_CLOSER_INTERVAL` | `10s` | How often the background monitor sweeps for expired auctions. |
| `BATCH_INSERT_INTERVAL` | `20s` | Bid batch flush interval. |
| `MAX_BATCH_SIZE` | `4` | Maximum bids per batch. |
| `MONGO_INITDB_ROOT_USERNAME` | `admin` | MongoDB root user (created on first run). |
| `MONGO_INITDB_ROOT_PASSWORD` | `admin` | MongoDB root password. |
| `MONGODB_URL` | `mongodb://admin:admin@mongodb:27017/auctions?authSource=admin` | Connection string. |
| `MONGODB_DB` | `auctions` | Database name. |

> Values accept any Go duration string (e.g. `20s`, `1m`, `1m30s`). If `AUCTION_INTERVAL` is missing or
> invalid the app falls back to `5m`; `AUCTION_CLOSER_INTERVAL` falls back to `10s`.

### See the automatic close in action

With the default `AUCTION_INTERVAL=20s`:

```bash
# 1. Create an auction
curl -X POST http://localhost:8080/auction \
  -H 'Content-Type: application/json' \
  -d '{"product_name":"Vintage Clock","category":"Decor","description":"A beautiful vintage wall clock from 1950","condition":1}'

# 2. List active auctions and copy the "id"
curl "http://localhost:8080/auction?status=0"

# 3. Wait > 20s, then fetch the auction by id — "status" is now 1 (Completed)
curl http://localhost:8080/auction/<auction-id>
```

## API Documentation

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auction` | Create an auction. Body: `product_name`, `category`, `description`, `condition` (`0`, `1` or `2`). |
| `GET` | `/auction?status=0` | List auctions by status (`0` = Active, `1` = Completed). `status` is required. |
| `GET` | `/auction/:auctionId` | Fetch an auction by id. |
| `GET` | `/auction/winner/:auctionId` | Fetch the winning bid for an auction. |
| `POST` | `/bid` | Place a bid. Body: `user_id`, `auction_id`, `amount`. |
| `GET` | `/bid/:auctionId` | List bids for an auction. |
| `GET` | `/user/:userId` | Fetch a user by id. |

`AuctionStatus`: `0 = Active`, `1 = Completed`. `ProductCondition`: `1 = New`, `2 = Used`, `3 = Refurbished`.

## Running Tests

Run unit tests (no external dependencies):

```bash
go test -race ./...
```

Run integration tests (requires a running Docker daemon — Testcontainers pulls `mongo:7` automatically):

```bash
go test -race -tags integration ./internal/infra/database/...
```

The integration suite covers:

- **Scheduled close** — `TestCreateAuction_ClosesAutomaticallyAfterInterval`: a created auction starts
  `Active` and transitions to `Completed` after `AUCTION_INTERVAL`.
- **Background monitor closes expired auction** — `TestStartAuctionCloser/closes_expired_auction`: an
  already-expired `Active` auction inserted directly into the database is closed by the monitor.
- **Completed auction is not reopened** — `TestStartAuctionCloser/completed_auction_is_not_reopened`:
  the `status = Active` filter makes updates idempotent; a `Completed` auction is never modified.
- **No expired auctions does nothing** — `TestStartAuctionCloser/no_expired_auctions_does_nothing`:
  an auction with a future timestamp is never closed by the monitor.
- **Context cancellation stops monitor** — `TestStartAuctionCloser/stops_on_context_cancel`: after the
  monitor context is cancelled, newly expired auctions are not closed.
- **Concurrent closers are idempotent** — `TestCreateAuction_ConcurrentClosers_Idempotent`: when the
  scheduled closer and the background monitor race to close the same auction, the final status is
  `Completed` without oscillation.

## Service Management

```bash
docker compose up --build      # build and start (foreground)
docker compose up --build -d   # build and start (detached)
docker compose logs -f app     # follow application logs
docker compose ps              # list services
docker compose down            # stop and remove containers/network
docker compose down -v         # also remove the MongoDB data volume
```

---

## Requirements Coverage (FullCycle)

> Maps the implementation to the original challenge brief — see **[docs/CHALLENGE.md](docs/CHALLENGE.md)** for the full
> assignment (pt-BR).

### Required by the challenge

- [x] A function that computes the auction duration from environment variables — `getAuctionInterval()` in
  `internal/infra/database/auction/create.go` (reads `AUCTION_INTERVAL`).
- [x] A goroutine that detects expired auctions and closes them via update —
  `StartAuctionCloser` / `closeExpiredAuctions` (background monitor), complemented by the per-auction
  `scheduleAuctionClose` / `closeAuction`.
- [x] A test validating that closing happens automatically — integration test suite with Testcontainers,
  covering scheduled close, background monitor, idempotency, context cancellation, and concurrency.
- [x] Documentation on how to run in dev + Docker/Docker Compose — this README.

### Beyond the base

- Hybrid closing strategy (scheduled close + background safety-net monitor) for durability across restarts.
- Idempotent updates filtered by `status = Active` and a mutex to guard concurrent closes.

---

## License

This project is licensed under the MIT License.