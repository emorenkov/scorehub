# Repository Guidelines

## Project Structure & Module Organization
- `cmd/<service>` contains binaries for the user, event, notification, and email services; keep service-specific wiring here.
- Domain logic lives in `pkg/<domain>` (e.g., `pkg/user`, `pkg/common`). Reuse shared helpers instead of duplicating code inside `cmd`.
- Infrastructure manifests are in `deploy/` (`docker-compose.yml`, `k8s/` specs). Database artifacts reside in `db/`; docs and diagrams live in `docs/`.
- Keep generated protobuf code colocated with their definitions under `pkg/*/proto`; re-run generators whenever `.proto` files change.

## Build, Test, and Development Commands
- `make build` compiles all services into `bin/`; run it before opening a PR.
- `make proto` regenerates gRPC stubs via `protoc` and the Go plugins—required after editing any proto files.
- `go test ./...` executes the full go test suite; pair with `GOFLAGS="-count=1"` when chasing flaky behavior.
- `docker compose -f deploy/docker-compose.yml up --build` brings up PostgreSQL, Kafka, and all services locally; stop with `down -v`.
- `kubectl apply -k deploy/k8s` deploys to a cluster; keep manifests in sync with Compose to avoid drift.

## Coding Style & Naming Conventions
- Target Go 1.25+ and run `go fmt ./...` plus `go vet ./...` before committing; keep imports organized via `gofmt`.
- Follow idiomatic Go: PascalCase for exported types/functions, camelCase for internals, ALL_CAPS only for actual constants.
- Keep packages single-purpose; prefer `NewX` constructors that accept interfaces where possible.
- Configuration structs should live in `pkg/common/config` (or closest peer) and load from `local.env` via `cleanenv`.

## Testing Guidelines
- Place tests alongside code as `*_test.go`; use table-driven cases with descriptive names such as `TestNotificationBuilder_Congrats`.
- Aim for meaningful coverage on core flows (event enrichment, Kafka producers/consumers). Validate both happy-path and failure retries.
- Run `go test ./pkg/... ./cmd/...` before pushing; attach `-race` when touching concurrency.
¸
## Commit & Pull Request Guidelines
- Keep commits focused, with imperative summaries similar to the existing `first commit` style (“Add notification fan-out”). Reference issues in the body.
- Each PR should include: concise description, testing evidence (`make build`, `go test ./...`), and screenshots/log excerpts if the change affects output.
- Link related tickets, call out config changes, and request reviews from owners of the touched service directory.

## Configuration & Security Tips
- Never commit secrets; load overrides via `local.env` and document new variables in `README.MD`.
- When adding brokers or databases, update both `deploy/docker-compose.yml` and `deploy/k8s/*.yaml` plus any seeding files in `db/` to keep environments consistent.
