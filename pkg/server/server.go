package server

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"strings"

	"appMetric/pkg/prometheus"
	"appMetric/pkg/util"
)

type MetricServer struct {
	port       int
	ip         string
	host       string
	promClient *prometheus.MetricRestClient
}

func NewMetricServer(port int, pclient *prometheus.MetricRestClient) *MetricServer {
	ip, err := util.ExternalIP()
	if err != nil {
		glog.Errorf("Failed to get server IP: %v", err)
		ip = "localhost"
	}

	host, err := os.Hostname()
	if err != nil {
		glog.Errorf("Failed to get hostname: %v", err)
		host = "localhost"
	}
	glog.V(2).Infof("Will server on %s:%d", ip, port)

	return &MetricServer{
		port:       port,
		ip:         ip,
		host:       host,
		promClient: pclient,
	}
}

func (s *MetricServer) Run() {
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s,
	}

	glog.V(1).Infof("HTTP server listens on: %s", server.Addr)
	panic(server.ListenAndServe())
}

func (s *MetricServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	glog.V(2).Infof("Begin to handle path: %v", path)

	if strings.EqualFold(path, "/favicon.ico") {
		s.faviconHandler(w, r)
		return
	}

	if strings.EqualFold(path, "/pod/metrics") {
		s.handlePodMetric(w, r)
		return
	}

	if strings.EqualFold(path, "/service/metrics") {
		s.handleServiceMetric(w, r)
		return
	}

	//if strings.EqualFold(path, "/metrics") {
	//	s.handleMetrics(w, r)
	//}

	s.handleWelcome(path, w, r)
	return
}
