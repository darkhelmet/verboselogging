// Copyright 2010 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package web

import (
    "bytes"
    "fmt"
    "path"
    "regexp"
    "strings"
)

// Router is a request handler that dispatches HTTP requests to other handlers
// using the request URL path and the request method.
//
// A router has a list of routes. A route is a request path pattern and a
// collection of (method, handler) pairs.
//
// A path pattern is a string with embedded parameters. A parameter has the syntax:
//
//  '<' name (':' regular-expression)? '>'
//
// If the regular expression is not specified, then the regular expression
// [^/]+ is used.
//
// The pattern must begin with the character '/'.
//
// A router dispatches requests by matching the request URL path against the
// registered route patterns in the order that the routes were registered. If a
// matching route is not found, then the router responds to the request with
// HTTP status 404.
//
// If a matching route is found, then the router looks for a handler using the
// request method, "GET" if the request method is "HEAD" and "*". If a handler
// is not found, then the router responds to the request with HTTP status 405.
//
// Any matching parameters are in route pattern are stored in the in the
// request URLParam field.
//
// If a pattern ends with '/', then the router redirects the URL without the
// trailing slash to the URL with the trailing slash.
//
type Router struct {
    routes []*route
}

type route struct {
    addSlash bool
    regexp   *regexp.Regexp
    names    []string
    handlers map[string]Handler
}

var parameterRegexp = regexp.MustCompile("<([A-Za-z0-9_]*)(:[^>]*)?>")

// compilePattern compiles the pattern to a regular expression and array of
// parameter names.
func compilePattern(pattern string, addSlash bool, sep string) (*regexp.Regexp,
    []string) {
    var buf bytes.Buffer
    names := make([]string, 8)
    i := 0
    buf.WriteString("^")
    for {
        a := parameterRegexp.FindStringSubmatchIndex(pattern)
        if len(a) == 0 {
            buf.WriteString(regexp.QuoteMeta(pattern))
            break
        } else {
            buf.WriteString(regexp.QuoteMeta(pattern[0:a[0]]))
            name := pattern[a[2]:a[3]]
            if name != "" {
                names[i] = pattern[a[2]:a[3]]
                i += 1
                buf.WriteString("(")
            }
            if a[4] >= 0 {
                buf.WriteString(pattern[a[4]+1 : a[5]])
            } else {
                buf.WriteString("[^" + sep + "]+")
            }
            if name != "" {
                buf.WriteString(")")
            }
            pattern = pattern[a[1]:]
        }
    }
    if addSlash {
        buf.WriteString("?")
    }
    buf.WriteString("$")
    return regexp.MustCompile(buf.String()), names[0:i]
}

// Register the route with the given pattern and handlers. The structure of the
// handlers argument is:
//
//  (method handler)+
//
// where method is a string and handler is a Handler or a
// func(*Request). Use "*" to match all methods.
func (router *Router) Register(pattern string, handlers ...interface{}) *Router {
    if pattern == "" || pattern[0] != '/' {
        panic("twister: Invalid route pattern " + pattern)
    }
    if len(handlers)%2 != 0 || len(handlers) == 0 {
        panic("twister: Invalid handlers for pattern " + pattern +
            ". Structure of handlers is [method handler]+.")
    }
    r := route{}
    r.addSlash = pattern[len(pattern)-1] == '/'
    r.regexp, r.names = compilePattern(pattern, r.addSlash, "/")
    r.handlers = make(map[string]Handler)
    for i := 0; i < len(handlers); i += 2 {
        method, ok := handlers[i].(string)
        if !ok {
            panic("twister: Bad method for pattern " + pattern)
        }
        switch handler := handlers[i+1].(type) {
        case Handler:
            r.handlers[method] = handler
        case func(*Request):
            r.handlers[method] = HandlerFunc(handler)
        default:
            panic(fmt.Sprintf("twister: Bad handler type %T for pattern %s and method %s", handler, pattern, method))
        }
    }
    router.routes = append(router.routes, &r)
    return router
}

type routerError int

func (status routerError) ServeWeb(req *Request) {
    req.Error(int(status), nil)
}

// addSlash redirects to the request URL with a trailing slash.
func addSlash(req *Request) {
    path := req.URL.Path + "/"
    if len(req.URL.RawQuery) > 0 {
        path = path + "?" + req.URL.RawQuery
    }
    req.Redirect(path, true)
}

// find the handler and path parameters given the path component of the request
// URL and the request method.
func (router *Router) find(path string, method string) (Handler, []string, []string) {
    for _, r := range router.routes {
        values := r.regexp.FindStringSubmatch(path)
        if len(values) == 0 {
            continue
        }
        if r.addSlash && path[len(path)-1] != '/' {
            return HandlerFunc(addSlash), nil, nil
        }
        values = values[1:]
        if handler := r.handlers[method]; handler != nil {
            return handler, r.names, values
        }
        if method == "HEAD" {
            if handler := r.handlers["GET"]; handler != nil {
                return handler, r.names, values
            }
        }
        if handler := r.handlers["*"]; handler != nil {
            return handler, r.names, values
        }
        return routerError(StatusMethodNotAllowed), nil, nil
    }
    return routerError(StatusNotFound), nil, nil
}

func cleanUrlPath(p string) string {
    if p == "" || p == "/" {
        return "/"
    }
    slash := p[len(p)-1] == '/'
    p = path.Clean(p)
    if slash {
        p += "/"
    }
    return p
}

// ServeWeb dispatches the request to a registered handler.
func (router *Router) ServeWeb(req *Request) {
    p := cleanUrlPath(req.URL.Path)
    if p != req.URL.Path {
        req.Redirect(p, true)
        return
    }
    handler, names, values := router.find(p, req.Method)
    if req.URLParam == nil {
        req.URLParam = make(map[string]string, len(values))
    }
    for i := 0; i < len(names); i++ {
        req.URLParam[names[i]] = values[i]
    }
    handler.ServeWeb(req)
}

// NewRouter allocates and initializes a new Router.
func NewRouter() *Router {
    return &Router{}
}

// HostRouter is a request handler that dispatches HTTP requests to other handlers
// using the host HTTP header.
//
// A host router has a list of routes where each route is a (pattern, handler)
// pair.  The router dispatches requests by matching the host header against
// the patterns in the order that the routes were registered. If a matching
// route is found, the request is dispatched to the route's handler.
//
// A pattern is a string with embedded parameters. A parameter has the syntax:
//
//  '<' name (':' regexp)? '>'
//
// If the regular expression is not specified, then the regular expression
// [^.]+ is used.
//
// Any matching parameters are in route pattern are stored in the in the
// request URLParam field.
type HostRouter struct {
    defaultHandler Handler
    routes         []hostRoute
}

type hostRoute struct {
    regexp  *regexp.Regexp
    names   []string
    handler Handler
}

// NewHostRouter allocates and initializes a new HostRouter.
func NewHostRouter(defaultHandler Handler) *HostRouter {
    if defaultHandler == nil {
        defaultHandler = NotFoundHandler()
    }
    return &HostRouter{defaultHandler: defaultHandler}
}

// Register a handler for the given pattern.
func (router *HostRouter) Register(hostPattern string, handler Handler) *HostRouter {
    regex, names := compilePattern(hostPattern, false, ".")
    router.routes = append(router.routes, hostRoute{regexp: regex, names: names, handler: handler})
    return router
}

func (router *HostRouter) find(host string) (Handler, []string, []string) {
    for _, r := range router.routes {
        values := r.regexp.FindStringSubmatch(host)
        if len(values) == 0 {
            continue
        }
        values = values[1:]
        return r.handler, r.names, values
    }
    return router.defaultHandler, nil, nil
}

// ServeWeb dispatches the request to a registered handler.
func (router *HostRouter) ServeWeb(req *Request) {
    host := strings.ToLower(req.URL.Host)
    handler, names, values := router.find(host)
    if req.URLParam == nil {
        req.URLParam = make(map[string]string, len(values))
    }
    for i := 0; i < len(names); i++ {
        req.URLParam[names[i]] = values[i]
    }
    handler.ServeWeb(req)
}
