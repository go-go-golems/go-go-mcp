# Implementing the Actual API Improvements

*Document path: `ttmp/2025-06-05/03-implementing-the-actual-api-improvements.md`*

---

## 1  Purpose & Scope
This document explains **how to transform `cmd/experiments/js-web-server/internal/engine/handlers.go` into a semi-drop-in replacement for Express.js**, fulfilling the requirements captured in `ttmp/2025-06-05/api-improvements.md`.

It is written for a developer who is new to the project yet familiar with Go and JavaScript. By the end, you should know **what to change, where to start, and how to verify success**.

---

## 2  What We Have So Far
- A Go-based JavaScript runtime (Goja) that already exposes an `app` object with basic routing (`get`, `post`, …) and minimal `app.use`.
- A thin Express-style `req`/`res` pair (`ExpressRequest`, `ExpressResponse`).
- Basic path matching & parameter extraction.

> **Key finding:** The skeleton is functional but lacks middleware chaining, centralised error handling, safe request defaults, and several quality-of-life helpers.

---

## 3  Target Feature Matrix (✅ = done, 🟡 = partial, ❌ = missing)
| Area | Requirement | File/Package | Status |
|------|-------------|--------------|--------|
| Request Safety | `req.query` etc. never `undefined` | `handlers.go:createExpressRequestObject` | ❌ |
| Response API | `status`, `send`, `json`, `redirect`, `set`, `cookie`, `end` | `ExpressResponse` | 🟡 |
| Middleware | `app.use(path?, fn)` with `next()` & error chaining | router layer | ❌ |
| Error Types | `ValidationError`, `DatabaseError`, global `handleError` | JS utilities | ❌ |
| Router | Path-to-regex (`/user/:id`, wildcards) | `pathMatches` | ❌ |
| Helpers | `Validate`, `Cache`, `Security`, `Test` stubs | runtime bindings | ❌ |

---

## 4  Work Packages & Detailed Steps

Each **Work Package (WP)** is a self-contained chunk of engineering work.  
The paragraphs explain _why_ the WP exists and _what end-state_ we expect.  
After each paragraph the individual tasks (A1, B2, …) are broken down with enough
context that an intern can start coding right away.

### WP-A  Request Object Safety
Express guarantees that properties such as `req.query`, `req.params`, `req.headers`
are _always_ defined (at worst an empty object). Our Go shim currently leaves many
of them `nil`, which translates to `undefined` in JavaScript and causes _Cannot read
property ‘foo’ of undefined_ errors. WP-A hardens the request object so even naïve
handler code is safe.

- **A1  Default-fill optional fields** – Inside
  `createExpressRequestObject` ensure every optional map/slice is initialised.
  ```go
  if req.Query == nil   { req.Query   = map[string]interface{}{} }
  if req.Params == nil  { req.Params  = map[string]string{}      }
  if req.Headers == nil { req.Headers = map[string]interface{}{} }
  if req.Cookies == nil { req.Cookies = map[string]string{}      }
  // Body can stay nil – handlers should test for it explicitly.
  ```
  _Test tip:_ add a Go test that constructs an HTTP request with **no** query or cookies and
  executes a tiny JS handler `registerHandler("GET","/", r=>r.query && true)` – it should not throw.

- **A2  Expose `safeReq` helper in JS** – Add a runtime binding so users can wrap
  external objects (e.g. when writing unit tests) and still enjoy the same safety:
  ```go
  e.rt.Set("safeReq", func(call goja.FunctionCall) goja.Value {
      r := call.Argument(0).ToObject(e.rt)
      // Merge with defaults similar to A1 …
      return r
  })
  ```
  Tell documentation writers to reference this helper in examples.

### WP-B  Middleware Stack & `next()`
Express is famous for its onion-layer middleware model – every request flows
through zero or more functions, each calling `next()` (optionally with an error)
to continue the chain.  Our implementation stores **one** function per route and
ignores `next()`.  WP-B builds a real stack so existing Express middlewares
(ported to pure JS) run unmodified.

- **B1  Change `HandlerInfo`** – Replace `Fn goja.Callable` with
  `Fns []goja.Callable`.
- **B2  Append, don't overwrite** – In `registerHandler`, if a route is registered
  multiple times, push the new callable to the slice (`append`). That lets
  `app.use()` inject global pre-filters _before_ route handlers.
- **B3  Execution trampoline** – At request time compute `stack := resolvedFns` and
  run a tiny recursive closure:
  ```go
  var exec func(int)
  exec = func(i int) {
      if i == len(stack) { return }
      // Convert Go closure to JS function passed as `next`
      next := e.rt.ToValue(func(args ...goja.Value) goja.Value {
          if len(args) > 0 && !goja.IsUndefined(args[0]) && !goja.IsNull(args[0]) {
              // Error path – save error and jump to error handlers (WP-C)
              savedErr = args[0]
              execErr(0)
              return nil
          }
          exec(i + 1)
          return nil
      })
      stack[i].Call(goja.Undefined(), reqVal, resVal, next)
  }
  exec(0)
  ```
- **B4  Error-handler detection** – A middleware that expects four parameters
  `(err, req, res, next)` is considered an _error handler_.  Collect those into a
  separate slice `errFns` and invoke via `execErr` when `savedErr != nil`.

### WP-C  Centralised Error Handling
When a handler throws, Express allows developers to capture the error, map it to
HTTP status codes, and keep the server healthy.  Today any uncaught JS error
kills the Go handler with a 500 but no structured JSON. WP-C introduces typed
errors and a single conversion point.

- **C1  JS error classes** – During `setupHTTPUtilities` inject:
  ```js
  class ValidationError extends Error { constructor(msg){super(msg);this.status=400;} }
  class DatabaseError   extends Error { constructor(msg){super(msg);this.status=500;} }
  globalThis.ValidationError = ValidationError;
  globalThis.DatabaseError   = DatabaseError;
  ```
- **C2  Wrap Callables with `runtime.Try`** – Goja offers `Try(fn, catchFn)`;
  catchFn receives a thrown value. Use it to intercept every handler/middleware.
- **C3  Status mapping** – If the thrown value has a `status` property (check via
  `.Has("status")`) coerce it to `int`; else default to 500.
- **C4  Error middleware** – Implemented in WP-B step B4; make sure
  `errFns` receive the saved error as first argument.

### WP-D  Extended Response Helpers
Beyond `res.json()` and `res.send()`, production apps rely on helpers like
`res.type()`, `res.append()`, and streaming with `res.write() / res.flush()`.  WP-D
fills these gaps so ported code compiles.

- **D1  Helper methods**
  - `Type(mime)` – sets `Content-Type` header and returns `*ExpressResponse` for chaining.
  - `Append(name, value)` – `Header().Add` rather than `Set`.
  - `Format(map[string]func())` – pick a response formatter based on `Accept`.
  - `Download(path, filename)` – send `Content-Disposition: attachment` and file bytes (stub OK).
- **D2  Streaming** – Embed `http.Flusher` detection:
  ```go
  func (r *ExpressResponse) Write(chunk []byte){ if f,ok:=r.writer.(http.Flusher); ok{ r.writer.Write(chunk); f.Flush() } }
  ```
  This unlocks server-sent events and large file responses.

### WP-E  Router Upgrade
Our current `pathMatches` compares segments one-by-one and fails for wildcards or
optional parts.  That is good enough for tutorials but breaks the moment someone
writes `app.get('/users/:userId/books/:bookId?', ...)`.  WP-E swaps in a mature
router or implements a mini path-to-regex utility.

- **E1  Choose a library or roll your own** – `github.com/dimfeld/httptreemux` is
  small (~300 LOC), supports parameters and wildcards, and is MIT-licensed.
- **E2  Populate `req.params`** – Extract captured values from the router's match
  result and store them as strings in the JS request object.
- **E3  Wildcard & optional segments** – Ensure `*` grabs the remainder of the
  path and `?` marks a segment optional. Add unit tests mirroring Express docs.

### WP-F  Utility Injection (Validate, Cache, …)
The long tail of enhancements (input validation, caching, security helpers,
mini-test framework) lives in pure JavaScript and can run inside Goja.  WP-F
packages those scripts and exposes them as globals so user handlers can
`const {Validate} = globalThis`.

- **F1  Move snippets into files** – Create new directory
  `cmd/experiments/js-web-server/internal/engine/bindings/` and add
  `validate.js`, `cache.js`, `security.js`, etc. Copy code blocks from
  `api-improvements.md` verbatim.
- **F2  Auto-load at Engine boot** – When a new `Engine` is created, iterate that
  directory, read each file, and `rt.RunString()` its contents so they are
  available before any user script executes.  Provide a debug log statement so
  we know which helper was loaded.

---

## 5  Implementation Checklist (🗹 = done)
- [ ] WP-A completed & unit-tested
- [ ] WP-B middleware chains work with `next()`
- [ ] WP-C error mapping returns correct status codes
- [ ] WP-D response helper unit-tests green
- [ ] WP-E router passes parameterised route tests
- [ ] WP-F utilities accessible from user scripts

Use `go test ./... && node tests/*.mjs` as your CI gate.

---

## 6  Milestones & Suggested Timeline
1. **Day 1-2**  → WP-A, WP-C skeleton
2. **Day 3-4**  → WP-B full middleware stack
3. **Day 5**    → WP-E router integration
4. **Day 6**    → WP-D niceties + docs update
5. **Day 7**    → WP-F utilities & sample apps

---

## 7  Key Resources
- `cmd/experiments/js-web-server/internal/engine/handlers.go`
- `ttmp/2025-06-05/api-improvements.md`
- [Express.js API reference](https://expressjs.com/en/4x/api.html)
- [Goja README](https://github.com/dop251/goja)

---

## 8  Next Steps for New Contributors
1. **Clone & build** the project (`make dev`).
2. **Run** the sample server in `examples/express-clone` and observe failing tests.
3. Pick the **top unchecked TODO** from §5 and create a PR.
4. Commit messages: `WP-A: ensure req.query defaults to {}` style.
5. Save any further technical notes in `ttmp/2025-06-05/0X-*.md` as per team conventions.

---

> Remember: _If you feel stuck in a deep debugging rabbit-hole, take a breath and **TOUCH GRASS**._ 