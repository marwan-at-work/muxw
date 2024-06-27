# Muxw

A simple wrapper around the new Go 1.22 muxer. It decorates the net/http muxer with the following:

1. Middleware functionality.
2. Trailing slash footgun protection.
3. Subrouting.
4. HTTP methods as Go methods such as `muxer.Post("/hello", helloHandler)` instead of `muxer.Handle("POST /hello", helloHandler)`.
