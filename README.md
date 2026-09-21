# River observability demo

This is a local Go CLI for observing three River job outcomes:

- a successful job;
- a job that returns an error;
- a job that panics.

The workers never call `recover()`. River rescues the panic and invokes the
configured `river.ErrorHandler`, which writes application-owned diagnostics to
the console. The handler is where an application can add telemetry without
changing worker error handling.

## Code map

- `cmd/river-demo/main.go` starts the CLI and converts Ctrl-C into context
  cancellation for graceful queue and UI shutdown.
- `internal/demo/demo.go` is the public entry point and defines the lifecycle
  result and summary output.
- `internal/demo/scenarios.go` contains the three core examples: successful,
  returned-error, and panicking workers.
- `internal/demo/telemetry.go` implements River's error and panic callbacks,
  writes concise diagnostics, and cancels failed jobs.
- `internal/demo/river_setup.go` opens the persistent SQLite database, applies
  River migrations, starts the client, inserts jobs, and observes final states.
- `internal/demo/river_ui.go` embeds and serves River UI against that same
  River client on loopback.
- `internal/demo/demo_test.go` is the table-driven lifecycle regression suite;
  it uses a temporary SQLite file per test.
- `.gitignore` excludes the persistent SQLite database and its WAL sidecar
  files from version control.

## Run it

The demo uses a persistent `river-demo.sqlite3` SQLite queue, runs migrations
in-process, and does not require PostgreSQL or any other service:

```sh
go run ./cmd/river-demo
```

The handler returns `SetCancelled: true` for the failing jobs. The
output therefore contains one error diagnostic, one concise panic diagnostic,
and a summary where the success job is completed and the two failed jobs are
cancelled. It then serves River UI until you press Ctrl-C.

## View jobs in River UI

River UI is embedded in the demo and is available at
`http://127.0.0.1:8080/riverui/` after the three jobs finish. It uses the same
SQLite client and database, so the jobs and their final states are visible in
the UI.

The standalone `riverui` binary and container are for PostgreSQL connections;
they cannot inspect this SQLite database. The embedded UI listens only on
loopback. The database is ignored by Git and remains available after Ctrl-C.

River UI includes job-management actions, including manual retry. This demo
does not configure automatic retries; use the UI's actions only when you want
to explore them.

## Verify

```sh
go test ./...
go vet ./...
```
