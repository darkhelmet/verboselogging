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
    "errors"
    "io"
)

type filterResponder struct {
    Responder
    filter func(status int, header Header) (int, Header)
}

func (rf *filterResponder) Respond(status int, header Header) io.Writer {
    return rf.Responder.Respond(rf.filter(status, header))
}

// FilterRespond replaces the request's responder with one that filters the
// arguments to Respond through the supplied filter. This function is intended
// to be used by middleware.
func FilterRespond(req *Request, filter func(status int, header Header) (int, Header)) {
    req.Responder = &filterResponder{req.Responder, filter}
}

// SetErrorHandler returns a handler that sets the request's error handler e.
func SetErrorHandler(e ErrorHandler, h Handler) Handler {
    return HandlerFunc(func(req *Request) {
        defer func() {
            if r := recover(); r != nil {
                var err error
                switch r := r.(type) {
                case string:
                    err = errors.New(r)
                case error:
                    err = r
                default:
                    err = errors.New("unknown")
                }
                e(req, StatusInternalServerError, err, NewHeader())
                panic(r)
            }
        }()
        req.ErrorHandler = e
        h.ServeWeb(req)
    })
}

// ProxyHeaderHandler returns a handler that overrides the Request.RemoteAddr field
// with the value of the header specified by addrName and the
// Request.URL.Scheme field with the value of the header specified by
// schemeName. No fix up is done for a field if the header name equals "" or the
// header is not present.
//
// The header names must be in canonical header name format.
// 
// Here's an example of how to use this handler with Nginx. In the nginx proxy
// configuration, specify a header for the IP address and scheme. The host
// header should also be passed through the proxy:
//
//    location / {
//        proxy_set_header X-Real-IP $remote_addr;
//        proxy_set_header X-Scheme $scheme;
//        proxy_set_header Host $http_host;
//        proxy_pass http://127.0.0.1:8080;
//    }       
//
// In the main function for the application, wrap the application handler with
// the proxy fix up:
//  
//  import (
//      "github.com/garyburd/twister/web"
//      "github.com/garyburd/twister/server"
//  )
//
//  func main() {
//      var h web.Handler
//      ... setup the application handler
//      h = web.ProxyHeaderHandler("X-Scheme", "X-Real-Ip", h)
//	    server.Run(":8080", h)
//  }
//
// The original values are added to the request Env with the keys
// "twister.web.OriginalRemoteAddr" and "twister.web.OriginalScheme".
func ProxyHeaderHandler(addrName, schemeName string, h Handler) Handler {
    return proxyHeaderHandler{
        addrName:   addrName,
        schemeName: schemeName,
        h:          h,
    }
}

type proxyHeaderHandler struct {
    addrName, schemeName string
    h                    Handler
}

func (h proxyHeaderHandler) ServeWeb(req *Request) {
    if s := req.Header.Get(h.addrName); s != "" {
        req.Env["twister.web.OriginalRemoteAddr"] = req.RemoteAddr
        req.RemoteAddr = s
    }
    if s := req.Header.Get(h.schemeName); s != "" {
        req.Env["twister.web.OriginalScheme"] = req.URL.Scheme
        req.URL.Scheme = s
    }
    h.h.ServeWeb(req)
}

// Name of XSRF cookie and request parameter.
const (
    XSRFCookieName = "xsrf"
    XSRFParamName  = "xsrf"
)

// FormHandler returns a handler that parses form encoded request bodies.
//
// If xsrfCheck is true, then cross-site request forgery protection is enabled
// using the cookie name XSRFCookieName and the parameter name
// XSRFParameterName. See CheckXSRF() for more information on cross-site
// request forgery protection.
func FormHandler(maxRequestBodyLen int, checkXSRF bool, h Handler) Handler {
    return formHandler{
        maxRequestBodyLen: maxRequestBodyLen,
        checkXSRF:         checkXSRF,
        h:                 h,
    }
}

type formHandler struct {
    maxRequestBodyLen int
    checkXSRF         bool
    h                 Handler
}

func (h formHandler) ServeWeb(req *Request) {
    if err := req.ParseForm(h.maxRequestBodyLen); err != nil {
        status := StatusBadRequest
        if err == ErrRequestEntityTooLarge {
            status = StatusRequestEntityTooLarge
            if e := req.Header.Get(HeaderExpect); e != "" {
                status = StatusExpectationFailed
            }
        }
        req.Error(status, errors.New("twister: Error reading or parsing form."))
        return
    }

    if h.checkXSRF {
        if err := CheckXSRF(req, XSRFCookieName, XSRFParamName); err != nil {
            req.Error(StatusNotFound, err)
            return
        }
    }

    h.h.ServeWeb(req)
}
