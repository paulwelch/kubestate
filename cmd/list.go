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
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

func List(c *cli.Context) error {
	raw, err := getRawMetricsFn(c.String("config"), c.String("metrics-namespace"), c.Bool("insecure-skip-tls-verify"))
	if err != nil {
		return err
	}

	typeByName := make(map[string]string)
	helpByName := make(map[string]string)

	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "# TYPE ") {
			parts := strings.SplitN(strings.TrimPrefix(line, "# TYPE "), " ", 2)
			if len(parts) == 2 {
				typeByName[parts[0]] = strings.ToUpper(strings.TrimSpace(parts[1]))
			}
			continue
		}
		if strings.HasPrefix(line, "# HELP ") {
			parts := strings.SplitN(strings.TrimPrefix(line, "# HELP "), " ", 2)
			if len(parts) == 2 {
				helpByName[parts[0]] = strings.TrimSpace(parts[1])
			}
		}
	}

	names := make([]string, 0, len(helpByName))
	for name := range helpByName {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("%s\t%s\t%s\n", typeByName[name], name, helpByName[name])
	}

	return nil
}
