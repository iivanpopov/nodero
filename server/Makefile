MAKEFLAGS += --environment-overrides
-include ../.env
ENV ?= release
export

.PHONY: all build run migrate-up migrate-down migrate-status clean

GOOS := linux
GOARCH := amd64

all: build

build:
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(if $(filter dev,$(ENV)),-gcflags="all=-N -l") -o ./dist/$(ENV)/server ./cmd/api

run:
	@go run ./cmd/api

migrate-up:
	@goose up

migrate-down:
	@goose down

migrate-status:
	@goose status

clean:
	@rm -rf dist

sqlc:
	sqlc generate
