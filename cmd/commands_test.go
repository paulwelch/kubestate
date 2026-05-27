package cmd

import (
	"errors"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/urfave/cli/v2"
)

func TestGetCommandPositionalMetricRaw(t *testing.T) {
	restore := stubRawMetrics(t, func(string, string, bool) (string, error) {
		return "" +
			`kube_target_metric{namespace="kube-system",pod="p1"} 1` + "\n" +
			`kube_other_metric{namespace="kube-system",pod="p2"} 1` + "\n", nil
	})
	defer restore()

	ctx := newTestContext(t, testContextOptions{
		stringFlags: map[string]string{
			"config":            "",
			"output":            "raw",
			"metric":            "*",
			"namespace":         "*",
			"metrics-namespace": "",
		},
		boolFlags: map[string]bool{
			"insecure-skip-tls-verify": false,
		},
		args: []string{"kube_target_metric"},
	})

	out, err := captureStdout(func() error { return Get(ctx) })
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !strings.Contains(out, "kube_target_metric") {
		t.Fatalf("expected output to include target metric, got %q", out)
	}
	if strings.Contains(out, "kube_other_metric") {
		t.Fatalf("expected output to exclude non-target metric, got %q", out)
	}
}

func TestListCommand(t *testing.T) {
	restore := stubRawMetrics(t, func(string, string, bool) (string, error) {
		return "" +
			"# HELP kube_second second metric\n" +
			"# TYPE kube_second counter\n" +
			"# HELP kube_first first metric\n" +
			"# TYPE kube_first gauge\n", nil
	})
	defer restore()

	ctx := newTestContext(t, testContextOptions{
		stringFlags: map[string]string{
			"config":            "",
			"metrics-namespace": "",
		},
		boolFlags: map[string]bool{
			"insecure-skip-tls-verify": false,
		},
	})

	out, err := captureStdout(func() error { return List(ctx) })
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if !strings.Contains(out, "GAUGE\tkube_first\tfirst metric") {
		t.Fatalf("expected kube_first line, got %q", out)
	}
	if !strings.Contains(out, "COUNTER\tkube_second\tsecond metric") {
		t.Fatalf("expected kube_second line, got %q", out)
	}
}

func TestTopCommands(t *testing.T) {
	restore := stubMetrics(t, func(string, string, bool) ([]*dto.MetricFamily, error) {
		return sampleTopMetricFamilies(), nil
	})
	defer restore()

	tests := []struct {
		name        string
		commandName string
		wantHeader  string
		wantValue   string
	}{
		{
			name:        "pods",
			commandName: "pods",
			wantHeader:  "Namespace",
			wantValue:   "kube-system",
		},
		{
			name:        "deployments",
			commandName: "deployments",
			wantHeader:  "Deployment",
			wantValue:   "metrics-server",
		},
		{
			name:        "nodes",
			commandName: "nodes",
			wantHeader:  "Node",
			wantValue:   "node1",
		},
	}

	for _, tc := range tests {
		ctx := newTestContext(t, testContextOptions{
			stringFlags: map[string]string{
				"config":            "",
				"namespace":         "kube-system",
				"metrics-namespace": "",
			},
			boolFlags: map[string]bool{
				"insecure-skip-tls-verify": false,
			},
			commandName: tc.commandName,
		})

		out, err := captureStdout(func() error { return Top(ctx) })
		if err != nil {
			t.Fatalf("Top(%s) returned error: %v", tc.commandName, err)
		}
		if !strings.Contains(out, tc.wantHeader) {
			t.Fatalf("Top(%s) expected header %q, got %q", tc.commandName, tc.wantHeader, out)
		}
		if !strings.Contains(out, tc.wantValue) {
			t.Fatalf("Top(%s) expected value %q, got %q", tc.commandName, tc.wantValue, out)
		}
	}
}

func TestWatchCommandRunsExecuteGet(t *testing.T) {
	sentinelErr := errors.New("watch stop")
	restore := stubExecuteGet(t, func(string, string, string, string, string, bool) error {
		return sentinelErr
	})
	defer restore()

	ctx := newTestContext(t, testContextOptions{
		stringFlags: map[string]string{
			"config":            "",
			"output":            "json",
			"metric":            "*",
			"namespace":         "*",
			"metrics-namespace": "",
		},
		intFlags: map[string]int{
			"interval": 1,
		},
		boolFlags: map[string]bool{
			"insecure-skip-tls-verify": false,
		},
	})

	err := Watch(ctx)
	if !errors.Is(err, sentinelErr) {
		t.Fatalf("Watch() error = %v, want %v", err, sentinelErr)
	}
}

func newTestContext(t *testing.T, opts testContextOptions) *cli.Context {
	t.Helper()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	for name, value := range opts.stringFlags {
		fs.String(name, value, "")
	}
	for name, value := range opts.boolFlags {
		fs.Bool(name, value, "")
	}
	for name, value := range opts.intFlags {
		fs.Int(name, value, "")
	}
	if err := fs.Parse(opts.args); err != nil {
		t.Fatalf("failed parsing args: %v", err)
	}

	ctx := cli.NewContext(nil, fs, nil)
	if opts.commandName != "" {
		ctx.Command = &cli.Command{Name: opts.commandName}
	}

	return ctx
}

type testContextOptions struct {
	stringFlags map[string]string
	boolFlags   map[string]bool
	intFlags    map[string]int
	args        []string
	commandName string
}

func captureStdout(fn func() error) (string, error) {
	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	runErr := fn()
	_ = w.Close()
	os.Stdout = originalStdout

	data, readErr := io.ReadAll(r)
	_ = r.Close()
	if readErr != nil {
		return "", readErr
	}

	return string(data), runErr
}

func stubRawMetrics(t *testing.T, fn func(string, string, bool) (string, error)) func() {
	t.Helper()
	original := getRawMetricsFn
	getRawMetricsFn = fn
	return func() {
		getRawMetricsFn = original
	}
}

func stubMetrics(t *testing.T, fn func(string, string, bool) ([]*dto.MetricFamily, error)) func() {
	t.Helper()
	original := getMetricsFn
	getMetricsFn = fn
	return func() {
		getMetricsFn = original
	}
}

func stubExecuteGet(t *testing.T, fn func(string, string, string, string, string, bool) error) func() {
	t.Helper()
	original := executeGetFn
	executeGetFn = fn
	return func() {
		executeGetFn = original
	}
}

func sampleTopMetricFamilies() []*dto.MetricFamily {
	return []*dto.MetricFamily{
		newMetricFamily("kube_pod_container_resource_requests", []*dto.Metric{
			newGaugeMetric(0.1, map[string]string{
				"namespace": "kube-system",
				"pod":       "metrics-server-abc",
				"container": "metrics-server",
				"node":      "node1",
				"resource":  "cpu",
			}),
			newGaugeMetric(104857600, map[string]string{
				"namespace": "kube-system",
				"pod":       "metrics-server-abc",
				"container": "metrics-server",
				"node":      "node1",
				"resource":  "memory",
			}),
		}),
		newMetricFamily("kube_pod_container_resource_limits", []*dto.Metric{
			newGaugeMetric(0.2, map[string]string{
				"namespace": "kube-system",
				"pod":       "metrics-server-abc",
				"container": "metrics-server",
				"node":      "node1",
				"resource":  "cpu",
			}),
			newGaugeMetric(209715200, map[string]string{
				"namespace": "kube-system",
				"pod":       "metrics-server-abc",
				"container": "metrics-server",
				"node":      "node1",
				"resource":  "memory",
			}),
		}),
		newMetricFamily("kube_node_status_capacity", []*dto.Metric{
			newGaugeMetric(8, map[string]string{"node": "node1", "resource": "cpu"}),
			newGaugeMetric(17179869184, map[string]string{"node": "node1", "resource": "memory"}),
		}),
		newMetricFamily("kube_node_status_allocatable", []*dto.Metric{
			newGaugeMetric(7.5, map[string]string{"node": "node1", "resource": "cpu"}),
			newGaugeMetric(16106127360, map[string]string{"node": "node1", "resource": "memory"}),
		}),
		newMetricFamily("kube_deployment_spec_replicas", []*dto.Metric{
			newGaugeMetric(2, map[string]string{"namespace": "kube-system", "deployment": "metrics-server"}),
		}),
		newMetricFamily("kube_deployment_status_replicas_available", []*dto.Metric{
			newGaugeMetric(2, map[string]string{"namespace": "kube-system", "deployment": "metrics-server"}),
		}),
		newMetricFamily("kube_deployment_status_replicas_unavailable", []*dto.Metric{
			newGaugeMetric(0, map[string]string{"namespace": "kube-system", "deployment": "metrics-server"}),
		}),
	}
}

func newMetricFamily(name string, metrics []*dto.Metric) *dto.MetricFamily {
	n := name
	t := dto.MetricType_GAUGE
	return &dto.MetricFamily{
		Name:   &n,
		Type:   &t,
		Metric: metrics,
	}
}

func newGaugeMetric(value float64, labels map[string]string) *dto.Metric {
	v := value
	return &dto.Metric{
		Label: labelPairs(labels),
		Gauge: &dto.Gauge{Value: &v},
	}
}

func labelPairs(in map[string]string) []*dto.LabelPair {
	out := make([]*dto.LabelPair, 0, len(in))
	for k, v := range in {
		kv, vv := k, v
		out = append(out, &dto.LabelPair{Name: &kv, Value: &vv})
	}
	return out
}
