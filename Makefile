VERSION=0.0.1
NAME=terraform-provider-vra7
TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=vra7
WEBSITE_REPO=github.com/hashicorp/terraform-website

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
TEST?=$(shell go list ./... |grep -v 'vendor')

.PHONY: all build check clean dev test testacc fmt fmtcheck websitefmtcheck lint tools  simplify race release

all: check build

build:
	for os in darwin linux windows; do \
	  GOARCH=amd64 GOOS=$$os go build -o ${NAME}-$$os; \
	done

check:
	@gofmt -d ${SRC}
	@test -z "$(shell gofmt -l ${SRC} | tee /dev/stderr)" || { echo "Fix formatting issues with 'make fmt'"; exit 1; }
	@ret=0 && for d in $$(go list ./... | grep -v /vendor/); do \
		test -z "$$(golint $${d} | tee /dev/stderr)" || ret=1; \
		done; exit $$ret
	@go tool vet main.go
	@go tool vet vra7
	go test ./...

clean:
	rm -rf pkg
	rm terraform-provider-vra7*

dev:
	GOARCH=$$(go env GOARCH) GOOS=$$(go env GOOS) go install

test: fmtcheck
	go test $(TEST) -timeout=30s -parallel=4

testacc:
	echo "TEST: " $(TEST)
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 240m

fmt:
	@gofmt -l -w $(SRC)
	
# Currently required by tf-deploy compile
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

websitefmtcheck:
	@sh -c "'$(CURDIR)/scripts/websitefmtcheck.sh'"

lint:
	@echo "==> Checking source code against linters..."
	@GOGC=30 golangci-lint run ./$(PKG_NAME)

tools:
	GO111MODULE=on go install github.com/client9/misspell/cmd/misspell
	GO111MODULE=on go install github.com/golangci/golangci-lint/cmd/golangci-lint

simplify:
	@gofmt -s -l -w $(SRC)

race:
	go test -race ./...

release:
	for os in darwin linux windows; do \
	  mkdir -p pkg/$$os; \
	  GOARCH=amd64 GOOS=$$os go build -o pkg/$$os/${NAME}; \
	  (cd pkg/$$os; zip ../../${NAME}_${VERSION}_$${os}_amd64.zip ${NAME}); \
	done
	

