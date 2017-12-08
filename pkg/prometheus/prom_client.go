package prometheus

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"appMetric/pkg/util"
)

const (
	API_PATH       = "/api/v1/"
	API_QUERY_PATH = "/api/v1/query"
	API_RANGE_PATH = "/api/v1/query_range"

	defaultTimeOut = time.Duration(60 * time.Second)

	TURBO_SVC_LATENCY_SUM   = "turbo_service_latency_time_ms_sum"
	TURBO_SVC_LATENCY_COUNT = "turbo_service_latency_time_ms_count"
	TURBO_SVC_REQUEST_COUNT = "turbo_service_request_count"

	TURBO_POD_LATENCY_SUM   = "turbo_pod_latency_time_ms_sum"
	TURBO_POD_LATENCY_COUNT = "turbo_pod_latency_time_ms_count"
	TURBO_POD_REQUEST_COUNT = "turbo_pod_request_count"

	TURBO_LATENCY_DURATION = "3m"

	K8S_PREFIX     = "kubernetes://"
	K8S_PREFIX_LEN = len(K8S_PREFIX)
)

type MetricRestClient struct {
	client *http.Client
	host   string
}

func NewRestClient(host string) (*MetricRestClient, error) {
	//1. get http client
	client := &http.Client{
		Timeout: defaultTimeOut,
	}

	//2. check whether it is using ssl
	addr, err := url.Parse(host)
	if err != nil {
		glog.Errorf("Invalid url:%v, %v", host, err)
		return nil, err
	}
	if addr.Scheme == "https" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	glog.V(2).Infof("Prometheus server address is: %v", host)

	return &MetricRestClient{
		client: client,
		host:   host,
	}, nil
}

func (c *MetricRestClient) GetJobs() (string, error) {
	p := fmt.Sprintf("%v%v%v", c.host, API_PATH, "label/job/values")
	glog.V(2).Infof("path=%v", p)

	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return "", err
	}

	glog.V(3).Infof("resp: %++v", resp)
	return string(result), nil
}

func getLatencyExp(pod bool) string {
	name_sum := ""
	name_count := ""
	if pod {
		name_sum = TURBO_POD_LATENCY_SUM
		name_count = TURBO_POD_LATENCY_COUNT
	} else {
		name_sum = TURBO_SVC_LATENCY_SUM
		name_count = TURBO_SVC_LATENCY_COUNT
	}
	du := TURBO_LATENCY_DURATION

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])/rate(%v{response_code=\"200\"}[%v])", name_sum, du, name_count, du)
	return result
}

// exp = rate(turbo_request_count{response_code="200",  source_service="unknown"}[3m])
func getRPSExp(pod bool) string {
	name_count := ""
	if pod {
		name_count = TURBO_POD_REQUEST_COUNT
	} else {
		name_count = TURBO_SVC_REQUEST_COUNT
	}
	du := TURBO_LATENCY_DURATION

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])", name_count, du)
	return result
}

// convert the UID from "kubernetes://<podName>.<namespace>" to "<namespace>/<podName>"
// for example, "kubernetes://video-671194421-vpxkh.default" to "default/video-671194421-vpxkh"
func convertPodUID(uid string) (string, error) {
	if !strings.HasPrefix(uid, K8S_PREFIX) {
		return "", fmt.Errorf("Not start with %v", K8S_PREFIX)
	}

	items := strings.Split(uid[K8S_PREFIX_LEN:], ".")
	if len(items) < 2 {
		return "", fmt.Errorf("Not enough fields: %v", uid[K8S_PREFIX_LEN:])
	}

	if len(items) > 2 {
		glog.Warningf("expected 2, got %d for: %v", len(items), uid[K8S_PREFIX_LEN:])
	}

	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	if len(items[0]) < 1 || len(items[1]) < 1 {
		return "", fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}

// convert UID from "svcName.namespace.svc.cluster.local" to "svcName.namespace"
// for example, "productpage.default.svc.cluster.local" to "default/productpage"
func convertSVCUID(uid string) (string, error) {
	if uid == "unknown" {
		return "", fmt.Errorf("unknown")
	}

	//1. split it
	items := strings.Split(uid, ".")
	if len(items) < 3 {
		err := fmt.Errorf("Not enough fields %d Vs. 3", len(items))
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//2. check the 3rd field
	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	items[2] = strings.TrimSpace(items[2])
	if items[2] != "svc" {
		err := fmt.Errorf("%v fields[2] should be [svc]: [%v]", uid, items[2])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	//3. construct the new uid
	if len(items[0]) < 1 || len(items[1]) < 1 {
		err := fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
		glog.V(3).Infof(err.Error())
		return "", err
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}

func (c *MetricRestClient) GetPodMetrics() (util.MetricSet, error) {
	mset := util.NewMetricSet()

	//1. get latency
	exp := getLatencyExp(true)
	glog.V(2).Infof("Pod latency exp= %v", exp)

	if err := c.GetLatency(exp, mset); err != nil {
		glog.Errorf("Failed to get Pod Latency: %v", err)
	}

	//2. get RPS
	exp = getRPSExp(true)
	glog.V(2).Infof("Pod RPS exp=%v", exp)
	if err := c.GetRPS(exp, mset); err != nil {
		glog.Errorf("Failed to get Pod RequestPerSecond: %v", err)
	}

	//3. check the set
	glog.V(2).Infof("Get %d pod metrics", len(mset))
	if len(mset) < 1 {
		glog.Warningf("Failed to get any pod metrics.")
		return mset, nil
	}

	//4. convert the UID from "kubernetes://<podName>.<namespace>" to "<podName>/<namespace>"
	result := util.NewMetricSet()
	for k, v := range mset {
		newKey, err := convertPodUID(k)
		if err != nil {
			continue
		}

		result[newKey] = v
	}

	glog.V(3).Infof("Pod metrics:\n%v", result.String())
	glog.V(2).Infof("Get %d pod metrics in the end", len(result))
	return result, nil
}

func (c *MetricRestClient) GetServiceMetrics() (util.MetricSet, error) {
	mset := util.NewMetricSet()

	//1. get latency
	exp := getLatencyExp(false)
	glog.V(2).Infof("Service latency exp= %v", exp)

	if err := c.GetLatency(exp, mset); err != nil {
		glog.Errorf("Failed to get service Latency: %v", err)
	}

	//2. get RPS
	exp = getRPSExp(false)
	glog.V(2).Infof("Service RPS exp=%v", exp)
	if err := c.GetRPS(exp, mset); err != nil {
		glog.Errorf("Failed to get service RequestPerSecond: %v", err)
	}

	//3. check the set
	glog.V(2).Infof("Get %d service metrics", len(mset))
	if len(mset) < 1 {
		glog.Warningf("Failed to get any service metrics.")
		return mset, nil
	}

	//4. convert the UID from "svcName.namespace.svc.cluster.local" to "svcName.namespace"
	result := util.NewMetricSet()
	for k, v := range mset {
		newKey, err := convertSVCUID(k)
		if err != nil {
			continue
		}

		result[newKey] = v
	}

	glog.V(3).Infof("Service metrics:\n%v", result.String())
	glog.V(2).Infof("Get %d service metrics in the end", len(result))
	return result, nil
}

// curl 'http://10.60.1.16:9090/api/v1/query?query=rate%28service_latency_time_ms_sum%7Bresponse_code%3D%22200%22%7D%5B1m%5D%29%2Frate%28service_latency_time_ms_count%7Bresponse_code%3D%22200%22%7D%5B1m%5D%29%20'
//
func (c *MetricRestClient) GetLatency(exp string, mset util.MetricSet) error {
	//1. do query
	result, err := c.Query(exp)
	if err != nil {
		glog.Errorf("Failed to get response time: %v", err)
		return err
	}

	if result.ResultType != "vector" {
		err := fmt.Errorf("Unexpected result type: %v Vs. vector", result.ResultType)
		glog.Error(err.Error())
		return err
	}
	glog.V(4).Infof("result.type=%v, \n result: %+v", result.ResultType, string(result.Result))

	//3. parse/decode the response
	var resp []TurboLatency
	if err := json.Unmarshal(result.Result, &resp); err != nil {
		glog.Errorf("Failed to unmarshal response time: %v", err)
		return err
	}

	for i := range resp {
		value := float64(resp[i].Value.Value)
		if math.IsNaN(value) {
			continue
		}

		uid := resp[i].Metric.DestinationUID
		mset.AddorSetResponeTime(uid, value)
		glog.V(4).Infof("[%d] latency=%.5f, uid=%v", i, value, uid)
	}

	return nil
}

func (c *MetricRestClient) GetRPS(exp string, mset util.MetricSet) error {
	//2. do query
	result, err := c.Query(exp)
	if err != nil {
		glog.Errorf("Failed to get pod requqest-per-second: %v", err)
		return err
	}

	if result.ResultType != "vector" {
		err := fmt.Errorf("Unexpected result type: %v Vs. vector", result.ResultType)
		glog.Error(err.Error())
		return err
	}
	glog.V(4).Infof("result.type=%v, \n result: %+v",
		result.ResultType, string(result.Result))

	//3. parse/decode the response
	var resp []RequestPerSecond
	if err := json.Unmarshal(result.Result, &resp); err != nil {
		glog.Errorf("Failed to unmarshal: %v", err)
		return err
	}

	for i := range resp {
		value := float64(resp[i].Value.Value)
		if math.IsNaN(value) {
			continue
		}

		// convert it from seconds to milliseconds;
		value = value * 1000.0

		uid := resp[i].Metric.DestinationUID
		mset.AddorSetRPS(uid, value)
		glog.V(4).Infof("[%d] requestPerSecond=%.5f, uid=%v", i, value, uid)
	}

	return nil
}

func (c *MetricRestClient) Query(query string) (*PromData, error) {
	p := fmt.Sprintf("%v%v", c.host, API_QUERY_PATH)
	glog.V(2).Infof("path=%v", p)

	req, err := http.NewRequest("GET", p, nil)
	if err != nil {
		glog.Errorf("Failed to generate a http.request: %v", err)
		return nil, err
	}

	//1. set query
	q := req.URL.Query()
	q.Set("query", query)
	req.URL.RawQuery = q.Encode()

	//2. set headers
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		glog.Errorf("Failed to send http request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read response: %v", err)
		return nil, err
	}

	ss := PromeResponse{}
	if err := json.Unmarshal(result, &ss); err != nil {
		glog.Errorf("Failed to unmarshall respone: %v", err)
		return nil, err
	}

	if ss.Status == "error" {
		return nil, fmt.Errorf(ss.Error)
	}

	glog.V(4).Infof("resp: %++v", resp)
	glog.V(4).Infof("metric: %+++v", ss)
	return ss.Data, nil
}
