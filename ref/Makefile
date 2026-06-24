# PolyBot — pre-deploy gates.
#
# Run `make check` before SCP/deploy. Catches the kind of typos and missing
# dependencies that otherwise only surface inside a Docker build on EC2.

GO_SERVICES := services/ingestion strategies/strategy-arb

.PHONY: check tidy build-ingestion build-strategy-arb

check:
	@set -e; \
	for d in $(GO_SERVICES); do \
		echo "==> go vet $$d"; (cd $$d && go vet ./...); \
		echo "==> go build $$d"; (cd $$d && go build ./...); \
	done
	@echo "OK — all Go services compile."

tidy:
	@set -e; \
	for d in $(GO_SERVICES); do \
		echo "==> go mod tidy $$d"; (cd $$d && go mod tidy); \
	done

build-ingestion:
	docker compose build ingestion

build-strategy-arb:
	docker compose build strategy-arb-cross-market
