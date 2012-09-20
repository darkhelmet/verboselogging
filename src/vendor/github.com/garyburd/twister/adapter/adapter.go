// Copyright 2011 Gary Burd
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

// Package adapter provides types for adapting Twister request handlers to
// "het/http" request handlers.
//
// A simple example of running Twister handlers in a "net/http" server is:
//
//    package main
//
//    import (
//        "github.com/garyburd/twister/adapter"
//        "github.com/garyburd/twister/web"
//        "io"
//        "log"
//        "net/http"
//    )
//
//    func serveHello(req *web.Request) {
//        w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain; charset=\"utf-8\"")
//        io.WriteString(w, "Hello World!")
//    }
//
//    func main() {
//        http.Handle("/", adapter.HTTPHandlerFunc{serveHello})
//        err := http.ListenAndServe(":8080", nil)
//        if err != nil {
//            log.Fatal("ListenAndServe:", err)
//        }
//    }
package adapter

import (
    "bufio"
    "errors"
    "io"
    "net"
    "net/http"
    "vendor/github.com/garyburd/twister/web"
)

type responder struct{ w http.ResponseWriter }

func (r responder) Respond(status int, header web.Header) io.Writer {
    for k, v := range header {
        r.w.Header()[k] = v
    }
    r.w.WriteHeader(status)
    return r.w
}

func (r responder) Hijack() (conn net.Conn, br *bufio.Reader, err error) {
    return nil, nil, errors.New("not implemented")
}

func webRequestFromHTTPRequest(w http.ResponseWriter, r *http.Request) *web.Request {
    header := web.Header(r.Header)

    url := *r.URL
    if url.Host == "" {
        url.Host = r.Host
    }

    req, _ := web.NewRequest(
        r.RemoteAddr,
        r.Method,
        url.RequestURI(),
        web.ProtocolVersion(r.ProtoMajor, r.ProtoMinor),
        &url,
        header)

    req.Body = r.Body
    req.Responder = responder{w}
    req.ContentLength = int(r.ContentLength)
    if r.Form != nil {
        req.Param = web.Values(r.Form)
    }
    req.Env["twister.adapter.request"] = r
    return req
}

// HTTPRequest returns the original http.Request for this web.Request.
func HTTPRequest(req *web.Request) *http.Request {
    return req.Env["twister.adapter.request"].(*http.Request)
}

// HTTPHandler adapts a Twister request handler to a standard "net/http" handler.
type HTTPHandler struct{ Handler web.Handler }

func (h HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.Handler.ServeWeb(webRequestFromHTTPRequest(w, r))
}

// HTTPHandlerFunc adapts a Twister request handler function to a standard "net/http" handler.
type HTTPHandlerFunc struct{ Func func(*web.Request) }

func (h HTTPHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    h.Func(webRequestFromHTTPRequest(w, r))
}
