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

type podSortKey struct {
	key   podKey
	value float64
}

type sortedPodKeys []*podSortKey

// sort.Interface implementation
func (s sortedPodKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedPodKeys) Len() int {
	return len(s)
}

func (s sortedPodKeys) Less(i, j int) bool {

	if s[i].value < s[j].value {
		return true
	}
	return false
}

func topPods(metricFamilies []dto.MetricFamily, namespaceFlag string) {
	pods := make(map[podKey]*pod)
	nodes := make(map[string]*node)

	for i := 0; i < len(metricFamilies); i++ {

		var ns, po, co, re, n string

		if *metricFamilies[i].Name == "kube_pod_container_resource_requests" ||
			*metricFamilies[i].Name == "kube_pod_container_resource_limits" ||
			*metricFamilies[i].Name == "kube_node_status_capacity_cpu_cores" ||
			*metricFamilies[i].Name == "kube_node_status_capacity_memory_bytes" ||
			*metricFamilies[i].Name == "kube_node_status_allocatable_memory_bytes" ||
			*metricFamilies[i].Name == "kube_node_status_allocatable_cpu_cores" {
			for _, f := range metricFamilies[i].Metric {

				for _, l := range f.Label {
					switch *l.Name {
					case "namespace":
						ns = *l.Value
					case "pod":
						po = *l.Value
					case "container":
						co = *l.Value
					case "resource":
						re = *l.Value
					case "node":
						n = *l.Value
					}
				}

				//need to parse if kube_node_* - there's no namespace on those metrics
				//might want to consider splitting into separate loops
				if namespaceFlag == "*" || namespaceFlag == ns ||
					*metricFamilies[i].Name == "kube_node_status_capacity_cpu_cores" ||
					*metricFamilies[i].Name == "kube_node_status_capacity_memory_bytes" ||
					*metricFamilies[i].Name == "kube_node_status_allocatable_memory_bytes" ||
					*metricFamilies[i].Name == "kube_node_status_allocatable_cpu_cores" {
					if n != "" && nodes[n] == nil {
						nodes[n] = &node{}
					}

					if ns != "" && po != "" && co != "" {
						if pods[podKey{ns, po, co}] == nil {
							pods[podKey{ns, po, co}] = &pod{}
							pods[podKey{ns, po, co}].node = n
						}
					}

					switch *metricFamilies[i].Name {
					case "kube_pod_container_resource_requests":
						if re == "cpu" {
							pods[podKey{ns, po, co}].cpuRequest += *f.Gauge.Value
						} else if re == "memory" {
							pods[podKey{ns, po, co}].memoryRequest += *f.Gauge.Value
						}
					case "kube_pod_container_resource_limits":
						if re == "cpu" {
							pods[podKey{ns, po, co}].cpuLimit += *f.Gauge.Value
						} else if re == "memory" {
							pods[podKey{ns, po, co}].memoryLimit += *f.Gauge.Value
						}
					case "kube_node_status_capacity_memory_bytes":
						nodes[n].memoryCapacity = *f.Gauge.Value
					case "kube_node_status_capacity_cpu_cores":
						nodes[n].cpuCapacity = *f.Gauge.Value
					case "kube_node_status_allocatable_memory_bytes":
						nodes[n].memoryAllocatable = *f.Gauge.Value
					case "kube_node_status_allocatable_cpu_cores":
						nodes[n].cpuAllocatable = *f.Gauge.Value
					}
				}
			}
		}
	}

	s := make(sortedPodKeys, 0, len(pods))
	for k, v := range pods {
		//load factor is equally weighted average of cpu and memory requested as percentage of allocatable
		load := ((v.memoryRequest / nodes[v.node].memoryAllocatable) + (v.cpuRequest / nodes[v.node].cpuAllocatable)) / 2
		s = append(s, &podSortKey{k, load})
	}
	sort.Sort(sort.Reverse(s))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 1, 1, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "Namespace", "Pod", "Container", "CPU (Req / Lim)", "Memory  (Req / Lim)", "Node", "Load")

	for _, v := range s {
		fmt.Fprintf(w, "%s\t%s\t%s\t(%.0fm / %.0fm)\t(%.0fMi / %.0fMi)\t%s\t%.0f%%\n", v.key.namespace, v.key.pod, v.key.container, pods[v.key].cpuRequest*1000, pods[v.key].cpuLimit*1000, (pods[v.key].memoryRequest / 1048576), (pods[v.key].memoryLimit / 1048576), pods[v.key].node, v.value*100)
	}

	w.Flush()

}
