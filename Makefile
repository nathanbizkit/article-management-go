.PHONY: unittest coveragetest e2etest

unittest:
	go test -v ./... -parallel 4

coveragetest:
	go test -v ./... -parallel 4 -coverprofile=cover.out && go tool cover -html cover.out

e2etest:
	bash test/run-api-tests.sh