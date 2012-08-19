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
    "bytes"
    "strconv"
    "time"
)

// parseCookieValues parses cookies from values and adds them to m. The
// function supports the Netscape draft specification for cookies
// (http://goo.gl/1WSx3). 
func parseCookieValues(values []string, m Values) error {
    for _, s := range values {
        key := ""
        begin := 0
        end := 0
        for i := 0; i < len(s); i++ {
            switch s[i] {
            case ' ', '\t':
                // leading whitespace?
                if begin == end {
                    begin = i + 1
                    end = begin
                }
            case '=':
                if key == "" {
                    key = s[begin:end]
                    begin = i + 1
                    end = begin
                } else {
                    end += 1
                }
            case ';':
                if len(key) > 0 && begin < end {
                    value := s[begin:end]
                    m.Add(key, value)
                }
                key = ""
                begin = i + 1
                end = begin
            default:
                end = i + 1
            }
        }
        if len(key) > 0 && begin < end {
            m.Add(key, s[begin:end])
        }
    }
    return nil
}

// Cookie is a helper for constructing Set-Cookie header values. 
// 
// Cookie supports the ancient Netscape draft specification for cookies
// (http://goo.gl/1WSx3) and the modern HttpOnly attribute
// (http://www.owasp.org/index.php/HttpOnly). Cookie does not attempt to
// support any RFC for cookies because the RFCs are not supported by popular
// browsers.
//
// As a convenience, the NewCookie function returns a cookie with the path
// attribute set to "/" and the httponly attribute set to true. 
type Cookie struct {
    name     string
    value    string
    path     string
    domain   string
    maxAge   time.Duration
    secure   bool
    httpOnly bool
}

// NewCookie returns a new cookie with the given name and value, the path
// attribute set to "/" and the httponly attribute set to true.
func NewCookie(name, value string) *Cookie {
    return &Cookie{name: name, value: value, path: "/", httpOnly: true}
}

// Path sets the cookie path attribute. The path must either be "" or start with a
// '/'.  The NewCookie function initializes the path to "/". If the path is "",
// then the path attribute is not included in the header value. 
func (c *Cookie) Path(path string) *Cookie { c.path = path; return c }

// Domain sets the cookie domain attribute. If the host is "", then the domain
// attribute is not included in the header value. 
func (c *Cookie) Domain(domain string) *Cookie { c.domain = domain; return c }

// MaxAge specifies the maximum age for a cookie. If the maximum age is 0, then
// the expiration time is not included in the header value and the browser will
// handle the cookie as a "session" cookie. To support Internet Explorer, the
// maximum age is also rendered as an absolute expiration time.
func (c *Cookie) MaxAge(maxAge time.Duration) *Cookie { c.maxAge = maxAge; return c }

// Delete sets the expiration date to a time in the past. 
func (c *Cookie) Delete() *Cookie { return c.MaxAge(-30 * 24 * time.Hour).HTTPOnly(false) }

// Secure sets the secure attribute. 
func (c *Cookie) Secure(secure bool) *Cookie { c.secure = secure; return c }

// HTTPOnly sets the httponly attribute. The NewCookie function
// initializes the httponly attribute to true.
func (c *Cookie) HTTPOnly(httpOnly bool) *Cookie {
    c.httpOnly = httpOnly
    return c
}

// String renders the Set-Cookie header value as a string.
func (c *Cookie) String() string {
    var buf bytes.Buffer

    buf.WriteString(c.name)
    buf.WriteByte('=')
    buf.WriteString(c.value)

    if c.path != "" {
        buf.WriteString("; path=")
        buf.WriteString(c.path)
    }

    if c.domain != "" {
        buf.WriteString("; domain=")
        buf.WriteString(c.domain)
    }

    if c.maxAge != 0 {
        buf.WriteString("; max-age=")
        buf.WriteString(strconv.Itoa(int(c.maxAge / time.Second)))
        buf.WriteString("; expires=")
        buf.WriteString(formatExpiration(c.maxAge))
    }

    if c.secure {
        buf.WriteString("; secure")
    }

    if c.httpOnly {
        buf.WriteString("; HttpOnly")
    }

    return buf.String()
}
