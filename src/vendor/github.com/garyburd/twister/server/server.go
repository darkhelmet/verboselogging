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

// Package server implements Twister's HTTP server.
package server

import (
    "bufio"
    "bytes"
    "errors"
    "io"
    "log"
    "net"
    "net/url"
    "runtime/debug"
    "strconv"
    "strings"
    "vendor/github.com/garyburd/twister/web"
)

var errBadRequestLine = errors.New("twister.server: could not parse request line")

// Server defines parameters for running an HTTP server.
type Server struct {
    // The server accepts incoming connections on this listener. The
    // application is required to set this field.
    Listener net.Listener

    // The server dispatches requests to this handler. The application is
    // required to set this field.
    Handler web.Handler

    // If true, then set the request URL protocol to HTTPS.
    Secure bool

    // Set request URL host to this string if host is not specified in the
    // request or headers.
    DefaultHost string

    // Log the request.
    Logger Logger

    // If true, do not recover from handler panics.
    NoRecoverHandlers bool
}

// Logger defines an interface for logging a request.
type Logger interface {
    Log(lr *LogRecord)
}

// LoggerFunc is a type adapter to allow the use of ordinary functions as Logger.
type LoggerFunc func(*LogRecord)

// Log calls f(lr).
func (f LoggerFunc) Log(lr *LogRecord) { f(lr) }

// transaction represents a single request-response transaction.
type transaction struct {
    server             *Server
    conn               net.Conn
    br                 *bufio.Reader
    responseBody       responseBody
    chunkedResponse    bool
    chunkedRequest     bool
    closeAfterResponse bool
    hijacked           bool
    req                *web.Request
    requestAvail       int
    requestErr         error
    requestConsumed    bool
    respondCalled      bool
    responseErr        error
    write100Continue   bool
    status             int
    header             web.Header
    headerSize         int
}

var httpslash = []byte("HTTP/")

// nextNum scans the next decimal number from p.
func nextNum(p []byte) (n int, rest []byte, err error) {
    for i, b := range p {
        switch {
        case '0' <= b && b <= '9':
            n = n*10 + int(b-'0')
            if n > 1000 {
                err = errBadRequestLine
                return
            }
        case i == 0:
            err = errBadRequestLine
            return
        default:
            rest = p[i:]
            return
        }
    }
    return
}

// nextWord scans to the next space in p.
func nextWord(p []byte) (s string, rest []byte, err error) {
    i := bytes.IndexByte(p, ' ')
    if i < 0 {
        err = errBadRequestLine
        return
    }
    s = string(p[:i])
    rest = p[i+1:]
    return
}

func readRequestLine(b *bufio.Reader) (method string, urlStr string, version int, err error) {
    var p []byte
    var isPrefix bool

    p, isPrefix, err = b.ReadLine()
    if isPrefix {
        err = web.ErrLineTooLong
    }
    if err != nil {
        return
    }

    method, p, err = nextWord(p)
    if err != nil {
        return
    }

    urlStr, p, err = nextWord(p)
    if err != nil {
        return
    }

    if !bytes.HasPrefix(p, httpslash) {
        err = errBadRequestLine
        return
    }

    var major int
    major, p, err = nextNum(p[len(httpslash):])
    if err != nil {
        return
    }

    if len(p) == 0 || p[0] != '.' {
        err = errBadRequestLine
        return
    }

    var minor int
    minor, p, err = nextNum(p[1:])
    if err != nil {
        return
    }

    if len(p) != 0 {
        err = errBadRequestLine
        return
    }

    version = web.ProtocolVersion(major, minor)
    return
}

func (t *transaction) prepare() (err error) {
    method, requestURI, version, err := readRequestLine(t.br)
    if err != nil {
        return err
    }

    header := web.Header{}
    err = header.ParseHttpHeader(t.br)
    if err != nil {
        return err
    }

    u, err := url.ParseRequestURI(requestURI)
    if err != nil {
        return err
    }

    if u.Host == "" {
        u.Host = header.Get(web.HeaderHost)
        if u.Host == "" {
            u.Host = t.server.DefaultHost
        }
    }

    if t.server.Secure {
        u.Scheme = "https"
    } else {
        u.Scheme = "http"
    }

    req, err := web.NewRequest(t.conn.RemoteAddr().String(), method, requestURI, version, u, header)
    if err != nil {
        return
    }
    t.req = req

    if s := req.Header.Get(web.HeaderExpect); s != "" {
        t.write100Continue = strings.ToLower(s) == "100-continue"
    }

    connection := strings.ToLower(req.Header.Get(web.HeaderConnection))
    if version >= web.ProtocolVersion(1, 1) {
        t.closeAfterResponse = connection == "close"
    } else if version == web.ProtocolVersion(1, 0) && req.ContentLength >= 0 {
        t.closeAfterResponse = connection != "keep-alive"
    } else {
        t.closeAfterResponse = true
    }

    req.Responder = t

    te := header.GetList(web.HeaderTransferEncoding)
    chunked := len(te) > 0 && te[0] == "chunked"

    switch {
    case req.Method == "GET" || req.Method == "HEAD":
        req.Body = identityReader{t}
        t.requestConsumed = true
    case chunked:
        req.Body = chunkedReader{t}
    case req.ContentLength >= 0:
        req.Body = identityReader{t}
        t.requestAvail = req.ContentLength
        t.requestConsumed = req.ContentLength == 0
    default:
        req.Body = identityReader{t}
        t.closeAfterResponse = true
    }

    return nil
}

func (t *transaction) checkRead() error {
    if t.requestErr != nil {
        if t.requestErr == web.ErrInvalidState {
            log.Println("twister: Request Read after response started.")
        }
        return t.requestErr
    }
    if t.write100Continue {
        t.write100Continue = false
        io.WriteString(t.conn, "HTTP/1.1 100 Continue\r\n\r\n")
    }
    return nil
}

type identityReader struct{ *transaction }

func (t identityReader) Read(p []byte) (int, error) {
    if err := t.checkRead(); err != nil {
        return 0, err
    }
    if t.requestAvail <= 0 {
        t.requestErr = io.EOF
        return 0, t.requestErr
    }
    if len(p) > t.requestAvail {
        p = p[:t.requestAvail]
    }
    var n int
    n, t.requestErr = t.br.Read(p)
    t.requestAvail -= n
    if t.requestAvail == 0 {
        t.requestConsumed = true
    }
    return n, t.requestErr
}

type chunkedReader struct{ *transaction }

func (t chunkedReader) Read(p []byte) (n int, err error) {
    if err = t.checkRead(); err != nil {
        return 0, err
    }
    if t.requestAvail == 0 {
        // We delay reading the first chunk length to this point to ensure that
        // we don't read the body until 100-continue is send (if needed).
        t.requestAvail, t.requestErr = readChunkFraming(t.br, true)
        if t.requestErr != nil {
            return 0, t.requestErr
            if t.requestErr == io.EOF {
                t.requestConsumed = true
            }
        }
    }
    if len(p) > t.requestAvail {
        p = p[:t.requestAvail]
    }
    n, err = t.br.Read(p)
    t.requestErr = err
    t.requestAvail -= n
    if err == nil && t.requestAvail == 0 {
        // We read the next chunk length here to ensure that the entire request
        // body encoding is consumed in case where the application reads
        // exactly the number of bytes in the decoded body.
        t.requestAvail, t.requestErr = readChunkFraming(t.br, false)
        if t.requestErr == io.EOF {
            t.requestConsumed = true
        }
    }
    return n, err
}

func readChunkFraming(br *bufio.Reader, first bool) (int, error) {
    if !first {
        // trailer from previous chunk
        p := make([]byte, 2)
        if _, err := io.ReadFull(br, p); err != nil {
            return 0, err
        }
        if p[0] != '\r' && p[1] != '\n' {
            return 0, errors.New("twister: bad chunked format")
        }
    }

    line, isPrefix, err := br.ReadLine()
    if err != nil {
        return 0, err
    }
    if isPrefix {
        return 0, errors.New("twister: bad chunked format")
    }
    n, err := strconv.ParseUint(string(line), 16, 64)
    if err != nil {
        return 0, err
    }
    if n == 0 {
        for {
            line, isPrefix, err = br.ReadLine()
            if err != nil {
                return 0, err
            }
            if isPrefix {
                return 0, errors.New("twister: bad chunked format")
            }
            if len(line) == 0 {
                return 0, io.EOF
            }
        }
    }
    return int(n), nil
}

func (t *transaction) Respond(status int, header web.Header) (body io.Writer) {
    if t.hijacked {
        log.Println("twister: Respond called on hijacked connection")
        return &nullResponseBody{err: web.ErrInvalidState}
    }
    if t.respondCalled {
        log.Println("twister: Multiple calls to Respond")
        return &nullResponseBody{err: web.ErrInvalidState}
    }
    t.respondCalled = true
    t.requestErr = web.ErrInvalidState
    t.status = status
    t.header = header

    if te := header.Get(web.HeaderTransferEncoding); te != "" {
        log.Println("twister: transfer encoding not allowed")
        delete(header, web.HeaderTransferEncoding)
    }

    if !t.requestConsumed {
        t.closeAfterResponse = true
    }

    if header.Get(web.HeaderConnection) == "close" {
        t.closeAfterResponse = true
    }

    t.chunkedResponse = true
    contentLength := -1

    if status == web.StatusNotModified {
        delete(header, web.HeaderContentType)
        delete(header, web.HeaderContentLength)
        t.chunkedResponse = false
    } else if s := header.Get(web.HeaderContentLength); s != "" {
        contentLength, _ = strconv.Atoi(s)
        t.chunkedResponse = false
    } else if t.req.ProtocolVersion < web.ProtocolVersion(1, 1) {
        t.closeAfterResponse = true
    }

    if t.closeAfterResponse {
        header.Set(web.HeaderConnection, "close")
        t.chunkedResponse = false
    }

    if t.req.Method == "HEAD" {
        t.chunkedResponse = false
    }

    if t.chunkedResponse {
        header.Set(web.HeaderTransferEncoding, "chunked")
    }

    proto := "HTTP/1.0"
    if t.req.ProtocolVersion >= web.ProtocolVersion(1, 1) {
        proto = "HTTP/1.1"
    }
    statusString := strconv.Itoa(status)
    text := web.StatusText(status)

    var b bytes.Buffer
    b.WriteString(proto)
    b.WriteString(" ")
    b.WriteString(statusString)
    b.WriteString(" ")
    b.WriteString(text)
    b.WriteString("\r\n")
    header.WriteHttpHeader(&b)
    t.headerSize = b.Len()

    const bufferSize = 4096
    switch {
    case t.req.Method == "HEAD" || status == web.StatusNotModified:
        t.responseBody, _ = newNullResponseBody(t.conn, b.Bytes())
    case t.chunkedResponse:
        t.responseBody, _ = newChunkedResponseBody(t.conn, b.Bytes(), bufferSize)
    default:
        t.responseBody, _ = newIdentityResponseBody(t.conn, b.Bytes(), bufferSize, contentLength)
    }
    return t.responseBody
}

func (t *transaction) Hijack() (conn net.Conn, br *bufio.Reader, err error) {
    if t.respondCalled {
        return nil, nil, web.ErrInvalidState
    }

    conn = t.conn
    br = t.br

    if t.server.Logger != nil {
        t.server.Logger.Log(&LogRecord{
            Request:  t.req,
            Header:   t.header,
            Hijacked: true,
        })
    }

    t.hijacked = true
    t.requestErr = web.ErrInvalidState
    t.responseErr = web.ErrInvalidState
    t.req = nil
    t.br = nil
    t.conn = nil

    return
}

func (t *transaction) invokeHandler() {
    if !t.server.NoRecoverHandlers {
        defer func() {
            if r := recover(); r != nil {
                urlStr := "none"
                if t.req != nil && t.req.URL != nil {
                    urlStr = t.req.URL.String()
                }
                stack := string(debug.Stack())
                log.Printf("Panic while serving \"%s\": %v\n%s", urlStr, r, stack)
                t.closeAfterResponse = true
            }
        }()
    }
    t.server.Handler.ServeWeb(t.req)
}

// Finish the HTTP request
func (t *transaction) finish() error {
    if !t.respondCalled {
        urlStr := "unknown"
        if t.req != nil && t.req.URL != nil {
            urlStr = t.req.URL.String()
        }
        return errors.New("twister: handler did not call respond while serving " + urlStr)
    }
    var written int
    if t.responseErr == nil {
        written, t.responseErr = t.responseBody.finish()
    }
    if t.responseErr != nil {
        t.closeAfterResponse = true
    } else {
        t.responseErr = web.ErrInvalidState
    }
    if t.server.Logger != nil {
        err := t.responseErr
        if err == web.ErrInvalidState {
            err = t.requestErr
            if err == web.ErrInvalidState {
                err = nil
            }
        }
        t.server.Logger.Log(&LogRecord{
            Written:    written,
            Request:    t.req,
            Header:     t.header,
            HeaderSize: t.headerSize,
            Status:     t.status,
            Error:      err})
    }
    t.conn = nil
    t.br = nil
    t.responseBody = nil
    return nil
}

func (s *Server) serveConnection(conn net.Conn) {
    defer conn.Close()
    br := bufio.NewReader(conn)
    for {
        t := &transaction{
            server: s,
            conn:   conn,
            br:     br}
        if err := t.prepare(); err != nil {
            if err != io.EOF {
                log.Println("twister: prepare failed", err)
                io.WriteString(conn, "HTTP/1.1 400 Bad Request\r\n\r\n")
            }
            break
        }

        t.invokeHandler()
        if t.hijacked {
            return
        }
        if err := t.finish(); err != nil {
            log.Println("twister: finish failed", err)
            break
        }
        if t.closeAfterResponse {
            break
        }
    }
}

// Serve accepts incoming HTTP connections on s.Listener, creating a new
// goroutine for each. The goroutines read requests and then call s.Handler to
// respond to the request.
//
// The "Hello World" server using Serve() is:
//
//  package main
//
//  import (
//      "github.com/garyburd/twister/web"
//      "github.com/garyburd/twister/server"
//      "io"
//      "log"
//      "net"
//  )
//
//  func helloHandler(req *web.Request) {
//      w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain")
//      io.WriteString(w, "Hello, World!\n")
//  }
//
//  func main() {
//      handler := web.NewRouter().Register("/", "GET", helloHandler)
//      listener, err := net.Listen("tcp", ":8080")
//      if err != nil {
//          log.Fatal("Listen", err)
//      }
//      defer listener.Close()
//      err = (&server.Server{Listener: listener, Handler: handler}).Serve()
//      if err != nil {
//          log.Fatal("Server", err)
//      }
//  }
func (s *Server) Serve() error {
    for {
        conn, e := s.Listener.Accept()
        if e != nil {
            if e, ok := e.(net.Error); ok && e.Temporary() {
                log.Printf("twister.server: accept error %v", e)
                continue
            }
            return e
        }
        go s.serveConnection(conn)
    }
    return nil
}

// Run is a convenience function for running an HTTP server. Run listens on the
// TCP address addr, initializes a server object and calls the server's Serve()
// method to handle HTTP requests. Run logs a fatal error if it encounters an
// error.
//
// The Server object is initialized with the handler argument and listener. If
// the application needs to set any other Server fields or if the application
// needs to create the listener, then the application should directly create
// the Server object and call the Serve() method.
//
// The "Hello World" server using Run() is:
//
//  package main
//
//  import (
//      "github.com/garyburd/twister/web"
//      "github.com/garyburd/twister/server"
//      "io"
//  )
//
//  func helloHandler(req *web.Request) {
//      w := req.Respond(web.StatusOK, web.HeaderContentType, "text/plain")
//      io.WriteString(w, "Hello, World!\n")
//  }
//
//  func main() {
//      server.Run(":8080", web.NewRouter().Register("/", "GET", helloHandler))
//  }
//
func Run(addr string, handler web.Handler) {
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatal("Listen", err)
        return
    }
    defer listener.Close()
    err = (&Server{Logger: LoggerFunc(ShortLogger), Listener: listener, Handler: handler}).Serve()
    if err != nil {
        log.Fatal("Server", err)
    }
}
