package demo

import (
	"context"
	"errors"

	"github.com/riverqueue/river"
)

const (
	successKind = "demo_success"
	errorKind   = "demo_error"
	panicKind   = "demo_panic"
)

var errIntentional = errors.New("intentional River job error")

// The following args and workers are the demo itself: one successful job, one
// returned error, and one panic. Replace their Work methods to demonstrate an
// application's real job behavior.
type successArgs struct{}

func (successArgs) Kind() string { return successKind }

type errorArgs struct{}

func (errorArgs) Kind() string { return errorKind }

type panicArgs struct{}

func (panicArgs) Kind() string { return panicKind }

type successWorker struct {
	river.WorkerDefaults[successArgs]
}

func (*successWorker) Work(context.Context, *river.Job[successArgs]) error { return nil }

type errorWorker struct {
	river.WorkerDefaults[errorArgs]
}

func (*errorWorker) Work(context.Context, *river.Job[errorArgs]) error { return errIntentional }

type panicWorker struct {
	river.WorkerDefaults[panicArgs]
}

func (*panicWorker) Work(context.Context, *river.Job[panicArgs]) error {
	panic("intentional River job panic")
}
