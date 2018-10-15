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

	//TODO: expand on filter flag - possible regex, state values, by namespace, by label
	//TODO: maybe match kubectl command pattern: get, describe, watch
	//      ideas for metric views - rolling update state; deployment state, hpa's, jobs, etc
	//TODO: add output format options (raw, json, table)
	//TODO: add reasonable defaults with no command or flags - maybe a 'top' display

	app := cli.NewApp()
	app.Name = "kubestate"
	app.Usage = "Show kubernetes state metrics"
	app.Version = "0.0.1"
	app.Compiled = time.Now()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       "~/.kube/config",
			Usage:       "path to config",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "get",
			Usage: "Get metrics",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "output, o",
					Value:       "json",
					Usage:       "Output format. Valid formats: json, raw, table",
				},
				cli.StringFlag{
					//TODO: should filter be a subcommand?
					Name:        "filter, f",
					Value:       "*",
					Usage:       "Metric filter family to show",
				},
			},
			Action: cmd.Get,
		},
		cli.Command{
			Name: "top",
			Usage: "Show top resource consumption by deployment",
			Action: cmd.Top,
		},
		cli.Command{
			Name: "watch",
			Usage: "Watch metric",
			Action: cmd.Watch,
		},
		cli.Command{
			Name: "list",
			Usage: "List metric families",
			Action: cmd.List,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
