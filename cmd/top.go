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
	"github.com/urfave/cli"
)

// other top rollup ideas: RC/RS / Service, Job/CronJob, Resource Quotas, HPA (network??), Storage (may not have right metrics for it)

type podKey struct {
	namespace, pod, container string
}

type pod struct {
	node                                             string
	cpuRequest, cpuLimit, memoryRequest, memoryLimit float64
}

type node struct {
	cpuCapacity, cpuAllocatable, memoryCapacity, memoryAllocatable float64
}

func Top(c *cli.Context) error {

	config := c.Parent().Parent().String("config")

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	switch c.Command.Name {
	case "deployments":
		topDeployments(metricFamilies)
	case "pods":
		topPods(metricFamilies)
	case "nodes":
		topNodes(metricFamilies)
	}

	return nil
}
