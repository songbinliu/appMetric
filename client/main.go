package main

import (
	"flag"
)

var (
	host = "http://localhost:8081"
)

func parseFlags() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&host, "serverUrl", "http://localhost:8081", "the address of app metrics server")
	flag.Parse()
}

func main() {
	parseFlags()
	c := &AppMetricClientConfig{
		Host: host,
	}

	client, _ := NewAppMetricClient(c)

	client.GetPodAppMetrics()
	//client.GetMetrics(API_PATH_SERVICE)
}
