MAKEFLAGS += --environment-overrides
-include ../.env
ENV ?= release
export

.PHONY: all build build-web dev check clean

all: build

build:
	@bun run tauri build $(if $(filter dev,$(ENV)),--debug)

build-web:
	@bun run build

dev:
	@bun run tauri dev

check:
	@bun run check

clean:
	@rm -rf build .svelte-kit
	@cargo clean -p desktop
