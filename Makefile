PROJECT_NAME := "go-start"
EXEC_NAME := app

.PHONY: help ## print this
help:
	@echo ""
	@echo "$(PROJECT_NAME) Development CLI"
	@echo ""
	@echo "Usage:"
	@echo "  make <command>"
	@echo ""
	@echo "Commands:"
	@grep '^.PHONY: ' Makefile | sed 's/.PHONY: //' | awk '{split($$0,a," ## "); printf "  \033[34m%0-10s\033[0m %s\n", a[1], a[2]}'

.PHONY: run ## Run the project
run:
	go run .

.PHONY: doctor ## checks if local environment is ready for development
doctor:
	@echo "Checking local environment..."
	@if ! command -v go &> /dev/null; then \
		echo "`go` is not installed. Please install it first."; \
		exit 1; \
	fi
	@if [[ ! ":$$PATH:" == *":$$HOME/go/bin:"* ]]; then \
		echo "GOPATH/bin is not in PATH. Please add it to your PATH variable."; \
		exit 1; \
	fi
	@if ! command -v cobra-cli &> /dev/null; then \
		echo "Cobra-cli is not installed. Please run 'make deps'."; \
		exit 1; \
	fi
	@if ! command -v sqlc &> /dev/null; then \
		echo "`sqlc` is not installed. Please run 'make deps'."; \
		exit 1; \
	fi

	@if ! command -v docker &> /dev/null; then \
		echo "`docker` is not installed. Please install it first."; \
		exit 1; \
	fi
	@echo "Local environment OK"


.PHONY: deps ## install dependencies used for development
deps:
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: build ## build the project
build:
	$(MAKE) generate
	go build -o ./bin/$(EXEC_NAME) ./cmd/main.go

.PHONY: clean ## delete generated code
clean:
	rm -rf generated

.PHONY: generate ## generate database code
generate:
	sqlc generate -f sqlc.yaml

.PHONY: lint ## run golangci-lint
lint:
	golangci-lint run

.PHONY: test ## run tests
test:
	go test -v ./...

.PHONY: fmt ## format the project
fmt:
	go fmt ./...