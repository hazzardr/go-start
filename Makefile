PROJECT_NAME := "go-start"
EXEC_NAME := gostart
SSH_USER := ansible
DEPLOY_TARGET_IP := 100.100.77.57

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

.PHONY: init ## initialize the project
init:
	go run . init

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

.PHONY: clean ## delete generated code
clean:
	rm -rf generated

.PHONY: generate ## generate database code
generate:
	sqlc generate -f sqlc.yaml

.PHONY: build ## builds the project
build:
	$(MAKE) generate
	go build -ldflags='-s' -o=./bin/${PROJECT_NAME} .
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/${PROJECT_NAME} .

.PHONY: lint ## run golangci-lint
lint:
	golangci-lint run

.PHONY: test ## run tests
test:
	go test -v ./...

.PHONY: fmt ## format the project
fmt:
	go fmt ./...


.PHONY: production/connect ## connects to production deployment server
production/connect:
	ssh ${SSH_USER}@${DEPLOY_TARGET_IP}

.PHONY: production/deploy ## deploys to production deployment server
production/deploy:
	$(MAKE) build
	rsync -P ./bin/linux_amd64/${PROJECT_NAME} ${SSH_USER}@${DEPLOY_TARGET_IP}:~
	rsync -P ./remote/${PROJECT_NAME}.service ${SSH_USER}@${DEPLOY_TARGET_IP}:~
	rsync -P ./remote/${PROJECT_NAME}.timer ${SSH_USER}@${DEPLOY_TARGET_IP}:~
	rsync -P .env ${SSH_USER}@${DEPLOY_TARGET_IP}:~
	ssh ${SSH_USER}@${DEPLOY_TARGET_IP} 'chmod 600 ~/.env'
	ssh -t ${SSH_USER}@${DEPLOY_TARGET_IP} '\
	  sudo mv ~/${PROJECT_NAME}.service /etc/systemd/system/ \
	  && sudo mv ~/${PROJECT_NAME}.timer /etc/systemd/system/ \
	  && sudo systemctl daemon-reload \
	  && sudo systemctl enable ${PROJECT_NAME}.timer \
	  && sudo systemctl restart ${PROJECT_NAME}.timer \
	'

.PHONY: production/logs ## gets the logs for the service
production/logs:
	ssh -t ${SSH_USER}@${DEPLOY_TARGET_IP} 'sudo journalctl -u ${PROJECT_NAME}.service'
