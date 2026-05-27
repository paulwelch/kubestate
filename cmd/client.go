/*
 * Copyright 2018 Paul Welch
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
 */

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/urfave/cli/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type metricsServiceRef struct {
	namespace string
	name      string
	proxyName string
}

func getRawMetrics(config, metricsNamespace string, insecureSkipTLSVerify bool) (string, error) {
	cfg, k8sclient, serviceRef, err := getClient(config, metricsNamespace, insecureSkipTLSVerify)
	if err != nil {
		return "", err
	}

	var r rest.Result
	r = k8sclient.RESTClient().Get().RequestURI(cfg.Host + "/api/v1/namespaces/" + serviceRef.namespace + "/services/" + serviceRef.proxyName + "/proxy/metrics").Do(context.Background())
	if r.Error() != nil {
		return "", r.Error()
	}
	resp, _ := r.Raw()

	return string(resp), nil
}

func getMetrics(config, metricsNamespace string, insecureSkipTLSVerify bool) ([]*dto.MetricFamily, error) {
	cfg, k8sclient, serviceRef, err := getClient(config, metricsNamespace, insecureSkipTLSVerify)
	if err != nil {
		return nil, err
	}

	var r rest.Result
	const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

	r = k8sclient.RESTClient().Get().SetHeader("Accept", acceptHeader).RequestURI(cfg.Host + "/api/v1/namespaces/" + serviceRef.namespace + "/services/" + serviceRef.proxyName + "/proxy/metrics").Do(context.Background())
	if r.Error() != nil {
		return nil, r.Error()
	}
	resp, _ := r.Raw()

	metricFamilies, err := parseMetricsResponse(resp)
	if err != nil {
		return nil, err
	}

	return metricFamilies, nil
}

func parseMetricsResponse(resp []byte) ([]*dto.MetricFamily, error) {
	metricFamilies := make([]*dto.MetricFamily, 0)
	reader := bytes.NewReader(resp)
	parseErr := error(nil)
	for {
		mf := &dto.MetricFamily{}
		if _, err := pbutil.ReadDelimited(reader, mf); err != nil {
			if err == io.EOF {
				break
			}
			parseErr = err
			metricFamilies = nil
			break
		}
		metricFamilies = append(metricFamilies, mf)
	}

	if len(metricFamilies) > 0 {
		return metricFamilies, nil
	}

	textParser := expfmt.NewTextParser(model.NameValidationScheme)
	parsed, err := textParser.TextToMetricFamilies(bytes.NewReader(resp))
	if err != nil {
		if parseErr != nil {
			return nil, fmt.Errorf("Error reading metric family protobuf: %v; text parse fallback failed: %v", parseErr, err)
		}
		return nil, fmt.Errorf("Error reading metric family text: %v", err)
	}

	names := make([]string, 0, len(parsed))
	for name := range parsed {
		names = append(names, name)
	}
	sort.Strings(names)

	metricFamilies = make([]*dto.MetricFamily, 0, len(names))
	for _, name := range names {
		metricFamilies = append(metricFamilies, parsed[name])
	}

	return metricFamilies, nil
}

func getClient(config, metricsNamespace string, insecureSkipTLSVerify bool) (*rest.Config, *kubernetes.Clientset, metricsServiceRef, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", expandHome(config))
	if err != nil {
		return nil, nil, metricsServiceRef{}, err
	}
	if insecureSkipTLSVerify {
		cfg.TLSClientConfig.Insecure = true
		cfg.TLSClientConfig.CAData = nil
		cfg.TLSClientConfig.CAFile = ""
	}

	k8sclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, metricsServiceRef{}, err
	}

	serviceRef, err := resolveMetricsService(k8sclient, metricsNamespace)
	if err != nil {
		return nil, nil, metricsServiceRef{}, err
	}

	req := k8sclient.RESTClient().Get().RequestURI(cfg.Host + "/api/v1/namespaces/" + serviceRef.namespace + "/services/" + serviceRef.proxyName + "/proxy/healthz")
	r := req.Do(context.Background())
	if r.Error() != nil {
		return nil, nil, metricsServiceRef{}, r.Error()
	}
	resp, _ := r.Raw()
	if !strings.EqualFold(strings.TrimSpace(string(resp)), "ok") {
		return nil, nil, metricsServiceRef{}, cli.Exit("Error: kube-state-metrics service is not healthy", 98)
	}

	return cfg, k8sclient, serviceRef, nil
}

func resolveMetricsService(k8sclient *kubernetes.Clientset, metricsNamespace string) (metricsServiceRef, error) {
	ctx := context.Background()

	if metricsNamespace != "" {
		svc, err := k8sclient.CoreV1().Services(metricsNamespace).Get(ctx, "kube-state-metrics", metav1.GetOptions{})
		if err != nil {
			return metricsServiceRef{}, cli.Exit(fmt.Sprintf("Error: kube-state-metrics service not found in namespace %q", metricsNamespace), 99)
		}
		return buildMetricsServiceRef(svc)
	}

	svcs, err := k8sclient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return metricsServiceRef{}, err
	}

	matches := make([]metricsServiceRef, 0)
	for _, svc := range svcs.Items {
		if svc.Name == "kube-state-metrics" {
			serviceRef, err := buildMetricsServiceRef(&svc)
			if err != nil {
				return metricsServiceRef{}, err
			}
			matches = append(matches, serviceRef)
		}
	}

	if len(matches) == 0 {
		return metricsServiceRef{}, cli.Exit("Error: kube-state-metrics service not found. Use --metrics-namespace if it is not discoverable.", 99)
	}

	// Prefer common namespaces if there are multiple installs.
	for _, preferred := range []string{"monitoring", "kube-system"} {
		for _, svc := range matches {
			if svc.namespace == preferred {
				return svc, nil
			}
		}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].namespace < matches[j].namespace })
	return matches[0], nil
}

func buildMetricsServiceRef(svc *corev1.Service) (metricsServiceRef, error) {
	if svc == nil {
		return metricsServiceRef{}, fmt.Errorf("nil service reference")
	}
	if len(svc.Spec.Ports) == 0 {
		return metricsServiceRef{}, cli.Exit(fmt.Sprintf("Error: kube-state-metrics service %q in namespace %q has no ports", svc.Name, svc.Namespace), 97)
	}
	port := svc.Spec.Ports[0]
	proxyName := "http:" + svc.Name + ":" + strconv.Itoa(int(port.Port))
	return metricsServiceRef{namespace: svc.Namespace, name: svc.Name, proxyName: proxyName}, nil
}

func expandHome(path string) string {
	if path == "" || path[0] != '~' {
		return path
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
