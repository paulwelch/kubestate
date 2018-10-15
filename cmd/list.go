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
	"github.com/urfave/cli"
	"fmt"
)

func List(c *cli.Context) error {

	config := c.Parent().String("config")

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	for i := 0; i < len(metricFamilies); i++ {
		fmt.Printf("%s\t%s\t%s\n", *metricFamilies[i].Type, *metricFamilies[i].Name, *metricFamilies[i].Help)
	}

	return nil
}
