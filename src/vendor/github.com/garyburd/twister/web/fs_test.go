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
    "os"
    "reflect"
    "strconv"
    "testing"
)

var testEtag = computeTestEtag()
var testContentLength = computeTestContentLength()

func computeTestEtag() string {
    info, _ := os.Stat("fs_test.go")
    return QuoteHeaderValue(strconv.FormatInt(info.ModTime().UnixNano(), 36))
}

func computeTestContentLength() string {
    info, _ := os.Stat("fs_test.go")
    return strconv.FormatInt(info.Size(), 10)
}

var fileHandlerTests = []struct {
    options        *ServeFileOptions
    method         string
    requestHeader  Header
    responseHeader Header
    status         int
    noBody         bool
    url            string
}{
    {
        // Simple GET
        method: "GET",
        status: StatusOK,
        responseHeader: NewHeader(
            HeaderEtag, testEtag,
            HeaderContentLength, testContentLength),
    },
    {
        // GET with ?v=
        method: "GET",
        status: StatusOK,
        responseHeader: NewHeader(
            HeaderEtag, testEtag,
            HeaderCacheControl, "max-age=315360000",
            HeaderContentLength, testContentLength),
        url: "http://example.com/?v=10",
    },
    {
        // GET with ?v= and cache control header in options
        method:  "GET",
        status:  StatusOK,
        options: &ServeFileOptions{Header: NewHeader(HeaderCacheControl, "foo, max-age=2, bar")},
        responseHeader: NewHeader(
            HeaderEtag, testEtag,
            HeaderCacheControl, "foo, bar, max-age=315360000",
            HeaderContentLength, testContentLength),
        url: "http://example.com/?v=10",
    },
    {
        // Simple HEAD
        method: "HEAD",
        status: StatusOK,
        responseHeader: NewHeader(
            HeaderEtag, testEtag,
            HeaderContentLength, testContentLength),
        noBody: true,
    },
    {
        // If-None-Match
        method: "GET",
        status: StatusNotModified,
        requestHeader: NewHeader(
            HeaderIfNoneMatch, testEtag),
        responseHeader: NewHeader(
            HeaderEtag, testEtag),
        noBody: true,
    },
    {
        // If-None-Match with entity headers in options.
        method:  "GET",
        status:  StatusNotModified,
        options: &ServeFileOptions{Header: NewHeader(HeaderContentType, "text/plain")},
        requestHeader: NewHeader(
            HeaderIfNoneMatch, testEtag),
        responseHeader: NewHeader(
            HeaderEtag, testEtag),
        noBody: true,
    },
    {
        // If-None-Match with extra stuff in header
        method: "GET",
        status: StatusNotModified,
        requestHeader: NewHeader(
            HeaderIfNoneMatch, "random, "+testEtag+", junk"),
        responseHeader: NewHeader(
            HeaderEtag, testEtag),
        noBody: true,
    },
}

func TestFileHandler(t *testing.T) {
    for _, tt := range fileHandlerTests {

        url := tt.url
        if url == "" {
            url = "http://example.com/"
        }

        fh := FileHandler("fs_test.go", tt.options)
        status, header, body := RunHandler(url, tt.method, tt.requestHeader, nil, fh)

        if status != tt.status {
            t.Errorf("%s %s %v %v status=%d, want %d", tt.method, url, tt.options, tt.requestHeader, status, tt.status)
        }

        delete(header, HeaderExpires)
        if !reflect.DeepEqual(header, tt.responseHeader) {
            t.Errorf("%s %s %v %v header=%v, want %v", tt.method, url, tt.options, tt.requestHeader, header, tt.responseHeader)
        }

        noBody := len(body) == 0
        if noBody != tt.noBody {
            t.Errorf("%s %s %v %v no body=%v", tt.method, url, tt.options, tt.requestHeader, noBody)
        }
    }
}
