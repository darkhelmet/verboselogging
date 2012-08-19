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
    "net/url"
    "reflect"
    "strings"
    "testing"
)

var multiPartTests = []struct {
    body  string
    param Values
    parts []Part
}{
    {
        // field
        body: "--deadbeef" +
            "\r\nContent-Disposition: form-data; name=\"name\"\r\n" +
            "\r\n" +
            "value" +
            "\r\n--deadbeef--\r\n",
        param: NewValues("name", "value"),
        parts: []Part{},
    },
    {
        // field, field
        body: "--deadbeef\r\n" +
            "Content-Disposition: form-data; name=\"name\"\r\n" +
            "\r\n" +
            "value" +
            "\r\n--deadbeef\r\n" +
            "Content-Disposition: form-data; name=\"hello\"\r\n" +
            "\r\n" +
            "world" +
            "\r\n--deadbeef--\r\n",
        param: NewValues("name", "value", "hello", "world"),
        parts: []Part{},
    },
    {
        // field, file
        body: "--deadbeef\r\n" +
            "Content-Disposition: form-data; name=hello\r\n" +
            "\r\n" +
            "world" +
            "\r\n--deadbeef\r\n" +
            "Content-Disposition: form-data; filename=\"file.txt\"; name=file\r\n" +
            "Content-Type: text/plain\r\n" +
            "\r\n" +
            "file-content" +
            "\r\n--deadbeef--\r\n",
        param: NewValues("hello", "world"),
        parts: []Part{
            {
                Name:         "file",
                Filename:     "file.txt",
                ContentType:  "text/plain",
                ContentParam: map[string]string{},
                Data:         []byte("file-content"),
            }},
    },
    {
        // file, field
        body: "--deadbeef\r\n" +
            "Content-Disposition: form-data; filename=\"file.txt\"; name=file\r\n" +
            "Content-Type: text/plain\r\n" +
            "\r\n" +
            "file-content" +
            "\r\n--deadbeef\r\n" +
            "Content-Disposition: form-data; name=hello\r\n" +
            "\r\n" +
            "world" +
            "\r\n--deadbeef--\r\n",
        param: NewValues("hello", "world"),
        parts: []Part{
            {
                Name:         "file",
                Filename:     "file.txt",
                ContentType:  "text/plain",
                ContentParam: map[string]string{},
                Data:         []byte("file-content"),
            }},
    },
    {
        // large field, large field
        body: "--deadbeef\r\n" +
            "Content-Disposition: form-data; name=\"name\"\r\n" +
            "\r\n" +
            strings.Repeat("abcd", 1025) +
            "\r\n--deadbeef\r\n" +
            "Content-Disposition: form-data; name=\"hello\"\r\n" +
            "\r\n" +
            strings.Repeat("ijkl", 1025) +
            "\r\n--deadbeef--\r\n",
        param: NewValues("name", strings.Repeat("abcd", 1025), "hello", strings.Repeat("ijkl", 1025)),
        parts: []Part{},
    },
    {
        // large file
        body: "--deadbeef\r\n" +
            "Content-Disposition: form-data; filename=\"file.txt\"; name=file\r\n" +
            "Content-Type: text/plain\r\n" +
            "\r\n" +
            strings.Repeat("abcd", 1025) +
            "\r\n--deadbeef--\r\n",
        param: NewValues(),
        parts: []Part{
            {
                Name:         "file",
                Filename:     "file.txt",
                ContentType:  "text/plain",
                ContentParam: map[string]string{},
                Data:         []byte(strings.Repeat("abcd", 1025)),
            }},
    },
}

func TestMultiPart(t *testing.T) {
    for _, tt := range multiPartTests {
        req, err := NewRequest(
            "",
            "",
            "",
            ProtocolVersion11,
            &url.URL{},
            NewHeader(HeaderContentType, "multipart/form-data; boundary=deadbeef"))
        if err != nil {
            t.Fatal("error creating request")
        }
        req.Body = strings.NewReader(tt.body)
        parts, err := ParseMultipartForm(req, -1)
        if err != nil {
            t.Errorf("%q\n\tparse returned error %v", tt.body, err)
            continue
        }
        if !reflect.DeepEqual(req.Param, tt.param) {
            t.Errorf("%q\n\tparam=%v, want %v", req.Param, tt.param)
            continue
        }
        if !reflect.DeepEqual(parts, tt.parts) {
            t.Errorf("%q\n\tparts=%+v, want %+v", tt.body, parts, tt.parts)
            continue
        }
    }
}
