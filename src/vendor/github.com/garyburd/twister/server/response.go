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
    "errors"
    "io"
    "vendor/github.com/garyburd/twister/web"
)

type responseBody interface {
    io.Writer
    web.Flusher

    // finish the response body and return an error if the connection should be
    // closed due to a write error.
    finish() (int, error)
}

// nullResponseBody discards the response body.
type nullResponseBody struct {
    err     error
    written int
}

func newNullResponseBody(wr io.Writer, header []byte) (*nullResponseBody, error) {
    w := &nullResponseBody{}
    w.written, w.err = wr.Write(header)
    return w, w.err
}

func (w *nullResponseBody) Write(p []byte) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    return len(p), nil
}

func (w *nullResponseBody) WriteString(p string) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    return len(p), nil
}

func (w *nullResponseBody) Flush() error {
    return w.err
}

func (w *nullResponseBody) finish() (int, error) {
    err := w.err
    if w.err == nil {
        w.err = web.ErrInvalidState
    }
    return w.written, err
}

// identityResponseBody implements identity encoding of the response body.
type identityResponseBody struct {
    err error
    bw  *bufio.Writer
    wr  io.Writer

    // Value of Content-Length header.
    contentLength int

    // Number of body bytes written.
    written int

    // Number of header bytes written.
    headerWritten int
}

func newIdentityResponseBody(wr io.Writer, header []byte, bufferSize, contentLength int) (*identityResponseBody, error) {
    w := &identityResponseBody{wr: wr, contentLength: contentLength}

    w.bw = bufio.NewWriterSize(wr, bufferSize)
    w.headerWritten, w.err = w.bw.Write(header)
    return w, w.err
}

type writerOnly struct{ io.Writer }

func (w *identityResponseBody) ReadFrom(src io.Reader) (n int64, err error) {
    if rf, ok := w.wr.(io.ReaderFrom); ok {
        err = w.bw.Flush()
        if err != nil {
            return
        }
        n, err = rf.ReadFrom(src)
        w.written += int(n)
        return
    }
    // Fall back to default io.Copy implementation.
    // Use wrapper to hide r.ReadFrom from io.Copy.
    return io.Copy(writerOnly{w}, src)
}

func (w *identityResponseBody) Write(p []byte) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    var n int
    n, w.err = w.bw.Write(p)
    w.written += n
    if w.err == nil && w.contentLength >= 0 && w.written > w.contentLength {
        w.err = errors.New("twister: long write by handler")
    }
    return n, w.err
}

func (w *identityResponseBody) WriteString(p string) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    var n int
    n, w.err = w.bw.WriteString(p)
    w.written += n
    if w.err == nil && w.contentLength >= 0 && w.written > w.contentLength {
        w.err = errors.New("twister: long write by handler")
    }
    return n, w.err
}

func (w *identityResponseBody) Flush() error {
    if w.err != nil {
        return w.err
    }
    w.err = w.bw.Flush()
    return w.err
}

func (w *identityResponseBody) finish() (int, error) {
    w.Flush()
    if w.err != nil {
        return w.headerWritten + w.written, w.err
    }
    if w.contentLength >= 0 && w.written < w.contentLength {
        w.err = errors.New("twister: short write by handler")
    }
    err := w.err
    if w.err == nil {
        w.err = web.ErrInvalidState
    }
    return w.headerWritten + w.written, err
}

type chunkedResponseBody struct {
    err     error     // error from wr
    wr      io.Writer // write here
    buf     []byte    // buffered output
    s       int       // start of chunk in buf
    n       int       // current write position in buf
    ndigit  int       // number of hex digits in chunk size
    written int
}

func newChunkedResponseBody(wr io.Writer, header []byte, bufferSize int) (*chunkedResponseBody, error) {
    w := &chunkedResponseBody{wr: wr, buf: make([]byte, bufferSize)}

    for n := int32(bufferSize); n != 0; n >>= 4 {
        w.ndigit += 1
    }

    if len(header) < len(w.buf) {
        w.n = copy(w.buf, header)
    } else {
        w.written, w.err = w.wr.Write(header)
    }

    w.s = w.n
    w.n += w.ndigit + 2
    return w, w.err
}

func (w *chunkedResponseBody) writeBuf() {
    var n int
    n, w.err = w.wr.Write(w.buf[:w.n])
    w.written += n
}

func (w *chunkedResponseBody) finalizeChunk() {
    if w.s+w.ndigit+2 == w.n {
        // The chunk is empty. Reset back start of chunk.
        w.n = w.s
        return
    }

    n := w.n - w.s - w.ndigit - 2

    // CRLF after data.
    w.buf[w.n] = '\r'
    w.buf[w.n+1] = '\n'
    w.n += 2

    // CRLF before data.
    w.buf[w.s+w.ndigit] = '\r'
    w.buf[w.s+w.ndigit+1] = '\n'

    // Length with 0 padding
    for i := w.s + w.ndigit - 1; i >= w.s; i-- {
        w.buf[i] = "0123456789abcdef"[n&0xf]
        n = n >> 4
    }
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *chunkedResponseBody) Flush() error {
    if w.err != nil {
        return w.err
    }
    w.finalizeChunk()
    if w.n > 0 {
        w.writeBuf()
        if w.err != nil {
            return w.err
        }
    }
    w.s = 0
    w.n = w.ndigit + 2 // length CRLF
    return nil
}

func (w *chunkedResponseBody) finish() (int, error) {
    if w.err != nil {
        return w.written, w.err
    }
    w.finalizeChunk()
    const last = "0\r\n\r\n"
    if w.n+len(last) > len(w.buf) {
        w.writeBuf()
        if w.err != nil {
            return w.written, w.err
        }
        w.n = 0
    }
    copy(w.buf[w.n:], last)
    w.n += len(last)
    w.writeBuf()
    err := w.err
    if w.err == nil {
        w.err = web.ErrInvalidState
    }
    return w.written, err
}

func (w *chunkedResponseBody) ncopy(max int) int {
    n := len(w.buf) - w.n - 2 // 2 for CRLF after data
    if n <= 0 {
        w.Flush()
        if w.err != nil {
            return -1
        }
        n = len(w.buf) - w.n - 2 // 2 for CRLF after data
    }
    if n > max {
        n = max
    }
    return n
}

func (w *chunkedResponseBody) Write(p []byte) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    nn := 0
    for len(p) > 0 {
        n := w.ncopy(len(p))
        if n < 0 {
            break
        }
        copy(w.buf[w.n:], p[:n])
        w.n += n
        nn += n
        p = p[n:]
    }
    return nn, w.err
}

func (w *chunkedResponseBody) WriteString(p string) (int, error) {
    if w.err != nil {
        return 0, w.err
    }
    nn := 0
    for len(p) > 0 {
        n := w.ncopy(len(p))
        if n < 0 {
            break
        }
        copy(w.buf[w.n:], p[:n])
        w.n += n
        nn += n
        p = p[n:]
    }
    return nn, w.err
}
