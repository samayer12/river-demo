package demo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/riverqueue/river"
	"riverqueue.com/riverui"
)

const (
	uiAddress = "127.0.0.1:8080"
	uiPrefix  = "/riverui"
)

// serveRiverUI serves the embedded UI against the same River client and stops
// when ctx is cancelled. Tests disable it to keep lifecycle checks bounded.
func serveRiverUI(ctx context.Context, out io.Writer, client *river.Client[*sql.Tx], enabled bool) error {
	if !enabled {
		return nil
	}

	logger := slog.New(slog.NewTextHandler(out, nil))
	handler, err := riverui.NewHandler(&riverui.HandlerOpts{
		Endpoints: riverui.NewEndpoints(client, nil),
		Logger:    logger,
		Prefix:    uiPrefix,
	})
	if err != nil {
		return fmt.Errorf("create River UI handler: %w", err)
	}
	if err := handler.Start(ctx); err != nil {
		return fmt.Errorf("start River UI handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle(uiPrefix+"/", handler)
	server := &http.Server{
		Addr:              uiAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Fprintf(out, "River UI is available at http://%s%s/ (press Ctrl-C to stop)\n", uiAddress, uiPrefix)
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve River UI: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("stop River UI: %w", err)
		}
		return nil
	}
}
