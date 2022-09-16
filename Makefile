.PHONY: test local

-include .env

test:
	go test ./... -count=1

env:
	echo "VAKTOR_CLIENT_ID=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_APP_CLIENT_ID}' | base64 -d)" > .env
	echo "VAKTOR_CLIENT_SECRET=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_APP_CLIENT_SECRET}' | base64 -d)" >> .env
	echo "VAKTOR_TOKEN_ENDPOINT=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_OPENID_CONFIG_TOKEN_ENDPOINT}' | base64 -d)" >> .env

local:
	VAKTOR_CLIENT_ID=$(VAKTOR_CLIENT_ID) \
	VAKTOR_CLIENT_SECRET=$(VAKTOR_CLIENT_SECRET) \
	VAKTOR_TOKEN_ENDPOINT=$(VAKTOR_TOKEN_ENDPOINT) \
	go run .
