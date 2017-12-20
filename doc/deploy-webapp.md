We will show how to deploy a web service in kubernetes, and get its response time and request counts with `appMetric`.

## 1. Simple Web Service
Here is the yaml file(*music.yaml*) for the demo web service.

* Make sure the port name in the service is `http`. 


```yaml
---
kind: Service
apiVersion: v1
metadata:
  name: music-service
spec:
  selector:
    app: music-app-pods
  type: ClusterIP
  ports:
    - name: http
      protocol: TCP
      port: 8080
      targetPort: 8080
---
apiVersion: v1
kind: ReplicationController
metadata:
  name: music-app
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: "music-app-pods"
    spec:
      containers:
      - name: "web-app"
        image: beekman9527/webapp
        resources:
          limits:
            cpu: "10m"
          requests:
            cpu: "2m"
        ports: 
        - containerPort: 8080
```

## 2. Deploy the service with the injected sidecar container
Inject the sidecar container, and deploy the demo service in Kubernetes.
```console
kubectl apply -f <(istioctl kube-inject -f musice.yaml)
```

## 3. Access the service
One convenient way to access the service is to access the service from a container in the same Kubernetes cluster:
```yaml
apiVersion: v1
kind: ReplicationController
metadata:
  name: curl
spec:
  replicas: 1
  selector:
    app: "curl-pods"
  template:
    metadata:
      labels:
        app: "curl-pods"
    spec:
      containers:
      - name: "sleep"
        image: beekman9527/curl:latest
        args:
          - --v=2
        resources:
          limits:
            cpu: "10m"
          requests:
            cpu: "2m"
```

Deploy this rc directly (without sidecar injection), and login to the container:
```console
kubectl exec -it curl-bd6kd /bin/bash

## in the container, access appmetric service, and the simple service by:
$ curl music-service.default:8080
$ curl appmetric.default:8081/pod/metrics
$ curl appmetric.default:8081/service/metrics
```


## 4. Check the metrics
The result of `curl appmetric.default:8081/service/metrics`  will be something like:
```json
{"default/music-app-rd52s":
    {"response_time":215.59182417582406,
    "req_per_second":0.52}
}
```
