# Yabbi Junior Go — test task context

## Goal

Build a small ad exchange / auction request router in Go.

The service receives an auction request, selects DSP partners that match pretargeting rules, sends the request to all matched DSPs in parallel under one shared timeout, and returns which DSPs responded successfully.

The task is intentionally simplified: no need to parse real bid responses or choose an auction winner.

## Terminology

### AuctionRequest
One opportunity to show an ad.

Example:

```json
{
  "request_id": "3f0a1c9e-2b1d-4a8f-9c11-7e6b2d0a55f1",
  "country": "RU",
  "device_type": "mobile",
  "bid_floor": 1.5,
  "categories": ["news", "sport"]
}
```

### Partner / DSP
A DSP connected to our ad exchange.

Example config:

```json
{
  "uuid": "123e4567-e89b-12d3-a456-426655440000",
  "name": "DSP Alpha",
  "endpoint": "http://localhost:9001/bid",
  "is_enabled": true,
  "countries": ["RU", "KZ"],
  "device_types": ["mobile", "desktop"],
  "min_bid_floor": 0.5,
  "blocked_categories": ["gambling"]
}
```

Treat `partner` as "DSP partner connected to our AdEx", not as a third separate entity.

## Endpoint

```text
POST /auction
```

Handler responsibilities:
- decode JSON
- validate input
- map DTO to internal `AuctionRequest`
- call `RunAuction`
- map result/errors to HTTP
- encode JSON response

No business logic in handler.

## Pretargeting rules

A partner matches only if all conditions are true:

1. `partner.is_enabled == true`
2. Country matches `partner.countries`, or `countries` is empty
3. Device matches `partner.device_types`, or `device_types` is empty
4. `request.bid_floor >= partner.min_bid_floor`
5. No request category appears in `partner.blocked_categories`

Pretargeting should be implemented in Go, not hidden in SQL.

Suggested function:

```go
func MatchPartner(req AuctionRequest, partner Partner) bool
```

or:

```go
func FilterPartners(req AuctionRequest, partners []Partner) []Partner
```

This is the highest-priority area for table-driven unit tests.

## Chosen architecture

Simple three-layer architecture:

```text
handler -> usecase -> storage
              |
              -> DSPClient
```

### Handler
Transport only.

### Usecase
`RunAuction` is responsible for:

1. Get partners from storage
2. Filter them using pretargeting
3. Create one shared auction timeout
4. Call matched DSPs in parallel
5. Ignore individual DSP failures/timeouts
6. Collect successful results
7. Build `AuctionResult`

Concurrency belongs in usecase because it is part of the application scenario.

### Storage
PostgreSQL.

Minimal interface:

```go
type PartnerStorage interface {
    GetPartners(ctx context.Context) ([]Partner, error)
}
```

Do not build unnecessary CRUD.

### DSPClient
External dependency of usecase.

Recommended interface:

```go
type DSPClient interface {
    SendBidRequest(
        ctx context.Context,
        partner Partner,
        req AuctionRequest,
    ) error
}
```

`SendBidRequest` is preferred over vague `SendMessage`.

Boundary:

```text
Usecase:
- decides whom to call
- decides calls are parallel
- owns shared timeout
- aggregates results

DSPClient:
- performs one HTTP request to one DSP
- serializes request
- calls partner.Endpoint
- checks HTTP status/error
```

DSPClient itself should not spawn goroutines.

## Concurrency model

Conceptually:

```text
RunAuction
    |
    +--> goroutine --> DSP Alpha
    +--> goroutine --> DSP Beta
    +--> goroutine --> DSP Gamma
    |
    +--> aggregate results
```

One DSP failure must not fail the whole auction.

Example:

```text
DSP Alpha -> success
DSP Beta  -> timeout
DSP Gamma -> success
```

Return Alpha and Gamma as successful.

Important: timeout is shared for the entire fan-out, not restarted for every partner.

Conceptually:

```go
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```

Pass the same derived context to all `SendBidRequest` calls.

Possible tools:
- goroutines
- buffered result channel
- optionally `sync.WaitGroup`

Avoid `errgroup` unless clearly beneficial because a single DSP error should not cancel other requests.

Be careful to avoid goroutine leaks or blocked sends after timeout.

## PostgreSQL decision

PostgreSQL is intentional because the task explicitly says real DB + migrations is a bonus.

Recommended:
- `github.com/jackc/pgx/v5`
- `pgxpool`

Possible schema:

```sql
CREATE TABLE partners (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    countries TEXT[] NOT NULL DEFAULT '{}',
    device_types TEXT[] NOT NULL DEFAULT '{}',
    min_bid_floor DOUBLE PRECISION NOT NULL DEFAULT 0,
    blocked_categories TEXT[] NOT NULL DEFAULT '{}'
);
```

Migrations:

```text
migrations/
  001_create_partners.sql
  002_seed_partners.sql
```

Seed at least 3-4 partners with different configurations.

Pretargeting still stays in Go.

## Expected models

```go
type AuctionRequest struct {
    RequestID  string
    Country    string
    DeviceType string
    BidFloor   float64
    Categories []string
}
```

```go
type Partner struct {
    ID                string
    Name              string
    Endpoint          string
    IsEnabled         bool
    Countries         []string
    DeviceTypes       []string
    MinBidFloor       float64
    BlockedCategories []string
}
```

Exact shape of `AuctionResult` and internal DSP result still needs to be designed.

Expected HTTP response from the task:

```json
{
  "request_id": "3f0a1c9e-2b1d-4a8f-9c11-7e6b2d0a55f1",
  "matched_dsps": ["dsp-alpha", "dsp-gamma"],
  "sent": 2,
  "succeeded": 1,
  "duration_ms": 42
}
```

Need to confirm exact semantics of:
- `matched_dsps`
- `sent`
- `succeeded`
- `duration_ms`

using the original task as source of truth.

## Error handling

Expected behavior:

- invalid input -> 4xx, likely 400
- storage failure -> 500
- one DSP error -> do not fail auction
- one DSP timeout -> ignore that DSP
- no matching partners -> `200 OK`, empty list, clear counts

Do not swallow infrastructure errors blindly.

## Logging

At least:

```text
request_id
number of partners loaded
number filtered out
number matched
number successfully contacted/responded
```

Optional: reason why a partner was filtered.

## Tests

Highest priority: pretargeting.

Recommended table-driven cases:

```text
- partner disabled
- exact country match
- country mismatch
- empty countries -> any country
- exact device match
- device mismatch
- empty device_types -> any device
- bid floor equal to min
- bid floor greater than min
- bid floor lower than min
- blocked-category intersection
- no blocked-category intersection
- multiple conditions combined
```

If time remains:
- usecase with fake storage
- usecase with fake DSPClient
- partial DSP failure
- timeout case
- no matched partners
- all DSPs successful

## Optional improvements

Original task says these are bonuses:

- real DB + migrations
- Dockerfile / docker-compose
- metrics
- healthcheck
- graceful shutdown
- env/flags configuration

Current priority:
- PostgreSQL + migrations: yes
- Dockerfile / docker-compose: probably yes
- graceful shutdown: yes if core complete
- env config: likely yes
- healthcheck: only if time remains
- metrics: low priority

Do not sacrifice core correctness for optional infrastructure.

## Time constraints

Deadline: 2 days.

Rough estimate:
- MVP: 3-4 hours
- clean submission: 5-7 hours
- with extras: 8-10 hours

Priority order:

```text
1. models/contracts
2. PostgreSQL storage
3. pretargeting
4. pretargeting tests
5. DSPClient
6. RunAuction concurrency
7. handler
8. manual/integration check
9. README
10. Docker / graceful shutdown / healthcheck if time remains
```

## Design principles

1. Code must be explainable line by line in an interview.
2. Avoid unnecessary abstractions.
3. Prefer explicit Go.
4. Keep HTTP concerns out of usecase.
5. Keep SQL concerns out of usecase.
6. Keep business filtering out of storage.
7. Keep concurrency orchestration in usecase.
8. Keep one-DSP transport concerns in DSPClient.
9. Propagate `context.Context`.
10. Treat partial DSP failure as normal auction behavior.

## Current open questions

### 1. Exact RunAuction contract

Likely:

```go
type AuctionUsecase interface {
    RunAuction(ctx context.Context, req AuctionRequest) (AuctionResult, error)
}
```

Need to define exact `AuctionResult`.

### 2. Internal DSP result

Possible shape:

```go
type DSPResult struct {
    PartnerID string
    Err       error
}
```

Need to decide whether more fields are useful.

### 3. HTTP response semantics

Confirm exact meaning of:
- `matched_dsps`
- `sent`
- `succeeded`
- `duration_ms`

### 4. Outbound DSP request body

Need to decide simplest valid request body:
- forward auction request
- or use a small DSP-specific DTO

### 5. Validation

Need exact rules for:
- `request_id`
- country format / allowed values
- device type
- `bid_floor >= 0`
- categories

Keep validation proportional to the task.

## Instructions for Codex

Use this file as architectural context.

When proposing changes:

1. Preserve the chosen `handler -> usecase -> storage` architecture.
2. Treat `DSPClient` as an external usecase dependency.
3. Do not move pretargeting into SQL unless explicitly discussed first.
4. Do not move DSP fan-out concurrency into handler or DSPClient.
5. Do not add frameworks or abstractions without concrete benefit.
6. Explain tradeoffs before changing an agreed design.
7. Use `net/http` for HTTP as required by the task.
8. PostgreSQL/pgx is intentional.
9. Prefer code that is easy to defend in a technical interview.
10. Focus first on task correctness, then optional production extras.

Next step: design exact interfaces and internal result types for `PartnerStorage`, `DSPClient`, and `AuctionUsecase`, then implement incrementally.
