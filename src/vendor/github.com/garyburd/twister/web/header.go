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
    "sort"
    "strconv"
    "strings"
)

// Octet types from RFC 2616
var (
    isCtl   [256]bool
    isToken [256]bool
    isSpace [256]bool
)

func init() {
    // OCTET      = <any 8-bit sequence of data>
    // CHAR       = <any US-ASCII character (octets 0 - 127)>
    // CTL        = <any US-ASCII control character (octets 0 - 31) and DEL (127)>
    // CR         = <US-ASCII CR, carriage return (13)>
    // LF         = <US-ASCII LF, linefeed (10)>
    // SP         = <US-ASCII SP, space (32)>
    // HT         = <US-ASCII HT, horizontal-tab (9)>
    // <">        = <US-ASCII double-quote mark (34)>
    // CRLF       = CR LF
    // LWS        = [CRLF] 1*( SP | HT )
    // TEXT       = <any OCTET except CTLs, but including LWS>
    // separators = "(" | ")" | "<" | ">" | "@" | "," | ";" | ":" | "\" | <"> 
    //              | "/" | "[" | "]" | "?" | "=" | "{" | "}" | SP | HT
    // token      = 1*<any CHAR except CTLs or separators>
    // qdtext     = <any TEXT except <">>

    for c := 0; c < 256; c++ {
        isCtl[c] = (0 <= c && c <= 31) || c == 127
        isChar := 0 <= c && c <= 127
        isSpace[c] = strings.IndexRune(" \t\r\n", rune(c)) >= 0
        isSeparator := strings.IndexRune(" \t\"(),/:;<=>?@[]\\{}", rune(c)) >= 0
        isToken[c] = isChar && !isCtl[c] && !isSeparator
    }
}

var (
    ErrLineTooLong    = errors.New("HTTP header line too long")
    ErrBadHeaderLine  = errors.New("could not parse HTTP header line")
    ErrHeaderTooLong  = errors.New("HTTP header value too long")
    ErrHeadersTooLong = errors.New("too many HTTP headers")
)

// Header maps header names to a slice of header values. 
// 
// The header names must be in canonical format: the first letter and letters
// following '-' are uppercase and all other letters are lowercase.  The
// Header* constants are in canonical format. Use the function HeaderName to
// convert a string to canonical format.
type Header map[string][]string

// NewHeader returns a map initialized with the given key-value pairs.
func NewHeader(kvs ...string) Header {
    if len(kvs)%2 == 1 {
        panic("twister: even number args required for NewHeader")
    }
    m := Header{}
    for i := 0; i < len(kvs); i += 2 {
        m.Add(kvs[i], kvs[i+1])
    }
    return m
}

// Add appends value to slice for given key.
func (m Header) Add(key string, value string) {
    m[key] = append(m[key], value)
}

// Set value for given key, discarding previous values if any.
func (m Header) Set(key string, value string) {
    m[key] = []string{value}
}

// Get returns the first value for given key or "" if the key is not found.
func (m Header) Get(key string) string {
    values := m[key]
    if len(values) == 0 {
        return ""
    }
    return values[0]
}

// GetValueParam returns a value and optional semi-colon prefixed name-value
// pairs for header with name key. The value and parameter keys are converted
// to lowercase. All whitespace is trimmed. This format is used by the
// Content-Type and Content-Disposition headers.
func (m Header) GetValueParam(key string) (value string, param map[string]string) {
    value, param, _ = splitValueParam(m.Get(key))
    return
}

// GetList returns list of comma separated values over multiple header values
// for the given key. Commas are ignored in quoted strings. Quoted values are
// not unescaped or unquoted. Whitespace is trimmed.
func (m Header) GetList(key string) []string {
    var result []string
    for _, s := range m[key] {
        begin := 0
        end := 0
        escape := false
        quote := false
        for i := 0; i < len(s); i++ {
            b := s[i]
            switch {
            case escape:
                escape = false
                end = i + 1
            case quote:
                switch b {
                case '\\':
                    escape = true
                case '"':
                    quote = false
                }
                end = i + 1
            case b == '"':
                quote = true
                end = i + 1
            case isSpace[b]:
                if begin == end {
                    begin = i + 1
                    end = begin
                }
            case b == ',':
                if begin < end {
                    result = append(result, s[begin:end])
                }
                begin = i + 1
                end = begin
            default:
                end = i + 1
            }
        }
        if begin < end {
            result = append(result, s[begin:end])
        }
    }
    return result
}

// ValueParams represents a value with parameters.
type ValueParams struct {
    Value string
    Param map[string]string
}

type byQuality []ValueParams

func (p byQuality) Len() int      { return len(p) }
func (p byQuality) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byQuality) Less(i, j int) bool {
    qi := float64(1)
    if s, ok := p[i].Param["q"]; ok {
        qi, _ = strconv.ParseFloat(s, 64)
    }
    qj := float64(1)
    if s, ok := p[j].Param["q"]; ok {
        qj, _ = strconv.ParseFloat(s, 64)
    }
    return qj < qi
}

// GetAccept returns a parsed Accept-* header in descending quality order.
func (m Header) GetAccept(key string) []ValueParams {
    parts := m.GetList(key)
    result := make([]ValueParams, len(parts))
    for i, part := range parts {
        value, param, _ := splitValueParam(part)
        result[i].Value = value
        result[i].Param = param
    }
    sort.Sort(byQuality(result))
    return result
}

// WriteHttpHeader writes the map in HTTP header format.
func (m Header) WriteHttpHeader(w io.Writer) error {
    keys := make([]string, 0, len(m))
    for key := range m {
        keys = append(keys, key)
    }
    sort.Strings(keys)

    for _, key := range keys {
        keyBytes := []byte(key)
        for _, value := range m[key] {
            if _, err := w.Write(keyBytes); err != nil {
                return err
            }
            if _, err := w.Write(colonSpaceBytes); err != nil {
                return err
            }
            valueBytes := []byte(value)
            // Convert \r, \n and other control characters to space to 
            // prevent response splitting attacks.
            for i, c := range valueBytes {
                if isCtl[c] {
                    valueBytes[i] = ' '
                }
            }
            if _, err := w.Write(valueBytes); err != nil {
                return err
            }
            if _, err := w.Write(crlfBytes); err != nil {
                return err
            }
        }
    }
    _, err := w.Write(crlfBytes)
    return err
}

// ParseHttpHeader parses the HTTP headers and appends the values to the
// supplied map. Header names are converted to canonical format.
func (m Header) ParseHttpHeader(br *bufio.Reader) (err error) {

    const (
        // Max size for header line
        maxLineSize = 4096
        // Max size for header value
        maxValueSize = 4096
        // Maximum number of headers 
        maxHeaderCount = 256
    )

    lastKey := ""
    headerCount := 0

    for {
        p, isPrefix, err := br.ReadLine()
        switch {
        case err == io.EOF:
            return io.ErrUnexpectedEOF
        case err != nil:
            return err
        case isPrefix:
            return ErrLineTooLong
        }

        // End of headers?
        if len(p) == 0 {
            break
        }

        // Don't allow huge header lines.
        if len(p) > maxLineSize {
            return ErrLineTooLong
        }

        if isSpace[p[0]] {

            if lastKey == "" {
                return ErrBadHeaderLine
            }

            p = trimBytes(p)

            if len(p) > 0 {
                values := m[lastKey]
                value := values[len(values)-1]
                value = value + " " + string(p)
                if len(value) > maxValueSize {
                    return ErrHeaderTooLong
                }
                values[len(values)-1] = value
            }

        } else {

            // New header
            headerCount = headerCount + 1
            if headerCount > maxHeaderCount {
                return ErrHeadersTooLong
            }

            // Key
            i := 0
            for i < len(p) && isToken[p[i]] {
                i += 1
            }
            if i < 1 {
                return ErrBadHeaderLine
            }
            key := HeaderNameBytes(p[:i])
            p = p[i:]
            lastKey = key

            p = trimBytesLeft(p)

            // Colon
            if p[0] != ':' {
                return ErrBadHeaderLine
            }
            p = p[1:]

            // Value 
            value := string(trimBytes(p))
            m.Add(key, value)
        }
    }
    return nil
}

func trimBytesLeft(p []byte) []byte {
    var i int
    for i = 0; i < len(p); i++ {
        if !isSpace[p[i]] {
            break
        }
    }
    return p[i:]
}

func trimBytes(p []byte) []byte {
    p = trimBytesLeft(p)
    var i int
    for i = len(p); i > 0; i-- {
        if !isSpace[p[i-1]] {
            break
        }
    }
    return p[0:i]
}

// Header names in canonical format.
const (
    HeaderAccept             = "Accept"
    HeaderAcceptCharset      = "Accept-Charset"
    HeaderAcceptEncoding     = "Accept-Encoding"
    HeaderAcceptLanguage     = "Accept-Language"
    HeaderAcceptRanges       = "Accept-Ranges"
    HeaderAge                = "Age"
    HeaderAllow              = "Allow"
    HeaderAuthorization      = "Authorization"
    HeaderCacheControl       = "Cache-Control"
    HeaderConnection         = "Connection"
    HeaderContentDisposition = "Content-Disposition"
    HeaderContentEncoding    = "Content-Encoding"
    HeaderContentLanguage    = "Content-Language"
    HeaderContentLength      = "Content-Length"
    HeaderContentLocation    = "Content-Location"
    HeaderContentMD5         = "Content-Md5"
    HeaderContentRange       = "Content-Range"
    HeaderContentType        = "Content-Type"
    HeaderCookie             = "Cookie"
    HeaderDate               = "Date"
    HeaderETag               = "Etag"
    HeaderEtag               = "Etag"
    HeaderExpect             = "Expect"
    HeaderExpires            = "Expires"
    HeaderFrom               = "From"
    HeaderHost               = "Host"
    HeaderIfMatch            = "If-Match"
    HeaderIfModifiedSince    = "If-Modified-Since"
    HeaderIfNoneMatch        = "If-None-Match"
    HeaderIfRange            = "If-Range"
    HeaderIfUnmodifiedSince  = "If-Unmodified-Since"
    HeaderLastModified       = "Last-Modified"
    HeaderLocation           = "Location"
    HeaderMaxForwards        = "Max-Forwards"
    HeaderOrigin             = "Origin"
    HeaderPragma             = "Pragma"
    HeaderProxyAuthenticate  = "Proxy-Authenticate"
    HeaderProxyAuthorization = "Proxy-Authorization"
    HeaderRange              = "Range"
    HeaderReferer            = "Referer"
    HeaderRetryAfter         = "Retry-After"
    HeaderServer             = "Server"
    HeaderSetCookie          = "Set-Cookie"
    HeaderTE                 = "Te"
    HeaderTrailer            = "Trailer"
    HeaderTransferEncoding   = "Transfer-Encoding"
    HeaderUpgrade            = "Upgrade"
    HeaderUserAgent          = "User-Agent"
    HeaderVary               = "Vary"
    HeaderVia                = "Via"
    HeaderWWWAuthenticate    = "Www-Authenticate"
    HeaderWarning            = "Warning"
    HeaderXXSRFToken         = "X-Xsrftoken"
)

// HeaderName returns the canonical format of the header name. 
func HeaderName(name string) string {
    return HeaderNameBytes([]byte(name))
}

// HeaderNameBytes returns the canonical format for the header name specified
// by the bytes in p. This function modifies the contents p.
func HeaderNameBytes(p []byte) string {
    upper := true
    for i, c := range p {
        if upper {
            if 'a' <= c && c <= 'z' {
                p[i] = c + 'A' - 'a'
            }
        } else {
            if 'A' <= c && c <= 'Z' {
                p[i] = c + 'a' - 'A'
            }
        }
        upper = c == '-'
    }
    return string(p)
}

// QuoteHeaderValue quotes s using quoted-string rules described in RFC 2616.
func QuoteHeaderValue(s string) string {
    var b bytes.Buffer
    b.WriteByte('"')
    for i := 0; i < len(s); i++ {
        c := s[i]
        switch c {
        case '\\', '"':
            b.WriteByte('\\')
        }
        b.WriteByte(c)
    }
    b.WriteByte('"')
    return b.String()
}

// QuoteHeaderValueOrToken quotes s if s is not a valid token per RFC 2616.
func QuoteHeaderValueOrToken(s string) string {
    for i := 0; i < len(s); i++ {
        if !isToken[s[i]] {
            return QuoteHeaderValue(s)
        }
    }
    return s
}

// UnquoteHeaderValue unquotes s if s is surrounded by quotes, otherwise s is
// returned.
func UnquoteHeaderValue(s string) string {
    if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
        return s
    }
    s = s[1 : len(s)-1]
    for i := 0; i < len(s); i++ {
        if s[i] == '\\' {
            var buf bytes.Buffer
            buf.WriteString(s[:i])
            escape := true
            for j := i + 1; j < len(s); j++ {
                b := s[j]
                switch {
                case escape:
                    escape = false
                    buf.WriteByte(b)
                case b == '\\':
                    escape = true
                default:
                    buf.WriteByte(b)
                }
            }
            s = buf.String()
            break
        }
    }
    return s
}

// indexFunc returns the index in s of the first byte satisfying f(c), or -1 if
// none do.
func indexFunc(s string, f func(b byte) bool) int {
    for i := 0; i < len(s); i++ {
        if f(s[i]) {
            return i
        }
    }
    return -1
}

// splitValueParam parses a value followed by optional semi-colon prefixed
// name-value pairsParameters. It returns the value, parameters and the
// reminder of the string.
func splitValueParam(s string) (value string, param map[string]string, rest string) {
    i := indexFunc(s, func(b byte) bool { return b == ';' })
    if i < 0 {
        value = s
        rest = ""
    } else {
        value = s[:i]
        rest = s[i+1:]
    }
    value = toLowerToken(trimRight(value))
    param, rest = splitParam(rest)
    return value, param, rest
}

// splitParam returns map of RFC 2616 parameters parsed from string and the
// remainder of the string.
func splitParam(s string) (param map[string]string, rest string) {
    param = make(map[string]string)
    for {
        var name string
        name, s = splitToken(skipSpace(s))
        if name == "" {
            break
        }
        if len(s) == 0 || s[0] != '=' {
            break
        }
        var value string
        value, s = splitTokenOrQuoted(s[1:])
        if value == "" {
            break
        }
        param[toLowerToken(name)] = value
        s = skipSpace(s)
        if len(s) == 0 || s[0] != ';' {
            return param, s
        }
        s = s[1:]
    }
    return param, s
}

// skipSpace returns remainder of s following any RFC 2616 whitespace.
func skipSpace(s string) (rest string) {
    i := indexFunc(s, func(b byte) bool { return !isSpace[b] })
    if i < 0 {
        return ""
    }
    return s[i:]
}

// splitToken returns RFC 2616 token at start of s and the remainder of s.
func splitToken(s string) (token, rest string) {
    i := indexFunc(s, func(b byte) bool { return !isToken[b] })
    if i < 0 {
        return s, ""
    }
    return s[:i], s[i:]
}

// splitQuoted returns RFC 2616 quoted value at the start of s and the
// remainder of s. The value is unescaped and quotes are removed. 
func splitQuoted(s string) (value, rest string) {
    if len(s) == 0 || s[0] != '"' {
        return "", ""
    }
    s = s[1:]
    for i := 0; i < len(s); i++ {
        switch s[i] {
        case '"':
            return s[:i], s[i+1:]
        case '\\':
            p := make([]byte, len(s)-1)
            j := copy(p, s[:i])
            escape := true
            for i = i + i; i < len(s); i++ {
                b := s[i]
                switch {
                case escape:
                    escape = false
                    p[j] = b
                    j += 1
                case b == '\\':
                    escape = true
                case b == '"':
                    return string(p[:j]), s[i+1:]
                default:
                    p[j] = b
                    j += 1
                }
            }
            return "", ""
        }
    }
    return "", ""
}

func splitTokenOrQuoted(s string) (string, string) {
    if len(s) == 0 {
        return "", ""
    }
    if s[0] == '"' {
        return splitQuoted(s)
    }
    return splitToken(s)
}

// toLowerToken converts RFC 2616 token bytes to lowercase.
func toLowerToken(s string) string {
    for i := 0; i < len(s); i++ {
        b := s[i]
        if 'A' <= b && b <= 'Z' {
            p := make([]byte, len(s))
            copy(p, s[:i])
            p[i] = b + ('a' - 'A')
            for i = i + 1; i < len(s); i++ {
                b := s[i]
                if 'A' <= b && b <= 'Z' {
                    b = b + ('a' - 'A')
                }
                p[i] = b
            }
            return string(p)
        }
    }
    return s
}

// trimRight removes RFC 2616 whitespace bytes from the end of s.
func trimRight(s string) string {
    var i int
    for i = len(s); i > 0; i-- {
        if !isSpace[s[i-1]] {
            break
        }
    }
    return s[0:i]
}
