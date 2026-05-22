# go-http

A from-scratch HTTP server in Go, built to learn how HTTP works at the wire level. No `net/http`.

```sh
go run .
```

Serves `app/` on `:8090`.

## Initial server

Built a simple server using `net` package from go.

Listening on `:8090` with `net.Listen`, accepting in a loop, and handing each connection to its own goroutine so multiple clients can be served at once.

Reading raw bytes off the connection and parsing them into a struct with method, path, query params, and form body. Learned the shape of an HTTP request — request line, headers, blank `\r\n\r\n`, then body — and that query strings and form bodies share the same `key=value&...` encoding.

Mapping paths to files under `app/`, defaulting `/` to `index.html`, falling back to `.html` if the raw path doesn't exist, and blocking `..` to avoid directory traversal. Setting `Content-Type` based on extension so CSS actually gets applied.

Building the response string by hand with status line, headers, and body. `Content-Length` has to match the body exactly or the client hangs.

Handling form posts and responding with `303 See Other` so the browser issues a fresh GET — the Post/Redirect/Get pattern, which avoids re-submitting on refresh.
