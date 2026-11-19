.PHONY: all build proto clean

ENV ?= release
GOOS := linux
GOARCH := amd64

all: build

build: proto
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(if $(filter dev,$(ENV)),-gcflags="all=-N -l") -o ./dist/$(ENV)/server ./src

proto:
	@mkdir -p gen
	@protoc -I ../proto --go_out=gen --go_opt=paths=source_relative ../proto/*.proto

clean:
	@rm -rf dist
