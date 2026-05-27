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
	"bufio"
	"fmt"
	"strings"

	"github.com/json-iterator/go"
	dto "github.com/prometheus/client_model/go"
	"github.com/urfave/cli/v2"
)

var (
	getRawMetricsFn = getRawMetrics
	getMetricsFn    = getMetrics
	executeGetFn    = executeGet
)

func Get(c *cli.Context) error {
	metricFilter := c.String("metric")
	if metricFilter == "*" && c.Args().Len() > 0 {
		metricFilter = c.Args().First()
	}

	return executeGet(
		c.String("config"),
		c.String("output"),
		metricFilter,
		c.String("namespace"),
		c.String("metrics-namespace"),
		c.Bool("insecure-skip-tls-verify"),
	)
}

func executeGet(config, outputFormat, metricFilterFlag, namespaceFlag, metricsNamespace string, insecureSkipTLSVerify bool) error {
	if outputFormat == "raw" {
		resp, err := getRawMetricsFn(config, metricsNamespace, insecureSkipTLSVerify)
		if err != nil {
			return err
		}

		if metricFilterFlag == "*" && namespaceFlag == "*" {
			fmt.Println(resp)
		} else {
			filtered := filterRawMetrics(resp, metricFilterFlag, namespaceFlag)
			if filtered != "" {
				fmt.Print(filtered)
			}
		}

		return nil
	}

	if outputFormat == "json" {
		metricFamilies, err := getMetricsFn(config, metricsNamespace, insecureSkipTLSVerify)
		if err != nil {
			return err
		}

		matches := make(map[int]*dto.MetricFamily)
		cnt := 0
		for i := 0; i < len(metricFamilies); i++ {
			if metricFilterFlag == "*" || metricFamilies[i].GetName() == metricFilterFlag {
				var found bool
				if namespaceFlag != "*" {
					for _, m := range metricFamilies[i].Metric {
						for _, l := range m.Label {
							if l.GetName() == "namespace" {
								found = true
							}
						}
					}
				}
				if namespaceFlag == "*" || found {
					matches[cnt] = metricFamilies[i]
					cnt++
				}
			}
		}
		if cnt > 1 {
			fmt.Print("[")
		}
		for i := 0; i < cnt; i++ {
			s, _ := jsoniter.MarshalToString(matches[i])
			fmt.Print(s)
			if cnt > 1 && i < cnt-1 {
				fmt.Print(",")
			}
		}
		if cnt > 1 {
			fmt.Println("]")
		}
		return nil
	}

	return cli.Exit("invalid output format; valid formats are: json, raw", 2)
}

func filterRawMetrics(raw, metricFilterFlag, namespaceFlag string) string {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	var b strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if shouldIncludeRawMetricLine(line, metricFilterFlag, namespaceFlag) {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String()
}

func shouldIncludeRawMetricLine(line, metricFilterFlag, namespaceFlag string) bool {
	if line == "" || strings.HasPrefix(line, "#") {
		return false
	}

	metricName := line
	if i := strings.Index(line, "{"); i >= 0 {
		metricName = line[:i]
	} else if i := strings.Index(line, " "); i >= 0 {
		metricName = line[:i]
	}

	if metricFilterFlag != "*" && metricName != metricFilterFlag {
		return false
	}

	if namespaceFlag == "*" {
		return true
	}

	start := strings.Index(line, "{")
	end := strings.Index(line, "}")
	if start < 0 || end <= start {
		return false
	}

	labels := strings.Split(line[start+1:end], ",")
	for _, label := range labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], "\"")
		if key == "namespace" && value == namespaceFlag {
			return true
		}
	}

	return false
}
