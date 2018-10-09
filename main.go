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
	"fmt"
	"github.com/json-iterator/go"
	"github.com/kubicorn/kubicorn/pkg/local"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/urfave/cli"
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"
)

func main() {
	var config string
	var filterFlag string
	var rawFlag bool
	var jsonFlag bool

	//TODO: expand on filter flag - possible regex, state values, by namespace, by label
	//TODO: maybe match kubectl command pattern: get, describe, watch
	//      ideas for metric views - rolling update state; deployment state, hpa's, jobs, etc
	//TODO: add output format options (raw, json, table)
	//TODO: add reasonable defaults with no command or flags - maybe a 'top' display
	//TODO: list of metrics with help text

	app := cli.NewApp()
	app.Name = "kubestate"
	app.Usage = "Show kubernetes state metrics"
	app.Version = "0.0.1"
	app.Compiled = time.Now()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       "~/.kube/config",
			Usage:       "path to kubeconfig",
			Destination: &config,
		},
		cli.StringFlag{
			//TODO: should filter be a subcommand?
			Name:        "filter, f",
			Value:       "*",
			Usage:       "Metric filter to show",
			Destination: &filterFlag,
		},
	}

	//TODO: refactor to command func's
	app.Commands = []cli.Command{
		cli.Command{
			Name: "get",
			Usage: "Get metrics",

			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:        "json, j",
					Usage:       "Show JSON format",
					Destination: &jsonFlag,
				},
				//TODO: should raw be a command?
				cli.BoolFlag{
					Name:        "raw, r",
					Usage:       "Show raw response data format",
					Destination: &rawFlag,
				},
			},

			Action: func(c *cli.Context) error {

				//Raw output
				if rawFlag {
					resp, err := getRawMetrics(config)
					if err != nil {
						return err
					}
					fmt.Println(resp)
					return nil
				} else if jsonFlag {
					metricFamilies, err := getMetrics(config)
					if err != nil {
						return err
					}

					for i := 0; i < len(metricFamilies); i++ {
						if filterFlag == "*" || *metricFamilies[i].Name == filterFlag {
							if i == 0 {
								fmt.Print("[")
							}
							s, _ := jsoniter.MarshalToString(metricFamilies[i])
							fmt.Print(s)
							if i < (len(metricFamilies) - 1) {
								fmt.Print(",")
							} else {
								fmt.Println("]")
							}
						}
					}
				} else {
					//TODO: default formatted output
					metricFamilies, err := getMetrics(config)
					if err != nil {
						return err
					}

					for i := 0; i < len(metricFamilies); i++ {
						fmt.Println("---------------")
						fmt.Println(*metricFamilies[i].Name)
						fmt.Println(*metricFamilies[i].Type)
						fmt.Println(*metricFamilies[i].Help)

						//for debugging
						for i := 0; i < len(metricFamilies[i].Metric); i++ {
							for j := 0; j < len((*metricFamilies[i].Metric[i]).Label); j++ {

								fmt.Println("---------------")
								fmt.Printf("Metric %d: Label %d:  %s  value: %s\n", i, j, *metricFamilies[i].Metric[i].Label[0].Name, *metricFamilies[i].Metric[i].Label[0].Value)

								switch *metricFamilies[i].Type {

								case dto.MetricType_COUNTER:
									fmt.Printf("Counter Value: %f", *metricFamilies[i].Metric[i].Counter.Value)

								case dto.MetricType_GAUGE:
									fmt.Printf("Gauge Value: %f", *metricFamilies[i].Metric[i].Gauge.Value)

								case dto.MetricType_SUMMARY:
									fmt.Println(*metricFamilies[i].Metric[i].Summary.Quantile[0].Value)
									fmt.Println(*metricFamilies[i].Metric[i].Summary.Quantile[0].Quantile)
									fmt.Println(*metricFamilies[i].Metric[i].Summary.SampleCount)
									fmt.Println(*metricFamilies[i].Metric[i].Summary.SampleSum)

								}
							}
						}
					}
				}
				return nil
			},
		},
		cli.Command{
			Name: "top",
			Usage: "Show top resource consumption by deployment",
			Action: func(c *cli.Context) error {
				fmt.Println("Implement Top Here")
				return nil
			},
		},
		cli.Command{
			Name: "watch",
			Usage: "Watch metric",
			Action: func(c *cli.Context) error {
				fmt.Println("Implement Watch Here")
				return nil
			},
		},

	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

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
