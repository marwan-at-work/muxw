package muxw

import (
	"net/http"
	"slices"
	"strings"
	"sync"
)

// Router wraps the net/http ServeMux with
// helpre methods.
type Router struct {
	m        *http.ServeMux
	mws      []Middleware
	once     sync.Once
	handler  http.Handler
	initted  bool
	notFound http.HandlerFunc
}

// NewRouter returns a ready ServeMux wrapper.
func NewRouter() *Router {
	return &Router{
		m: http.NewServeMux(),
	}
}

var _ http.Handler = &Router{}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.once.Do(r.init)
	// Potentially remove this if statement when this issue resolves:
	// https://github.com/golang/go/issues/65648
	if r.notFound != nil { // check notFound first so we don't call m.Handler twice.
		if _, p := r.m.Handler(req); p == "" {
			r.notFound(w, req)
			return
		}
	}
	r.handler.ServeHTTP(w, req)
}

func (r *Router) init() {
	if r.initted {
		panic("already initialized")
	}
	r.handler = chain(r.m, r.mws...)
	r.initted = true
}

func chain(h http.Handler, mws ...Middleware) http.Handler {
	slices.Reverse(mws)
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}

// Get register a handler func under the given path for
// the http method GET.
func (r *Router) Get(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodGet, path, hf)
}

// Head register a handler func under the given path for
// the http method HEAD.
func (r *Router) Head(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodHead, path, hf)
}

// Post register a handler func under the given path for
// the http method POST.
func (r *Router) Post(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodPost, path, hf)
}

// Put register a handler func under the given path for
// the http method PUT.
func (r *Router) Put(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodPut, path, hf)
}

// Patch register a handler func under the given path for
// the http method PATCH.
func (r *Router) Patch(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodPatch, path, hf)
}

// Delete register a handler func under the given path for
// the http method DELETE.
func (r *Router) Delete(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodDelete, path, hf)
}

// Connect register a handler func under the given path for
// the http method CONNECT.
func (r *Router) Connect(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodConnect, path, hf)
}

// Options register a handler func under the given path for
// the http method OPTIONS.
func (r *Router) Options(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodOptions, path, hf)
}

// Trace register a handler func under the given path for
// the http method TRACE.
func (r *Router) Trace(path string, hf http.HandlerFunc) {
	r.Handle(http.MethodTrace, path, hf)
}

// SetNotFoundHandler sets a handler that will be called if no routes were found
// by the serve mux.
func (r *Router) SetNotFoundHandler(hf http.HandlerFunc) {
	r.notFound = hf
}

// Middleware describes the signature of an http
// handler middleware.
type Middleware func(http.Handler) http.Handler

// Use applies the given middleware in the given order
func (r *Router) Use(mws ...Middleware) {
	if r.initted {
		panic("already initialized")
	}
	r.mws = append(r.mws, mws...)
}

// Mount mounts the given path prefix to the given http handler.
func (r *Router) Mount(pathPrefix string, h http.Handler) {
	r.m.Handle(strings.TrimSuffix(pathPrefix, "/")+"/", h)
}

// Handle registers the given handler to the given method and path.
func (r *Router) Handle(method, path string, handler http.Handler) {
	if r.initted {
		panic("already initialized")
	}

	// While net/http catches some of those cases, make it explicitly
	// that we only receive the expected patterns.
	if path == "" {
		panic("path cannot be empty")
	}
	if method == "" {
		panic("method cannot be empty")
	}

	// remainder paths are special, you can't put a /
	// in front of them nor can you put a /{$}. They basically
	// work like a path prefix or [*Router.Mount]
	if isRemainderPattern(path) {
		r.m.Handle(method+" "+path, handler)
		return
	}
	// Every given path might have the following suffixes:
	// /hello => no trailing slash
	// /hello/ => trailing slash
	// /hello/{$} => no trailing slash, but wildcard strict ending.
	//
	// Note that {$} must be preceeded with a slash and must come at the end.
	// So we will never get /hello{$} or /hello/{$}/.
	if path == "/" {
		r.m.Handle(method+" /{$}", handler)
	} else {
		path = strings.TrimSuffix(path, "/")
		path = strings.TrimSuffix(path, "/{$}")
		r.m.Handle(method+" "+path, handler)
		path += "/{$}"
		r.m.Handle(method+" "+path, handler)
	}
}

// Handler returns the handler to use for the given request, consulting
// r.Method, r.Host, and r.URL.Path. It always returns a non-nil handler. If the
// path is not in its canonical form, the handler will be an
// internally-generated handler that redirects to the canonical path. If the
// host contains a port, it is ignored when matching handlers.
//
// The path and host are used unchanged for CONNECT requests.
//
// Handler also returns the registered pattern that matches the request or, in
// the case of internally-generated redirects, the path that will match after
// following the redirect.
//
// If there is no registered handler that applies to the request, Handler
// returns a “page not found” handler and an empty pattern.
func (r *Router) Handler(req *http.Request) (h http.Handler, pattern string) {
	return r.m.Handler(req)
}

func isRemainderPattern(path string) bool {
	segments := strings.Split(path, "/")
	last := segments[len(segments)-1]
	// a remainder pattern is like {name...} which
	// means it should have at least six characters.
	if len(last) < 6 {
		return false
	}
	return last[0] == '{' && strings.HasSuffix(last, "...}")
}
