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
    "errors"
    "io"
    "io/ioutil"
    "math"
)

var scratch [1024]byte

func skipReader(r io.Reader, n int) error {
    for n > 0 {
        m := n
        if m > len(scratch) {
            m = len(scratch)
        }
        m, err := r.Read(scratch[:m])
        if err != nil {
            return err
        }
        n -= m
    }
    return nil
}

// Part represents an element of a multi-part request entity.
type Part struct {
    Name         string
    Filename     string
    ContentType  string
    ContentParam map[string]string
    Data         []byte
}

// ParseMultipartForm parses a multipart/form-data body. Form fields are added
// to the request Param. This function loads the entire request body in memory.
// If this is not appropriate, then the application should use MultipartReader
// to read the request body incrementally.
func ParseMultipartForm(req *Request, maxRequestBodyLen int) ([]Part, error) {
    m, err := NewMultipartReader(req, maxRequestBodyLen)
    if err != nil {
        return nil, err
    }
    parts := make([]Part, 0)
    var buf bytes.Buffer
    for {
        header, r, err := m.Next()
        if err == io.EOF {
            break
        } else if err != nil {
            return nil, err
        }
        if disp, dispParam := header.GetValueParam(HeaderContentDisposition); disp == "form-data" {
            if name := dispParam["name"]; name != "" {
                if filename := dispParam["filename"]; filename != "" {
                    contentType, contentParam := header.GetValueParam(HeaderContentType)
                    data, err := ioutil.ReadAll(r)
                    if err != nil {
                        return nil, err
                    }
                    parts = append(parts, Part{
                        ContentType:  contentType,
                        ContentParam: contentParam,
                        Name:         name,
                        Filename:     filename,
                        Data:         data})
                } else {
                    buf.Reset()
                    _, err := buf.ReadFrom(r)
                    if err != nil {
                        return nil, err
                    }
                    req.Param.Add(name, buf.String())
                }
            }
        }
    }
    return parts, nil
}

// MultipartReader reads a multipart/form-data request body.
type MultipartReader struct {
    br       *bufio.Reader
    err      error
    boundary []byte
    avail    int
    r        *partReader
}

var ErrNotMultipartFormData = errors.New("twister: request not multipart/form-data")

// NewMultipartReader returns a multipart/form-data reader. 
func NewMultipartReader(req *Request, maxRequestBodyLen int) (*MultipartReader, error) {

    if req.ContentType != "multipart/form-data" {
        return nil, ErrNotMultipartFormData
    }

    boundary := req.ContentParam["boundary"]
    if boundary == "" {
        return nil, errors.New("twister: multipart/form-data boundary missing")
    }

    if len(boundary) > 512 {
        return nil, errors.New("twister: multipart/form-data boundary too long")
    }

    if maxRequestBodyLen < 0 {
        maxRequestBodyLen = math.MaxInt32
    }

    body := req.Body
    if req.ContentLength > maxRequestBodyLen {
        return nil, ErrRequestEntityTooLarge
    } else if req.ContentLength < 0 {
        body = io.LimitReader(body, int64(maxRequestBodyLen))
    }

    m := &MultipartReader{
        br:       bufio.NewReader(body),
        boundary: []byte("\r\n--" + boundary),
    }

    p, isPrefix, err := m.br.ReadLine()
    if err != nil {
        return nil, err
    }

    if isPrefix || !bytes.Equal(p, m.boundary[2:]) {
        return nil, errors.New("twister: multipart/form-data body malformed")
    }

    return m, nil
}

// Next returns the next part of a multipart/form-data body.  Next returns
// os.EOF if no more parts remain. 
func (m *MultipartReader) Next() (Header, io.Reader, error) {
    if m.r != nil {
        skipReader(m.r, math.MaxInt32)
        m.r = nil
    }

    if m.err != nil {
        return nil, nil, m.err
    }

    header := Header{}
    m.err = header.ParseHttpHeader(m.br)
    if m.err != nil {
        return nil, nil, m.err
    }

    m.avail = 0
    m.r = &partReader{m, nil}
    return header, m.r, nil
}

func (m *MultipartReader) fill() error {
    if m.err != nil {
        return m.err
    }

    // To avoid unnecessary buffer sliding, don't peek more than the buffered
    // amount unless we are getting close to the end of the buffered data (size
    // of boundary + 20 bytes of fluff).
    n := m.br.Buffered()
    if n <= len(m.boundary)+20 {
        n = 4096
    }
    p, err := m.br.Peek(n)

    // 4 = len("--\r\n")
    if len(p) < len(m.boundary)+4 {
        if err == nil || err == io.EOF {
            err = io.ErrUnexpectedEOF
        }
        m.err = err
        return err
    }

    i := bytes.Index(p, m.boundary)
    switch {
    case i == 0:
        switch {
        case bytes.HasPrefix(p[len(m.boundary):], crlfBytes):
            skipReader(m.br, len(m.boundary)+len(crlfBytes))
            return io.EOF
        case bytes.HasPrefix(p[len(m.boundary):], dashDashCrlfBytes):
            // Skip final boundary and up to 4096 bytes of junk following the boundary.
            skipReader(m.br, len(m.boundary)+len(dashDashCrlfBytes)+4096)
            m.err = io.EOF
            return io.EOF
        default:
            m.avail = len(m.boundary)
        }
    case i < 0:
        m.avail = len(p) - len(m.boundary) + 1
    default:
        m.avail = i
    }
    return nil
}

type partReader struct {
    m   *MultipartReader
    err error
}

func (r *partReader) Read(p []byte) (int, error) {
    if r.err != nil {
        return 0, r.err
    }
    nn := 0
    for len(p) > 0 {
        if r.m.avail == 0 {
            r.err = r.m.fill()
            if r.err != nil {
                break
            }
        }
        n := len(p)
        if n > r.m.avail {
            n = r.m.avail
        }
        n, _ = r.m.br.Read(p[:n])
        nn += n
        r.m.avail -= n
        p = p[n:]
    }
    if nn > 0 {
        return nn, nil
    }
    return 0, r.err
}
