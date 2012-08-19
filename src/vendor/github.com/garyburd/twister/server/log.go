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

package server

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net"
    "sync"
    "time"
    "vendor/github.com/garyburd/twister/web"
)

// LogRecord records information about a request for logging.
type LogRecord struct {
    // The request, possibly modified by handlers.
    Request *web.Request

    // Errors encountered while handling request.
    Error error

    // Response status.
    Status int

    // Response headers.
    Header web.Header

    // Number of bytes written to output including headers and transfer encoding.
    Written int

    // Size of the header in bytes.
    HeaderSize int

    // True if connection hijacked.
    Hijacked bool
}

func writeStringMap(w io.Writer, title string, m map[string][]string) {
    first := true
    for key, values := range m {
        if first {
            fmt.Fprintf(w, "  %s:\n", title)
            first = false
        }
        for _, value := range values {
            fmt.Fprintf(w, "    %s: %s\n", key, value)
        }
    }
}

// ShortLogger logs a short summary of the request.
func ShortLogger(lr *LogRecord) {
    if lr.Error != nil {
        log.Printf("%d %s %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL, lr.Error)
    } else {
        log.Printf("%d %s %s\n", lr.Status, lr.Request.Method, lr.Request.URL)
    }
}

// VerboseLogger prints out just about everything about the request and response.
func VerboseLogger(lr *LogRecord) {
    var b = &bytes.Buffer{}
    fmt.Fprintf(b, "REQUEST\n")
    fmt.Fprintf(b, "  %s HTTP/%d.%d %s\n", lr.Request.Method, lr.Request.ProtocolVersion/1000, lr.Request.ProtocolVersion%1000, lr.Request.URL)
    fmt.Fprintf(b, "  RemoteAddr:  %s\n", lr.Request.RemoteAddr)
    fmt.Fprintf(b, "  ContentType:  %s\n", lr.Request.ContentType)
    fmt.Fprintf(b, "  ContentLength:  %d\n", lr.Request.ContentLength)
    writeStringMap(b, "Header", map[string][]string(lr.Request.Header))
    writeStringMap(b, "Param", map[string][]string(lr.Request.Param))
    writeStringMap(b, "Cookie", map[string][]string(lr.Request.Cookie))
    if lr.Hijacked {
        fmt.Fprintf(b, "HIJACKED\n")
    } else {
        fmt.Fprintf(b, "RESPONSE\n")
        fmt.Fprintf(b, "  Error: %v\n", lr.Error)
        fmt.Fprintf(b, "  Status: %d\n", lr.Status)
        fmt.Fprintf(b, "  Written: %d\n", lr.Written)
        fmt.Fprintf(b, "  HeaderSize: %d\n", lr.HeaderSize)
        writeStringMap(b, "Header", lr.Header)
    }
    log.Print(b.String())
}

// ApacheCombinedLogger writes Apache Combined Log style logs to the given writer.
//
// Example usage:
//
//   logFile, err := os.Open("access.log", os.O_CREAT|os.O_WRONLY|os.O_APPEND, 0644)
//   if err != nil {
//       panic(fmt.Sprintf("Failed to open \"access.log\": %s", err.String()))
//   }
//
//   defer logFile.Close()
//   logger := server.NewApacheCombinedLogger(logFile)
type ApacheCombinedLogger struct {
    mutex sync.Mutex
    w     io.Writer
}

const apacheTimeFormat = "02/Jan/2006:15:04:05 -0700"

// NewApacheCombinedLogger creates a new Apache logger.
func NewApacheCombinedLogger(w io.Writer) *ApacheCombinedLogger {
    return &ApacheCombinedLogger{w: w}
}

// SwitchFiles switches the output of the logger to the new writer.
func (acl *ApacheCombinedLogger) SwitchFiles(w io.Writer) {
    acl.mutex.Lock()
    defer acl.mutex.Unlock()

    acl.w = w
}

func (acl *ApacheCombinedLogger) Log(lr *LogRecord) {
    if acl.w == nil {
        return
    }

    host, _, err := net.SplitHostPort(lr.Request.RemoteAddr)
    if err != nil {
        log.Print(fmt.Sprintf("Failed to resolve \"%s\": %s", lr.Request.RemoteAddr, err.Error()))
        return
    }

    var b = &bytes.Buffer{}
    fmt.Fprintf(b, "%s - - [%s] ", host, time.Now().Format(apacheTimeFormat))
    fmt.Fprintf(b, "\"%s %s HTTP/%d.%d\" ",
        lr.Request.Method, lr.Request.URL, lr.Request.ProtocolVersion/1000, lr.Request.ProtocolVersion%1000)
    fmt.Fprintf(b, "%d %d \"%s\" \"%s\"\n",
        lr.Status, lr.Written-lr.HeaderSize, lr.Request.Header.Get(web.HeaderReferer),
        lr.Request.Header.Get(web.HeaderUserAgent))

    // Lock to make sure that we don't write while log output is being changed.
    acl.mutex.Lock()
    defer acl.mutex.Unlock()

    acl.w.Write(b.Bytes())
}
