---
Title: Implementation diary
Ticket: GO-GO-MCP-EMBEDDABLE-AUTH
Status: active
Topics:
    - mcp
    - oidc
    - authentication
    - introspection
    - dependencies
DocType: working-note
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Chronological record of moving the embedded OIDC server out of pkg/embeddable, fixing embedded_dev expiration, and adding token introspection to external_oidc."
LastUpdated: 2026-10-15T19:35:00-04:00
WhatFor: "Resuming or reviewing GO-GO-MCP-EMBEDDABLE-AUTH."
WhenToUse: "Before continuing the implementation or reviewing the pull request."
---

# Implementation diary

## Step 1: Worktree, ticket, analysis and design

The goal is to make go-go-mcp's `pkg/embeddable` a reasonable dependency for applications such as francois, and to let its `external_oidc` mode accept opaque access tokens. This step set up an isolated worktree, opened the ticket, analysed the auth code and wrote the design guide. It also found that `embedded_dev` mode rejects every token over HTTP on current `main`.

### Prompt Context

**User prompt (verbatim):** "ok, create a new worktree for ~/code/wesen/go-go-golems/go-go-mcp and create a new ticket there with docmgr --root ... and then create a design doc. then implement the go-go-mcp changes and commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill) and then create a PR. Write a report for the vault when done with the go-go-mcp changes."

**Assistant interpretation:** In a new go-go-mcp worktree, create a docmgr ticket and design doc for the two changes identified in francois FRANCOIS-007 (keep the embedded OIDC server out of `pkg/embeddable`; add token introspection), implement them with focused commits, keep this diary, open a pull request, then write an Obsidian vault report.

**Inferred user intent:** Make go-go-mcp the MCP library francois (and similar hosted applications behind ZITADEL) can adopt, by improving the library itself.

### What I did

- `git -C ~/code/wesen/go-go-golems/go-go-mcp worktree add -b task/embeddable-auth-introspection /home/manuel/workspaces/2026-10-15/embeddable-auth-introspection/go-go-mcp origin/main` (base `86db8ed`, equal to local `main`). The path follows the existing worktree convention (`~/workspaces/<date>/<task>/go-go-mcp`).
- `docmgr --root <worktree>/ttmp`: added topic `introspection` and doc type `working-note` to the vocabulary; created ticket `GO-GO-MCP-EMBEDDABLE-AUTH`, the design doc and this diary.
- Read `pkg/embeddable/auth_provider.go`, `auth_provider_external.go`, `auth_context.go`, `server.go:245-345`, `command.go:57-240`, `official_backend.go:300-429`, `pkg/auth/oidc/server.go` (structure, `Config`, `IntrospectAccessToken`), `cmd/go-go-mcp/cmds/oidc.go`, `cmd/go-go-mcp/cmds/server/start.go`, and go-sdk's `auth/auth.go:97-174`.
- Measured imports: `pkg/embeddable` 318 non-standard packages; without `pkg/auth/oidc` 61; `pkg/tools` brings `geppetto/pkg/helpers` (22 packages including watermill) for two reflection helpers.
- Baseline: `GOWORK=off go test ./pkg/embeddable/... ./pkg/auth/...` → `ok` for both packages.
- Probed `embedded_dev` through `authMiddleware` with a throwaway test (deleted afterwards).
- Wrote the design guide (`design-doc/01-…`).

### Why

- The design needed to know every place `pkg/auth/oidc` is used and how the HTTP middleware treats a verifier's result before choosing where the new seam goes.

### What worked

- The embedded mode's only code dependency on `pkg/auth/oidc` is the provider in `auth_provider.go:73-135`; options, constants and flags are plain data, so they can stay.

### What didn't work

- The probe on `main` showed `embedded_dev` is broken over HTTP:

  ```text
  $ GOWORK=off go test ./pkg/embeddable -run TestProbeEmbeddedStaticKeyThroughMiddleware -v
      zz_probe_test.go:20: status=401 body="token missing expiration\n"
  ```

  Cause: the embedded provider returns `AuthPrincipal` without `Expiration`; go-sdk `RequireBearerToken` rejects a zero expiration unless `AllowMissingExpiration` is set (`go-sdk@v1.7.0/auth/auth.go:162-166`). Existing tests miss it because they call the provider directly or use a stub that sets `Expiration`.

### What I learned

- `cmd/go-go-mcp`'s `server start` builds its configuration without auth; only the `oidc` admin commands use `pkg/auth/oidc`, so the go-go-mcp binary keeps working without the option.
- `WWWAuthenticateHeader()` is part of the verifier interface, but the backend lets go-sdk write the challenge from `ResourceMetadataURL`; the method is used by tests and custom verifiers.

### What was tricky to build

- Choosing the seam: the `mcp start` flags must still accept `--auth-mode embedded_dev`, so the mode name and options stay in `pkg/embeddable` while only the implementation moves behind an option.

### What warrants a second pair of eyes

- Removing `embeddable.WithOIDC` without an alias (AGENT.md forbids compatibility layers unless asked).

### What should be done in the future

- Implement the plan in design §8.

### Code review instructions

- Read design §3.2 (request path) and §4 (evidence) first.

### Technical details

- Measurement command: `GOWORK=off GOFLAGS=-mod=mod GOPROXY=off go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./pkg/embeddable | sort -u | wc -l`.

## Step 2: Link `embedded_dev` on request and fix its expiration

This step moved the embedded development OIDC provider out of `pkg/embeddable` into the new package `pkg/embeddable/embeddeddev`, added the `WithAuthModeProvider` seam, and fixed `embedded_dev` tokens being rejected over HTTP. `pkg/embeddable` now compiles 65 non-standard packages instead of 318.

### Prompt Context

**User prompt (verbatim):** see Step 1.

**Assistant interpretation:** Implement design §5.1–5.3.

**Inferred user intent:** A light `pkg/embeddable` that still offers `embedded_dev` to programs that ask for it.

### What I did

- `pkg/auth/oidc/server.go`: `IntrospectAccessToken` now returns `AccessTokenInfo{Subject, ClientID, ExpiresAt}`; fosite tokens take `ExpiresAt` from `ar.GetSession().GetExpiresAt(fosite.AccessToken)`, SQLite development tokens from the stored record.
- `pkg/embeddable/auth_provider.go`: removed the embedded provider and the `pkg/auth/oidc` import; added `AuthModeProvider`, `errEmbeddedDevNotLinked`, and the linked-provider lookup in `newHTTPAuthRuntime`; exported `BearerChallenge`.
- `pkg/embeddable/auth_provider_external.go`, `official_backend.go`: `protectedResourceMetadataURL` → exported `ProtectedResourceMetadataURL`.
- `pkg/embeddable/server.go`: `authModeProviders` map in `ServerConfig`; `WithAuthModeProvider`; removed `WithOIDC` and the `OIDCOptions` alias.
- New `pkg/embeddable/embeddeddev/embeddeddev.go`: `Provide()`, `WithOIDC(opts)`, the provider (static key reports `now + 1 minute`; issued tokens report their expiry). `make logcopter-generate` added its `logcopter.go`.
- `pkg/embeddable/examples/oidc/main.go`: uses `embeddeddev.WithOIDC(embeddable.EmbeddedOIDCOptions{…})`.
- Tests: in `pkg/embeddable`, `TestEmbeddedDevRequiresLinkedProvider`, `TestWithAuthModeProviderSuppliesVerifierAndRoutes`, `TestWithAuthModeProviderRejectsMissingArguments`, `TestImportGraphExcludesEmbeddedAuthorizationServer` (runs `go list -deps .`); in `embeddeddev`, static key and development token expiry tests, an HTTP test through `MountHTTPHandlers` with streamable HTTP, and a no-token 401 test.

### Why

- An explicit option keeps the dependency visible at the call site (design decision §5.1). The expiry fix belongs with the move because the moved provider is where the principal is built.

### What worked

- `GOWORK=off go build ./...`, `go vet`, `GOWORK=off go test ./...` (all packages `ok`), `golangci-lint run` → `0 issues.`, `make logcopter-check` clean.
- Dependency count: `go list -deps ./pkg/embeddable` → 65 non-standard packages (was 318); `./pkg/embeddable/embeddeddev` → 319.
- The HTTP test fails without the fix and passes with it. With the static key's `Expiration` line removed:

  ```text
  --- FAIL: TestEmbeddedDevAcceptsStaticKeyOverHTTP (0.04s)
      embeddeddev_test.go:103: static key: status 401, body "token missing expiration\n"
  ```

  After restoring the file the test passes (Go reported the cached result, confirming the file was byte-identical).

### What didn't work

- No failures beyond the deliberate check above.

### What I learned

- `MountHTTPHandlers` plus `httptest` exercises the real go-sdk middleware, which is what the old provider-level test missed.

### What was tricky to build

- `WithOIDC` must both link the provider and select the mode; it composes `Provide()` and `WithAuth(...)` in one option so option order does not matter for callers.

### What warrants a second pair of eyes

- Removing `embeddable.WithOIDC` and `embeddable.OIDCOptions` is a breaking API change for downstream users when they upgrade (AGENT.md: no compatibility layers).
- The fosite-issued token path (`GetExpiresAt`) is covered by reading, not by a test: obtaining a fosite token needs a full authorization-code flow.

### What should be done in the future

- Introspection (Step 3) and documentation (Step 4).

### Code review instructions

- Start with `pkg/embeddable/auth_provider.go` and `pkg/embeddable/embeddeddev/embeddeddev.go`; run `GOWORK=off go test ./pkg/embeddable/... ./pkg/auth/...`.

### Technical details

- Error text when the mode is not linked: `auth mode embedded_dev is not linked into this program; add embeddeddev.Provide() from github.com/go-go-golems/go-go-mcp/pkg/embeddable/embeddeddev to the server options`.
