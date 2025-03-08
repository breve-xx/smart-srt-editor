APP_NAME = editor
GO = go
TEMPL = templ
DOCKER = docker

dependencies:
	@$(GO) mod tidy

build:
	@$(GO) build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

clean:
	@rm -rf bin

generate:
	@$(TEMPL) generate

run: generate
	@$(GO) run cmd/$(APP_NAME)/main.go

docker-build:
	@$(DOCKER) buildx build -t smart-srt-editor:local .

docker-run:
	@$(DOCKER) run --rm -p 8080:8080 smart-srt-editor:local

all: dependencies

.PHONY: dependencies build clean generate run docker-build docker-run all