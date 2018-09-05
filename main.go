/*
 * Copyright 2018 Paul Welch
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
 */

package main

import (
	"bytes"
	_ "code.google.com/p/goprotobuf/proto"
	"fmt"
	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/urfave/cli"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "kubestate"
	app.Usage = "Show kubernetes state metrics"
	app.Action = func(c *cli.Context) error {

		//TODO: add command line help text
		//TODO: add optional command line arg for kube config
		configPath := local.Expand("~/.kube/config") //kube config path
		config, err := clientcmd.BuildConfigFromFlags("", configPath)
		if err != nil {
			return err
		}
		k8sclient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		//determine if kube-state-metrics is available service
		stateServiceFound := false
		svcs, err := k8sclient.CoreV1().Services("kube-system").List(v1.ListOptions{})
		if err != nil {
			return err
		}
		for _, svc := range svcs.Items {
			if svc.Name == "kube-state-metrics" {
				stateServiceFound = true
			}
		}
		if !stateServiceFound {
			return cli.NewExitError("Error: kube-state-metrics service not found", 99)
		}
		r := k8sclient.RESTClient().Get().RequestURI("/api/v1/namespaces/kube-system/services/kube-state-metrics:http-metrics/proxy/healthz").Do()
		if r.Error() != nil {
			return r.Error()
		}
		resp, _ := r.Raw()
		if string(resp) != "ok" {
			return cli.NewExitError("Error: kube-state-metrics service is not healthy", 98)
		}

		//request protobuf using Accept header
		const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

		//get kube-state-metrics raw data and parse
		r = k8sclient.RESTClient().Get().SetHeader("Accept", acceptHeader).RequestURI("/api/v1/namespaces/kube-system/services/kube-state-metrics:http-metrics/proxy/metrics").Do()
		if r.Error() != nil {
			return r.Error()
		}
		resp, _ = r.Raw()

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
				return fmt.Errorf("Error reading metric family protobuf: %v", err)
			}
			metricFamilies = append(metricFamilies, mf)
		}

		//TODO: add command line args to show state values - possibly: query, metric name, watch
		//fmt.Println(len(metricFamilies))
		//fmt.Println(metricFamilies[0])

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
