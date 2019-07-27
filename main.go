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
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

func main() {

	//ideas to expand on filter flag - regex; state values; by label
	//ideas for metric views - rolling update state; hpa's; jobs

	app := cli.NewApp()
	app.Name = "kubestate"
	app.Usage = "Show kubernetes state metrics"
	app.Version = "0.0.3"
	app.Compiled = time.Now()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       "~/.kube/config",
			Usage:       "path to config",
		},
		cli.StringFlag{
			Name:        "namespace, n",
			Value:       "*",
			Usage:       "namespace to show (default is all namespaces)",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "get",
			Usage: "Get metric",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "output, o",
					Value:       "json",
					Usage:       "Output format. Valid formats: json, raw, table",
				},
				cli.StringFlag{
					Name:        "metric, m",
					Value:       "*",
					Usage:       "Metric name to show",
				},
			},
			Action: cmd.Get,
		},
		cli.Command{
			Name: "top",
			Usage: "Show top resource consumption by deployment",
			Subcommands: []cli.Command{
				cli.Command{
					Name: "pods",
					Aliases: []string{"po"},
					Usage: "Get top resource usage for pods",
					Action: cmd.Top,
				},
				cli.Command{
					Name: "deployments",
					Aliases: []string{"deploy"},
					Usage: "Get top resource usage for deployments",
					Action: cmd.Top,
				},
				cli.Command{
					Name: "nodes",
					Aliases: []string{"nodes"},
					Usage: "Get top resource usage for nodes",
					Action: cmd.Top,
				},
			},
		},
		cli.Command{
			Name: "watch",
			Usage: "Watch metric",
			Action: cmd.Watch,
		},
		cli.Command{
			Name: "list",
			Usage: "List metrics",
			Action: cmd.List,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
