MAKEFLAGS += --environment-overrides
include ../.env
ENV ?= release
export

.PHONY: all build clean

all: build

build:
	@cargo build --locked --profile $(ENV) -p master

clean:
	@cargo clean -p master
