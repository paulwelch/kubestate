package cmd

import (
	"testing"

	"github.com/urfave/cli/v2"
)

func TestShouldIncludeRawMetricLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		metric string
		ns     string
		want   bool
	}{
		{
			name:   "skips comments",
			line:   "# HELP kube_metric something",
			metric: "*",
			ns:     "*",
			want:   false,
		},
		{
			name:   "matches metric and namespace",
			line:   `kube_metric{namespace="default",pod="p1"} 1`,
			metric: "kube_metric",
			ns:     "default",
			want:   true,
		},
		{
			name:   "rejects wrong namespace",
			line:   `kube_metric{namespace="kube-system",pod="p1"} 1`,
			metric: "kube_metric",
			ns:     "default",
			want:   false,
		},
		{
			name:   "accepts any namespace when wildcard",
			line:   `kube_metric{namespace="kube-system",pod="p1"} 1`,
			metric: "kube_metric",
			ns:     "*",
			want:   true,
		},
		{
			name:   "rejects wrong metric",
			line:   `kube_other{namespace="default"} 1`,
			metric: "kube_metric",
			ns:     "*",
			want:   false,
		},
		{
			name:   "handles unlabeled metrics for wildcard namespace",
			line:   `go_gc_duration_seconds 0.1`,
			metric: "*",
			ns:     "*",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIncludeRawMetricLine(tt.line, tt.metric, tt.ns)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestFilterRawMetrics(t *testing.T) {
	raw := `# HELP kube_metric something
# TYPE kube_metric gauge
kube_metric{namespace="default",pod="p1"} 1
kube_metric{namespace="kube-system",pod="p2"} 2
kube_other{namespace="default"} 3
`

	got := filterRawMetrics(raw, "kube_metric", "default")
	want := "kube_metric{namespace=\"default\",pod=\"p1\"} 1\n"

	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExecuteGetRejectsInvalidOutput(t *testing.T) {
	err := executeGet("", "table", "*", "*", "", false)
	if err == nil {
		t.Fatal("expected error for invalid output")
	}
	exitErr, ok := err.(cli.ExitCoder)
	if !ok {
		t.Fatalf("expected cli.ExitCoder, got %T", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Fatalf("got exit code %d want 2", exitErr.ExitCode())
	}
}
