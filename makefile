SHELL := /bin/bash

# Bookeeping transactions
# curl -il -X GET http://localhost:9080/v1/node/block/list/1/latest


# ==============================================================================
# Local support


scratch: 
	go run app/tooling/scratch/main.go 

up: 
	go run app/services/node/main.go -race



tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -v ./...
	go mod tidy
	go mod vendor	