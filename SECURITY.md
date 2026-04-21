# Security Policy

## Supported versions

The `syncmap` library follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html). Security fixes land on the most recent minor release of the current major version. Older majors (once a `v2.0.0` exists) are not supported.

| Version | Supported |
|---------|-----------|
| `v1.x` (latest minor) | Yes |
| Older `v1.x` minors | No |
| Pre-1.0 (`v0.x`) | Never released |

## Threat model

`github.com/axonops/syncmap` is a type-safe generic wrapper around the Go standard library's [`sync.Map`](https://pkg.go.dev/sync#Map). It exposes the same set of operations with compile-time type safety in place of call-site type assertions. It has **zero runtime dependencies** outside the standard library.

**In scope:**

- Correctness of the wrapper under concurrent use. Every method is safe for concurrent use by multiple goroutines without additional locking, inherited directly from `sync.Map`.
- Type-assertion safety at the `sync.Map` boundary. All internal `any → V` assertions are guarded so that the library cannot panic on the documented public API surface.
- Zero-value distinction. `Load`, `LoadAndDelete`, and `Swap` correctly distinguish "value V is the zero value of its type" from "no entry is present" via the `ok` / `loaded` return, matching the stdlib `sync.Map` contract.
- `CompareAndSwap` and `CompareAndDelete` — exposed as package-level generic functions with a tighter `V comparable` constraint so non-comparable value types (slice, map, func) are rejected at compile time rather than panicking at runtime inside `sync.Map`.
- No orphaned goroutines: the library spawns none of its own.
- Build and release supply chain: reproducible builds, pinned dependencies, signed releases via CI.

**Out of scope:**

- Denial of service from pathological key distributions — `sync.Map` itself makes no complexity guarantees about hashing, and this wrapper does not change that.
- Comparison panics when `V` is an interface type whose dynamic value is itself not comparable. This matches Go's `==` semantics for interfaces and is documented on `CompareAndSwap`.
- Memory exhaustion from unbounded insertion — the library provides no eviction policy. Bound the key space at the caller.
- Use of the map to cache security-sensitive material. Clearing a value from the map does not guarantee the underlying memory is zeroed; the Go runtime may retain it until garbage collection.

## Reporting a vulnerability

**Do not open a public issue for a suspected vulnerability.**

Email **oss@axonops.com** with:

- A concise description of the issue.
- Steps to reproduce, including the Go version and OS/architecture.
- Any proof-of-concept code, crash reports, or `go test -race` output.
- Your preferred attribution (name, handle, or anonymous).

We will:

- Acknowledge receipt within **3 business days**.
- Share a mitigation plan within **14 business days**.
- Coordinate an embargoed release with you if a fix requires a new tag.
- Credit you in the release notes and in this repository's security advisories unless you request otherwise.

## Dependency security

Runtime dependencies: **none**. Test dependencies are pinned in `go.mod`:

- `github.com/stretchr/testify`
- `github.com/cucumber/godog`
- `go.uber.org/goleak`

CI runs [`govulncheck`](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) on every push and pull request and fails the build on any vulnerability in called code. Dependabot tracks upstream advisories weekly.
