SHELL := /bin/bash

# Wallets 
# Quincy: 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA
# Adil: 0x3cA5ddA619be35CE3b7943c5492F7c26f99A3A85
# Kennedy: 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6
# Pavel: 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f
# Valerie: 0xb088B6aD396a87D826676C78c21D385dbe555Fca
# Adam: 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b
# Bookeeping transactions
# curl -il -X GET http://localhost:9080/v1/node/block/list/1/latest
# curl -il -X GET http://localhost:8080/v1/tx/uncommitted/list
# curl -il -X GET http://localhost:8080/v1/start/mining
# curl -il -X GET http://localhost:8080/v1/accounts/list
# ==============================================================================
# Local support


scratch: 
	go run app/tooling/scratch/main.go 

up:
	go run app/services/node/main.go -race | go run app/tooling/logfmt/main.go

up2:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7281 --web-public-host 0.0.0.0:8280 --web-private-host 0.0.0.0:9280 --state-beneficiary=testminer2 --state-db-path zblock/testminer2/ | go run app/tooling/logfmt/main.go

up3:
	go run app/services/node/main.go -race --web-debug-host 0.0.0.0:7381 --web-public-host 0.0.0.0:8380 --web-private-host 0.0.0.0:9380 --state-beneficiary=miner3 --state-db-path zblock/miner3/ | go run app/tooling/logfmt/main.go
	
down:
	kill -INT `ps -x | grep "main -race" | grep -v grep | sed -n 1,1p | cut -c1-5`



accounts:
	curl -il -X GET http://localhost:8080/v1/accounts/list


tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	# go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	go get -u -v ./...
	go mod tidy
	go mod vendor


load10:
	go run app/tooling/cli/main.go send -a quincy -n 1 -f 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -t 0x3cA5ddA619be35CE3b7943c5492F7c26f99A3A85 -v 10000
	go run app/tooling/cli/main.go send -a quincy -n 2 -f 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -t 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6 -v 10000
	go run app/tooling/cli/main.go send -a quincy -n 3 -f 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -t 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f -v 10000
	go run app/tooling/cli/main.go send -a quincy -n 4 -f 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -t 0xb088B6aD396a87D826676C78c21D385dbe555Fca -v 10000
	go run app/tooling/cli/main.go send -a quincy -n 5 -f 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -t 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b -v 10000


load11:
	go run app/tooling/cli/main.go send -a adil -n 1 -f 0x3cA5ddA619be35CE3b7943c5492F7c26f99A3A85 -t 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6 -v 900
	go run app/tooling/cli/main.go send -a kennedy -n 1 -f 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6 -t 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f -v 700
	go run app/tooling/cli/main.go send -a pavel -n 1 -f 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f -t 0xb088B6aD396a87D826676C78c21D385dbe555Fca -v 500
	go run app/tooling/cli/main.go send -a valerie -n 1 3 -f 0xb088B6aD396a87D826676C78c21D385dbe555Fca -t 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b -v 300
	go run app/tooling/cli/main.go send -a adam -n 1 -f 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b -t 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -v 50	

load12:
	go run app/tooling/cli/main.go send -a adil -n 2 -f 0x3cA5ddA619be35CE3b7943c5492F7c26f99A3A85 -t 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6 -v 800
	go run app/tooling/cli/main.go send -a kennedy -n 2 -f 0xDcccd88102dB4719E95BE81e0bAC6D586f3FbEF6 -t 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f -v 600
	go run app/tooling/cli/main.go send -a pavel -n 2 -f 0xA48326a46FebCC7FE6fFB4f7F96E609CfEe4388f -t 0xb088B6aD396a87D826676C78c21D385dbe555Fca -v 400
	go run app/tooling/cli/main.go send -a valerie -n 2  -f 0xb088B6aD396a87D826676C78c21D385dbe555Fca -t 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b -v 200
	go run app/tooling/cli/main.go send -a adam -n 2 -f 0x90dBE80D1430994b9874348615c0c0AbfDbcAf5b -t 0x6e4397Fc40dA776f1b27edb115C53b7fCd6AABbA -v 10