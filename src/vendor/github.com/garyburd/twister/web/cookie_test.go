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
    "reflect"
    "testing"
)

var ParseCookieValuesTests = []struct {
    values []string
    m      Values
}{
    {[]string{"a=b"}, Values{"a": []string{"b"}}},
    {[]string{"a=b; c"}, Values{"a": []string{"b"}}},
    {[]string{"a=b; =c"}, Values{"a": []string{"b"}}},
    {[]string{"a=b; ; "}, Values{"a": []string{"b"}}},
    {[]string{"a=b; c=d"}, Values{"a": []string{"b"}, "c": []string{"d"}}},
    {[]string{"a=b; c=d"}, Values{"a": []string{"b"}, "c": []string{"d"}}},
    {[]string{"a=b;c=d"}, Values{"a": []string{"b"}, "c": []string{"d"}}},
    {[]string{" a=b;c=d "}, Values{"a": []string{"b"}, "c": []string{"d"}}},
    {[]string{"a=b", "c=d"}, Values{"a": []string{"b"}, "c": []string{"d"}}},
    {[]string{"a=b", "c=x=y"}, Values{"a": []string{"b"}, "c": []string{"x=y"}}},
}

func TestParseCookieValues(t *testing.T) {
    for _, pt := range ParseCookieValuesTests {
        m := make(Values)
        if err := parseCookieValues(pt.values, m); err != nil {
            t.Errorf("parseCookieValues(%q) error %q", pt.values, err)
        }
        if !reflect.DeepEqual(pt.m, m) {
            t.Errorf("parseCookieValues(%q) = %q, want %q", pt.values, m, pt.m)
        }
    }
}
