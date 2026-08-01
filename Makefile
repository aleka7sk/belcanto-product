.PHONY: corpus contract-check mobile-check api-format api-tidy api-vet api-build api-test api-race api-check migrate api-run

corpus:
	python3 .claude/scripts/validate.py

contract-check:
	pnpm contract:lint

mobile-check:
	pnpm mobile:check

api-format:
	@files="$$(find apps/api -type f -name '*.go' -exec gofmt -l {} +)"; test -z "$$files" || { echo "$$files"; exit 1; }

api-tidy:
	cd apps/api && go mod tidy -diff

api-vet:
	cd apps/api && go vet ./...

api-build:
	cd apps/api && go build ./...

api-test:
	cd apps/api && go test ./... -count=1

api-race:
	cd apps/api && go test ./... -race -count=1

api-check: api-format api-tidy api-vet api-build api-test api-race

migrate:
	cd apps/api && go run ./cmd/migrate up

api-run:
	cd apps/api && go run ./cmd/api
