SQLC_VERSION := v1.31.1
SQLC_BIN := $(CURDIR)/bin/tools/sqlc/$(SQLC_VERSION)/sqlc

.PHONY: setup authentik-up authentik-down authentik-logs authentik-status authentik-check run build check sqlc-install sqlc-generate sqlc-check

$(SQLC_BIN):
	GOBIN="$(dir $(SQLC_BIN))" go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

sqlc-install: $(SQLC_BIN)
	@"$(SQLC_BIN)" version

sqlc-generate: $(SQLC_BIN)
	"$(SQLC_BIN)" generate -f sqlc.yaml

sqlc-check: $(SQLC_BIN)
	"$(SQLC_BIN)" compile -f sqlc.yaml
	"$(SQLC_BIN)" diff -f sqlc.yaml

setup:
	@sh scripts/setup-authentik.sh

authentik-up: setup
	@sh scripts/authentik.sh up -d

authentik-down:
	@sh scripts/authentik.sh down

authentik-logs:
	@sh scripts/authentik.sh logs -f --tail=100 server worker

authentik-status:
	@sh scripts/authentik.sh ps

authentik-check:
	@sh scripts/authentik.sh config >/dev/null

run:
	go run ./cmd/api

build:
	go build -o bin/tenantledger-api ./cmd/api

check:
	go test ./...
	go vet ./...
