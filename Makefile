MAKEFLAGS += --environment-overrides
-include .env
ENV ?= release
export

.PHONY: all build check clean \
        master-build master-clean \
        slave-build slave-clean \
        server-build server-run server-clean \
        web-build web-dev web-preview web-check web-clean \
        desktop-build desktop-build-web desktop-dev desktop-check desktop-clean

all: build

build: master-build slave-build server-build web-build

check: web-check desktop-check

clean: master-clean slave-clean server-clean web-clean desktop-clean

master-build:
	@$(MAKE) -C master build

master-clean:
	@$(MAKE) -C master clean

slave-build:
	@$(MAKE) -C slave build

slave-clean:
	@$(MAKE) -C slave clean

server-build:
	@$(MAKE) -C server build

server-run:
	@$(MAKE) -C server run

server-clean:
	@$(MAKE) -C server clean

web-build:
	@$(MAKE) -C web build

web-dev:
	@$(MAKE) -C web dev

web-preview:
	@$(MAKE) -C web preview

web-check:
	@$(MAKE) -C web check

web-clean:
	@$(MAKE) -C web clean

desktop-build:
	@$(MAKE) -C desktop build

desktop-build-web:
	@$(MAKE) -C desktop build-web

desktop-dev:
	@$(MAKE) -C desktop dev

desktop-check:
	@$(MAKE) -C desktop check

desktop-clean:
	@$(MAKE) -C desktop clean
