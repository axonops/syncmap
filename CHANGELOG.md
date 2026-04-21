# Changelog

All notable changes to `github.com/axonops/syncmap` are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

No unreleased changes.

## Upgrading

From `v1.0.0` onwards `syncmap` follows the standard Go semantic-versioning
compatibility promise: breaking changes to the public API only in a new
major version. Minor and patch releases are always backwards-compatible
for the API surface documented on [pkg.go.dev](https://pkg.go.dev/github.com/axonops/syncmap).
Pin a specific tag in your `go.mod`, review the release notes for the
target version, and run your test suite with `-race` against the new
version before rolling to production.

## [1.0.0] — 2026-04-21

Initial AxonOps release. Forked from [`github.com/rgooding/go-syncmap`](https://github.com/rgooding/go-syncmap) by Richard Gooding so the library can be consumed under AxonOps engineering controls — reproducible release workflow, signed releases, CI quality gates, security scanning, and CLA governance. The upstream project is fully usable on its own; this fork exists only to package it for AxonOps. Library behaviour is unchanged apart from a handful of small API additions tracked below.

### Added

- **Core API** mirroring [`sync.Map`](https://pkg.go.dev/sync#Map): `Load`, `Store`, `LoadOrStore`, `LoadAndDelete`, `Delete`, `Range` — each generic over `K comparable, V any`, returning the typed zero value of `V` on miss so callers never deal with untyped `any`.
- **Extension methods** beyond `sync.Map`: `Len`, `Map` (snapshot as a plain Go map), `Keys`, `Values` — each documented as `O(n)` and a point-in-time approximation under concurrent mutation.
- **`Swap` method** wrapping `sync.Map.Swap` (Go 1.20) with the same typed-zero-on-miss guard as `LoadAndDelete`.
- **`Clear` method** wrapping `sync.Map.Clear` (Go 1.23).
- **`CompareAndSwap` and `CompareAndDelete`** as package-level generic functions with a tighter `[K, V comparable]` constraint, so non-comparable value types (slice, map, func) are rejected at compile time rather than panicking at runtime inside `sync.Map`. The `SyncMap[K, V any]` type signature is unchanged.
- **Package documentation** (`doc.go`) covering relationship to `sync.Map`, when to use `SyncMap` vs `sync.Map` vs `map + sync.RWMutex`, thread safety, zero-value usability, and a runnable Quick Start.
- **Unit tests** — external black-box package (`syncmap_test`) using `testify` and [`go.uber.org/goleak`](https://pkg.go.dev/go.uber.org/goleak). Every test runs under `-race` with `t.Parallel()`. Line coverage is **100%** of the library package.
- **Runnable godoc examples** covering every public symbol, each ending with a deterministic `// Output:` block.
- **Benchmarks** for every public method, plus a concurrent 90/10 read-write pattern and overhead pairs comparing the generic wrapper against raw `sync.Map`. The committed `bench.txt` baseline is the reference the CI `benchstat-regression-guard` job diffs against.
- **BDD suite** — [`godog`](https://pkg.go.dev/github.com/cucumber/godog) feature files under `tests/bdd/` exercising every public symbol plus a concurrent Store scenario. Runs under strict mode enforced by a CI guard.
- **Fuzz targets** — `FuzzLoadStore` (round-trip invariant) and `FuzzConcurrent` (4 goroutines over random op sequences, race-clean).
- **CI** (`.github/workflows/ci.yml`): format check, vet, golangci-lint, unit + BDD tests, 95% coverage threshold, module tidy, govulncheck, cross-platform builds (`linux/amd64`, `darwin/arm64`, `windows/amd64`), benchstat regression guard, BDD strict-mode guard, Apache-header guard, no-local-paths guard, no-AI-attribution guard, Makefile-targets guard, markdown lint, and `llms-full.txt` drift guard.
- **Release workflow** (`.github/workflows/release.yml`): `workflow_dispatch` only; verifies, tags, publishes via GoReleaser, warms the Go module proxy. Local tag creation is forbidden.
- **Dependabot** configuration with weekly updates and auto-merge for patch-level test dependencies.
- **LLM documentation bundle**: `llms.txt` (concise summary) and `llms-full.txt` (concatenated corpus) for AI-assistant ingestion, with a CI guard that fails the build on drift.
- `LICENSE` (Apache 2.0, preserved from upstream), `NOTICE` (crediting Richard Gooding as the upstream author), `SECURITY.md`.

### Changed

- Module path: `github.com/rgooding/go-syncmap` → `github.com/axonops/syncmap`.
- Minimum Go toolchain raised to **1.26**.

### Breaking

- Renamed the `Items()` method on `SyncMap` to `Values()` to match Go stdlib convention (`maps.Values`, Go 1.23). No deprecation shim — the rename lands pre-v1.0 under the new module path.

### Attribution

This release is a fork of [`github.com/rgooding/go-syncmap`](https://github.com/rgooding/go-syncmap) by Richard Gooding, which is distributed under Apache 2.0; this fork continues under the same licence. The original upstream copyright is preserved in git history and credited in `NOTICE`.

[Unreleased]: https://github.com/axonops/syncmap/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/axonops/syncmap/releases/tag/v1.0.0
