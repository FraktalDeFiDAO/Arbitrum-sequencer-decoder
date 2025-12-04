# Repository Guidelines

## Project Structure & Module Organization
- `cmd/sequencer-reader`, `cmd/sequencer-capture`: CLIs for reading sequencer feeds and capturing transactions.
- `pkg/`: core libraries (decoder, simulator, oracle, arb-engine, executor, types). Keep packages small, lowercase.
- `agent_docs/`, `docs/`: architecture specs, plan, and workflow standards—review before coding.
- `testdata/sequencer/`: JSONL fixtures for integration/replay tests; name by DEX and scenario (e.g., `uniswapv3_swap_smoke.jsonl`).
- `monitoring/`, `audit_reports/`: operational and audit notes; update when adding metrics/checks.

## Build, Test, and Development Commands
- Build binaries: `./dev.sh build` (writes to `bin/`). In-container: `go build -o bin/sequencer-reader ./cmd/sequencer-reader`.
- Tests: `./dev.sh test` or `go test ./... -v`; target a package with `go test ./pkg/decoder -run TestUniswapV3Decoder -v`.
- Lint/format: `./dev.sh lint` runs `golangci-lint`; run `gofmt ./...` before committing.
- Security scan: `./dev.sh security` (gosec). Start/stop dev stack via `./dev.sh dev` / `./dev.sh down`; open a tool shell with `./dev.sh shell`.

## Coding Style & Naming Conventions
- Go 1.25+. Follow `gofmt`; keep imports ordered. `golangci-lint run` must be clean before PR.
- Packages/files lowercase; tests mirror sources with `_test.go`. Exported names use CamelCase; wrap errors with `%w` for context.
- Env/config keys are uppercase snake case (e.g., `ARBITRUM_RPC_URL`). Prefer structured logs with clear component fields.

## Testing Guidelines
- Use table-driven tests in Go `testing`. Aim for ≥80% coverage on new code. Add benchmarks for performance-sensitive paths (`go test ./pkg/simulator -bench=. -benchmem`).
- Reuse `testdata/sequencer/` fixtures; when adding, document scenario and expected outcome in test comments. Integration/replay tests should validate decoder accuracy and price simulation outputs.

## Commit & Pull Request Guidelines
- Branches: `feature/<issue>-slug`, `fix/<issue>-slug`, `release/vX.Y.Z`, `hotfix/<issue>-slug`.
- Commits: `<type>: summary` (feat, fix, docs, style, refactor, perf, test, chore); optional body ≤72 chars/line; include `Fixes #<issue>` when relevant.
- PRs target `develop` (or `main` for release/hotfix). Include summary, testing performed, breaking changes, related issues; attach screenshots for UI/metrics dashboards when applicable. Ensure lint, tests, and security checks are green before review.

## Security & Configuration Tips
- Never commit secrets; use environment variables or injected configs. Run `./dev.sh security` before pushing.
- Add dependencies sparingly and note rationale in PR description. Use Podman Compose stack; registry config lives at `~/.config/containers/registries.conf` as in README.
