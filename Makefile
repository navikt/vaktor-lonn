.PHONY: test local

include .env
export

test:
	go test ./... -count=1

env:
	echo "AZURE_APP_CLIENT_ID=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_APP_CLIENT_ID}' | base64 -d)" > .env
	echo "AZURE_APP_CLIENT_SECRET=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_APP_CLIENT_SECRET}' | base64 -d)" >> .env
	echo "AZURE_OPENID_CONFIG_TOKEN_ENDPOINT=$(shell kubectl get --context=dev-gcp --namespace=navdig `kubectl get secret --context=dev-gcp --namespace=navdig --sort-by='{.metadata.creationTimestamp}' -l app=vaktor-lonn,type=azurerator.nais.io -o name | tail -1` -o jsonpath='{.data.AZURE_OPENID_CONFIG_TOKEN_ENDPOINT}' | base64 -d)" >> .env
	echo "MINWINTID_ENDPOINT=http://localhost:8079/json/Hr/Vaktor/Vaktor_Tiddata" >> .env
	echo "MINWINTID_USERNAME=dummy" >> .env
	echo "MINWINTID_PASSWORD=dummy" >> .env
	echo "MINWINTID_INTERVAL=5s" >> .env
	echo "VAKTOR_PLAN_ENDPOINT=dummy" >> .env

local:
	go run .

mock:
	uvicorn MWTmock.app.main:app --port 8079