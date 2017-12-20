We will show how to deploy a web service in kubernetes, and get its response time and request counts with `appMetric`.

## 1. Simple Web Service
Here is the yaml file for the demo web service.

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
        purpose: "service-test"
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
