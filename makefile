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




load:
	go run app/tooling/cli/main.go send -a quincy -n 1 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 100
	go run app/tooling/cli/main.go send -a adil -n 1 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76 -v 75
	go run app/tooling/cli/main.go send -a kennedy -n 2 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 150
	go run app/tooling/cli/main.go send -a pavel -n 2 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 125
	go run app/tooling/cli/main.go send -a valerie -n 3 -f 0xF01813E4B85e178A83e29B8E7bF26BD830a25f32 -t 0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0 -v 200
	go run app/tooling/cli/main.go send -a adam -n 3 -f 0xdd6B972ffcc631a62CAE1BB9d80b7ff429c8ebA4 -t 0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9 -v 250	