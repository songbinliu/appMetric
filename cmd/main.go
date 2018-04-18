package main

import (
	"flag"
	"github.com/golang/glog"

	myp "appMetric/pkg/prometheus"
	"appMetric/pkg/server"
	"github.com/songbinliu/xfire/pkg/prometheus"
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

func getJobs(mclient *prometheus.RestClient) {
	msg, err := mclient.GetJobs()
	if err != nil {
		glog.Errorf("Failed to get jobs: %v", err)
		return
	}
	glog.V(1).Infof("jobs: %v", msg)
}

func test_prometheus(mclient *prometheus.RestClient) {
	glog.V(2).Infof("Begin to test prometheus client...")
	getJobs(mclient)
	glog.V(2).Infof("End of testing prometheus client.")
	return
}

func main() {
	parseFlags()
	pclient, err := prometheus.NewRestClient(prometheusHost)
	if err != nil {
		glog.Fatal("Failed to generate client: %v", err)
	}
	//mclient.SetUser("", "")
	test_prometheus(pclient)

	appGetter := myp.NewIstioEntityGetter("istio.app.metric")
	appGetter.SetType(false)
	redisGetter := myp.NewRedisEntityGetter("redis.app.metric")
	appClient := myp.NewAggregator(pclient)
	appClient.AddGetter(appGetter)
	appClient.AddGetter(redisGetter)

	vappGetter := myp.NewIstioEntityGetter("istio.vapp.metric")
	vappGetter.SetType(true)
	vappClient := myp.NewAggregator(pclient)
	vappClient.AddGetter(vappGetter)

	s := server.NewMetricServer(port, appClient, vappClient)
	s.Run()
	return
}
