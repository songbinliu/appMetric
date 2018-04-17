package util

import (
	"bytes"
	"fmt"
)

type EntityMetric struct {
	UID     string             `json:"uid,omitempty"`
	Type    int32              `json:"type,omitempty"`
	Labels  map[string]string  `json:"labels,omitempty"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

func NewEntityMetric(id string, t int32) *EntityMetric {
	m := &EntityMetric{
		UID:     id,
		Type:    t,
		Labels:  make(map[string]string),
		Metrics: make(map[string]float64),
	}

	return m
}

func (e *EntityMetric) SetLabel(name, value string) {
	e.Labels[name] = value
}

func (e *EntityMetric) SetMetric(name string, value float64) {
	e.Metrics[name] = value
}

type MetricResponse struct {
	Status  int             `json:"status"`
	Message string          `json:"message:omitemtpy"`
	Data    []*EntityMetric `json:"data:omitempty"`
}

func NewMetricResponse() *MetricResponse {
	return &MetricResponse{
		Status:  0,
		Message: "",
		Data:    []*EntityMetric{},
	}
}

func (r *MetricResponse) SetStatus(v int, msg string) {
	r.Status = v
	r.Message = msg
}

func (r *MetricResponse) AddMetric(m *EntityMetric) {
	r.Data = append(r.Data, m)
}

// --------------- old stuff ----------------
type ObjectMetric struct {
	UID              string  `json:"uid,omitempty"`
	Latency          float64 `json:"response_time,omitempty"`
	RequestPerSecond float64 `json:"req_per_second,omitempty"`
}

func (o *ObjectMetric) String() string {
	buffer := bytes.NewBufferString("")
	buffer.WriteString(fmt.Sprintf("latency=%.5f, rps=%.5f", o.Latency, o.RequestPerSecond))
	if len(o.UID) > 0 {
		buffer.WriteString(fmt.Sprintf(", uid=%v", o.UID))
	}

	return buffer.String()
}

func NewObjectMetric(uid string, resp, rps float64) *ObjectMetric {
	return &ObjectMetric{
		UID:              uid,
		Latency:          resp,
		RequestPerSecond: rps,
	}
}

type MetricSet map[string]*ObjectMetric

func NewMetricSet() MetricSet {
	ret := make(MetricSet)
	return ret
}

func (mset MetricSet) AddMetric(uid string, resp, rps float64) {
	obj := NewObjectMetric(uid, resp, rps)
	obj.UID = ""
	mset[uid] = obj
}

func (mset MetricSet) AddorSetResponeTime(uid string, resp float64) {
	obj, exist := mset[uid]
	if !exist {
		mset.AddMetric(uid, resp, 0.0)
	} else {
		if obj.Latency < resp {
			obj.Latency = resp
		}
	}
}

func (mset MetricSet) AddorSetRPS(uid string, rps float64) {
	obj, exist := mset[uid]
	if !exist {
		mset.AddMetric(uid, 0.0, rps)
	} else {
		if obj.RequestPerSecond < rps {
			obj.RequestPerSecond = rps
		}
	}
}

func (mset MetricSet) String() string {
	buffer := bytes.NewBufferString("")

	buffer.WriteString(fmt.Sprintf("size = %d\n", len(mset)))
	for k, v := range mset {
		buffer.WriteString(fmt.Sprintf("%s, uid=%v\n", v.String(), k))
	}

	return buffer.String()
}
