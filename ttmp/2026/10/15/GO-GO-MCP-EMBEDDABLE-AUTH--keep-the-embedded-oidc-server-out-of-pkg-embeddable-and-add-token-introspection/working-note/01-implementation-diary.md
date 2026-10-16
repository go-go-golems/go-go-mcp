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

## Step 3: Token introspection in `external_oidc`

This step added RFC 7662 token introspection to the `external_oidc` mode with three validation settings (`jwt`, `introspection`, `auto`), four `mcp start` flags, startup checks, and a distinction between "the token is invalid" (401) and "the identity provider could not answer" (500).

### Prompt Context

**User prompt (verbatim):** see Step 1.

**Assistant interpretation:** Implement design §5.4.

**Inferred user intent:** Let go-go-mcp verify opaque tokens such as those ZITADEL issues to dynamically registered MCP clients.

### What I did

- `pkg/embeddable/server.go`: `TokenValidation` (`jwt`, `introspection`, `auto`) and `ExternalOIDCOptions.TokenValidation`, `IntrospectionURL`, `IntrospectionClientID`, `IntrospectionClientSecret`.
- `pkg/embeddable/auth_provider_external.go`:
  - discovery reads `introspection_endpoint`;
  - `configureTokenValidation` checks the mode, endpoint URL rules and credentials at startup;
  - `validateBearerToken` dispatches, and the previous body became `validateJWT`;
  - `introspect` posts `token` and `token_type_hint=access_token` with `client_secret_basic` (credentials form-encoded first, RFC 6749 §2.3.1), limits the body to 1 MiB, and checks `active`, `iss`, `exp` (required), `nbf`, `aud`, `token_type != refresh_token` and required scopes;
  - `looksLikeJWS` and `isBase64URLRune`;
  - `errAuthorizationServerUnavailable` for transport, status and decoding failures.
- `pkg/embeddable/official_backend.go`: `officialExternalOIDCMiddleware` returns `errAuthorizationServerUnavailable` errors unwrapped (go-sdk answers 500) and logs them; other failures stay `ErrInvalidToken` (401).
- `pkg/embeddable/command.go`: `--oidc-token-validation`, `--oidc-introspection-url`, `--oidc-introspection-client-id`, `--oidc-introspection-client-secret-file` (read and trimmed).
- `pkg/embeddable/auth_provider_introspection_test.go`: a fake identity provider (discovery, JWKS, introspection recording Basic credentials and form) and eight tests: active token mapping; ten rejection cases; 500 and invalid JSON; `auto` routing (valid JWT and forged JWT never introspected, opaque token introspected); `looksLikeJWS`; startup checks; the HTTP middleware (204, 401, 500); flag parsing.
- Design §5.4 gained the 500-versus-401 paragraph.

### Why

- Opaque tokens can only be verified by the issuer. `auto` lets one deployment accept both JWT and opaque tokens without sending JWTs, valid or forged, to the issuer.
- A 401 makes MCP clients restart sign-in, which is the wrong reaction to an unreachable identity provider.

### What worked

- `GOWORK=off go test ./pkg/embeddable/ -run 'Introspection|AutoValidation|LooksLikeJWS|MCPStartIntrospection' -v` → all eight tests pass.
- `GOWORK=off go test -race ./pkg/embeddable/ ./pkg/embeddable/embeddeddev/` → `ok`, no races.
- `GOWORK=off go test ./...` → 7 packages `ok`, no failures; `golangci-lint run` → `0 issues.`; `make logcopter-check` clean.

### What didn't work

- First test draft declared a package-level `type url struct`, which collides with the `net/url` import in the package's other files (an identifier may not be declared in both the file block and the package block). Renamed to `introspectionCall` before compiling.
- `golangci-lint run` → `pkg/embeddable/auth_provider_external.go:376:7: QF1001: could apply De Morgan's law (staticcheck)`. Replaced the negated condition with `isBase64URLRune`.
- `gosec` is not installed locally (`which gosec` empty); CI will run whatever security scan the repository configures.

### What I learned

- go-jose's `jwt.NumericDate` and `jwt.Audience` decode RFC 7662 `exp`/`nbf` and string-or-array `aud` directly, so the response struct needs no custom decoding.

### What was tricky to build

- Where scope enforcement happens: the middleware calls `validateBearerToken(..., false)` and lets go-sdk enforce `RequiredScopes` (so the 403 carries a scope challenge), while direct `ValidateBearerToken` enforces them itself. Introspection follows the same split.

### What warrants a second pair of eyes

- `exp` is required in introspection responses. RFC 7662 makes it optional; go-sdk's middleware rejects tokens without an expiration anyway, so requiring it gives a precise error rather than changing behaviour.
- No caching: one identity-provider round trip per MCP request (design §5.4).

### What should be done in the future

- Optional TTL cache; `private_key_jwt` client authentication; documentation (Step 4).

### Code review instructions

- `pkg/embeddable/auth_provider_external.go`: `configureTokenValidation`, `validateBearerToken`, `introspect`, `looksLikeJWS`.
- `GOWORK=off go test ./pkg/embeddable/ -run Introspection -v`.

### Technical details

- Request: `POST <introspection_endpoint>`, `Authorization: Basic base64(urlencode(id):urlencode(secret))`, body `token=…&token_type_hint=access_token`.

## Step 4: Documentation

This step brought the user-facing documentation in line with the code: the help pages for embedded/external OIDC and for hosted deployments, the README's authorization section, and the repository changelog.

### Prompt Context

**User prompt (verbatim):** see Step 1.

**Assistant interpretation:** Design §8 step 4.

**Inferred user intent:** Users of the library learn the new option and flags from the shipped help, not from this ticket.

### What I did

- `pkg/doc/topics/07-embedded-oidc.md`: the embedded server is linked through `embeddeddev`; `Provide()`/`WithOIDC(...)` replace the removed compatibility API; the programmatic example adds `embeddeddev.Provide()`; the four introspection flags; a new "Opaque access tokens and introspection" section (validation table, acceptance rules, 500-on-outage, no cache, CLI example); expiry note for embedded tokens; checklist updated.
- `pkg/doc/topics/08-oidc-keycloak-coolify-hosted-mcp.md`: new flags in the front matter and the core-flags list; when tokens are JWTs (Keycloak) versus opaque (ZITADEL default); pseudocode with the introspection branch.
- `README.md`: one paragraph in "HTTP authorization boundary".
- `changelog.md`: new entry at the top.

### Why

- The help pages are embedded in the binary (`go-go-mcp help …`), so they are what users read.

### What worked

- `GOWORK=off go run ./cmd/go-go-mcp help embedded-oidc` renders the new section ("### Opaque access tokens and introspection", line 349 of the output) and the `embeddeddev.Provide()` text; `help oidc-keycloak-coolify-hosted-mcp` lists `--oidc-token-validation`.

### What didn't work

- No failures.

### What I learned

- The hosted guide still describes the smailnail/Keycloak deployment; it now states that ZITADEL needs introspection, but a ZITADEL walkthrough would be a separate document.

### What was tricky to build

- N/A.

### What warrants a second pair of eyes

- The claim in the docs that ZITADEL issues opaque tokens by default to dynamically registered clients comes from zitadel#12658; verify against the ZITADEL version in use.

### What should be done in the future

- Open the pull request (Step 5).

### Code review instructions

- Read the rendered help: `GOWORK=off go run ./cmd/go-go-mcp help embedded-oidc`.

### Technical details

- N/A.

## Step 5: Pull request

The branch was pushed to the `wesen` fork, following the most recent merged pull request (#88, `wesen:task/add-plot-editor`), and opened against `go-go-golems/go-go-mcp`.

### Prompt Context

**User prompt (verbatim):** see Step 1.

**Assistant interpretation:** Open the pull request for the go-go-mcp changes.

**Inferred user intent:** Review and merge through the usual GitHub flow.

### What I did

- Checked remotes (`origin` = go-go-golems, `wesen` = fork) and that no git hooks are installed (`.git/hooks` has only samples, so lefthook's pre-push `make goreleaser` does not run).
- `git push -u wesen task/embeddable-auth-introspection`.
- `gh pr create --repo go-go-golems/go-go-mcp --base main --head wesen:task/embeddable-auth-introspection …` → https://github.com/go-go-golems/go-go-mcp/pull/89. The body lists the measurement, the bug, the introspection rules, breaking changes, validation and follow-ups.

### Why

- Fork-based pull requests match the latest merged PR in this repository.

### What worked

- Push and PR creation succeeded on the first attempt.

### What didn't work

- N/A.

### What I learned

- N/A.

### What was tricky to build

- N/A.

### What warrants a second pair of eyes

- The breaking-change list in the PR body.

### What should be done in the future

- Watch CI; write the vault report.

### Code review instructions

- Review commits in order: `db0cc5d` (ticket), `83d18b8` (embeddeddev), `84388d1` (introspection), `34c6555` (docs).

### Technical details

- Branch: `task/embeddable-auth-introspection`; worktree `/home/manuel/workspaces/2026-10-15/embeddable-auth-introspection/go-go-mcp`.

## Step 6: Vault report

A project report for the Obsidian vault (go-go-parc) explains the three changes as a technical deep dive: concepts, the measurements and the bug, the request path before and after, the seam, the expiry fix, introspection, verification, failures and open work.

### Prompt Context

**User prompt (verbatim):** see Step 1 ("Write a report for the vault when done with the go-go-mcp changes.").

**Assistant interpretation:** Write a durable project note in go-go-parc and commit and push it.

**Inferred user intent:** Keep the reasoning and results findable outside this repository.

### What I did

- Wrote `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/10/15/PROJECT REPORT - go-go-mcp Embeddable Auth - Linking the Embedded OIDC Server on Request and Verifying Opaque Tokens - A Technical Deep Dive.md` (front matter with repo, worktree, ticket, branch, PR; two Mermaid diagrams; links to the Francois reports).
- Staged only that file; vault commit `5529b84`, pushed.

### Why

- The vault's convention for finished engineering work is a `PROJECT REPORT - …` note in the day's folder.

### What worked

- Commit and push succeeded; the vault working tree was otherwise clean.

### What didn't work

- N/A.

### What I learned

- N/A.

### What was tricky to build

- N/A.

### What warrants a second pair of eyes

- N/A.

### What should be done in the future

- Update the report's status line when PR #89 merges.

### Code review instructions

- N/A.

### Technical details

- Vault commit: `5529b84`.

## Step 7: Review fix: accept only bearer access tokens from introspection

Codex reviewed PR #89 and raised one P1 finding: the introspection check rejected refresh tokens only when the issuer wrote the literal `token_type: refresh_token`. RFC 7662 makes `token_type` optional, and `token_type_hint=access_token` is only a lookup hint, so an issuer could report an active refresh token without a type, and if it carried the audience and scopes it would be accepted as an access token. The finding is correct. The check is now an allowlist: the response must positively identify a bearer access token.

### Prompt Context

**User prompt (verbatim):** "Address code review comment: https://github.com/go-go-golems/go-go-mcp/pull/89"

**Assistant interpretation:** Fix the review finding on PR #89, with tests, and answer the comment.

**Inferred user intent:** Merge-ready PR without a token-confusion hole.

### What I did

- Read the review: one inline comment (id 4289358470) on `pkg/embeddable/auth_provider_external.go:321` from `chatgpt-codex-connector[bot]`, "Reject refresh tokens without relying on token_type".
- Checked what the main target returns: `zitadel/oidc/v3@v3.47.8` leaves the response fields to the server's storage (`pkg/op/token_intospection.go`), and ZITADEL's documentation states `token_type` is always `Bearer` in introspection responses. An allowlist therefore does not break ZITADEL.
- `auth_provider_external.go`: replaced the `refresh_token` comparison with `checkIntrospectedTokenType`:
  - `token_type` must equal `Bearer` case-insensitively (RFC 6749 §5.1);
  - a missing `token_type` is rejected unless `IntrospectionAllowMissingTokenType` is set;
  - a Keycloak-style `typ` other than `Bearer` is rejected even with that opt-out;
  - DPoP is rejected because proofs are not verified.
- `server.go`: `ExternalOIDCOptions.IntrospectionAllowMissingTokenType`. `command.go`: `--oidc-introspection-allow-missing-token-type`.
- Tests:
  - new rejection cases: missing `token_type` (the review scenario), `Refresh_Token`, `DPoP`, `N_A`, Keycloak `typ` `Refresh` and `Offline`;
  - lowercase `bearer` is accepted;
  - a new test shows the opt-out accepts a missing type but still rejects `typ: Refresh` and an explicit `refresh_token`;
  - the flag test covers the new flag.
- Docs: design §5.4 rule 6, help page 07 (rule, rationale and flag), help page 08 front matter, changelog.

### Why

- A denylist of one optional literal cannot fail closed. The only signals RFC 7662 responses offer are `token_type` and, for Keycloak, `typ`; requiring a positive bearer type is the reliable discriminator, and an explicit opt-out keeps issuers that omit the field usable, with the risk named in the flag's help.

### What worked

- With the old check restored temporarily, all six new cases fail, each with `ValidateBearerToken() error = <nil>, want errUnauthorizedToken` (a first run piped through `head -10` hid the sixth; a full run listed all of them):

  ```text
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/DPoP_token
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/keycloak_refresh_typ
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/keycloak_offline_typ
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/missing_token_type
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/N_A_token_type
  --- FAIL: TestIntrospectionRejectsTokensItMustNotAccept/refresh_token_mixed_case
  ```

  With the fix they pass.
- `GOWORK=off go test ./...` → 7 packages `ok`; `go test -race ./pkg/embeddable/` → `ok`; `golangci-lint run` → `0 issues.`; `make logcopter-check` clean.

### What didn't work

- `gofmt -l` flagged the edited test file (map literal alignment after adding longer keys); `gofmt -w` fixed it.

### What I learned

- ZITADEL's introspection `active` is also false when the calling client is not in the token's audience, which binds tokens to the resource server independently of the `aud` check here.

### What was tricky to build

- Keeping the opt-out narrow: it relaxes only the missing-field case; explicit non-bearer types and Keycloak refresh `typ` values stay rejected.

### What warrants a second pair of eyes

- Whether Keycloak's introspection responses include `token_type`; if not, Keycloak deployments using introspection need the opt-out (its `typ` check still protects them).

### What should be done in the future

- Verify against a live ZITADEL and Keycloak introspection response.

### Code review instructions

- `checkIntrospectedTokenType` in `pkg/embeddable/auth_provider_external.go`; `GOWORK=off go test ./pkg/embeddable/ -run Introspection -v`.

### Technical details

- Review comment: https://github.com/go-go-golems/go-go-mcp/pull/89#discussion_r4289358470
