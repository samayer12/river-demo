package demo

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// riverErrorHandler is deliberately the only failure-observation mechanism.
// River captures worker panics before invoking HandlePanic; workers do not use
// recover.
type riverErrorHandler struct {
	mu  sync.Mutex
	out io.Writer
}

func (h *riverErrorHandler) HandleError(_ context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	h.write("ERROR river_job_error job_id=%d job_kind=%s attempt=%d error=%q\n", job.ID, job.Kind, job.Attempt, err)
	return &river.ErrorHandlerResult{SetCancelled: true}
}

func (h *riverErrorHandler) HandlePanic(_ context.Context, job *rivertype.JobRow, panicValue any, trace string) *river.ErrorHandlerResult {
	// River provides trace for production telemetry. The demo prints only the
	// panic value so the lifecycle output stays readable.
	_ = trace
	h.write("PANIC river_job_panic job_id=%d job_kind=%s attempt=%d panic=%v\n", job.ID, job.Kind, job.Attempt, panicValue)
	return &river.ErrorHandlerResult{SetCancelled: true}
}

func (h *riverErrorHandler) write(format string, args ...any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	fmt.Fprintf(h.out, format, args...)
}
