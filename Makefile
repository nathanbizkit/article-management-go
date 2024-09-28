.PHONY: unittest coveragetest mock

unittest:
	go test ./...

coveragetest:
	go test ./... -coverprofile=cover.out && go tool cover -html cover.out

mock:
	mockery