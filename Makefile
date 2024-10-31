.PHONY: unittest integrationtest coverage e2etest

unittest:
	go test -v ./... -short -parallel 4

integrationtest:
	go test -v ./... -parallel 4

coverage:
	go test -v ./... -short -parallel 4 -coverprofile=cover.out && go tool cover -html cover.out

e2etest:
	bash test/run-api-tests.sh