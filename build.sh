#!/bin/bash 
docker build -t 10.29.230.150:31381/library/sms-test:20251103 .
docker push 10.29.230.150:31381/library/sms-test:20251103


kubectl apply -f deploy.yml
kubectl get -n sms-webhook deploy
kubectl get -n sms-webhook svc