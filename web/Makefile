MAKEFLAGS += --environment-overrides
-include ../.env
export

.PHONY: all build dev preview check clean

all: build

build:
	@bun run build

dev:
	@bun run dev

preview:
	@bun run preview

check:
	@bun run check

clean:
	@rm -rf build .svelte-kit
