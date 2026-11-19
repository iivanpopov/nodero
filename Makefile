MAKEFLAGS += --environment-overrides
include .env
ENV ?= release
export

.PHONY: all build clean master-build master-clean slave-build \
        slave-clean server-proto server-build server-clean

all: build

build: master-build slave-build server-build

clean: master-clean slave-clean server-clean

master-build:
	@$(MAKE) -C master build

master-clean:
	@$(MAKE) -C master clean

slave-build:
	@$(MAKE) -C slave build

slave-clean:
	@$(MAKE) -C slave clean

server-proto:
	@$(MAKE) -C server proto

server-build:
	@$(MAKE) -C server build

server-clean:
	@$(MAKE) -C server clean
