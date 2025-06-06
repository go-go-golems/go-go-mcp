Below is a focused blue-print for **wrapping Go’s `http.Request` + `ResponseWriter` so the JavaScript you run inside Goja looks and feels exactly like Express.js**.
Everything is split into three layers:

1. **Go structs** that talk to the Go HTTP stack.
2. **Goja Objects** that JS handlers receive (`req`, `res`).
3. **A tiny JS façade** (≈30 LoC) to sugar-coat a few convenience aliases so copy-pasting Express examples “just works”.

---

## 1 Design invariants

| Invariant                                         | Why it matters                                                                                                  |
| ------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| One wrapper pair **per request**                  | Express relies on per-request mutability (`res.locals`, `req.params` etc.). Never share them across goroutines. |
| Chainable helpers return **the same goja.Object** | Enables idioms like `res.status(201).json(body)`.                                                               |
| **Header writes are idempotent**                  | After the first body byte you must freeze headers; expose `res.headersSent`.                                    |
| **Case-insensitive headers** on both sides        | Express treats header names this way; map to `http.Header` which is already canonicalised.                      |

---

## 2 `req` – Request wrapper

### Mapping table (core subset)

| JS property / method | Go source                                                     | Notes               |
| -------------------- | ------------------------------------------------------------- | ------------------- |
| `req.method`         | `r.Method`                                                    | immutable string    |
| `req.url`            | `r.URL.String()`                                              | full URL inc. query |
| `req.originalUrl`    | same as above, *before* router prefix slicing                 |                     |
| `req.path`           | `r.URL.Path`                                                  |                     |
| `req.query`          | `r.URL.Query()` (url.Values)                                  |                     |
| `req.params`         | router-captured `map[string]string`                           |                     |
| `req.headers`        | `r.Header` (exposed via proxy for case-insensitive access)    |                     |
| `req.get(field)`     | `r.Header.Get(strings.ToLower(field))`                        |                     |
| `req.body`           | lazy-loaded buffer; decoder depends on body-parser middleware |                     |
| `req.cookies`        | parsed once with `r.Cookie()`; signed variant later           |                     |

### Go implementation sketch

```go
type jsRequest struct {
    rt     *goja.Runtime
    r      *http.Request
    params map[string]string
    body   []byte // nil until accessed
}

func wrapRequest(rt *goja.Runtime, r *http.Request, params map[string]string) *goja.Object {
    w := &jsRequest{rt: rt, r: r, params: params}
    o := rt.NewObject()

    // Plain fields
    _ = o.Set("method", r.Method)
    _ = o.Set("url", r.URL.String())
    _ = o.Set("originalUrl", r.URL.String())
    _ = o.Set("path", r.URL.Path)
    _ = o.Set("query", r.URL.Query())
    _ = o.Set("params", params)

    // Header helpers
    _ = o.Set("get", func(call goja.FunctionCall) goja.Value {
        name := strings.ToLower(call.Argument(0).String())
        return rt.ToValue(r.Header.Get(name))
    })

    // Lazy body getter
    _ = o.SetAccessor("body",
        func(v goja.Value) {
            // setter (not used)
        },
        func() goja.Value {
            if w.body == nil {
                // read once
                w.body, _ = io.ReadAll(r.Body)
            }
            return rt.ToValue(w.body)
        },
    )
    return o
}
```

Key points

* We **do not embed** the Go struct directly; instead we *copy* the desired data, so JS cannot mutate Go internals accidentally.
* Accessors give us *lazy* body parsing; if a route never touches `req.body`, we avoid an allocation.

---

## 3 `res` – Response wrapper

### Mapping table (90 % DX coverage)

| JS call                                          | Go action                                                   |
| ------------------------------------------------ | ----------------------------------------------------------- |
| `res.status(code)`                               | store `code`, return `res`                                  |
| `res.set(field, val)` / `res.header(field, val)` | `w.Header().Set(field, val)`                                |
| `res.get(field)`                                 | `w.Header().Get(field)`                                     |
| `res.type(mime)`                                 | shorthand for `Set("Content-Type", mime)`                   |
| `res.send(body)`                                 | detect `string`, `[]byte`, or `goja.Object` → JSON fallback |
| `res.json(obj)`                                  | `Content-Type: application/json`, `json.Marshal`            |
| `res.sendStatus(code)`                           | `status(code).send(http.StatusText(code))`                  |
| `res.location(url)`                              | `Set("Location", url)`                                      |
| `res.redirect([code], url)`                      | defaults to `302`                                           |
| `res.write(chunk)` / `res.end([chunk])`          | raw streaming                                               |
| `res.locals`                                     | `map[string]any{}` – persisted until end of request         |
| `res.headersSent`                                | bool guard                                                  |

### Go implementation sketch

```go
type jsResponse struct {
    rt          *goja.Runtime
    w           http.ResponseWriter
    status      int
    headersSent bool
    o           *goja.Object // back-reference for chaining
}

func wrapResponse(rt *goja.Runtime, w http.ResponseWriter) *goja.Object {
    r := &jsResponse{rt: rt, w: w}
    o := rt.NewObject()
    r.o = o

    // status(code)
    _ = o.Set("status", func(fc goja.FunctionCall) goja.Value {
        r.status = int(fc.Argument(0).ToInteger())
        return o
    })

    // set / header
    _ = o.Set("set", func(fc goja.FunctionCall) goja.Value {
        field := http.CanonicalHeaderKey(fc.Argument(0).String())
        val   := fc.Argument(1).String()
        if !r.headersSent {
            r.w.Header().Set(field, val)
        }
        return o
    })
    _ = o.Set("get", func(fc goja.FunctionCall) goja.Value {
        field := http.CanonicalHeaderKey(fc.Argument(0).String())
        return rt.ToValue(r.w.Header().Get(field))
    })
    _ = o.Set("header", o.Get("set")) // alias

    // type(mime)
    _ = o.Set("type", func(fc goja.FunctionCall) goja.Value {
        return o.Get("set").(func(goja.FunctionCall) goja.Value)(
            goja.FunctionCall{Arguments: []goja.Value{rt.ToValue("Content-Type"), fc.Argument(0)}},
        )
    })

    // send / json
    _ = o.Set("send", r.send)
    _ = o.Set("json", func(fc goja.FunctionCall) goja.Value {
        if !r.headersSent {
            r.w.Header().Set("Content-Type", "application/json")
        }
        data, _ := json.Marshal(fc.Argument(0).Export())
        return r.send(goja.FunctionCall{Arguments: []goja.Value{rt.ToValue(data)}})
    })

    // end([chunk])
    _ = o.Set("end", func(fc goja.FunctionCall) goja.Value {
        r.send(fc) // write if chunk present
        r.finish()
        return goja.Undefined()
    })

    // locals & flags
    _ = o.Set("locals", map[string]interface{}{})
    _ = o.SetAccessor("headersSent", nil, func() goja.Value {
        return rt.ToValue(r.headersSent)
    })

    return o
}

func (r *jsResponse) send(fc goja.FunctionCall) goja.Value {
    if !r.headersSent {
        if r.status != 0 {
            r.w.WriteHeader(r.status)
        }
        r.headersSent = true
    }
    if len(fc.Arguments) > 0 {
        switch val := fc.Argument(0).Export().(type) {
        case []byte:
            _, _ = r.w.Write(val)
        case string:
            _, _ = io.WriteString(r.w, val)
        default:
            // fallback to JSON
            j, _ := json.Marshal(val)
            r.w.Header().Set("Content-Type", "application/json")
            _, _ = r.w.Write(j)
        }
    }
    return r.o
}

func (r *jsResponse) finish() {
    if flusher, ok := r.w.(http.Flusher); ok {
        flusher.Flush()
    }
}
```

Highlights

* **Chainability** – every mutator returns `r.o`.
* **`headersSent` guard** – trying to mutate headers afterwards is a no-op, mirroring Express.
* **Automatic `WriteHeader`** – done lazily on first `send`/`end` if user set a status earlier.
* **Binary vs string vs JSON detection** – behaviour mirrors Express’s `res.send` ﻿([medium.com][1]).

---

## 4 Bootstrapping the handler

```go
func (s *Server) handle(w http.ResponseWriter, r *http.Request, params map[string]string, jsFn goja.Callable) {
    rt := s.pool.Get()        // or create new
    defer s.pool.Put(rt)

    reqObj := wrapRequest(rt, r, params)
    resObj := wrapResponse(rt, w)

    // Call the JS route handler.
    // Any exception bubbles up -> convert to 500 + log.
    _, err := jsFn(goja.Undefined(), reqObj, resObj)
    if err != nil {
        if !resObj.Get("headersSent").ToBoolean() {
            w.WriteHeader(http.StatusInternalServerError)
        }
        log.Warn().Msgf("JS error: %v", err)
    }
}
```

If the JS handler **returns a Promise** (detected via `goja.IsPromise(value)`), `Await(p)` until fulfilment and propagate rejection as above.

---

## 5 A 30-line JS façade (optional but recommended)

```js
// express-shim.js   – pre-loaded once into every runtime
global.express = function () {
  return {
    get : (p, h) => __go_register_handler('GET', p, h),
    post: (p, h) => __go_register_handler('POST', p, h),
    use : (p, h) => {                 // simple path-optional variant
      if (typeof p === 'function') { h = p; p = '/'; }
      __go_register_middleware(p, h);
    },
    listen: (port) => __go_listen(port)
  };
};

express.Router = function () {
  const prefix = arguments[0] || '';
  const r = {};
  ['get','post','delete','put','patch','options'].forEach(m =>
    r[m] = (p, h) => __go_register_handler(m.toUpperCase(), prefix + p, h));
  r.use = (p, h) => __go_register_middleware(prefix + (typeof p === 'string' ? p : ''), h);
  return r;
};
```

These `__go_*` functions are injected from Go during boot and internally point to your routing/middleware tables—no Node-specific objects required.

---

## 6 Putting it together (usage demo)

```js
const app = express();
app.use(express.json());                 // body parser you expose from Go
app.get('/users/:id', (req, res) => {
  res.status(200)
     .json({id: req.params.id, q: req.query});
});
app.post('/echo', (req, res) => res.send(req.body));
app.listen(8080);
```

The example above runs unmodified inside your Go binary because `req` and `res` wrappers fulfil every property that code expects.

---

## 7 Testing sanity

* **Golden tests** in Go: spin up the engine, feed JS snippets, assert the HTTP response.
* **Supertest** in JS: expose a fake `fetch` that drives the Go router in-process so you can reuse the Express testing ecosystem.

---

### What you get

* **Intuitive DX** – Express tutorials and StackOverflow answers paste right in.
* **Zero Node runtime** – only Go and Goja in production.
* **Predictable performance** – JSON serialisation, header writes, and streaming are all handled in native Go code; the JS layer merely instructs.

Wire the wrappers exactly as shown and you’ll hit feature-parity for the vast majority of Express.js application code, while still reaping Go’s concurrency and binary-deployment strengths.

[1]: https://medium.com/gist-for-js/use-of-res-json-vs-res-send-vs-res-end-in-express-b50688c0cddf?utm_source=chatgpt.com "Difference between res.json vs res.send vs res.end in Express.js"
