# Backend Testing Guide

This doc covers how to test the Dispatch backend: what's testable, what
helpers exist (and what should), and recipes for adding tests when you
add new code. It mirrors the shape of Coder's `docs/about/contributing/backend.md`,
but anchored in the architecture Dispatch actually has (a single Go
binary, an in-memory store, no Postgres).

If you just want to run what's already here:

```bash
go test ./...
go vet ./...
go build ./...
```

If you want to know where to write the next test, jump to **Recipes**.

## Architecture refresher

Dispatch's backend is a single Go binary, `dispatchd`. Every handler in
[cmd/dispatchd/main.go](../cmd/dispatchd/main.go) talks to one abstraction:
the `database.Store` interface from
[internal/database/store.go](../internal/database/store.go). At runtime
that interface is a stack of decorators:

```
handler → dbauthz → dbaudit → dbmetrics → memstore
            RBAC     audit     latency      backing
```

This shape is what makes the codebase pleasant to test: every layer can
be tested in isolation against the layer below it, and the entire stack
can be assembled in a test the same way it's assembled in `main()`.
There is no Postgres, no migrations, and no external services — every
test runs in-process with the in-memory store as the backing data.

## Tech stack

Dispatch's backend is intentionally narrow on dependencies. Knowing
which library does what is enough to navigate the codebase.

* [go-chi/chi](https://github.com/go-chi/chi): HTTP router.
* `log/slog`: structured logging (stdlib).
* `crypto/rand` + `crypto/sha256`: API key generation and hashing.
* `math/rand/v2`: simulation random number streams.
* `net/http` + `net/http/httptest`: server and end-to-end test scaffolding.

Notably absent — and on purpose:

* No ORM, no sqlc, no `database/sql`.
* No Prometheus client library: [internal/database/dbmetrics](../internal/database/dbmetrics)
  exposes its own per-method ring-buffer histogram via JSON.
* No OPA: [internal/database/dbauthz](../internal/database/dbauthz)
  uses plain Go switches with mandatory inline comments per check.
* No mock framework: the in-memory `memstore` is the test fixture.

## Repository structure

* [cmd/dispatchd](../cmd/dispatchd): server entrypoint, HTTP handlers,
  auth middleware, bootstrap that mints the operator and per-device
  agent keys on first run.
* [internal/analysis](../internal/analysis): pure statistical core
  (Welch's t-test, Cohen's d, regression verdict). Zero dependencies
  on the rest of the codebase, easiest layer to unit-test.
* [internal/database](../internal/database): the `Store` interface and
  every value type that crosses it. Importing this package and nothing
  else is enough to write a fake Store.
  * [memstore](../internal/database/memstore): in-memory backing
    implementation. Used in `main()` *and* as the test fixture for
    everything above it.
  * [dbauthz](../internal/database/dbauthz): RBAC wrapper. The
    load-bearing layer — every guard has an inline comment explaining
    why it exists, per [AGENTS.md](../AGENTS.md).
  * [dbaudit](../internal/database/dbaudit): audit log wrapper. Writes
    one audit entry per mutating operator action.
  * [dbmetrics](../internal/database/dbmetrics): per-method count,
    error count, and ring-buffer p50/p95/p99 latencies.
* [internal/simulation](../internal/simulation): scenario runner that
  drives the API as if it were a real fleet, switching between operator
  and per-device agent contexts mid-request.

## What's testable, and how

The layered architecture gives you four natural test scopes. Pick the
narrowest one that covers the change.

### 1. Pure analysis (`internal/analysis`)

No I/O, no state. Test against known datasets:

```go
func TestAnalyze_DetectsLidarRegression(t *testing.T) {
    canary  := []float64{91.5, 91.8, 91.3, 91.6, 91.4, 91.7, 91.5}
    control := []float64{98.0, 98.1, 97.9, 98.0, 98.2, 97.8, 98.1}
    r := analysis.Analyze("lidar_accuracy", analysis.HigherIsBetter, canary, control, analysis.DefaultAlpha)
    if r.Verdict != analysis.Regression {
        t.Fatalf("expected regression, got %s", r.Verdict)
    }
    if r.PValue > 0.001 {
        t.Errorf("expected p << 0.001, got %.4f", r.PValue)
    }
}
```

Cover the four verdicts (`Regression`, `Improvement`, `NoChange`,
`InsufficientData`) and the two directions. Edge cases worth a test:
zero variance in both groups, sample sizes below `minSampleSize`, and
statistical significance with effect size below `minEffectSize` (which
must return `NoChange`, not `Regression`).

### 2. The store, end-to-end (`internal/database/memstore`)

Construct a `memstore.New()` and call methods directly. No subject is
needed because memstore doesn't enforce authz — it's the backing layer.
This is the right scope for testing experiment lifecycle:

```go
func TestExperiment_AnalyzeAutoHoldsOnRegression(t *testing.T) {
    ctx := context.Background()
    s := memstore.New()

    exp, _ := s.CreateExperiment(ctx, database.CreateExperimentParams{
        DeployID:       "test",
        CanaryDevices:  []string{"car-1"},
        ControlDevices: []string{"car-2"},
        WindowMinutes:  5,
    })
    _ = s.SetMetricDirection(ctx, database.SetMetricDirectionParams{
        ExperimentID: exp.ID,
        MetricName:   "lidar_accuracy",
        Direction:    analysis.HigherIsBetter,
    })

    // Push 7 canary samples around 91.5, 7 control samples around 98.0.
    insertN(t, s, exp.ID, "car-1", database.GroupCanary, "lidar_accuracy", 91.5, 0.2, 7)
    insertN(t, s, exp.ID, "car-2", database.GroupControl, "lidar_accuracy", 98.0, 0.2, 7)

    out, err := s.AnalyzeExperiment(ctx, exp.ID)
    if err != nil {
        t.Fatal(err)
    }
    if out.Decision != database.DecisionAutoHold {
        t.Fatalf("expected auto_hold, got %s", out.Decision)
    }
}
```

### 3. The dbauthz boundary (`internal/database/dbauthz`)

This is the most important set of tests in the codebase. The only thing
standing between a compromised agent key and another device's data is
the per-method authz table in [dbauthz.go](../internal/database/dbauthz/dbauthz.go).
The test shape is a matrix: every method × every subject type → expect
allow or deny.

Use a table-driven test:

```go
func TestDBAuthz_InsertSample(t *testing.T) {
    cases := []struct {
        name     string
        subject  database.Subject
        deviceID string
        wantErr  error
    }{
        {"operator denied",      database.Subject{Type: database.SubjectOperator, APIKeyID: "k1"}, "car-1", database.ErrUnauthorized},
        {"agent own device ok",  database.Subject{Type: database.SubjectAgent, APIKeyID: "k2", DeviceID: "car-1"}, "car-1", nil},
        {"agent other device",   database.Subject{Type: database.SubjectAgent, APIKeyID: "k3", DeviceID: "car-1"}, "car-2", database.ErrUnauthorized},
        {"missing subject",      database.Subject{}, "car-1", database.ErrUnauthorized},
        {"system bypass",        database.SystemSubject, "car-1", nil},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            store := dbauthz.New(memstore.New())
            // bootstrap the experiment + device so the underlying call
            // doesn't fail for unrelated reasons.
            seedExperiment(t, store, "exp-1", "car-1", "car-2")

            ctx := context.Background()
            if c.subject != (database.Subject{}) {
                ctx = database.WithSubject(ctx, c.subject)
            }
            err := store.InsertSample(ctx, database.InsertSampleParams{
                ExperimentID: "exp-1",
                DeviceID:     c.deviceID,
                MetricName:   "cpu_usage",
                Group:        database.GroupCanary,
                Value:        33.0,
            })
            if !errors.Is(err, c.wantErr) {
                t.Fatalf("got %v, want %v", err, c.wantErr)
            }
        })
    }
}
```

When you add a new Store method, **you must add a row to this table**.
The dbauthz file has a check; the test file proves it. A Store method
without a corresponding authz test is the kind of thing a senior
reviewer will catch and call out.

### 4. End-to-end (`cmd/dispatchd`)

Use `net/http/httptest.NewServer` to spin up the whole stack, including
the auth middleware. This is the scope where you verify that handlers
correctly pull the subject from the context, that errors map to the
right HTTP status codes, and that the bootstrap actually mints usable
keys.

A `curl`-driven smoke test is a good first-pass for a new endpoint;
the Go equivalent looks like:

```go
func TestE2E_AgentCannotPushForOtherDevice(t *testing.T) {
    srv, op, agent := startTestServer(t) // returns server + plaintext keys
    defer srv.Close()

    // operator creates the experiment
    must(t, doJSON(srv.URL+"/api/experiments", "POST", op, map[string]any{
        "deploy_id":       "boundary",
        "canary_devices":  []string{"car-1"},
        "control_devices": []string{"car-2"},
        "window_minutes":  5,
    }))

    // car-1 agent tries to push as car-2
    code, body := doJSON(srv.URL+"/api/metrics", "POST", agent["car-1"], map[string]any{
        "experiment_id": "exp-boundary",
        "device_id":     "car-2",
        "metric_name":   "cpu_usage",
        "value":         33.0,
        "group":         "control",
    })
    if code != http.StatusForbidden {
        t.Fatalf("expected 403, got %d: %s", code, body)
    }
}
```

End-to-end tests are slower and noisier than the unit-scope tests
above, so use them sparingly — once per HTTP endpoint is enough to
prove the wiring. The detailed coverage belongs in the lower layers.

## Test helpers (the ones to build)

Dispatch doesn't have a testutil package yet. When you write more than
a few tests, you'll find yourself reaching for the same primitives.
Here's the shape they should take, modeled loosely on Coder's helpers
but adapted to what we actually need.

### `internal/testutil` (proposed)

A general-purpose helpers package. Suggested initial contents:

* `MustNoError(t *testing.T, err error)` — fails the test if `err != nil`.
* `Eventually(t *testing.T, want func() bool, timeout time.Duration)` —
  poll-with-backoff for time-sensitive assertions. Replaces `time.Sleep`,
  which AGENTS.md forbids.
* `OperatorContext(t *testing.T) (context.Context, database.Subject)` —
  fabricates an operator subject and attaches it to a context. Saves
  five lines per test.
* `AgentContext(t *testing.T, deviceID string) context.Context` — same,
  but for an agent scoped to one device.

### `internal/database/dbtestutil` (proposed)

Database-scoped helpers. Suggested contents:

* `NewStore(t *testing.T) database.Store` — returns a fresh
  `dbauthz.New(dbaudit.New(dbmetrics.New(memstore.New()), log))` so
  every test exercises the full wrapper stack by default. If you only
  want one layer, instantiate the lower piece directly.
* `Seed(t *testing.T, s database.Store) SeedResult` — bootstraps a
  ready-to-use world: one operator key, three devices, three agent
  keys. Returns a struct with the keys and contexts already wired.
* `MustInsertSamples(t *testing.T, s database.Store, ctx context.Context, args ...InsertSampleParams)` —
  one-call bulk insert that fails the test on the first error.

### `cmd/dispatchd/dispatchdtest` (proposed)

The end-to-end equivalent of Coder's `coderdtest`. Spins up an
`httptest.NewServer` wrapped around the same router `main()` builds,
runs the same bootstrap, and returns the URL + plaintext keys so tests
don't have to repeat the wiring. Suggested signature:

```go
func New(t *testing.T) (*httptest.Server, BootstrapKeys)

type BootstrapKeys struct {
    Operator string            // plaintext operator API key
    Agents   map[string]string // device_id → plaintext agent API key
}
```

## Recipes

Practical step-by-steps for the most common test-related changes. If
you find yourself doing one of these, follow the recipe top-to-bottom.

### Adding a new Store method

When you add a method to the `database.Store` interface, you have to
update **every** implementation and wrapper, in this order. Skipping
one will break the build, which is good — that's the type system
enforcing the architecture.

1. Add the method to `Store` in [internal/database/store.go](../internal/database/store.go).
2. Define any new param/return types in the same file.
3. Implement it in [memstore/memstore.go](../internal/database/memstore/memstore.go)
   under the appropriate section (Experiments / Samples / Devices / etc.).
4. Add a passthrough in [dbmetrics/dbmetrics.go](../internal/database/dbmetrics/dbmetrics.go)
   following the existing pattern: `start := time.Now(); … s.record("MethodName", start, err)`.
5. Decide whether the new method is mutating. If yes, add a wrapper in
   [dbaudit/dbaudit.go](../internal/database/dbaudit/dbaudit.go) that
   calls `s.audit(ctx, "verb.noun", "target_type", target_id, metadata)`
   on success. If no, add a plain passthrough.
6. Add an authz guard in [dbauthz/dbauthz.go](../internal/database/dbauthz/dbauthz.go).
   **Every guard must have an inline comment explaining why it exists**
   (per AGENTS.md, removing it would silently change access).
7. Add a row to the dbauthz test matrix for the new method covering
   each subject type.
8. If the method is exposed via HTTP, add a handler in
   [cmd/dispatchd/main.go](../cmd/dispatchd/main.go) and one
   end-to-end test that proves both the success path and the 403 path.

### Adding a new entity (table)

Adding `deploys`, `metric_directions` as their own entity, etc.

1. Define the value type in [internal/database/store.go](../internal/database/store.go).
2. Add the row type and the storage map/slice to memstore. Pick a map
   if lookups are by ID, a slice if you scan with filters.
3. Add CRUD methods to the `Store` interface and implement them.
4. Run through the "Adding a new Store method" recipe for each new
   method, including the audit/authz/test obligations.
5. If the new entity needs an authz rule beyond
   "operators can do anything," add it explicitly to dbauthz.go and
   write a test that proves the boundary.

### Adding a new wrapper

You probably won't, but if you do (e.g., a `dbcache` wrapper for
read-heavy methods):

1. Create `internal/database/dbcache/dbcache.go`.
2. Define `type Store struct { inner database.Store; … }` and a
   `New(inner)` constructor.
3. Implement every method in the `Store` interface. Most will be
   passthroughs; the few you actually cache get custom logic.
4. Wire it into `main()` in the right position. Order matters:
   `dbauthz → dbcache → dbaudit → dbmetrics → memstore` would let
   the cache see authorized requests but skip the audit/metrics
   layers on cache hits, which is usually what you want. Document
   the choice in the wrapper's package doc.
5. The compiler will tell you if you missed a method.

### Testing time-sensitive logic

**Do not use `time.Sleep` in tests.** AGENTS.md is explicit and the
GO.md guide points at the same rule. The two acceptable approaches:

1. Inject a clock. memstore already has a `now func() time.Time` field
   that defaults to `time.Now`. A test can construct a memstore and
   reach in to override it (or you can add a `WithClock` option later).
2. Use `testing/synctest` for end-to-end scopes. It gives you a
   deterministic fake clock that advances when all goroutines are
   blocked, which is the right primitive for testing soak timers and
   the eventual real-time SSE fan-out.

### Asserting against the audit log

Because dbaudit writes audit entries through the same `Store` interface,
asserting on them is just another `ListAuditEntries` call:

```go
entries, _ := store.ListAuditEntries(opCtx, database.AuditFilter{Limit: 10})
if len(entries) == 0 || entries[0].Action != "experiment.create" {
    t.Fatalf("expected experiment.create at top of audit log, got %+v", entries)
}
```

The order is newest-first by design (security logs are read most-recent
first), so the entry you just produced is at index 0.

### Asserting against the metrics snapshot

`dbmetrics.Store` exposes a `Snapshot()` method that returns a sorted
slice of per-method stats. In tests you usually only care that the
counter incremented:

```go
metrics := dbmetrics.New(memstore.New())
_, _ = metrics.CreateExperiment(opCtx, params)
snap := metrics.Snapshot()
got := findMethod(snap, "CreateExperiment")
if got.Count != 1 {
    t.Fatalf("expected 1 call, got %d", got.Count)
}
```

Don't assert on absolute latencies — they're machine-dependent and
flaky. If you need to test the percentile math, test
[dbmetrics.go](../internal/database/dbmetrics/dbmetrics.go)'s
`percentileIdx` helper directly.

## Open questions

Things this guide doesn't yet answer because the codebase doesn't yet
need to:

* **Real database backing.** When (if) we move off memstore to Postgres,
  this guide needs a `dbtestutil.WillUsePostgres` section like Coder
  has. Until then, the in-memory store is the only backing.
* **Real-time fan-out.** When the SSE/pub-sub follow-up lands, this
  guide needs a section on testing concurrent broker subscribers.
  `testing/synctest` is the right primitive.
* **Frontend tests.** [site/AGENTS.md](../site/AGENTS.md) covers
  Storybook for component testing; this doc is backend-only.

If you find yourself needing one of these and there's no section yet,
add it.
