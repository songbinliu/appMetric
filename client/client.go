package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"appMetric/pkg/util"
)

const (
	defaultTimeOut   = time.Duration(60 * time.Second)
	API_PATH_POD     = "/pod/metrics"
	API_PATH_SERVICE = "/service/metrics"
)

type AppMetricClientConfig struct {
	// a string with hostname and port, http://localhost:8081
	Host string
}

type AppMetricClient struct {
	client *http.Client
	config *AppMetricClientConfig
}

func NewAppMetricClient(conf *AppMetricClientConfig) (*AppMetricClient, error) {
	//1. get http client
	client := &http.Client{
		Timeout: defaultTimeOut,
	}

	//2. check whether it is using ssl
	addr, err := url.Parse(conf.Host)
	if err == nil && addr.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	glog.V(2).Infof("AppMetrics server address is: %v", conf.Host)
	app := &AppMetricClient{
		client: client,
		config: conf,
	}

	return app, nil
}

func (c *AppMetricClient) GetMetrics(path string) (util.MetricSet, error) {
	p := fmt.Sprintf("%v%v", c.config.Host, path)
	glog.V(2).Infof("path=%v", p)

	//1. set up request setting
	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	//2. send request and get response
	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	//3. parse response
	if resp.StatusCode != 200 {
		glog.Errorf("response code != 200: %d", resp.StatusCode)
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return nil, err
	}

	var mset util.MetricSet
	if err = json.Unmarshal(result, &mset); err != nil {
		glog.Errorf("Unmarshal result failed: %v\n%v", err, string(result))
		return nil, err
	}

	glog.V(3).Infof("Get %d metrics for %v", len(mset), path)

	return mset, nil
}

func (c *AppMetricClient) GetPodAppMetrics() (util.MetricSet, util.MetricSet, error) {
	var podMetric util.MetricSet
	var svcMetric util.MetricSet
	var err1 error
	var err2 error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if podMetric, err1 = c.GetMetrics(API_PATH_POD); err1 != nil {
			glog.Errorf("Failed to get Pod metrics: %v", err1)
			podMetric = util.NewMetricSet()
			return
		}

		glog.V(2).Infof("Got %d pod metrics.", len(podMetric))
		glog.V(4).Infof("Pod metrics: %v", podMetric.String())
	}()

	go func() {
		defer wg.Done()
		if svcMetric, err2 = c.GetMetrics(API_PATH_SERVICE); err2 != nil {
			glog.Errorf("Failed to get Service metrics: %v", err2)
			svcMetric = util.NewMetricSet()
			return
		}

		glog.V(2).Infof("Got %d service metrics.", len(svcMetric))
		glog.V(4).Infof("Service metrics: %v", svcMetric.String())
	}()

	wg.Wait()
	if err1 != nil && err2 != nil {
		err := fmt.Errorf("Not able to get Pod metrics, nor Service metrics.")
		glog.Errorf(err.Error())
		return podMetric, svcMetric, err
	}

	return podMetric, svcMetric, nil
}
