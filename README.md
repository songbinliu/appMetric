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

# Prerequisites
** Kubernetes 1.7.3 +
** Istio 0.2.12 +

