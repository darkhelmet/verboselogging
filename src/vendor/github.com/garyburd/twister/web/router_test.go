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
    "sort"
    "testing"
)

type routeTestHandler string

func (h routeTestHandler) ServeWeb(req *Request) {
    w := req.Respond(StatusOK)
    var keys []string
    for key := range req.URLParam {
        keys = append(keys, key)
    }
    sort.Strings(keys)
    w.Write([]byte(string(h)))
    for _, key := range keys {
        w.Write([]byte(" "))
        w.Write([]byte(key))
        w.Write([]byte(":"))
        w.Write([]byte(req.URLParam[key]))
    }
}

var routeTests = []struct {
    url    string
    method string
    status int
    body   string
}{
    {url: "/Bogus/Path", method: "GET", status: StatusNotFound, body: ""},
    {url: "/Bogus/Path", method: "POST", status: StatusNotFound, body: ""},
    {url: "/", method: "GET", status: StatusOK, body: "home-get"},
    {url: "/", method: "HEAD", status: StatusOK, body: "home-get"},
    {url: "/", method: "POST", status: StatusMethodNotAllowed, body: ""},
    {url: "/a", method: "GET", status: StatusOK, body: "a-get"},
    {url: "/a", method: "HEAD", status: StatusOK, body: "a-get"},
    {url: "/a", method: "POST", status: StatusOK, body: "a-*"},
    {url: "/a/", method: "GET", status: StatusNotFound, body: ""},
    {url: "/b", method: "GET", status: StatusOK, body: "b-get"},
    {url: "/b", method: "HEAD", status: StatusOK, body: "b-get"},
    {url: "/b", method: "POST", status: StatusOK, body: "b-post"},
    {url: "/b", method: "PUT", status: StatusMethodNotAllowed, body: ""},
    {url: "/c", method: "GET", status: StatusOK, body: "c-*"},
    {url: "/c", method: "HEAD", status: StatusOK, body: "c-*"},
    {url: "/d", method: "GET", status: StatusMovedPermanently, body: ""},
    {url: "/d/", method: "GET", status: StatusOK, body: "d"},
    {url: "/e/foo", method: "GET", status: StatusOK, body: "e x:foo"},
    {url: "/e/foo/", method: "GET", status: StatusNotFound, body: ""},
    {url: "/f/foo/bar", method: "GET", status: StatusMovedPermanently, body: ""},
    {url: "/f/foo/bar/", method: "GET", status: StatusOK, body: "f x:foo y:bar"},
    {url: "/g/foo", method: "GET", status: StatusNotFound, body: ""},
    {url: "/g/99", method: "GET", status: StatusOK, body: "g x:99"},
}

func TestRouter(t *testing.T) {
    r := NewRouter()
    r.Register("/", "GET", routeTestHandler("home-get"))
    r.Register("/a", "GET", routeTestHandler("a-get"), "*", routeTestHandler("a-*"))
    r.Register("/b", "GET", routeTestHandler("b-get"), "POST", routeTestHandler("b-post"))
    r.Register("/c", "*", routeTestHandler("c-*"))
    r.Register("/d/", "GET", routeTestHandler("d"))
    r.Register("/e/<x>", "GET", routeTestHandler("e"))
    r.Register("/f/<x>/<y>/", "GET", routeTestHandler("f"))
    r.Register("/g/<x:[0-9]+>", "GET", routeTestHandler("g"))

    for _, rt := range routeTests {
        status, _, body := RunHandler(rt.url, rt.method, nil, nil, r)
        if status != rt.status {
            t.Errorf("url=%s method=%s, status=%d, want %d", rt.url, rt.method, status, rt.status)
        }
        if status == StatusOK {
            if string(body) != rt.body {
                t.Errorf("url=%s method=%s body=%q, want %q", rt.url, rt.method, string(body), rt.body)
            }
        }
    }
}

var hostRouteTests = []struct {
    url    string
    status int
    body   string
}{
    {url: "http://www.example.com/", status: StatusOK, body: "www.example.com"},
    {url: "http://foo.example.com/", status: StatusOK, body: "*.example.com x:foo"},
    {url: "http://example.com/", status: StatusOK, body: "default"},
}

func TestHostRouter(t *testing.T) {
    r := NewHostRouter(routeTestHandler("default"))
    r.Register("www.example.com", routeTestHandler("www.example.com"))
    r.Register("<x>.example.com", routeTestHandler("*.example.com"))

    for _, rt := range hostRouteTests {
        status, _, body := RunHandler(rt.url, "GET", nil, nil, r)
        if status != rt.status {
            t.Errorf("url=%s, status=%d, want %d", rt.url, status, rt.status)
        }
        if status == StatusOK {
            if string(body) != rt.body {
                t.Errorf("url=%s, body=%q, want %q", rt.url, string(body), rt.body)
            }
        }
    }
}
