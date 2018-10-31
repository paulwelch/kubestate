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
	"github.com/urfave/cli"
	"os"
	"sort"
	"text/tabwriter"
)

//TODO: is there a more accurate command name than top for this?
//TOP Ideas
// top rollup by: Deployment / RC/RS / Service / Pod, Job/CronJob, Resource Quotas, HPA (network??), Storage (may not have right metrics for it)

//pods---
type rowKey struct {
	namespace, pod, container string
}

type row struct {
	node string
	cpuRequest, cpuLimit, memoryRequest, memoryLimit float64
}

type node struct {
	cpuCapacity, cpuAllocatable, memoryCapacity, memoryAllocatable float64
}

type sortKey struct {
	key rowKey
	value float64
}

type sortedKeys []*sortKey

// sort.Interface implementation
func (s sortedKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortedKeys) Len() int {
	return len(s)
}

func (s sortedKeys) Less(i, j int) bool {

	if s[i].value < s[j].value {
		return true
	}
	return false
}
//---pods

//deployments---
type deployKey struct {
	namespace, deployment string
}

type deploy struct {
	requested, available, unavailable float64
}

type deploySortKey struct {
	key deployKey
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
//---deployments

func Top(c *cli.Context) error {

	config := c.Parent().Parent().String("config")

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	//TODO: do we aggregate by node?
	switch c.Command.Name {
	case "deployments":
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

		s := make(sortedDeployKeys, 0, len(table))
		for k, v := range table {
			s = append(s, &deploySortKey{k, v.requested})
		}
		sort.Sort(sort.Reverse(s))

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 1, 1, ' ', 0)

		fmt.Fprintf(w, "%s\t%s\t%s\n", "Namespace", "Deployment", "(Req / Avail / Unavail)")

		for _, v := range s {
			fmt.Fprintf(w, "%s\t%s\t(%.0f / %.0f / %.0f)\n", v.key.namespace, v.key.deployment, table[v.key].requested, table[v.key].available, table[v.key].unavailable)
		}

		w.Flush()

	case "pods":
		table := make(map[rowKey]*row)
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

					if n != "" && nodes[n] == nil {
						nodes[n] = &node{}
					}

					if ns != "" && po != "" && co != "" {
						if table[rowKey{ns, po, co}] == nil {
							table[rowKey{ns, po, co}] = &row{}
							table[rowKey{ns, po, co}].node = n
						}
					}

					switch *metricFamilies[i].Name  {
					case "kube_pod_container_resource_requests":
						if re == "cpu" {
							table[rowKey{ns, po, co}].cpuRequest += *f.Gauge.Value
						} else if re == "memory" {
							table[rowKey{ns, po, co}].memoryRequest += *f.Gauge.Value
						}
					case "kube_pod_container_resource_limits":
						if re == "cpu" {
							table[rowKey{ns, po, co}].cpuLimit += *f.Gauge.Value
						} else if re == "memory" {
							table[rowKey{ns, po, co}].memoryLimit += *f.Gauge.Value
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

		s := make(sortedKeys, 0, len(table))
		for k, v := range table {
			//load factor is equally weighted average of cpu and memory requested as percentage of allocatable
			load := ((v.memoryRequest / nodes[v.node].memoryAllocatable) + (v.cpuRequest / nodes[v.node].cpuAllocatable)) / 2
			s = append(s, &sortKey{k, load})
		}
		sort.Sort(sort.Reverse(s))


		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 1, 1, ' ', 0)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "Namespace", "Pod", "Container", "CPU (Req / Lim)", "Memory  (Req / Lim)", "Node", "Load")

		for _, v := range s {
			fmt.Fprintf(w, "%s\t%s\t%s\t(%.0fm / %.0fm)\t(%.0fMi / %.0fMi)\t%s\t%.0f%%\n", v.key.namespace, v.key.pod, v.key.container, table[v.key].cpuRequest*1000, table[v.key].cpuLimit*1000, (table[v.key].memoryRequest/1048576), (table[v.key].memoryLimit/1048576), table[v.key].node, v.value*100)
		}

		w.Flush()

	}

	return nil
}

