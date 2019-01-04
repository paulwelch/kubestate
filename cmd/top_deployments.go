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
	"fmt"
	dto "github.com/prometheus/client_model/go"
	"os"
	"sort"
	"text/tabwriter"
)

type deployKey struct {
	namespace, deployment string
}

type deploy struct {
	requested, available, unavailable float64
}

type deploySortKey struct {
	key   deployKey
	value float64
}

type sortedDeployKeys []*deploySortKey

// sort.Interface implementation
func (s sortedDeployKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedDeployKeys) Len() int {
	return len(s)
}

func (s sortedDeployKeys) Less(i, j int) bool {

	if s[i].value < s[j].value {
		return true
	}
	return false
}

func topDeployments(metricFamilies []dto.MetricFamily, namespaceFlag string) {
	//TODO: add rolling update metrics
	table := make(map[deployKey]*deploy)

	for i := 0; i < len(metricFamilies); i++ {

		var ns, d string

		if *metricFamilies[i].Name == "kube_deployment_spec_replicas" ||
			*metricFamilies[i].Name == "kube_deployment_status_replicas_available" ||
			*metricFamilies[i].Name == "kube_deployment_status_replicas_unavailable" {
			for _, f := range metricFamilies[i].Metric {

				for _, l := range f.Label {
					switch *l.Name {
					case "namespace":
						ns = *l.Value
					case "deployment":
						d = *l.Value
					}
				}

				if namespaceFlag == "*" || namespaceFlag == ns {
					if table[deployKey{ns, d}] == nil {
						table[deployKey{ns, d}] = &deploy{}
					}

					switch *metricFamilies[i].Name {
					case "kube_deployment_spec_replicas":
						table[deployKey{ns, d}].requested += *f.Gauge.Value
					case "kube_deployment_status_replicas_available":
						table[deployKey{ns, d}].available += *f.Gauge.Value
					case "kube_deployment_status_replicas_unavailable":
						table[deployKey{ns, d}].unavailable += *f.Gauge.Value
					}
				}
			}
		}
	}

	s := make(sortedDeployKeys, 0, len(table))
	for k, v := range table {
		s = append(s, &deploySortKey{k, v.requested})
	}
	sort.Sort(sort.Reverse(s))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 1, 1, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\n", "Namespace", "Deployment", "Replicas (Req / Avail / Unavail)")

	for _, v := range s {
		fmt.Fprintf(w, "%s\t%s\t(%.0f / %.0f / %.0f)\n", v.key.namespace, v.key.deployment, table[v.key].requested, table[v.key].available, table[v.key].unavailable)
	}

	w.Flush()
}
