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

type nodeSortKey struct {
	key   string
	value float64
}

type sortedNodeKeys []*nodeSortKey

// sort.Interface implementation
func (s sortedNodeKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedNodeKeys) Len() int {
	return len(s)
}

func (s sortedNodeKeys) Less(i, j int) bool {

	if s[i].value < s[j].value {
		return true
	}
	return false
}

func topNodes(metricFamilies []dto.MetricFamily, namespaceFlag string) {
	podAllocated := make(map[string]*pod)
	nodes := make(map[string]*node)

	for i := 0; i < len(metricFamilies); i++ {

		var re, n, ns string

		if *metricFamilies[i].Name == "kube_pod_container_resource_requests" ||
			*metricFamilies[i].Name == "kube_pod_container_resource_limits" ||
			*metricFamilies[i].Name == "kube_node_status_capacity_cpu_cores" ||
			*metricFamilies[i].Name == "kube_node_status_capacity_memory_bytes" ||
			*metricFamilies[i].Name == "kube_node_status_allocatable_memory_bytes" ||
			*metricFamilies[i].Name == "kube_node_status_allocatable_cpu_cores" {

			for _, f := range metricFamilies[i].Metric {

				for _, l := range f.Label {
					switch *l.Name {
					case "resource":
						re = *l.Value
					case "node":
						n = *l.Value
					case "namespace":
						ns = *l.Value
					}
				}

				if n != "" && nodes[n] == nil {
					nodes[n] = &node{}
					if podAllocated[n] == nil {
						podAllocated[n] = &pod{}
					}
					podAllocated[n].node = n
				}

				switch *metricFamilies[i].Name {
				case "kube_pod_container_resource_requests":
					if namespaceFlag == "*" || namespaceFlag == ns {
						if re == "cpu" {
							podAllocated[n].cpuRequest += *f.Gauge.Value
						} else if re == "memory" {
							podAllocated[n].memoryRequest += *f.Gauge.Value
						}
					}
				case "kube_pod_container_resource_limits":
					if namespaceFlag == "*" || namespaceFlag == ns {
						if re == "cpu" {
							podAllocated[n].cpuLimit += *f.Gauge.Value
						} else if re == "memory" {
							podAllocated[n].memoryLimit += *f.Gauge.Value
						}
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

	s := make(sortedNodeKeys, 0, len(podAllocated))
	for _, v := range podAllocated {
		//load factor is equally weighted average of cpu and memory requested as percentage of allocatable
		load := ((v.memoryRequest / nodes[v.node].memoryAllocatable) + (v.cpuRequest / nodes[v.node].cpuAllocatable)) / 2
		s = append(s, &nodeSortKey{v.node, load})
	}
	sort.Sort(sort.Reverse(s))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 1, 1, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Node", "CPU (Req / Lim / Cap)", "Memory (Req / Lim / Cap)", "Load")

	for _, v := range s {
		fmt.Fprintf(w, "%s\t(%.0fm / %.0fm / %.0fm)\t(%.0fMi / %.0fMi / %.0fMi)\t%.0f%%\n", podAllocated[v.key].node, podAllocated[v.key].cpuRequest*1000, podAllocated[v.key].cpuLimit*1000, nodes[v.key].cpuCapacity*1000, podAllocated[v.key].memoryRequest/1048576, podAllocated[v.key].memoryLimit/1048576, nodes[v.key].memoryCapacity/1048576, v.value*100)
	}

	w.Flush()

}
