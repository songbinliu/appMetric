package util

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMetricSet_AddMetric2(t *testing.T) {
	a := make(MetricSet)
	o1 := &ObjectMetric{
		UID:              "o1",
		Latency:          0.1,
		RequestPerSecond: 10,
	}

	o2 := &ObjectMetric{
		UID:              "o2",
		Latency:          0.01,
		RequestPerSecond: 12.5,
	}

	a[o1.UID] = o1
	a[o2.UID] = o2

	enbytes, err := json.Marshal(a)
	if err != nil {
		t.Errorf("Failed to marshal MetricSet: %v", err)
		return
	}

	fmt.Printf("marshalled: %v\n", string(enbytes))

	var b MetricSet
	if err := json.Unmarshal(enbytes, &b); err != nil {
		t.Errorf("Failed to unmarshall MetricSet: %v", err)
		return
	}
	fmt.Printf("Unmarshaled: %++v\n", b)

	if len(b) != len(a) {
		t.Errorf("unmarshal failed: %d Vs. %d", len(a), len(b))
		return
	}

	for k, v := range b {
		v2, exist := a[k]
		if !exist {
			t.Errorf("key %v does not exist.", k)
			return
		}

		if v.Latency != v2.Latency {
			t.Errorf("key %v 's value dismatch %v Vs. %v", k, v.Latency, v2.Latency)
			return
		}

		if v.RequestPerSecond != v2.RequestPerSecond {
			t.Errorf("key %v 's value dismatch %v Vs. %v", k, v.RequestPerSecond, v2.RequestPerSecond)
			return
		}
	}

	fmt.Println("MetricSet Test1 success.")
}

func TestMetricSet_AddMetric(t *testing.T) {
	a := NewMetricSet()

	a.AddMetric("o1", 10, 11)
	a.AddMetric("o2", 11, 12)

	//1. encode it
	ebytes, err := json.Marshal(a)
	if err != nil {
		t.Errorf("Marshal MetricSet failed: %v", err)
		return
	}

	fmt.Println(string(ebytes))

	//2. decode it
	var b MetricSet
	if err := json.Unmarshal(ebytes, &b); err != nil {
		t.Errorf("Unmarshal MetricSet failed: %v", err)
		return
	}

	for k, v := range b {
		fmt.Printf("k=%v, v=%+v\n", k, v)
	}
}

func TestNewMetricSet_JSON(t *testing.T) {
	a := NewMetricSet()

	a.AddMetric("default/video-xff1", 0.1, 0.2)
	a.AddMetric("default/video-xff2", 0.2, 0.3)
	a.AddMetric("default/details", 0.0076, 0.0628)

	//1. encode it
	ebytes, err := json.Marshal(a)
	if err != nil {
		t.Errorf("Marshal MetricSet failed: %v", err)
		return
	}

	fmt.Println(string(ebytes))

	//2. decode it
	var b MetricSet
	if err := json.Unmarshal(ebytes, &b); err != nil {
		t.Errorf("Unmarshal MetricSet failed: %v", err)
		return
	}

	//3. compare them
	if len(b) != len(a) {
		t.Errorf("unmarshal failed: %d Vs. %d", len(a), len(b))
		return
	}

	for k, v := range b {
		v2, exist := a[k]
		if !exist {
			t.Errorf("key %v does not exist.", k)
			return
		}

		if v.Latency != v2.Latency {
			t.Errorf("key %v 's value dismatch %v Vs. %v", k, v.Latency, v2.Latency)
			return
		}

		if v.RequestPerSecond != v2.RequestPerSecond {
			t.Errorf("key %v 's value dismatch %v Vs. %v", k, v.RequestPerSecond, v2.RequestPerSecond)
			return
		}
	}

}

func TestMetricSet_JSON2(t *testing.T) {
	content := `{"default/details":{"response_time":0.007645090909090926,"req_per_second":0.06285714285714285},
	             "default/inception-be-pods":{"response_time":0.9387358181818174,"req_per_second":0.06285714285714285},
	             "default/productpage":{"response_time":0.06333772727272725,"req_per_second":0.06285714285714285},
	             "default/ratings":{"response_time":0.005956999999999998,"req_per_second":0.039999999999999994},
	             "default/reviews":{"response_time":0.021384818181818168,"req_per_second":0.06285714285714285},
	             "default/video":{"response_time":0.11201881818181818,"req_per_second":0.06285714285714285}}`

	var m MetricSet
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		t.Errorf("Unmrshal MetricSet failed: %v", err)
		return
	}

	for k, v := range m {
		fmt.Printf("k=%v, v=%+v\n", k, v)
	}
}

func TestEntityMetric_Marshall(t *testing.T) {
	em := NewEntityMetric("aid1")
	em.SetLabel("name", "default/curl-1xfj")
	em.SetLabel("ip", "10.0.2.3")
	em.SetLabel("scope", "k8s1")

	em.SetMetric("latency", 133.2)
	em.SetMetric("tps", 12)
	em.SetMetric("readLatency", 50)

	//1. marshal
	ebytes, err := json.Marshal(em)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", em)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var em2 EntityMetric
	if err = json.Unmarshal(ebytes, &em2); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	fmt.Printf("em2=%+v\n", em2)
}

func TestNewMetricResponse(t *testing.T) {
	em := NewEntityMetric("aid1")
	em.SetLabel("name", "default/curl-1xfj")
	em.SetLabel("ip", "10.0.2.3")
	em.SetLabel("scope", "k8s1")

	em.SetMetric("latency", 133.2)
	em.SetMetric("tps", 12)
	em.SetMetric("readLatency", 50)

	em2 := NewEntityMetric("aid2")
	em2.SetLabel("name", "istio/music-ftaf2")
	em2.SetLabel("ip", "10.0.3.2")
	em2.SetLabel("scope", "k8s1")

	em2.SetMetric("latency", 13.2)
	em2.SetMetric("tps", 10)
	em2.SetMetric("readLatency", 5)

	res := NewMetricResponse()
	res.SetStatus(0, "good")
	res.AddMetric(em)
	res.AddMetric(em2)

	//1. marshal it
	ebytes, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", res)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var mr MetricResponse
	if err = json.Unmarshal(ebytes, &mr); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	if mr.Status != 0 || len(mr.Data) < 1{
		t.Errorf("Failed to un-marshal MetricResponse: %+v", res)
		return
	}

	fmt.Printf("mr=%+v, len=%d\n", mr, len(mr.Data))
	for i, e := range mr.Data {
		fmt.Printf("[%d] %+v\n", i, e)
	}
}

func TestNewMetricResponse2(t *testing.T) {
	res := NewMetricResponse()
	res.SetStatus(-1, "error")

	//1. marshal it
	ebytes, err := json.Marshal(res)
	if err != nil {
		t.Errorf("Failed to marshall EntityMetric %+v", res)
		return
	}

	fmt.Println(string(ebytes))

	//2. unmarshal it
	var mr MetricResponse
	if err = json.Unmarshal(ebytes, &mr); err != nil {
		t.Errorf("Failed to un-marshal bytes: %v", string(ebytes))
		return
	}
	if mr.Status == 0 || len(mr.Data) > 0{
		t.Errorf("Failed to un-marshal MetricResponse: %+v", res)
		return
	}

	fmt.Printf("%+v\n", mr)
}
