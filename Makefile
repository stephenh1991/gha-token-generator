IMG ?= steveh1991/gha-token-generator
TAG ?= latest
LATEST_RELEASE = $(shell git describe --tags --abbrev=0)
GORELEASER_VERSION = v0.173.2
GORELEASER_ARCH = $(shell uname -m)
GORELEASER_OS = $(shell uname -s)

generate-test-data:
	ssh-keygen -t rsa -b 4096 -f ./cmd/token-generator/testdata/key.pem -m pem -q -N ""; \
	cat ./cmd/token-generator/testdata/key.pem | base64 > ./cmd/token-generator/testdata/key_64.pem;

test:
	go mod download; \
	go test -v ./cmd/token-generator;\

build:
	go mod download; \
	go build -o ./bin/token-generator ./cmd/token-generator

docker-build:
	docker build -t $(IMG):$(TAG) .
	docker tag $(IMG):$(TAG) $(IMG):$(LATEST_RELEASE)
	docker tag $(IMG):$(TAG) $(IMG):latest

docker-push:
	docker push $(IMG):$(TAG)
	docker push $(IMG):$(LATEST_RELEASE)
	docker push $(IMG):latest

install-goreleaser:
	mkdir -p bin
	curl -sSL https://github.com/goreleaser/goreleaser/releases/download/${GORELEASER_VERSION}/goreleaser_${GORELEASER_OS}_${GORELEASER_ARCH}.tar.gz | tar -xz -C ./bin goreleaser

release-dryrun:
	./bin/goreleaser -f ./.goreleaser.yml --snapshot --skip-publish --rm-dist --debug

release:
	./bin/goreleaser -f ./.goreleaser.yml --rm-dist
