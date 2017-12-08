# appMetric
Get [Kubernestes](https://kubernetes.io) service and pod `latency` and `request-per-seconds` from [Istio](https://istio.io).

# Kubernetes service latency and request-per-second
format: <namespace/svcName>: {"response_time": <ms>, "req_per_second": <rps>}
  
```json
{
	"default/details": {
		"response_time": 8.339090909086881,
		"req_per_second": 0.06285714285714285
	},
	"default/inception-be-pods": {
		"response_time": 958.370636363146,
		"req_per_second": 0.06285714285714285
	},
	"default/productpage": {
		"response_time": 54.26372727271717,
		"req_per_second": 0.06285714285714285
	},
	"default/ratings": {
		"response_time": 6.075428571431222,
		"req_per_second": 0.039999999999999994
	},
	"default/reviews": {
		"response_time": 20.26118181818116,
		"req_per_second": 0.06285714285714285
	},
	"default/video": {
		"response_time": 108.67336363635745,
		"req_per_second": 0.06285714285714285
	}
}
```

# Deploy

## Prerequisites
* Kubernetes 1.7.3 +
* Istio 0.2.12 +

## Deploy metrics and rules in Istio
Isito metrics, handlers and rules are defined in [script](https://github.com/songbinliu/appMetric/blob/master/scripts/istio/turbo.metric.yaml), deploy it with:
```console
istioctl create -f scripts/istio/turbo.metric.yaml
```
**Four Metrics**: pod latency, pod request count, service latency and service request count.
**One Handler**: a `Prometheus handler]` to consume the four metrics, and generate metrics in [Prometheus](https://prometheus.io) format. This server will provide REST API to get the metrics from Prometheus.
**One Rule**: Only the `http` based metrics will be handled by the defined handler.

## Run REST API Server
build and run this go application:
```console
make build
./_output/appMetric --v=3 --promUrl=http://localhost:9090 --port=8081
```

Then the server will serve on port `8081`; access the REST API by:
```console
curl http://localhost:8081/service/metrics
```
```json
{"default/details":{"response_time":13.135999999995173,"req_per_second":0.06285714285714285},"default/inception-be-pods":{"response_time":953.5242727268435,"req_per_second":0.06285714285714285},"default/productpage":{"response_time":76.38181818180617,"req_per_second":0.06285714285714285},"default/ratings":{"response_time":8.805875000001961,"req_per_second":0.04571428571428571},"default/reviews":{"response_time":28.504636363632844,"req_per_second":0.06285714285714285},"default/video":{"response_time":111.38272727271216,"req_per_second":0.06285714285714285}}
```

Alternately, this REST API service can also deployed in Kubernetes:
```console
kubectl create -f scripts/k8s/deploy.yaml

# Access it in Kubernetes by service name:
curl http://appmetric.default:8081/service/metrics
```


