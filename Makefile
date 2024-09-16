SHELL:=/bin/bash

ifdef test_run
	TEST_ARGS := -run $(test_run)
endif

migrate_up=go run main.go migrate --direction=up --step=0
migrate_down=go run main.go migrate --direction=down --step=0
run_worker_command=go run main.go worker
run_command=go run main.go server

run: check-modd-exists
	@modd -f ./.modd/server.modd.conf

lint: check-cognitive-complexity
	golangci-lint run --print-issued-lines=false --exclude-use-default=false --enable=revive --enable=goimports  --enable=unconvert --enable=unparam --concurrency=2

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif

ifdef test_run
	$(eval TEST_ARGS := -run $(test_run))
endif
	$(eval test_command=$(TEST_BIN) ./... $(TEST_ARGS) -v --cover)

test-only: check-gotest mockgen
	SVC_DISABLE_CACHING=true $(test_command)

test: lint test-only

check-modd-exists:
	@modd --version > /dev/null	

run-worker: check-modd-exists
	@modd -f ./.modd/worker.modd.conf

run-telegram-bot: check-modd-exists
	@modd -f ./.modd/telegram-bot.modd.conf

check-cognitive-complexity:
	find . -type f -name '*.go' -not -name "*.pb.go" -not -name "mock*.go" -not -name "generated.go" -not -name "federation.go" \
      -exec gocognit -over 15 {} +

migrate:
	@if [ "$(DIRECTION)" = "" ] || [ "$(STEP)" = "" ]; then\
    	$(migrate_up);\
	else\
		go run main.go migrate --direction=$(DIRECTION) --step=$(STEP);\
    fi