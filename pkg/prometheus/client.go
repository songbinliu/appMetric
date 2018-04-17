package prometheus

import (
	"appMetric/pkg/util"
	"github.com/golang/glog"
	"github.com/songbinliu/client_prometheus/pkg/prometheus"
)

type EntityMetricGetter interface {
	GetEntityMetric(client *prometheus.RestClient) ([]*util.EntityMetric, error)
	Name() string
	Category() string
}

type Aggregator struct {
	pclient *prometheus.RestClient
	Getters map[string]EntityMetricGetter
}

func NewAggregator(pclient *prometheus.RestClient) *Aggregator {
	result := &Aggregator{
		pclient: pclient,
		Getters: make(map[string]EntityMetricGetter),
	}

	return result
}

func (c *Aggregator) AddGetter(getter EntityMetricGetter) bool {
	name := getter.Name()
	if _, exist := c.Getters[name]; exist {
		glog.Errorf("Entity Metric Getter: %v already exists", name)
		return false
	}

	c.Getters[name] = getter
	return true
}

func (c *Aggregator) GetEntityMetrics() ([]*util.EntityMetric, error) {
	result := []*util.EntityMetric{}
	for _, getter := range c.Getters {
		dat, err := getter.GetEntityMetric(c.pclient)
		if err != nil {
			glog.Errorf("Failed to get entity metrics: %v", err)
			continue
		}

		result = append(result, dat...)
	}

	return result, nil
}
