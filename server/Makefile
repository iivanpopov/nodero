MAKEFLAGS += --environment-overrides
-include ../.env
ENV ?= release
export

.PHONY: all build run proto migrate-up migrate-down migrate-status clean

GOOS := linux
GOARCH := amd64

all: build

build: proto
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(if $(filter dev,$(ENV)),-gcflags="all=-N -l") -o ./dist/$(ENV)/server ./cmd/api

run:
	@go run ./cmd/api

proto:
	@mkdir -p gen
	@protoc -I ../proto --go_out=gen --go_opt=paths=source_relative ../proto/*.proto

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
