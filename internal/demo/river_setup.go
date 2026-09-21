package demo

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riversqlite"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"
	_ "modernc.org/sqlite"
)

const databasePath = "river-demo.sqlite3"

// run contains the reusable River and SQLite plumbing. The scenario-specific
// behavior is in scenarios.go; warning handling is in telemetry.go.
func run(ctx context.Context, out io.Writer, path string, serveUI bool) (Result, error) {
	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return Result{}, fmt.Errorf("open SQLite database: %w", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	driver := riversqlite.New(db)
	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return Result{}, fmt.Errorf("create River migrator: %w", err)
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return Result{}, fmt.Errorf("migrate River schema: %w", err)
	}

	workers := river.NewWorkers()
	river.AddWorker(workers, &successWorker{})
	river.AddWorker(workers, &errorWorker{})
	river.AddWorker(workers, &panicWorker{})

	client, err := river.NewClient(driver, &river.Config{
		ErrorHandler: &riverErrorHandler{out: out},
		Logger:       slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.Level(9)})),
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 3},
		},
		Workers: workers,
	})
	if err != nil {
		return Result{}, fmt.Errorf("create River client: %w", err)
	}

	events, unsubscribe := client.Subscribe(river.EventKindJobCancelled, river.EventKindJobCompleted, river.EventKindJobFailed)
	defer unsubscribe()
	if err := client.Start(ctx); err != nil {
		return Result{}, fmt.Errorf("start River client: %w", err)
	}
	defer client.Stop(context.Background())

	success, err := client.Insert(ctx, successArgs{}, nil)
	if err != nil {
		return Result{}, fmt.Errorf("insert success job: %w", err)
	}
	errorJob, err := client.Insert(ctx, errorArgs{}, nil)
	if err != nil {
		return Result{}, fmt.Errorf("insert error job: %w", err)
	}
	panicJob, err := client.Insert(ctx, panicArgs{}, nil)
	if err != nil {
		return Result{}, fmt.Errorf("insert panic job: %w", err)
	}

	const expectedEvents = 3
	latest := make(map[int64]*rivertype.JobRow, expectedEvents)
	for range expectedEvents {
		select {
		case event := <-events:
			latest[event.Job.ID] = event.Job
		case <-ctx.Done():
			return Result{}, fmt.Errorf("wait for jobs to finish: %w", ctx.Err())
		}
	}

	result := Result{Scenarios: []ScenarioResult{
		scenarioResult("success", success.Job.ID, latest[success.Job.ID]),
		scenarioResult("error", errorJob.Job.ID, latest[errorJob.Job.ID]),
		scenarioResult("panic", panicJob.Job.ID, latest[panicJob.Job.ID]),
	}}
	printSummary(out, result)
	if err := serveRiverUI(ctx, out, client, serveUI); err != nil {
		return Result{}, err
	}
	return result, nil
}

func sqliteDSN(path string) string {
	return fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_txlock=immediate", path)
}

func scenarioResult(name string, jobID int64, job *rivertype.JobRow) ScenarioResult {
	if job == nil {
		return ScenarioResult{Name: name, JobID: jobID, State: "not-observed"}
	}
	return ScenarioResult{Name: name, JobID: jobID, State: string(job.State), Attempts: job.Attempt}
}
