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
	"text/tabwriter"
)

//TODO: is there a more accurate command name than top for this?
//TOP Ideas
// top rollup by: Deployment / RC/RS / Service / Pod, Job/CronJob, Resource Quotas, HPA (network??), Storage (may not have right metrics for it)


func Top(c *cli.Context) error {

	config := c.Parent().Parent().String("config")

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	//TODO: do we aggregate by node?
	switch c.Command.Name {
	case "deployments":
		//TODO: deployment pods - desired (kube_deployment_spec_replicas), available(kube_deployment_status_replicas_available), unavailable(kube_deployment_status_replicas_unavailable)
	case "pods":
		//TODO: should we show pods that have no request or limit set?

		type rowKey struct {
			namespace, pod, container string
		}
		type row struct {
			cpuRequest, cpuLimit, memoryRequest, memoryLimit float64
		}
		table := make(map[rowKey]*row)

		for i := 0; i < len(metricFamilies); i++ {

			var ns, po, co, re string

			if *metricFamilies[i].Name == "kube_pod_container_resource_requests" ||
				*metricFamilies[i].Name == "kube_pod_container_resource_limits" {
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
						}
					}

					if table[rowKey{ns, po, co}] == nil {
						table[rowKey{ns, po, co}] = &row{}
					}

					if *metricFamilies[i].Name == "kube_pod_container_resource_requests" {
						if re == "cpu" {
							table[rowKey{ns, po, co}].cpuRequest += *f.Gauge.Value
						} else if re == "memory" {
							table[rowKey{ns, po, co}].memoryRequest += *f.Gauge.Value
						}
					} else if *metricFamilies[i].Name == "kube_pod_container_resource_limits" {
						if re == "cpu" {
							table[rowKey{ns, po, co}].cpuLimit += *f.Gauge.Value
						} else if re == "memory" {
							table[rowKey{ns, po, co}].memoryLimit += *f.Gauge.Value
						}
					}
				}

			}

		}

		//TODO: Sort highest to lowest resources
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 10, 10, 0, '\t', 0)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "Namespace", "Pod", "Container", "CPU (Requested / Limit)", "Memory  (Requested / Limit)")

		for k, v := range table {
			fmt.Fprintf(w, "%s\t%s\t%s\t(%.2f / %.2f)\t(%.0f Mi / %.0f Mi)\n", k.namespace, k.pod, k.container, v.cpuRequest, v.cpuLimit, (v.memoryRequest/1048576), (v.memoryLimit/1048576))
		}

		w.Flush()

	}

	return nil
}
