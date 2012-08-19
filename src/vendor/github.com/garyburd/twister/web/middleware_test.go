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

package web

import (
    "io"
    "strings"
    "testing"
)

const testToken = "12345678"

var xsrfTests = []struct {
    url    string // request URL
    method string // request method
    header Header // request headers
    body   string // request body
    status int    // expected response status
    cookie bool   // true if cookie expected in response
}{
    {url: "/",
        method: "GET",
        status: StatusOK,
        cookie: true},
    {url: "/?xsrf=" + testToken,
        method: "POST",
        header: NewHeader(HeaderCookie, "xsrf="+testToken),
        status: StatusOK},
    {url: "/",
        method: "POST",
        header: NewHeader(
            HeaderCookie, "xsrf="+testToken,
            HeaderContentType, "application/x-www-form-urlencoded"),
        body:   "xsrf=" + testToken,
        status: StatusOK},
    {url: "/",
        method: "POST",
        header: NewHeader(
            HeaderXXSRFToken, testToken,
            HeaderCookie, "xsrf="+testToken,
            HeaderContentType, "application/x-www-form-urlencoded"),
        body:   "junk",
        status: StatusOK},
    {url: "/",
        method: "POST",
        header: NewHeader(
            HeaderCookie, "xsrf="+testToken,
            HeaderContentType, "application/x-www-form-urlencoded"),
        body:   "junk",
        status: StatusNotFound},
    {url: "/",
        method: "PUT",
        header: NewHeader(
            HeaderCookie, "xsrf="+testToken,
            HeaderContentType, "application/x-www-form-urlencoded"),
        body:   "junk",
        status: StatusNotFound},
    {url: "/",
        method: "DELETE",
        header: NewHeader(
            HeaderCookie, "xsrf="+testToken,
            HeaderContentType, "application/x-www-form-urlencoded"),
        status: StatusNotFound},
}

func xsrfErrorHandler(req *Request, status int, reason error, header Header) {
    io.WriteString(req.Responder.Respond(status, header), req.Param.Get("xsrf"))
}

func xsrfHandler(req *Request) {
    io.WriteString(req.Respond(StatusOK), req.Param.Get("xsrf"))
}

func TestXSRF(t *testing.T) {
    h := SetErrorHandler(xsrfErrorHandler, FormHandler(1000, true, HandlerFunc(xsrfHandler)))

    for i, tt := range xsrfTests {
        status, header, body := RunHandler(tt.url, tt.method, tt.header, []byte(tt.body), h)
        if status != tt.status {
            t.Errorf("test %d, exepected status %d, actual status %d", i, tt.status, status)
        }
        if tt.cookie {
            c := header.Get(HeaderSetCookie)
            if c == "" {
                t.Errorf("test %d, cookie not set", i)
            } else if !strings.HasPrefix(c, "xsrf="+string(body)+";") {
                t.Errorf("test %d, set cookie != param", i)
            }
        } else {
            if string(body) != testToken {
                t.Errorf("test %d, testToken != param", i)
            }
            c := header.Get(HeaderSetCookie)
            if c != "" {
                t.Errorf("test %d, unepxected cookie", i)
            }
        }
    }
}
