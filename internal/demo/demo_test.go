package demo

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name             string
		wantStates       []string
		wantOutput       []string
		absentFromOutput []string
	}{
		{
			name:       "cancels failures after one error or panic",
			wantStates: []string{"completed", "cancelled", "cancelled"},
			wantOutput: []string{
				"ERROR river_job_error", "PANIC river_job_panic",
				"SUMMARY scenario=success", "SUMMARY scenario=error", "SUMMARY scenario=panic",
			},
			absentFromOutput: []string{"trace="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var output bytes.Buffer

			result, err := run(ctx, &output, filepath.Join(t.TempDir(), "river-demo.sqlite3"), false)
			if err != nil {
				t.Fatal(err)
			}

			assertScenarioStates(t, result, tt.wantStates...)
			assertOutput(t, output.String(), tt.wantOutput, tt.absentFromOutput)
		})
	}
}

func assertScenarioStates(t *testing.T, result Result, want ...string) {
	t.Helper()
	if len(result.Scenarios) != len(want) {
		t.Fatalf("scenario count = %d, want %d", len(result.Scenarios), len(want))
	}
	for index, state := range want {
		if got := result.Scenarios[index].State; got != state {
			t.Errorf("scenario %q state = %q, want %q", result.Scenarios[index].Name, got, state)
		}
	}
}

func assertOutput(t *testing.T, text string, expected, absent []string) {
	t.Helper()

	for _, value := range expected {
		if !strings.Contains(text, value) {
			t.Errorf("output missing %q:\n%s", value, text)
		}
	}
	for _, value := range absent {
		if strings.Contains(text, value) {
			t.Errorf("output included %q:\n%s", value, text)
		}
	}
}
