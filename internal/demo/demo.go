// Package demo runs a small River queue locally to demonstrate successful,
// errored, and panicked job execution.
package demo

import (
	"context"
	"fmt"
	"io"
)

// ScenarioResult describes a job's terminal state at the end of the
// demonstration.
type ScenarioResult struct {
	Name     string
	JobID    int64
	State    string
	Attempts int
}

// Result is the observable outcome of a Run.
type Result struct {
	Scenarios []ScenarioResult
}

// Run executes the three scenarios, prints diagnostics and a summary to out, then
// serves River UI until ctx is cancelled. River setup is kept in river_setup.go.
func Run(ctx context.Context, out io.Writer) (Result, error) {
	return run(ctx, out, databasePath, true)
}

func printSummary(out io.Writer, result Result) {
	for _, scenario := range result.Scenarios {
		fmt.Fprintf(out, "SUMMARY scenario=%s job_id=%d state=%s attempts=%d\n", scenario.Name, scenario.JobID, scenario.State, scenario.Attempts)
	}
}
