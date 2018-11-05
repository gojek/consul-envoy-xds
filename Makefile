.PHONY: all
all: build fmt vet lint test

APP=consul-envoy-xds
ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")
UNIT_TEST_PACKAGES=$(shell go list ./... | grep -v "featuretests")

APP_EXECUTABLE="./out/$(APP)"

setup-circleci:
	sudo apt-get install unzip
	wget https://releases.hashicorp.com/consul/1.2.3/consul_1.2.3_linux_amd64.zip
	sudo unzip consul_1.2.3_linux_amd64.zip -d /usr/local/bin	

setup:
	go get -u github.com/golang/lint/golint
	go get github.com/DATA-DOG/godog/cmd/godog
	go get -u github.com/go-playground/overalls
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v
	mkdir -p out/
	go build -o $(APP_EXECUTABLE)
	@echo "consul-envoy-xds is setup!! Run make test to run tests"

copy-config:
	cp application.yml.sample application.yml

build-deps:
	go install

update-deps:
	dep ensure -update

compile:
	mkdir -p out/
	go build -o $(APP_EXECUTABLE)

build: build-deps compile fmt vet lint

install:
	go install ./...

fmt:
	go fmt ./...

vet:
	go vet $(ALL_PACKAGES)

lint:
	@for p in $(UNIT_TEST_PACKAGES); do \
		echo "==> Linting $$p"; \
		golint $$p | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; } \
	done

test: compile copy-config
	ENVIRONMENT=test go test $(UNIT_TEST_PACKAGES) -p=1

test-coverage: compile
	@echo "mode: count" > out/coverage-all.out
	env ENVIRONMENT=test overalls -project consul-envoy-xds
	find . -name "profile.coverprofile" -exec rm "{}" \;
	mv overalls.coverprofile out/overalls.coverprofile
	go tool cover -html=out/overalls.coverprofile -o out/coverage.html
