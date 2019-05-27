#!/bin/bash
SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
cd ${SCRIPTPATH}

kubectl apply -f deployment.yaml
kubectl apply -f ingress.yaml
kubectl set image --namespace dev deployment alice-web alice-web=alicews/alice-web:${DRONE_COMMIT}