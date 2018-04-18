package prometheus

import (
	"github.com/golang/glog"

	"appMetric/pkg/inter"
	"github.com/songbinliu/xfire/pkg/prometheus"
)

type EntityMetricGetter interface {
	GetEntityMetric(client *prometheus.RestClient) ([]*inter.EntityMetric, error)
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

func (c *Aggregator) GetEntityMetrics() ([]*inter.EntityMetric, error) {
	result := []*inter.EntityMetric{}
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
