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
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	dto "github.com/prometheus/client_model/go"
	"github.com/json-iterator/go"
	"github.com/urfave/cli"
	"fmt"
)

func Get(c *cli.Context) error {

	config := c.Parent().String("config")
	outputFormat := c.String("output")
	filterFlag := c.String("filter")

	//Raw output
	if outputFormat == "raw" {
		resp, err := getRawMetrics(config)
		if err != nil {
			return err
		}

		fmt.Println(resp)
		return nil
	} else if outputFormat == "json" {
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
		//TODO: table format
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
}
