package cmd

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	dto "github.com/prometheus/client_model/go"
	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"bytes"
	"fmt"
	"io"
)

func getRawMetrics(config string) (string, error) {
	cfg, k8sclient, err := getClient(config)
	if err != nil {
		return "", err
	}

	//get kube-state-metrics raw data and parse
	var r rest.Result

	r = k8sclient.RESTClient().Get().RequestURI(cfg.Host + "/api/v1/namespaces/kube-system/services/kube-state-metrics:http-metrics/proxy/metrics").Do()
	if r.Error() != nil {
		return "", r.Error()
	}
	resp, _ := r.Raw()

	return string(resp), nil
}

func getMetrics(config string) ([]dto.MetricFamily, error) {
	cfg, k8sclient, err := getClient(config)
	if err != nil {
		return nil, err
	}

	//get kube-state-metrics raw data and parse
	var r rest.Result

	//request protobuf using Accept header
	const acceptHeader= `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

	r = k8sclient.RESTClient().Get().SetHeader("Accept", acceptHeader).RequestURI(cfg.Host + "/api/v1/namespaces/kube-system/services/kube-state-metrics:http-metrics/proxy/metrics").Do()
	if r.Error() != nil {
		return nil, r.Error()
	}
	resp, _ := r.Raw()

	//Parse protobuf into MetricFamily array, output family if filter specified
	//Might be faster with parallel go routine to parse, but with higher complexity.
	//Only ~100 families, so probably not worth it at this time.
	metricFamilies := make([]dto.MetricFamily, 0)
	reader := bytes.NewReader(resp)
	for {
		mf := dto.MetricFamily{}
		if _, err = pbutil.ReadDelimited(reader, &mf); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("Error reading metric family protobuf: %v", err)
		}
		metricFamilies = append(metricFamilies, mf)
	}

	return metricFamilies, nil
}

func getClient(config string) (*rest.Config, *kubernetes.Clientset, error) {

	cfg, err := clientcmd.BuildConfigFromFlags("", local.Expand(config))
	if err != nil {
		return nil, nil, err
	}
	k8sclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	//determine if kube-state-metrics service is available and healthy
	stateServiceFound := false
	svcs, err := k8sclient.CoreV1().Services("kube-system").List(v1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	for _, svc := range svcs.Items {
		if svc.Name == "kube-state-metrics" {
			stateServiceFound = true
		}
	}
	if !stateServiceFound {
		return nil, nil, cli.NewExitError("Error: kube-state-metrics service not found. See https://github.com/kubernetes/kube-state-metrics", 99)
	}
	req := k8sclient.RESTClient().Get().RequestURI(cfg.Host + "/api/v1/namespaces/kube-system/services/kube-state-metrics:http-metrics/proxy/healthz")
	r := req.Do()
	if r.Error() != nil {
		return nil, nil, r.Error()
	}
	resp, _ := r.Raw()
	if string(resp) != "ok" {
		return nil, nil, cli.NewExitError("Error: kube-state-metrics service is not healthy", 98)
	}

	return cfg, k8sclient, nil

}
