.PHONY: start stop restart unittest integrationtest unitcoverage integrationcoverage testall coverage e2etest

start:
	docker-compose up -d

stop:
	docker-compose down -v

restart:
	docker-compose restart

unittest:
	go test -v ./... -short

integrationtest:
	go test -v ./...

unitcoverage:
	{ \
	go test -v ./... -short -coverprofile="$$PWD/coverage/profile_unit.out" ;\
	go tool cover -html coverage/profile_unit.out ;\
	}

integrationcoverage:
	{ \
	go test -v ./... -coverprofile="$$PWD/coverage/profile_integration.out" ;\
	go tool cover -html coverage/profile_integration.out ;\
	}

testall:
	{ \
	go test -v ./... -short ;\
	go test -v ./... ;\
	}

coverage:
	{ \
	rm -rf coverage/unit ;\
	rm -rf coverage/integration ;\
	mkdir -p coverage/unit ;\
	mkdir -p coverage/integration ;\
	rm coverage/profile.out ;\
	go test -v ./... -short -cover -args -test.gocoverdir="$$PWD/coverage/unit" ;\
	go test -v ./... -cover -args -test.gocoverdir="$$PWD/coverage/integration" ;\
	go tool covdata textfmt -i=./coverage/unit,./coverage/integration -o coverage/profile.out ;\
	go tool cover -html coverage/profile.out ;\
	}

e2etest:
	bash test/e2e/run-api-tests.sh