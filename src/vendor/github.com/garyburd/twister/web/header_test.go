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
    "reflect"
    "testing"
)

var quoteHeaderValueTests = []struct {
    s            string
    quote        string
    quoteOrToken string
}{
    {s: "a", quote: "\"a\"", quoteOrToken: "a"},
    {s: "x\"y", quote: "\"x\\\"y\"", quoteOrToken: "\"x\\\"y\""},
    {s: "x\\y", quote: "\"x\\\\y\"", quoteOrToken: "\"x\\\\y\""},
}

func TestQuoteHeaderValue(t *testing.T) {
    for _, tt := range quoteHeaderValueTests {
        if quote := QuoteHeaderValue(tt.s); quote != tt.quote {
            t.Errorf("QuoteHeaderValue(%q) = %q, want %q", tt.s, quote, tt.quote)
        }
        if quoteOrToken := QuoteHeaderValueOrToken(tt.s); quoteOrToken != tt.quoteOrToken {
            t.Errorf("QuoteHeaderValueOrToken(%q) = %q, want %q", tt.s, quoteOrToken, tt.quoteOrToken)
        }
    }
}

var UnquoteHeaderValueTests = []struct {
    s       string
    unquote string
}{
    {s: "a", unquote: "a"},
    {s: "a b", unquote: "a b"},
    {s: "\"a\"", unquote: "a"},
    {s: "\"a \\\\ b \\\" c\"", unquote: "a \\ b \" c"},
}

func TestUnquoteHeaderValue(t *testing.T) {
    for _, tt := range UnquoteHeaderValueTests {
        if unquote := UnquoteHeaderValue(tt.s); unquote != tt.unquote {
            t.Errorf("UnquoteHeaderValue(%q) = %q, want %q", tt.s, unquote, tt.unquote)
        }
    }
}

var getHeaderListTests = []struct {
    s   string
    l   []string
}{
    {s: `a`, l: []string{`a`}},
    {s: `a, b , c `, l: []string{`a`, `b`, `c`}},
    {s: `a,, b , , c `, l: []string{`a`, `b`, `c`}},
    {s: `a,b,c`, l: []string{`a`, `b`, `c`}},
    {s: ` a b, c d `, l: []string{`a b`, `c d`}},
    {s: `"a, b, c", d `, l: []string{`"a, b, c"`, "d"}},
    {s: `","`, l: []string{`","`}},
    {s: `"\""`, l: []string{`"\""`}},
    {s: `" "`, l: []string{`" "`}},
}

func TestGetHeaderList(t *testing.T) {
    for _, tt := range getHeaderListTests {
        header := NewHeader("foo", tt.s)
        if l := header.GetList("foo"); !reflect.DeepEqual(tt.l, l) {
            t.Errorf("GetList for %q = %q, want %q", tt.s, l, tt.l)
        }
    }
}

var parseHTTPHeaderTests = []struct {
    name   string
    header Header
    s      string
}{
    {"multihdr", NewHeader(
        HeaderContentType, "text/html",
        HeaderCookie, "hello=world",
        HeaderCookie, "foo=bar"),
        `Content-Type: text/html
CoOkie: hello=world
Cookie: foo=bar

`},
    {"continuation", NewHeader(
        HeaderContentType, "text/html",
        HeaderCookie, "hello=world, foo=bar"),
        `Cookie: hello=world,
 foo=bar
Content-Type: text/html

`},
}

func TestParseHttpHeader(t *testing.T) {
    for _, tt := range parseHTTPHeaderTests {
        b := bufio.NewReader(bytes.NewBufferString(tt.s))
        header := Header{}
        err := header.ParseHttpHeader(b)
        if err != nil {
            t.Errorf("ParseHttpHeader error for %s = %v", tt.name, err)
        }
        if !reflect.DeepEqual(tt.header, header) {
            t.Errorf("ParseHttpHeader for %s = %q, want %q", tt.name, header, tt.header)
        }
    }
}

var getValueParamTests = []struct {
    s     string
    value string
    param map[string]string
}{
    {`text/html`, "text/html", map[string]string{}},
    {`text/html  `, "text/html", map[string]string{}},
    {`text/html ; `, "text/html", map[string]string{}},
    {`tExt/htMl`, "text/html", map[string]string{}},
    {`tExt/htMl; fOO=";"; hellO=world`, "text/html", map[string]string{
        "hello": "world",
        "foo":   `;`,
    }},
    {`"quoted"`, `"quoted"`, map[string]string{}},
    {`text/html; foo=bar, hello=world`, "text/html", map[string]string{"foo": "bar"}},
    {`text/html ; foo=bar `, "text/html", map[string]string{"foo": "bar"}},
    {`text/html ;foo=bar `, "text/html", map[string]string{"foo": "bar"}},
    {`text/html; foo="b\ar"`, "text/html", map[string]string{"foo": "bar"}},
    {`text/html; foo="b\"a\"r"`, "text/html", map[string]string{"foo": "b\"a\"r"}},
    {`text/html; foo="b;ar"`, "text/html", map[string]string{"foo": "b;ar"}},
    {`text/html; FOO="bar"`, "text/html", map[string]string{"foo": "bar"}},
    {`form-data; filename="file.txt"; name=file`, "form-data", map[string]string{"filename": "file.txt", "name": "file"}},
}

func TestGetValueParam(t *testing.T) {
    for _, tt := range getValueParamTests {
        header := NewHeader(HeaderContentType, tt.s)
        value, param := header.GetValueParam(HeaderContentType)
        if value != tt.value {
            t.Errorf("%q, value=%s, want %s", tt.s, value, tt.value)
        }
        if !reflect.DeepEqual(param, tt.param) {
            t.Errorf("%q, param=%v, want %v", tt.s, param, tt.param)
        }
    }
}

var getAcceptTests = []struct {
    s   string
    vps []ValueParams
}{
    {"audio/*; q=0.2, audio/basic", []ValueParams{
        {"audio/basic", map[string]string{}},
        {"audio/*", map[string]string{"q": "0.2"}},
    }},
    {"text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c", []ValueParams{
        {"text/html", map[string]string{}},
        {"text/x-c", map[string]string{}},
        {"text/x-dvi", map[string]string{"q": "0.8"}},
        {"text/plain", map[string]string{"q": "0.5"}},
    }},
}

func TestGetAccept(t *testing.T) {
    for _, tt := range getAcceptTests {
        header := NewHeader(HeaderAccept, tt.s)
        vps := header.GetAccept(HeaderAccept)
        if !reflect.DeepEqual(vps, tt.vps) {
            t.Errorf("accept(%q)=%v, want %v", tt.s, vps, tt.vps)
        }
    }
}
