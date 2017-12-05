package main

import (
	"flag"
	"github.com/golang/glog"

	"appMetric/pkg/prometheus"
	"appMetric/pkg/server"
)

var (
	prometheusHost string
	port           int
)

func parseFlags() {
	flag.Set("logtostderr", "true")
	flag.StringVar(&prometheusHost, "promUrl", "http://localhost:9090", "the address of prometheus server")
	flag.IntVar(&port, "port", 8081, "port to expose metrics")
	flag.Parse()
}

func getJobs(mclient *prometheus.MetricRestClient) {
	msg, err := mclient.GetJobs()
	if err != nil {
		glog.Errorf("Failed to get jobs: %v", err)
		return
	}
	glog.V(1).Infof("jobs: %v", msg)
}

func test_prometheus(mclient *prometheus.MetricRestClient) {
	glog.V(2).Infof("Begin to test prometheus client...")
	getJobs(mclient)
	mset, err := mclient.GetPodMetrics()
	if err != nil {
		glog.Errorf("Failed to get pod metrics: %v", err)
		return
	}

	glog.V(2).Infof("%v", mset.String())
	glog.V(2).Infof("End of testing prometheus client.")
	return
}

func main() {
	parseFlags()
	mclient, err := prometheus.NewRestClient(prometheusHost)
	if err != nil {
		glog.Fatal("Failed to generate client: %v", err)
	}
	test_prometheus(mclient)

	s := server.NewMetricServer(port, mclient)
	s.Run()
	return
}
