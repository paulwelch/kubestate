package cmd

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	dto "github.com/prometheus/client_model/go"
	"github.com/json-iterator/go"
	"github.com/urfave/cli"
	"fmt"
	"strings"
)

func Get(c *cli.Context) error {

	var config string
	var filterFlag string

	var rawFlag = false
	var jsonFlag = false

	for _, f := range c.App.Flags {
		switch strings.Split(f.GetName(), ",")[0] {
		case "config":
			config = f.(cli.StringFlag).Value
		case "filter":
			filterFlag = f.(cli.StringFlag).Value
		}
	}

	if c.IsSet("raw") {
		rawFlag = true
	}
	if c.IsSet("json") {
		jsonFlag = true
	}

	//Raw output
	if rawFlag {
		resp, err := getRawMetrics(config)
		if err != nil {
			return err
		}

		fmt.Println(resp)
		return nil
	} else if jsonFlag {
		metricFamilies, err := getMetrics(config)
		if err != nil {
			return err
		}

		for i := 0; i < len(metricFamilies); i++ {
			if filterFlag == "*" || *metricFamilies[i].Name == filterFlag {
				if i == 0 {
					fmt.Print("[")
				}
				s, _ := jsoniter.MarshalToString(metricFamilies[i])
				fmt.Print(s)
				if i < (len(metricFamilies) - 1) {
					fmt.Print(",")
				} else {
					fmt.Println("]")
				}
			}
		}
	} else {
		//TODO: default formatted output
		metricFamilies, err := getMetrics(config)
		if err != nil {
			return err
		}

		for i := 0; i < len(metricFamilies); i++ {
			fmt.Println("---------------")
			fmt.Println(*metricFamilies[i].Name)
			fmt.Println(*metricFamilies[i].Type)
			fmt.Println(*metricFamilies[i].Help)

			//for debugging
			for i := 0; i < len(metricFamilies[i].Metric); i++ {
				for j := 0; j < len((*metricFamilies[i].Metric[i]).Label); j++ {

					fmt.Println("---------------")
					fmt.Printf("Metric %d: Label %d:  %s  value: %s\n", i, j, *metricFamilies[i].Metric[i].Label[0].Name, *metricFamilies[i].Metric[i].Label[0].Value)

					switch *metricFamilies[i].Type {

					case dto.MetricType_COUNTER:
						fmt.Printf("Counter Value: %f", *metricFamilies[i].Metric[i].Counter.Value)

					case dto.MetricType_GAUGE:
						fmt.Printf("Gauge Value: %f", *metricFamilies[i].Metric[i].Gauge.Value)

					case dto.MetricType_SUMMARY:
						fmt.Println(*metricFamilies[i].Metric[i].Summary.Quantile[0].Value)
						fmt.Println(*metricFamilies[i].Metric[i].Summary.Quantile[0].Quantile)
						fmt.Println(*metricFamilies[i].Metric[i].Summary.SampleCount)
						fmt.Println(*metricFamilies[i].Metric[i].Summary.SampleSum)

					}
				}
			}
		}
	}
	return nil
}
