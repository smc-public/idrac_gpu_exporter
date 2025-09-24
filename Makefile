VERSION  = $(or $(shell git tag --points-at HEAD | grep -oP 'v\K[0-9.]+'), unknown)
REVISION = $(shell git rev-parse HEAD)

REPOSITORY := github.com/smc-public/idrac_gpu_exporter
LDFLAGS    := -X $(REPOSITORY)/internal/version.Version=$(VERSION)
LDFLAGS    += -X $(REPOSITORY)/internal/version.Revision=$(REVISION)
GOFLAGS    := -ldflags "$(LDFLAGS)"
RUNFLAGS   ?= -config config.yml -verbose

build:
	go build $(GOFLAGS) -o idrac_gpu_exporter ./cmd/idrac_gpu_exporter

run:
	go run ./cmd/idrac_gpu_exporter $(RUNFLAGS)
