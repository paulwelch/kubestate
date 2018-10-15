package cmd

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"github.com/urfave/cli"
	"fmt"
	"strings"
)

func List(c *cli.Context) error {

	var config string

	for _, f := range c.App.Flags {
		if strings.Split(f.GetName(), ",")[0] == "config" {
			config = f.(cli.StringFlag).Value
		}
	}

	metricFamilies, err := getMetrics(config)
	if err != nil {
		return err
	}

	for i := 0; i < len(metricFamilies); i++ {
		fmt.Printf("%s\t%s\t%s\n", *metricFamilies[i].Type, *metricFamilies[i].Name, *metricFamilies[i].Help)
	}

	return nil
}
