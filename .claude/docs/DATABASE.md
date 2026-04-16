# Database Layer

This doc describes how persistence works in Dispatch. The layer is
intentionally narrow: a single Go interface, two backing
implementations (in-memory and Postgres), and three decorator
wrappers. The in-memory store keeps the simulation fast and lets
tests run in-process; the Postgres store adds durability when
`DISPATCH_DATABASE_URL` is set.

For testing-specific guidance see [docs/TESTING.md](../../docs/TESTING.md).

## Architecture

Every handler in Dispatch holds one abstraction: the `database.Store`
interface defined in [internal/database/store.go](../../internal/database/store.go).
At runtime that interface is a stack of decorators wrapping an
in-memory backing implementation:

```
handler → dbauthz → dbaudit → dbmetrics → memstore OR pgstore
            RBAC     audit     latency       backing
```

The wrappers compose at startup in [cmd/dispatchd/main.go](../../cmd/dispatchd/main.go).
Every handler receives the dbauthz-wrapped store and cannot reach past
it — there is no API for unwrapping. That shape is what makes the
authz layer load-bearing instead of decorative: a future endpoint
that forgets to authorize a call has nowhere to fall through, because
the underlying store is unreachable except through the wrapper chain.

## Backing store selection

Set `DISPATCH_DATABASE_URL` to a Postgres connection string to use
the durable Postgres backing. When unset, dispatchd falls back to
the in-memory store — same behavior as before, no external deps
needed for the simulation.

```
# In-memory (default):
go run ./cmd/dispatchd

# Postgres:
DISPATCH_DATABASE_URL="postgres://dispatch:dispatch@localhost:5432/dispatch?sslmode=disable" \
  go run ./cmd/dispatchd
```

The pgstore runs embedded migrations on startup via a minimal
`schema_migrations` table. Migration SQL lives in
`internal/database/pgstore/migrations/` and is embedded with
`//go:embed`.

## Components

### The interface — `internal/database/store.go`

The `Store` interface is the single chokepoint. Its methods mirror
the shape of a sqlc-generated querier:

* Every method takes `context.Context` as the first argument.
* Every method takes a typed `Params` struct for arguments.
* Every method returns value types, never live pointers.

That uniformity is what lets the decorators wrap each call
generically. The Params/return shape is also what would let a future
sqlc swap drop in cleanly.

Value types in the same file: `Experiment`, `Sample`, `Device`,
`APIKey`, `AuditEntry`, plus the enums `Status`, `Decision`, `Group`,
`SubjectType`. JSON tags match the field names the dashboard already
expects, so the API contract is stable across the refactor.

### The backing — `internal/database/memstore/memstore.go`

In-memory, single mutex, maps and slices. Maps are keyed by ID for
entities with primary-key lookups (experiments, devices, api_keys);
slices are used for entities that are typically scanned with filters
(samples, audit_log, metric_directions). The store does not enforce
authorization or write audit entries — that's the wrappers' job.

The private `experimentRow` / `sampleRow` / `apiKeyRow` types embed
the public value types. They exist so a future change can carry
running aggregates (Welford, ring buffers) on the row without
breaking the public type's JSON contract. Today they're just
embeddings.

`AnalyzeExperiment` is the one method with non-trivial logic:
it pivots the flat sample slice into per-metric, per-group buckets,
calls `analysis.Analyze` per metric, and sets the experiment's
`Decision` based on whether any verdict was a regression. The pivot
is O(n) over samples for the experiment, which is fine at simulation
scale — if it ever isn't, switch to running aggregates updated inside
`InsertSample`.

### The Postgres backing — `internal/database/pgstore/`

Drop-in replacement for memstore. Uses pgx v5 connection pool,
same ID scheme (`exp-`, `smp-`, `key-`, `aud-` prefixes + random
hex), same validation logic, same error sentinels. The decorator
stack and all handlers are unchanged.

`pgstore.go` — the `Store` implementation. Each method maps to
straightforward SQL. `AnalyzeExperiment` runs inside a transaction
with `SELECT ... FOR UPDATE` to prevent concurrent analysis of the
same experiment.

`migrate.go` — embedded migration runner. SQL files in
`pgstore/migrations/` are applied in lexicographic order via a
`schema_migrations` tracking table. No external migration tool
dependency.

### The wrappers

Each wrapper implements the full `Store` interface and holds an
`inner database.Store`. They compose by being constructed inside-out
in `main()`:

```go
backing := memstore.New()
metrics := dbmetrics.New(backing)
audited := dbaudit.New(metrics, log)
store   := dbauthz.New(audited)
```

Order matters. Authz is outermost so denied calls don't get audited
or counted (a denied call shouldn't show up as a "successful insert"
in metrics). Metrics is innermost so it sees the actual store latency
without wrapper overhead in the measurement.

#### `dbauthz` — RBAC at the boundary

[internal/database/dbauthz/dbauthz.go](../../internal/database/dbauthz/dbauthz.go)

Three principals:

* `SubjectOperator` — human dashboard user. Allowed everything
  except inserting samples directly (operators must not be able to
  fabricate data, since that would let them hide a regression by
  forging a clean sample stream).
* `SubjectAgent` — device-side dispatch-agent. Restricted to
  pushing samples for its own `DeviceID` and reading its own device
  row. Cannot list other devices, mutate experiments, or read the
  audit log.
* `SubjectSystem` — bootstrap escape hatch. Bypasses all checks.
  Used in exactly two places: (1) `bootstrap()` mints the first
  operator key before any real subject exists, and (2) the auth
  middleware looks up an API key by plaintext, which is the very
  call that establishes the real subject. **Must not appear on any
  context that reached a request handler.**

Every guard in the file has an inline comment explaining **why** it
exists. Per [AGENTS.md](../../AGENTS.md), removing one would silently
change access — the comment is the breadcrumb that keeps a future
reader (human or LLM) from "cleaning up" a check they don't
understand.

#### `dbaudit` — audit log at the boundary

[internal/database/dbaudit/dbaudit.go](../../internal/database/dbaudit/dbaudit.go)

Writes one `AuditEntry` per mutating operator action: experiment
create/analyze/hold/promote, set metric direction, device create,
api key create/revoke. Reads are not audited (too noisy). Sample
inserts are not audited (too high-volume, and they're already
attributable through the agent's API key on the underlying sample
row).

Audit entries are written **after** the underlying op succeeds, so
a failed op doesn't leave a phantom audit row. If the audit insert
itself fails, the underlying op still returns success: a missing
audit row is bad, but failing the user's request because the audit
log misbehaved is worse. The failure is logged via slog so a
silent audit log can be detected.

The wrapper does **not** audit the `InsertAuditEntry` call itself,
which would recurse forever.

#### `dbmetrics` — per-method observability

[internal/database/dbmetrics/dbmetrics.go](../../internal/database/dbmetrics/dbmetrics.go)

Per-method count, error count, and a fixed-size ring buffer
(`ringSize = 1024`) of recent durations. `Snapshot()` computes
sorted p50/p95/p99 from the ring at call time and returns it as
JSON, exposed via `/api/admin/metrics`. No Prometheus client
dependency.

If the simulation ever needs longer-window aggregates, replace the
ring with HDR Histogram or t-digest. Don't add Prometheus client_golang
without first asking whether the JSON endpoint is enough — the SRE
story is "I built the observability layer, not pulled in the
standard one."

## Entities

The current set, all defined in [store.go](../../internal/database/store.go):

| Entity | Backing | Purpose |
|---|---|---|
| `Experiment` | map by ID | One deploy validation. Holds canary/control device IDs, status, decision, results. |
| `Sample` | flat slice | One metric value pushed by an agent. Append-only per AGENTS.md ("never mutate historical snapshot data"). |
| `MetricDirection` | flat slice | Tells `analysis.Analyze` whether higher or lower is "good" for a given metric on a given experiment. Stored separately from the experiment so the analyzer can scan all metrics in one place. |
| `Device` | map by ID | One fleet device. Required for agent authentication — the dbauthz wrapper checks the agent's `DeviceID` against the `device_id` it's trying to write. |
| `APIKey` | map by ID | One credential. Either an operator key (no scope) or an agent key (scoped to one `DeviceID`). Plaintext is shown once at creation and never persisted; the row stores only the SHA-256 hash. |
| `AuditEntry` | flat slice | One operator action. Newest-first for `ListAuditEntries`, since security logs are read most-recent first. |

When you add a new entity, ask which container fits. Map by ID is
cheaper for primary-key lookups; flat slices are cheaper to scan
when you frequently filter by non-PK fields (and they pivot to a
SQL table cleanly if/when we move to Postgres).

## The operator vs agent model

Dispatch has exactly two trust levels in the distributed system:

* **Operator** — authenticated user on the dashboard. Triggers
  deploys, promotes, holds, manages keys. API key is unscoped.
* **Device agent** — runs on each fleet device. Pushes metrics for
  its own device. Reads its own config. Nothing else.

The dbauthz wrapper is the **only** place this distinction is
enforced. Handlers don't check roles. Middleware doesn't check
roles. Even the simulation respects it — `runScenario` constructs
per-device agent contexts mid-request specifically so dbauthz fires
the same way it would for real device traffic.

This is the principle of least privilege from AGENTS.md applied at
the data-access boundary instead of at the HTTP boundary, which is
the load-bearing reason to do it. Two reasons HTTP-level checks
are weaker:

1. A new endpoint can forget to call the check.
2. A method that runs from inside server code (like the simulation)
   bypasses HTTP entirely.

The dbauthz wrapper catches both, because it sits at the only chokepoint
both paths share.

## Recipes

### Adding a new Store method

Follow this order. Skipping a step will break the build, which is
the type system enforcing the architecture — that's a feature.

1. Add the method to `Store` in [store.go](../../internal/database/store.go).
2. Define any new param/return types in the same file.
3. Implement it in [memstore](../../internal/database/memstore/memstore.go)
   under the appropriate section.
4. Add a passthrough in [dbmetrics](../../internal/database/dbmetrics/dbmetrics.go)
   following the pattern: `start := time.Now(); … s.record("MethodName", start, err)`.
5. Decide whether the new method is mutating. If yes, add a wrapper
   in [dbaudit](../../internal/database/dbaudit/dbaudit.go) that
   calls `s.audit(ctx, "verb.noun", "target_type", target_id, metadata)`
   on success. If no, add a plain passthrough.
6. Add an authz guard in [dbauthz](../../internal/database/dbauthz/dbauthz.go).
   **Every guard must have an inline comment explaining why it exists.**
   This is non-negotiable per AGENTS.md.
7. Add a row to the dbauthz test matrix in
   [docs/TESTING.md](../../docs/TESTING.md) covering each subject
   type. A method without a corresponding authz test is a reviewable
   defect.
8. If exposed via HTTP, add a handler in [main.go](../../cmd/dispatchd/main.go)
   and one end-to-end test for the success path and the 403 path.

### Adding a new entity

1. Define the value type in [store.go](../../internal/database/store.go)
   alongside the existing entities. Keep the JSON tags matching the
   API contract you want.
2. Add the storage container to memstore. Map by ID for primary-key
   lookups, slice for scan-with-filter.
3. Add CRUD methods to the `Store` interface and follow the
   "Adding a new Store method" recipe for each one.
4. If the entity has its own access-control rule beyond
   "operators can do anything", add it explicitly to dbauthz.go and
   write the test that proves the boundary.
5. If the entity is auditable, decide which mutations should produce
   audit entries. Don't audit reads. Don't audit high-volume writes.

### Adding an HTTP endpoint that does NOT route through the Store

Almost every handler in dispatchd takes a `database.Store` argument
and gets its authorization for free, because dbauthz wraps every
Store method. Some handlers don't have a Store call to make —
notably observability endpoints like `/api/admin/metrics`, which
read from a concrete wrapper type (`*dbmetrics.Store`) instead of a
data row. Those handlers sit **outside** the dbauthz wrapper.

If you don't add an explicit guard, any authenticated principal —
including device agents — can hit them. That's exactly the
cross-tenant leak the layered architecture is supposed to prevent.

The rule: **every handler that does not call a Store method must
call `database.RequireOperator(ctx)` (or a stricter check) before
doing any work.**

```go
func adminMetrics(metrics *dbmetrics.Store) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Required: this handler reads dbmetrics state directly,
        // not via the Store interface, so dbauthz can't gate it.
        if err := database.RequireOperator(r.Context()); err != nil {
            writeStoreErr(w, err)
            return
        }
        writeJSON(w, http.StatusOK, metrics.Snapshot())
    }
}
```

`RequireOperator` lives next to `SubjectFromContext` in
[subject.go](../../internal/database/subject.go). It rejects missing
subjects, the system subject (which must never escape the auth
middleware), and any non-operator principal.

If you find yourself needing finer-grained control here ("agents
allowed, but only for their own device"), don't reinvent dbauthz in
the handler. Either move the underlying data behind a Store method
so the existing wrapper layer fires, or write a sibling helper next
to `RequireOperator` and document the rule that justifies it.

### Adding a new wrapper

You probably won't need one, but if you do (e.g., a `dbcache` for
read-heavy methods):

1. Create `internal/database/dbcache/dbcache.go`.
2. Define `type Store struct { inner database.Store; … }` and a
   `New(inner)` constructor.
3. Implement every method in `Store`. Most are passthroughs; the
   few you actually cache get custom logic.
4. Wire it into `main()` in the right position. The order constraint:
   authz must stay outermost (denied calls should never reach
   anything else), audit must stay above metrics (so a failed audit
   write still gets timed). A cache between authz and audit would
   let cache hits skip the audit, which is usually what you want
   for reads. Document the ordering choice in the package doc.
5. The compiler will tell you if you missed a method.

### Storing data on the experiment row vs. its own entity

When you find yourself wanting to store something experiment-related,
ask: do I ever need to query this across experiments?

* **No** — put it on the `Experiment` value type. Example: `Status`,
  `Decision`, `HoldReason`. These are only read in the context of
  one experiment.
* **Yes** — give it its own slice. Example: `MetricDirection`. We
  don't currently query directions across experiments, but we might
  ("which deploys had a `cpu_usage` direction set?"), and it pivots
  to a SQL table cleanly.

Don't preemptively split everything into its own entity. The
`Results` slice lives on the experiment row even though it could be
its own entity, because we have never needed cross-experiment
analysis-result queries and don't anticipate one.

### Adding running aggregates instead of full sample history

Today `AnalyzeExperiment` walks the full sample slice. If that ever
gets slow (it won't at simulation scale, but the question is fair),
the right shape is per-metric Welford aggregates updated inside
`InsertSample`:

```go
type metricAgg struct {
    n          int
    mean       float64
    sumSquares float64 // for Welford's running variance
}
```

Carry one `map[metricName]map[group]*metricAgg` per experiment row.
`AnalyzeExperiment` reads the aggregates instead of walking samples.
The `analysis` package would need a sibling function that takes
`(mean, variance, n)` instead of `[]float64`. The `Sample` slice can
stay around for audit/replay and never be scanned in the hot path.

Don't do this work until there's an actual measurement showing the
walk is the bottleneck. Premature optimization. The current shape
is fine for thousands of samples per experiment.

## Key generation and hashing

API key plaintext is generated with `crypto/rand` (32 bytes,
hex-encoded → 64 chars, 256 bits of entropy). The store hashes the
plaintext with SHA-256 and stores only the hash. Plaintext is
returned exactly once from `CreateAPIKey` and never persisted; the
caller is responsible for showing it to the user.

The auth middleware looks keys up by hashing the incoming
`X-API-Key` header value and matching against stored hashes. This
is constant-time-ish enough for a simulation — if you move to
production, switch to a constant-time comparison and a slower
hash like Argon2 or bcrypt, and add per-key rate limiting.

## Sentinel errors

[internal/database/errors.go](../../internal/database/errors.go) defines
four sentinels that wrappers and the backing implementation return:

* `ErrNotFound` → HTTP 404
* `ErrUnauthorized` → HTTP 403
* `ErrInvalidArgument` → HTTP 400
* `ErrConflict` → HTTP 409

Handlers map them via `writeStoreErr` in
[cmd/dispatchd/main.go](../../cmd/dispatchd/main.go). When you return
an error from a Store method, wrap one of these with `fmt.Errorf` so
the HTTP layer maps it correctly. Custom errors get 500'd, which is
the right behavior — they're loud enough to investigate.

## Open questions

Things this layer doesn't yet do because the codebase doesn't yet
need them. If you find yourself wanting one, that's the signal to
build it.

* **Persistence across restarts.** The store is in-memory. Process
  restart loses all state. Acceptable for a simulation; not for
  production.
* **Pub/sub for live updates.** Real-time fan-out (SSE / WebSocket)
  is a separate concern. The layered Store doesn't preclude it but
  doesn't provide it. The right shape is a sibling `broker` package
  that handlers call after a successful Store mutation.
* **Soft delete.** No entity is currently deletable. When the first
  one is, AGENTS.md is explicit: ask the user whether to soft-delete
  or hard-delete each time. Don't assume.
* **Pagination.** `ListAuditEntries` has a `Limit` field but no
  cursor. Fine for the current dashboard; needs revisiting if the
  audit log gets a UI with infinite scroll.
* **Real database backing.** Out of scope for the simulation. When
  it's needed, the swap goes in `internal/database/sqlstore` and the
  `Store` interface should not change.
