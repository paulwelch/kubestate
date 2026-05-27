/*
 * Copyright 2018 Paul Welch
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.
 */

package main

import (
	"github.com/paulwelch/kubestate/cmd"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
)

var version = "dev"

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "kubestate"
	app.Usage = "Show kubernetes state metrics"
	app.Version = version
	app.Compiled = time.Now()

	app.Flags = []cli.Flag{
		&cli.StringFlag{Name: "config, c", Value: "~/.kube/config", Usage: "path to config"},
		&cli.StringFlag{Name: "namespace, n", Value: "*", Usage: "namespace to show (default is all namespaces)"},
		&cli.StringFlag{Name: "metrics-namespace", Usage: "namespace where kube-state-metrics service is running (auto-discovered if unset)"},
		&cli.BoolFlag{Name: "insecure-skip-tls-verify", Usage: "skip TLS certificate verification when connecting to Kubernetes API"},
	}

	app.Commands = []*cli.Command{
		{
			Name:  "get",
			Usage: "Get metric",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "output, o", Value: "json", Usage: "Output format. Valid formats: json, raw"},
				&cli.StringFlag{Name: "metric, m", Value: "*", Usage: "Metric name to show"},
			},
			Action: cmd.Get,
		},
		{
			Name:  "top",
			Usage: "Show top resource consumption by deployment",
			Subcommands: []*cli.Command{
				{Name: "pods", Aliases: []string{"po"}, Usage: "Get top resource usage for pods", Action: cmd.Top},
				{Name: "deployments", Aliases: []string{"deploy"}, Usage: "Get top resource usage for deployments", Action: cmd.Top},
				{Name: "nodes", Usage: "Get top resource usage for nodes", Action: cmd.Top},
			},
		},
		{
			Name:  "watch",
			Usage: "Watch metric",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "output, o", Value: "json", Usage: "Output format. Valid formats: json, raw"},
				&cli.StringFlag{Name: "metric, m", Value: "*", Usage: "Metric name to show"},
				&cli.IntFlag{Name: "interval, i", Value: 10, Usage: "Refresh interval in seconds"},
			},
			Action: cmd.Watch,
		},
		{Name: "list", Usage: "List metrics", Action: cmd.List},
	}

	return app
}

func main() {
	app := newApp()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
