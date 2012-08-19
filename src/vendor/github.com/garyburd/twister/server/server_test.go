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

package server

import (
    "bufio"
    "bytes"
    "io"
    "log"
    "net"
    "os"
    "testing"
    "time"
    "vendor/github.com/garyburd/twister/web"
)

type testAddr string

func (a testAddr) Network() string {
    return string(a)
}

func (a testAddr) String() string {
    return string(a)
}

var defaultErrs = []error{nil, io.EOF}

type testListener struct {
    in, out bytes.Buffer
    done    chan bool
    readAll bool
    errs    []error
}

func (l *testListener) Accept() (conn net.Conn, err error) {
    err = l.errs[0]
    if len(l.errs) > 1 {
        l.errs = l.errs[1:]
    }
    return testConn{l}, err
}

func (l *testListener) Close() error {
    return nil
}

func (l *testListener) Addr() net.Addr {
    return testAddr("listen")
}

type testConn struct {
    *testListener
}

func (c testConn) Read(b []byte) (int, error) {
    n, err := c.in.Read(b)
    if err == io.EOF {
        c.readAll = true
    }
    return n, err
}

func (c testConn) Write(b []byte) (int, error) {
    return c.out.Write(b)
}

func (c testConn) Close() error {
    c.done <- true
    return nil
}

func (c testConn) LocalAddr() net.Addr {
    return testAddr("local")
}

func (c testConn) RemoteAddr() net.Addr {
    return testAddr("remote")
}

func (c testConn) SetDeadline(nsec time.Time) error {
    return nil
}

func (c testConn) SetReadDeadline(nsec time.Time) error {
    return nil
}

func (c testConn) SetWriteDeadline(nsec time.Time) error {
    return nil
}

func testHandler(req *web.Request) {
    req.ParseForm(1000)
    header := make(web.Header)
    if req.Param.Get("panic") == "before" {
        panic("before")
    }
    if s := req.Param.Get("cl"); s != "" {
        header.Set(web.HeaderContentLength, s)
    }
    if req.Param.Get("connection") == "close" {
        header.Set(web.HeaderConnection, "close")
    }
    w := req.Responder.Respond(web.StatusOK, header)
    if s := req.Param.Get("w"); s != "" {
        w.Write([]byte(s))
    }
    if req.Param.Get("panic") == "after" {
        panic("after")
    }
}

var serverTests = []struct {
    in      string
    out     string
    readAll bool
    errs    []error
}{
    {in: "Garbage\r\n",
        out: "HTTP/1.1 400 Bad Request\r\n\r\n",
    },
    {
        in:  "GET / HTTP/1.0\r\n\r\n",
        out: "HTTP/1.0 200 OK\r\nConnection: close\r\n\r\n",
    },
    {
        in:  "GET / HTTP/1.0\r\n\r\n",
        out: "HTTP/1.0 200 OK\r\nConnection: close\r\n\r\n",
    },
    {
        in:  "GET /?w=Hello HTTP/1.0\r\n\r\n",
        out: "HTTP/1.0 200 OK\r\nConnection: close\r\n\r\nHello",
    },
    {
        in:  "GET /?w=Hello HTTP/1.0\r\nConnection: keep-alive\r\n\r\n",
        out: "HTTP/1.0 200 OK\r\nConnection: close\r\n\r\nHello",
    },
    {
        in:  "GET /?cl=5&w=Hello HTTP/1.0\r\n\r\n",
        out: "HTTP/1.0 200 OK\r\nConnection: close\r\nContent-Length: 5\r\n\r\nHello",
    },
    {
        in:      "GET /?cl=5&w=Hello HTTP/1.0\r\nConnection: keep-alive\r\n\r\n",
        out:     "HTTP/1.0 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // Handler forces connection to close.
        in:      "GET /?cl=5&w=Hello&connection=close HTTP/1.0\r\nConnection: keep-alive\r\n\r\n",
        out:     "HTTP/1.0 200 OK\r\nConnection: close\r\nContent-Length: 5\r\n\r\nHello",
        readAll: false,
    },
    {
        in:      "GET /?w=Hello HTTP/1.1\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n0005\r\nHello\r\n0\r\n\r\n",
        readAll: true,
    },
    {
        in:      "GET /?cl=5&w=Hello HTTP/1.1\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // POST
        in:      "POST /?cl=5 HTTP/1.1\r\nContent-Length: 7\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\nw=Hello",
        out:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // POST with chunked body
        in:      "POST /?cl=5 HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\n7\r\nw=Hello\r\n0\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // POST with very chunky body
        in:      "POST /?cl=5 HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\n1\r\nw\r\n1\r\n=\r\n5\r\nHello\r\n0\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // POST with expect
        in:      "POST /?cl=5 HTTP/1.1\r\nContent-Length: 7\r\nContent-Type: application/x-www-form-urlencoded\r\nExpect: 100-continue\r\n\r\nw=Hello",
        out:     "HTTP/1.1 100 Continue\r\n\r\nHTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // POST with expect and chunked body
        in:      "POST /?cl=5 HTTP/1.1\r\nTransfer-Encoding: chunked\r\nContent-Type: application/x-www-form-urlencoded\r\nExpect: 100-continue\r\n\r\n7\r\nw=Hello\r\n0\r\n\r\n",
        out:     "HTTP/1.1 100 Continue\r\n\r\nHTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // Expect connection close because request body not read by handler.
        in:  "POST /?cl=0 HTTP/1.1\r\nContent-Length: 7\r\n\r\nw=Hello",
        out: "HTTP/1.1 200 OK\r\nConnection: close\r\nContent-Length: 0\r\n\r\n",
    },
    {
        // Two requests with identity encoded response.
        in: "GET /?cl=5&w=Hello HTTP/1.1\r\n\r\n" +
            "GET /?cl=5&w=Hello HTTP/1.1\r\n\r\n",
        out: "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello" +
            "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
        readAll: true,
    },
    {
        // Two requests with chunked encoded response.
        in: "GET /?w=Hello HTTP/1.1\r\n\r\n" +
            "GET /?w=Hello HTTP/1.1\r\n\r\n",
        out: "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n0005\r\nHello\r\n0\r\n\r\n" +
            "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n0005\r\nHello\r\n0\r\n\r\n",
        readAll: true,
    },
    {
        // HEAD does not include body for identity encoded responses.
        in:      "HEAD /?cl=5&w=Hello HTTP/1.1\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n",
        readAll: true,
    },
    {
        // HEAD does not include body for chunked  encoded responses.
        in:      "HEAD /?w=Hello HTTP/1.1\r\n\r\n",
        out:     "HTTP/1.1 200 OK\r\n\r\n",
        readAll: true,
    },
    {
        // panic
        in: "GET /?cl=5&w=Hello&panic=before HTTP/1.1\r\n\r\n",
    },
    {
        // panic
        in:  "GET /?cl=5&w=Hello&panic=after HTTP/1.1\r\n\r\n",
        out: "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\nHello",
    },
    /*
    	{
    		// temporary error
    		in:      "GET /?w=Hello HTTP/1.1\r\n\r\n",
    		out:     "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n0005\r\nHello\r\n0\r\n\r\n",
    		readAll: true,
    		errs:    []error{os.EINTR, nil, io.EOF},
    	},
    */
}

type silentLogger struct {
    t *testing.T
}

func (l silentLogger) Write(p []byte) (int, error) {
    l.t.Log(string(p))
    return len(p), nil
}

func TestServer(t *testing.T) {
    log.SetOutput(silentLogger{t})
    defer log.SetOutput(os.Stdout)
    for _, st := range serverTests {
        l := &testListener{done: make(chan bool), errs: st.errs}
        l.in.WriteString(st.in)
        if l.errs == nil {
            l.errs = defaultErrs
        }
        err := (&Server{Listener: l, Handler: web.HandlerFunc(testHandler)}).Serve()
        if err != io.EOF {
            t.Errorf("Server() = %v", err)
        }
        <-l.done
        out := l.out.String()
        if out != st.out {
            t.Errorf("in=%q\ngot:  %q\nwant: %q", st.in, out, st.out)
        }
        if l.readAll != st.readAll {
            t.Errorf("in=%q readAll = %v, want %v", st.in, l.readAll, st.readAll)
        }
    }
}

var readRequestLineTests = []struct {
    line    string
    method  string
    url     string
    version int
}{
    {
        "GET / HTTP/1.0",
        "GET",
        "/",
        web.ProtocolVersion10,
    },
    {
        "GET / HTTP/1.1",
        "GET",
        "/",
        web.ProtocolVersion11,
    },
    {
        "GET /bax HTTP/01.01",
        "GET",
        "/bax",
        web.ProtocolVersion11,
    },
    {
        "GET /bax xTTP/01.01",
        "",
        "",
        0,
    },
    {
        "GET/bax HTTP/01.01",
        "",
        "",
        0,
    },
}

func TestReadRequestLine(t *testing.T) {
    for _, tt := range readRequestLineTests {
        r := bufio.NewReader(bytes.NewBuffer([]byte(tt.line + "\r\n")))
        method, url, version, err := readRequestLine(r)
        if (err != nil) != (tt.method == "") {
            v := tt.method == ""
            t.Errorf("%s, err=%v expectedErr=%v", tt.line, err, v)
        }
        if err != nil {
            continue
        }
        if method != tt.method {
            t.Errorf("%s, method=%s, want=%s", tt.line, method, tt.method)
        }
        if url != tt.url {
            t.Errorf("%s, url=%s, want=%s", tt.line, url, tt.url)
        }
        if version != tt.version {
            t.Errorf("%s, version=%d, want=%d", tt.line, version, tt.version)
        }
    }
}
