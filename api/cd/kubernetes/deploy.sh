#!/bin/bash
SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
cd ${SCRIPTPATH}

kubectl apply -f deployment.yaml
kubectl apply -f ingress.yaml
kubectl set image --namespace dev deployment alice alice=alicews/alice:${DRONE_COMMIT}