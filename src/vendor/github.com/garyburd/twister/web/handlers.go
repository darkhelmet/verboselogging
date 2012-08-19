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

import ()

type redirectHandler struct {
    url       string
    permanent bool
}

func (rh *redirectHandler) ServeWeb(req *Request) {
    req.Redirect(rh.url, rh.permanent)
}

// RedirectHandler returns a request handler that redirects to the given URL. 
func RedirectHandler(url string, permanent bool) Handler {
    return &redirectHandler{url, permanent}
}

var notFoundHandler = HandlerFunc(func(req *Request) { req.Error(StatusNotFound, nil) })

// NotFoundHandler returns a request handler that responds with 404 not found.
func NotFoundHandler() Handler {
    return notFoundHandler
}
