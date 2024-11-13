.PHONY: unittest integrationtest unitcoverage integrationcoverage coverage e2etest

unittest:
	go test -v ./... -short -parallel 4

integrationtest:
	go test -v ./... -parallel 4

unitcoverage:
	go test -v ./... -short -parallel 4 -coverprofile=cover.out && go tool cover -html cover.out

integrationcoverage:
	go test -v ./... -parallel 4 -coverprofile=cover.out && go tool cover -html cover.out

coverage:
	{ \
	mkdir -p coverage/unit ;\
	mkdir -p coverage/integration ;\
	go test -v ./... -short -parallel 4 -cover -args -test.gocoverdir="$$PWD/coverage/unit" ;\
	go test -v ./... -parallel 4 -cover -args -test.gocoverdir="$$PWD/coverage/integration" ;\
	go tool covdata textfmt -i=./coverage/unit,./coverage/integration -o coverage/profile.out ;\
	go tool cover -html coverage/profile.out ;\
	}

e2etest:
	bash test/run-api-tests.sh