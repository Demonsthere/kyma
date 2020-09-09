#!/usr/bin/env bash

set -e

KYMA_GATEWAY_CERTS_ISSUER=$(kubectl get secret/kyma-gateway-certs -n istio-system -o jsonpath='{.metadata.annotations.cert-manager\.io/issuer-name}' --ignore-not-found)

if [[ "$KYMA_GATEWAY_CERTS_ISSUER" != "kyma-ca-issuer" ]]; then
    echo Deleting secret kyma-gateway-certs
    kubectl delete secret -n istio-system kyma-gateway-certs --ignore-not-found
fi

APISERVER_PROXY_TLS_CERTS_ISSUER=$(kubectl get secret/apiserver-proxy-tls-cert -n kyma-system -o jsonpath='{.metadata.annotations.cert-manager\.io/issuer-name}' --ignore-not-found)

if [[ "$APISERVER_PROXY_TLS_CERTS_ISSUER" != "kyma-ca-issuer" ]]; then
    echo Deleting secret apiserver-proxy-tls-cert
    kubectl delete secret -n kyma-system apiserver-proxy-tls-cert --ignore-not-found
fi

echo Deleting CM net-global-overrides
kubectl delete cm -n kyma-installer net-global-overrides --ignore-not-found

CM=$(kubectl get cm -n kyma-installer net-global-overrides-copy -o yaml)
CM_NAME="net-global-overrides"
NEW_CM="${CM//net-global-overrides-copy/$CM_NAME}"

cat <<EOF | kubectl apply -f -
---
$NEW_CM
EOF

echo echo Deleting CM net-global-overrides-copy
kubectl delete cm -n kyma-installer net-global-overrides-copy --ignore-not-found
