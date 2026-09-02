.PHONY: test
test:
	@echo "Running tests with in-memory SQLite (fast)..."
	go test ./... -v

.PHONY: test-cloudflare
test-cloudflare:
	@if [ ! -f .env.test ]; then \
		echo "Error: .env.test not found."; \
		echo "Copy .env.test.example and fill in your Cloudflare credentials:"; \
		echo "  cp .env.test.example .env.test"; \
		exit 1; \
	fi
	@echo "Running tests with real Cloudflare D1/R2..."
	@set -a && . ./.env.test && set +a && USE_CLOUDFLARE_TESTS=1 go test ./... -v
