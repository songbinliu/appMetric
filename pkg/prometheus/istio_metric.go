package prometheus

import (
	"appMetric/pkg/util"
	"bytes"
	"fmt"
	"github.com/golang/glog"
	pclient "github.com/songbinliu/client_prometheus/pkg/prometheus"
	"math"
	"strings"
)

const (
	// NOTO: for istio 2.x, the prefix "istio_" should be removed
	turbo_SVC_LATENCY_SUM   = "istio_turbo_service_latency_time_ms_sum"
	turbo_SVC_LATENCY_COUNT = "istio_turbo_service_latency_time_ms_count"
	turbo_SVC_REQUEST_COUNT = "istio_turbo_service_request_count"

	turbo_POD_LATENCY_SUM   = "istio_turbo_pod_latency_time_ms_sum"
	turbo_POD_LATENCY_COUNT = "istio_turbo_pod_latency_time_ms_count"
	turbo_POD_REQUEST_COUNT = "istio_turbo_pod_request_count"

	turboLatencyDuration = "3m"

	k8sPrefix    = "kubernetes://"
	k8sPrefixLen = len(k8sPrefix)

	podTPS     = 0
	podLatency = 1
	svcTPS     = 2
	svcLatency = 3

	podType = 1
	svcType = 2
)

type IstioEntityGetter struct {
	name  string
	query *IstioQuery
	etype int //Pod(Application), or Service
}

func NewIstioEntityGetter(name string) *IstioEntityGetter {
	return &IstioEntityGetter{
		name:  name,
		etype: podType,
		query: NewIstioQuery(),
	}
}

func (istio *IstioEntityGetter) Name() string {
	return istio.name
}

func (istio *IstioEntityGetter) SetType(isVirtualApp bool) {
	if isVirtualApp {
		istio.etype = svcType
	} else {
		istio.etype = podType
	}
}

func (istio *IstioEntityGetter) Category() string {
	if istio.etype == podType {
		return "Istio.Application"
	}

	return "Istio.VirtualApplication"
}

func (istio *IstioEntityGetter) GetEntityMetric(client *pclient.RestClient) ([]*util.EntityMetric, error) {
	result := []*util.EntityMetric{}

	if istio.etype == podType {
		istio.query.SetQueryType(podTPS)
	} else {
		istio.query.SetQueryType(svcTPS)
	}
	tpsDat, err := client.GetMetrics(istio.query)
	if err != nil {
		glog.Errorf("Failed to get Pod Transaction metrics: %v", err)
		return result, err
	}

	if istio.etype == podType {
		istio.query.SetQueryType(podLatency)
	} else {
		istio.query.SetQueryType(svcLatency)
	}
	latencyDat, err := client.GetMetrics(istio.query)
	if err != nil {
		glog.Errorf("Failed to get pod Latency metrics: %v", err)
		return result, err
	}

	glog.V(4).Infof("len(TPS)=%d, len(Latency)=%d", len(tpsDat), len(latencyDat))

	result = istio.mergeTPSandLatency(tpsDat, latencyDat)

	return result, nil
}

func (istio *IstioEntityGetter) assignMetric(entity *util.EntityMetric, metric *IstioMetricData) {
	for k, v := range metric.Labels {
		entity.SetLabel(k, v)
	}

	//2. other information
	entity.SetLabel("metric.category", istio.Category())
}

func (istio *IstioEntityGetter) mergeTPSandLatency(tpsDat, latencyDat []pclient.MetricData) []*util.EntityMetric {
	result := []*util.EntityMetric{}
	midresult := make(map[string]*util.EntityMetric)
	etype := util.ApplicationType
	if istio.etype == svcType {
		etype = util.VirtualApplicationType
	}

	for _, dat := range tpsDat {
		tps, ok := dat.(*IstioMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for TPS: not an IstioMetricData")
			continue
		}

		entity := util.NewEntityMetric(tps.uuid, etype)

		istio.assignMetric(entity, tps)
		entity.SetMetric(util.TPS, tps.GetValue())
		midresult[entity.UID] = entity
		glog.V(5).Infof("uid=%v,uid2=%v, %+v", entity.UID, tps.uuid, entity)
	}

	for _, dat := range latencyDat {
		latency, ok := dat.(*IstioMetricData)
		if !ok {
			glog.Errorf("Type assertion failed for Latency: not an IstioMetricData")
			continue
		}

		entity, exist := midresult[latency.uuid]
		if !exist {
			glog.V(3).Infof("Some entity does not have TPS metric: %+v", latency)
			entity = util.NewEntityMetric(latency.uuid, etype)
			midresult[entity.UID] = entity
			istio.assignMetric(entity, latency)
		}
		entity.SetMetric(util.Latency, latency.GetValue())
		glog.V(5).Infof("uid=%v, %+v", entity.UID, entity)
	}

	glog.V(4).Infof("len(midResult) = %d", len(midresult))

	for _, entity := range midresult {
		result = append(result, entity)
	}

	return result
}

// IstioQuery : generate queries for Istio-Prometheus metrics
// qtype 0: pod.request-per-second
//       1: pod.latency
//       2: service.request-per-second
//       3: service.latency
type IstioQuery struct {
	qtype    int
	queryMap map[int]string
}

// IstioMetricData : hold the result of Istio-Prometheus data
type IstioMetricData struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	uuid   string
	dtype  int //0,1,2,3 same as qtype
}

// NewIstioQuery : create a new IstioQuery
func NewIstioQuery() *IstioQuery {
	q := &IstioQuery{
		qtype:    0,
		queryMap: make(map[int]string),
	}

	isPod := true
	q.queryMap[podTPS] = getRPSExp(isPod)
	q.queryMap[1] = getLatencyExp(isPod)
	isPod = false
	q.queryMap[2] = getRPSExp(isPod)
	q.queryMap[3] = getLatencyExp(isPod)

	return q
}

func (q *IstioQuery) SetQueryType(t int) error {
	if t < 0 {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
		glog.Error(err)
		return err
	}

	if t > len(q.queryMap) {
		err := fmt.Errorf("Invalid query type: %d, vs 0|1|2|3", t)
		glog.Error(err)
		return err
	}

	q.qtype = t

	return nil
}

func (q *IstioQuery) GetQueryType() int {
	return q.qtype
}

func (q *IstioQuery) GetQuery() string {
	return q.queryMap[q.qtype]
}

func (q *IstioQuery) Parse(m *pclient.RawMetric) (pclient.MetricData, error) {
	d := NewIstioMetricData()
	d.SetType(q.qtype)
	if err := d.Parse(m); err != nil {
		glog.Errorf("Failed to parse metrics: %s", err)
		return nil, err
	}

	return d, nil
}

func (q *IstioQuery) String() string {
	var buffer bytes.Buffer

	for k, v := range q.queryMap {
		tmp := fmt.Sprintf("qtype:%d, query=%s", k, v)
		buffer.WriteString(tmp)
	}

	return buffer.String()
}

func NewIstioMetricData() *IstioMetricData {
	return &IstioMetricData{
		Labels: make(map[string]string),
	}
}

func (d *IstioMetricData) Parse(m *pclient.RawMetric) error {
	d.Value = float64(m.Value.Value)
	if math.IsNaN(d.Value) {
		return fmt.Errorf("Failed to convert value: NaN")
	}

	labels := m.Labels

	//1. pod/svc Name
	v, ok := labels["destination_uid"]
	if !ok {
		err := fmt.Errorf("No content for destination uid: %v+", m.Labels)
		return err
	}
	uid, err := d.parseUID(v)
	if err != nil {
		glog.Errorf("Failed to parse UID(%v): %v", v, err)
		return err
	}
	d.Labels[util.Name] = uid
	d.uuid = uid

	//2. ip
	v, ok = labels["destination_ip"]
	if !ok {
		glog.Errorf("No destination_ip label: %v", labels)
		return nil
	}

	ip, err := d.parseIP(v)
	if err != nil {
		glog.Errorf("Failed to parse IP(%v): %v", v, err)
		return nil
	}
	d.Labels[util.IP] = ip

	//NOTO: set uuid to its IP if available
	d.uuid = ip
	return nil
}

func (d *IstioMetricData) parseUID(muid string) (string, error) {
	if d.dtype < 2 {
		return convertPodUID(muid)
	}

	return convertSVCUID(muid)
}

// input: [0 0 0 0 0 0 0 0 0 0 255 255 10 2 1 84]
// output: 10.2.1.84
func (d *IstioMetricData) parseIP(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < 7 {
		return "", fmt.Errorf("Illegal string")
	}

	content := raw[1 : len(raw)-1]
	items := strings.Split(content, " ")
	i := len(items) - 4

	result := fmt.Sprintf("%v.%v.%v.%v", items[i], items[i+1], items[i+2], items[i+3])
	return result, nil
}

func (d *IstioMetricData) SetType(t int) {
	d.dtype = t
}

func (d *IstioMetricData) GetEntityID() string {
	return d.uuid
}

func (d *IstioMetricData) GetValue() float64 {
	return d.Value
}

func (d *IstioMetricData) String() string {
	var buffer bytes.Buffer

	uid := d.GetEntityID()
	content := fmt.Sprintf("uid=%v, value=%.5f", uid, d.GetValue())
	buffer.WriteString(content)

	return buffer.String()
}

// GetIstioMetric :
//   An example to get the 4 kinds of metrics from Istio-Prometheus
func GetIstioMetric(client *pclient.RestClient) {
	q := NewIstioQuery()

	for i := 0; i < 4; i++ {
		q.SetQueryType(i)
		result, err := client.GetMetrics(q)
		if err != nil {
			glog.Errorf("Failed to get metric: %v", err)
		}

		msg := "Pod QPS"
		if i == 1 {
			msg = "Pod Latency"
		} else if i == 2 {
			msg = "Service QPS"
		} else if i == 3 {
			msg = "Service Latency"
		}

		glog.V(2).Infof("====== %v =========", msg)
		for i := range result {
			glog.V(2).Infof("\t[%d] %v", i, result[i])
		}
	}
}

func getLatencyExp(pod bool) string {
	name_sum := ""
	name_count := ""
	if pod {
		name_sum = turbo_POD_LATENCY_SUM
		name_count = turbo_POD_LATENCY_COUNT
	} else {
		name_sum = turbo_SVC_LATENCY_SUM
		name_count = turbo_SVC_LATENCY_COUNT
	}
	du := turboLatencyDuration

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])/rate(%v{response_code=\"200\"}[%v])", name_sum, du, name_count, du)
	return result
}

// exp = rate(turbo_request_count{response_code="200",  source_service="unknown"}[3m])
func getRPSExp(pod bool) string {
	name_count := ""
	if pod {
		name_count = turbo_POD_REQUEST_COUNT
	} else {
		name_count = turbo_SVC_REQUEST_COUNT
	}
	du := turboLatencyDuration

	result := fmt.Sprintf("rate(%v{response_code=\"200\"}[%v])", name_count, du)
	return result
}

// convert the UID from "kubernetes://<podName>.<namespace>" to "<namespace>/<podName>"
// for example, "kubernetes://video-671194421-vpxkh.default" to "default/video-671194421-vpxkh"
func convertPodUID(uid string) (string, error) {
	if !strings.HasPrefix(uid, k8sPrefix) {
		return "", fmt.Errorf("Not start with %v", k8sPrefix)
	}

	items := strings.Split(uid[k8sPrefixLen:], ".")
	if len(items) < 2 {
		return "", fmt.Errorf("Not enough fields: %v", uid[k8sPrefixLen:])
	}

	if len(items) > 2 {
		glog.Warningf("expected 2, got %d for: %v", len(items), uid[k8sPrefixLen:])
	}

	items[0] = strings.TrimSpace(items[0])
	items[1] = strings.TrimSpace(items[1])
	if len(items[0]) < 1 || len(items[1]) < 1 {
		return "", fmt.Errorf("Invalid fields: %v/%v", items[0], items[1])
	}

	nid := fmt.Sprintf("%s/%s", items[1], items[0])
	return nid, nil
}

// 10.10.172.236:9100
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