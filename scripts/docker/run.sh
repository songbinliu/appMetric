#!/bin/bash

url=localhost:9090
docker run -d -p 18081:8081 beekman9527/appmetric:v2 --promUrl=$url --v=3 --logtostderr
sleep 1
docker ps
