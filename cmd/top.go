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
)

//TOP Ideas
// top rollup by: Deployment / RC/RS / Service / Pod, Job/CronJob, Resource Quotas, HPA (network??), Storage (may not have right metrics for it)


func Top(c *cli.Context) error {

	config := c.Parent().Parent().String("config")

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	switch c.Command.Name {
	case "deployments":
	case "pods":
		//map key: namespace -> pod -> container
		type key struct {
			namespace, pod, container, resource string
		}
		resources := make(map[key]float64)

		for i := 0; i < len(metricFamilies); i++ {

			var ns, po, co, re string

			if *metricFamilies[i].Name == "kube_pod_container_resource_requests" {
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
					resources[key{ns, po, co, re}] += *f.Gauge.Value
				}

			}

		}

		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", "Namespace", "Pod", "Container", "CPU")
		for k, v := range resources {
			fmt.Printf("%s\t%s\t%s\t%s\t%f\n", k.namespace, k.pod, k.container, k.resource, v)
		}

	}

	return nil
}
