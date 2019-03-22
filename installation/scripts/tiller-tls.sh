#!/usr/bin/env bash -e

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

while :
do
  if [[ $(kubectl get -n kyma-installer secret helm-secret) ]]
    then
      echo "---> Secrets have been created"
      break
    else
      echo "---> Secrets not present. Waiting 5s..."
      sleep 5
    fi
done

mkdir -p "$(helm home)"

echo "---> Get Helm secrets and put then into $(helm home)"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.ca\.crt']}" | base64 --decode > "$(helm home)/ca.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.crt']}" | base64 --decode > "$(helm home)/cert.pem"
kubectl get -n kyma-installer secret helm-secret -o jsonpath="{.data['global\.helm\.tls\.key']}" | base64 --decode > "$(helm home)/key.pem"
