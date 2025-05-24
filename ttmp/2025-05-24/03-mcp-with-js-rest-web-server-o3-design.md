Below is a slice-by-slice design for a **Go-native JavaScript playground / web-server** that satisfies every bullet you gave—nothing more, nothing less.

---

## 1. Overall topology

```
┌──────────────────────────────┐
│    net/http front-door       │
│  (chi / stdlib ServeMux)     │
└───────────────┬──────────────┘
                │ route hit
┌───────────────▼──────────────┐
│  dispatcher goroutine        │  –––-- single owner of the JS runtime
│  (requests → chan EvalJob)   │
└───────────────┬──────────────┘
                │ <-chan
┌───────────────▼──────────────┐
│  jsEngine                     │  –––-- goja.Runtime + bindings
│  • global state lives here    │
│  • Open sqlite.Conn here      │
│  • registerHandler() etc.     │
└───────────────┬──────────────┘
                │ invokes
┌───────────────▼──────────────┐
│  Go callback wrappers         │  –––-- call JS     │
│   (return JSON / []byte)      │  –––-- sanitize    │
└──────────────────────────────┘
```

*Only one goroutine touches `goja.Runtime`; everything else talks to it through `chan EvalJob`.*

---

## 2. Packages & files

| Path                              | Responsibility                                                                                          |
| --------------------------------- | ------------------------------------------------------------------------------------------------------- |
| **cmd/playground/main.go**        | CLI flags, http server bootstrap                                                                        |
| **internal/engine/engine.go**     | `jsEngine` (goja runtime + sqlite handle)                                                               |
| **internal/engine/bindings.go**   | Expose Go functions to JS (`db.query`, `registerHandler`, `registerFile`)                               |
| **internal/engine/dispatcher.go** | `type EvalJob struct { code string; w http.ResponseWriter; r *http.Request }`                           |
| **internal/api/execute.go**       | REST endpoint `/v1/execute` that accepts JS, writes it to `scripts/ts-N.js`, then sends `EvalJob{code}` |
| **internal/web/router.go**        | Chi router + a map `handlers[path][method] -> handlerID` provided by JS                                 |
| **scripts/\***                    | Timestamped JS source that was fed through `execute_js`                                                 |

---

## 3. jsEngine bootstrap

```go
func NewEngine(dbPath string) *Engine {
    rt := goja.New()
    db, _ := sql.Open("sqlite3", dbPath)

    e := &Engine{rt: rt, db: db}

    // 1) Inject sqlite wrapper
    rt.Set("db", map[string]func(string, ...any) []map[string]any{
        "query": e.jsQuery,
    })

    // 2) Registration API
    rt.Set("registerHandler", e.registerHandler)
    rt.Set("registerFile",    e.registerFile)

    // 3) Minimal promise polyfill (optional if you want async/await)
    //    … omitted for brevity

    return e
}
```

### Binding shapes

```go
// JS side
registerHandler("GET", "/counter", () => ({count: ++globalCounter}));

registerFile("/logo.png", () => new Uint8Array([...]));

// Or a function that streams a file from disk:
registerFile("/download/db", () => {
    return readFileSync("data.sqlite");   // Node-style helper we expose
});
```

`registerHandler` pushes a closure into a Go map keyed by `method+path`.
`registerFile` same, but Go wrapper writes headers (`Content-Type`, `Content-Length`) and copies bytes.

---

## 4. Dispatcher & request flow

```go
type EvalJob struct {
    fn       goja.Callable   // pre-registered closure
    w        http.ResponseWriter
    r        *http.Request
}

var jobs = make(chan EvalJob, 1024) // unbounded enough for demo

// single goroutine:
for job := range jobs {
    res, err := job.fn(nil /*this*/, vm.ToValue(job.r))
    if err == nil {
        switch v := res.Export().(type) {
        case []byte:
            job.w.Write(v)
        default:
            json.NewEncoder(job.w).Encode(v)
        }
    } else {
        job.w.WriteHeader(500)
        job.w.Write([]byte(err.Error()))
    }
}
```

HTTP handler wrapper simply looks up `fn` in the registry and `jobs <- EvalJob{...}`.

---

## 5. `execute_js(code string)` external API

```go
func executeHandler(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    // 1. Persist source
    ts := time.Now().UTC().Format("20060102-150405")
    os.WriteFile("scripts/"+ts+".js", body, 0600)

    // 2. Evaluate (non-blocking to caller)
    jobs <- EvalJob{fn: nil, code: string(body)}

    w.WriteHeader(202)
}
```

*If `fn` is nil, the dispatcher just runs `rt.RunString(code)`.*

This lets you live-patch the VM while it is serving traffic; new routes appear instantly because JS executed `registerHandler`.

---

## 6. SQLite exposure details

```go
func (e *Engine) jsQuery(query string, args ...any) []map[string]any {
    rows, _ := e.db.Query(query, args...)
    cols, _ := rows.Columns()

    var out []map[string]any
    for rows.Next() {
        vals := make([]any, len(cols))
        scan := make([]any, len(cols))
        for i := range vals { scan[i] = &vals[i] }
        rows.Scan(scan...)

        rec := make(map[string]any)
        for i, c := range cols { rec[c] = vals[i] }
        out = append(out, rec)
    }
    return out
}
```

The slice-of-map translates cleanly into JSON when returned through `registerHandler` closures.

---

## 7. Concurrency & state model

* **Single runtime** → global JS variables work exactly like Node’s top-level.
* Because goja is not thread-safe, *all* evaluation goes through the dispatcher channel.
* Each HTTP request handler therefore incurs a hop through that channel; with fast JS this is still OK for moderate RPS. If you need more throughput, spin up *N* dispatcher/VM pairs bound to a consistent-hash partition of the route space.

---

## 8. Minimal bootstrap script (optional)

Create `bootstrap.js` shipped with the binary so empty DB + runtime can respond immediately:

```js
let globalCounter = 0;

registerHandler("GET", "/",      () => "JS playground online");
registerHandler("GET", "/health",() => ({ok: true}));
```

`engine.Init("bootstrap.js")` runs this before the HTTP server starts.

---

## 9. CLI usage

```bash
# start server on :8080, db in ./data.sqlite
playground -port 8080 -db data.sqlite

# hot-load new code
curl -XPOST localhost:8080/v1/execute --data-binary @newcode.js
```

---

## 10. Footprint

* go1.22+, goja v0.0.0-2025xxxx, mattn/go-sqlite3 (≅ 3 MB static build)
* Single binary, zero external deps beyond `sqlite3` C library (or use `modernc.org/sqlite` for pure Go).

---

That’s the complete skeleton—all moving parts, precise API shapes, and concurrency boundaries needed to implement your playground exactly as specified. Feel free to poke at any slice in more depth.
