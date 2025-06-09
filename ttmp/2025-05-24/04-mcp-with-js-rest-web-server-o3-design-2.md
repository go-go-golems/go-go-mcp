### 1  Overview

```
┌────────────────────────────┐
│          Browser           │  (POST /api/exec   HTTP requests)
└─────────────┬──────────────┘
              │
┌─────────────▼──────────────┐
│   Go HTTP Front Controller │  (chi / std `http`)
└─────────────┬──────────────┘
              │
  ┌───────────▼─────────┐   async calls (channel)
  │   JS Engine Loop    │───┐
  │  (goja Runtime)     │   │
  └───────────┬─────────┘   │
              │             │
  ┌───────────▼─────────┐   │
  │   SQLite Pool       │   │
  └─────────────────────┘   │
                            ▼
                   JS handlers & global state
```

* **Single goja runtime** keeps all JS objects, functions, and globals alive between requests.
* Every HTTP request (or `execute_js`) is marshalled to that runtime through a **channel-based event loop**—no locking inside JS.
* SQLite is accessed from Go and surfaced as a small JS module; the same pool is reused for Go and JS.

---

### 2  Directory layout

```
/cmd/server          main.go
/internal/engine     engine.go    (JS VM + loop)
/internal/bridge     sqlite.go   (JS <-> SQLite)
/internal/router     mux.go      (method/path → JS fn)
/scripts/            persisted JS snippets
/static/             optional static files
```

---

### 3  JavaScript-side standard library

```js
// sqlite
const db = sqlite.open("data.db");
db.exec("CREATE TABLE IF NOT EXISTS notes(id INTEGER, body TEXT)");
const rows = db.query("SELECT * FROM notes WHERE id = ?", 1);

// HTTP
registerRoute("GET", "/hello", (req) => {
  return { status: 200, body: `hi ${req.query.name ?? "world"}` };
});

registerFile("/logo.png", "/static/logo.png");

// Global state survives until server restart
global.counter = (global.counter ?? 0) + 1;
```

#### Symbols exposed to JS

| Name                              | Signature                                         | Purpose                          |
| --------------------------------- | ------------------------------------------------- | -------------------------------- |
| `registerRoute(method, path, fn)` | `fn(req) → {status, headers?, body}`              | Registers REST handler.          |
| `registerFile(path, src)`         | `src` = string path **or** `(req) → {mime, body}` | Serves static or generated file. |
| `sqlite.open(file)`               | returns *DB*                                      | Opens/creates SQLite file.       |
| *DB*`.query(sql, ...params)`      | → JS array of row objects                         | SELECT wrapper.                  |
| *DB*`.exec(sql, ...params)`       | → `{rowsAffected}`                                | INSERT/UPDATE/DDL wrapper.       |

---

### 4  Engine internals (`internal/engine/engine.go`)

```go
type Engine struct {
    vm          *goja.Runtime
    calls       chan func() (any, error)   // Event-loop work queue
    routes      map[string]map[string]goja.Value // method → path → fn
    files       map[string]fileEntry
    scriptDir   string
}

func New(db *sql.DB, scriptDir string) *Engine {
    e := &Engine{
        vm:      goja.New(),
        calls:   make(chan func() (any, error), 128),
        routes:  map[string]map[string]goja.Value{},
        files:   map[string]fileEntry{},
        scriptDir: scriptDir,
    }
    e.exposeStdlib(db)
    go e.loop()
    return e
}

// Public entry points -------------------------------------------------

func (e *Engine) ExecuteJS(code string) (goja.Value, error) {
    tsName := time.Now().Format("20060102T150405.000") + ".js"
    os.WriteFile(filepath.Join(e.scriptDir, tsName), []byte(code), 0644)
    return e.call(func() (any, error) {
        return e.vm.RunString(code)
    })
}

func (e *Engine) CallRoute(method, path string, req *http.Request, body []byte) (resp *Result, err error) {
    return e.call(func() (any, error) {
        fn := e.routes[method][path]
        if fn == nil { return nil, errNotFound }
        jsReq := e.goToJSRequest(req, body)
        val, err := fn(goja.Undefined(), jsReq)
        if err != nil { return nil, err }
        return e.jsToGoResponse(val), nil
    })
}

// Private helpers -----------------------------------------------------

func (e *Engine) call(f func() (any, error)) (res any, err error) {
    done := make(chan struct{})
    e.calls <- func() (any, error) { defer close(done); return f() }
    <-done
    return res, err
}

func (e *Engine) loop() {
    for work := range e.calls {
        work()
    }
}
```

*The loop ensures a single-threaded JS world; Go code stays concurrent.*

---

### 5  HTTP front controller (`internal/router/mux.go`)

```go
r := chi.NewRouter()

// Dynamic JS routes
r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
    if resp, err := eng.CallRoute(r.Method, r.URL.Path, r, readBody(r)); err == nil {
        writeHTTP(w, resp)
        return
    }
    http.NotFound(w, r)
})

// File routes
r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
    if f := eng.FileFor(r.URL.Path); f != nil {
        serveFileOrJS(w, r, f)
        return
    }
    http.NotFound(w, r)
})

// Exec API
r.Post("/api/exec", func(w http.ResponseWriter, r *http.Request) {
    src, _ := io.ReadAll(r.Body)
    val, err := eng.ExecuteJS(string(src))
    if err != nil { http.Error(w, err.Error(), 400); return }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(val.Export())
})
```

---

### 6  `sqlite` JS module bridge (`internal/bridge/sqlite.go`)

```go
func (e *Engine) exposeStdlib(db *sql.DB) {
    sqliteObj := e.vm.NewObject()
    sqliteObj.Set("open", func(call goja.FunctionCall) goja.Value {
        path := call.Argument(0).String()
        conn, _ := sql.Open("sqlite3", path)
        dbObj := e.vm.NewObject()
        dbObj.Set("query", func(fc goja.FunctionCall) goja.Value {
            rows, _ := conn.Query(fc.Argument(0).String(), args(fc.Arguments[1:])...)
            return rowsToJS(e.vm, rows)
        })
        dbObj.Set("exec", func(fc goja.FunctionCall) goja.Value {
            res, _ := conn.Exec(fc.Argument(0).String(), args(fc.Arguments[1:])...)
            affected, _ := res.RowsAffected()
            o := e.vm.NewObject(); _ = o.Set("rowsAffected", affected); return o
        })
        return dbObj
    })
    _ = e.vm.Set("sqlite", sqliteObj)
}
```

---

### 7  HTTP request/response marshaling

* **Go → JS**: `req` object exposes `method`, `path`, `headers` (plain object), `query` (object), `body` (`Uint8Array`), `json()` helper.
* **JS → Go**: handler returns plain object:

```ts
{
  status:   number,            // mandatory
  headers?: { [k: string]: string },
  body?:    string | Uint8Array
}
```

A missing body maps to 204 No Content.

---

### 8  Sample end-to-end flow

1. `POST /api/exec` with script above → stored under `/scripts/20250524T100212.123.js`, evaluated once, functions register themselves.
2. User hits `GET /hello?name=Manuel`

   * Go router → `Engine.CallRoute("GET","/hello", …)`
   * JS handler runs inside runtime, returns `{status:200, body:"hi Manuel"}`
   * Go translates and writes to client.

---

### 9  Deployment checklist (build only what is required)

* **Dependencies**: `github.com/dop251/goja`, `github.com/go-chi/chi/v5`, `github.com/mattn/go-sqlite3`.
* `go run ./cmd/server` starts everything; scripts can be added live via `/api/exec`.

This outlines all modules, interfaces, and data flow needed to implement the described Go-hosted JavaScript playground with SQLite access, REST/file handler registration, and an external `execute_js` endpoint.
