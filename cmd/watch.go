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
	"os"
	"os/signal"
	"time"

	"github.com/urfave/cli/v2"
)

func Watch(c *cli.Context) error {
	interval := c.Int("interval")
	if interval < 1 {
		return cli.Exit("interval must be >= 1", 2)
	}

	config := c.String("config")
	output := c.String("output")
	metric := c.String("metric")
	namespace := c.String("namespace")
	metricsNamespace := c.String("metrics-namespace")
	insecureSkipTLSVerify := c.Bool("insecure-skip-tls-verify")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Stop(interrupt)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	run := func() error {
		fmt.Print("\x1bc")
		fmt.Printf("kubestate watch (interval=%ds)\n\n", interval)
		return executeGetFn(config, output, metric, namespace, metricsNamespace, insecureSkipTLSVerify)
	}

	if err := run(); err != nil {
		return err
	}

	for {
		select {
		case <-interrupt:
			return nil
		case <-ticker.C:
			if err := run(); err != nil {
				return err
			}
		}
	}
}
