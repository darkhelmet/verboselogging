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
    "bufio"
    "bytes"
    "io"
    "net"
    "net/url"
    "time"
)

type testTransaction struct {
    in, out bytes.Buffer
    status  int
    header  Header
}

type testResponder struct {
    t *testTransaction
}

func (r testResponder) Respond(status int, header Header) io.Writer {
    r.t.status = status
    r.t.header = header
    return testResponseBody{r.t}
}

func (r testResponder) Hijack() (net.Conn, *bufio.Reader, error) {
    return testConn{r.t}, bufio.NewReader(&bytes.Buffer{}), nil
}

type testResponseBody struct {
    t *testTransaction
}

func (b testResponseBody) Flush() error {
    return nil
}

func (b testResponseBody) Write(p []byte) (int, error) {
    return b.t.out.Write(p)
}

type testConn struct {
    t *testTransaction
}

func (c testConn) Read(b []byte) (int, error) {
    return c.t.in.Read(b)
}

func (c testConn) Write(b []byte) (int, error) {
    return c.t.out.Write(b)
}

func (c testConn) Close() error {
    return nil
}

func (c testConn) LocalAddr() net.Addr {
    return testAddr("local")
}

func (c testConn) RemoteAddr() net.Addr {
    return testAddr("remote")
}

func (c testConn) SetDeadline(t time.Time) error {
    return nil
}

func (c testConn) SetReadDeadline(t time.Time) error {
    return nil
}

func (c testConn) SetWriteDeadline(t time.Time) error {
    return nil
}

type testAddr string

func (a testAddr) Network() string {
    return string(a)
}

func (a testAddr) String() string {
    return string(a)
}

// RunHandler runs the handler with a request created from the arguments and
// returns the response. This function is intended to be used in tests.
func RunHandler(urlStr string, method string, reqHeader Header, reqBody []byte, handler Handler) (status int, header Header, respBody []byte) {
    var t testTransaction
    if reqBody != nil {
        t.in.Write(reqBody)
    }
    remoteAddr := "1.2.3.4"
    protocolVersion := ProtocolVersion11
    if reqHeader == nil {
        reqHeader = make(Header)
    }
    u, err := url.Parse(urlStr)
    if err != nil {
        panic(err)
    }
    req, err := NewRequest(remoteAddr, method, u.RequestURI(), protocolVersion, u, reqHeader)
    if err != nil {
        panic(err)
    }
    req.Body = &t.in
    req.Responder = testResponder{&t}
    handler.ServeWeb(req)
    return t.status, t.header, t.out.Bytes()
}
