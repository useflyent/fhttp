// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/useflyent/fhttp/internal/race"
)

var headerWriteTests = []struct {
	h        Header
	exclude  map[string]bool
	expected string
}{
	{Header{}, nil, ""},
	{
		Header{
			"Content-Type":   {"text/html; charset=UTF-8"},
			"Content-Length": {"0"},
		},
		nil,
		"Content-Length: 0\r\nContent-Type: text/html; charset=UTF-8\r\n",
	},
	{
		Header{
			"Content-Length": {"0", "1", "2"},
		},
		nil,
		"Content-Length: 0\r\nContent-Length: 1\r\nContent-Length: 2\r\n",
	},
	{
		Header{
			"Expires":          {"-1"},
			"Content-Length":   {"0"},
			"Content-Encoding": {"gzip"},
		},
		map[string]bool{"Content-Length": true},
		"Content-Encoding: gzip\r\nExpires: -1\r\n",
	},
	{
		Header{
			"Expires":          {"-1"},
			"Content-Length":   {"0", "1", "2"},
			"Content-Encoding": {"gzip"},
		},
		map[string]bool{"Content-Length": true},
		"Content-Encoding: gzip\r\nExpires: -1\r\n",
	},
	{
		Header{
			"Expires":          {"-1"},
			"Content-Length":   {"0"},
			"Content-Encoding": {"gzip"},
		},
		map[string]bool{"Content-Length": true, "Expires": true, "Content-Encoding": true},
		"",
	},
	{
		Header{
			"Nil":          nil,
			"Empty":        {},
			"Blank":        {""},
			"Double-Blank": {"", ""},
		},
		nil,
		"Blank: \r\nDouble-Blank: \r\nDouble-Blank: \r\n",
	},
	// Tests header sorting when over the insertion sort threshold side:
	{
		Header{
			"k1": {"1a", "1b"},
			"k2": {"2a", "2b"},
			"k3": {"3a", "3b"},
			"k4": {"4a", "4b"},
			"k5": {"5a", "5b"},
			"k6": {"6a", "6b"},
			"k7": {"7a", "7b"},
			"k8": {"8a", "8b"},
			"k9": {"9a", "9b"},
		},
		map[string]bool{"k5": true},
		"k1: 1a\r\nk1: 1b\r\nk2: 2a\r\nk2: 2b\r\nk3: 3a\r\nk3: 3b\r\n" +
			"k4: 4a\r\nk4: 4b\r\nk6: 6a\r\nk6: 6b\r\n" +
			"k7: 7a\r\nk7: 7b\r\nk8: 8a\r\nk8: 8b\r\nk9: 9a\r\nk9: 9b\r\n",
	},
	// Test sorting headers by the special Header-Order header
	{
		Header{
			"a":            {"2"},
			"b":            {"3"},
			"e":            {"1"},
			"c":            {"5"},
			"d":            {"4"},
			HeaderOrderKey: {"e", "a", "b", "d", "c"},
		},
		nil,
		"e: 1\r\na: 2\r\nb: 3\r\nd: 4\r\nc: 5\r\n",
	},
	// Make sure that http 1.1 capitla letters are also sorted properly
	{
		Header{
			"X-NewRelic-ID":         {"12345"},
			"x-api-key":             {"ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
			"MESH-Commerce-Channel": {"android-app-phone"},
			"mesh-version":          {"cart=4"},
			"User-Agent":            {"size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)"},
			"X-Request-Auth":        {"hawkHeader"},
			"X-acf-sensor-data":     {"3456"},
			"Content-Type":          {"application/json; charset=UTF-8"},
			"Accept":                {"application/json"},
			"Transfer-Encoding":     {"chunked"},
			"Host":                  {"prod.jdgroupmesh.cloud"},
			"Connection":            {"Keep-Alive"},
			"Accept-Encoding":       {"gzip"},
			HeaderOrderKey: {
				"X-NewRelic-ID",
				"x-api-key",
				"MESH-Commerce-Channel",
				"mesh-version",
				"User-Agent",
				"X-Request-Auth",
				"X-acf-sensor-data",
				"Content-Type",
				"Accept",
				"Transfer-Encoding",
				"Host",
				"Connection",
				"Accept-Encoding",
			},
		},
		nil,
		"X-NewRelic-ID: 12345\r\nx-api-key: ABCDEFGHIJKLMNOPQRSTUVWXYZ\r\nMESH-Commerce-Channel: android-app-phone\r\n" +
			"mesh-version: cart=4\r\nUser-Agent: size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)\r\n" +
			"X-Request-Auth: hawkHeader\r\nX-acf-sensor-data: 3456\r\nContent-Type: application/json; charset=UTF-8\r\n" +
			"Accept: application/json\r\nTransfer-Encoding: chunked\r\nHost: prod.jdgroupmesh.cloud\r\nConnection: Keep-Alive\r\n" +
			"Accept-Encoding: gzip\r\n",
	},
}

func TestHeaderWrite(t *testing.T) {
	var buf bytes.Buffer
	for i, test := range headerWriteTests {
		test.h.WriteSubset(&buf, test.exclude)
		if buf.String() != test.expected {
			t.Errorf("#%d:\n got: %q\nwant: %q", i, buf.String(), test.expected)
		}
		buf.Reset()
	}
}

var parseTimeTests = []struct {
	h   Header
	err bool
}{
	{Header{"Date": {""}}, true},
	{Header{"Date": {"invalid"}}, true},
	{Header{"Date": {"1994-11-06T08:49:37Z00:00"}}, true},
	{Header{"Date": {"Sun, 06 Nov 1994 08:49:37 GMT"}}, false},
	{Header{"Date": {"Sunday, 06-Nov-94 08:49:37 GMT"}}, false},
	{Header{"Date": {"Sun Nov  6 08:49:37 1994"}}, false},
}

func TestParseTime(t *testing.T) {
	expect := time.Date(1994, 11, 6, 8, 49, 37, 0, time.UTC)
	for i, test := range parseTimeTests {
		d, err := ParseTime(test.h.Get("Date"))
		if err != nil {
			if !test.err {
				t.Errorf("#%d:\n got err: %v", i, err)
			}
			continue
		}
		if test.err {
			t.Errorf("#%d:\n  should err", i)
			continue
		}
		if !expect.Equal(d) {
			t.Errorf("#%d:\n got: %v\nwant: %v", i, d, expect)
		}
	}
}

type hasTokenTest struct {
	header string
	token  string
	want   bool
}

var hasTokenTests = []hasTokenTest{
	{"", "", false},
	{"", "foo", false},
	{"foo", "foo", true},
	{"foo ", "foo", true},
	{" foo", "foo", true},
	{" foo ", "foo", true},
	{"foo,bar", "foo", true},
	{"bar,foo", "foo", true},
	{"bar, foo", "foo", true},
	{"bar,foo, baz", "foo", true},
	{"bar, foo,baz", "foo", true},
	{"bar,foo, baz", "foo", true},
	{"bar, foo, baz", "foo", true},
	{"FOO", "foo", true},
	{"FOO ", "foo", true},
	{" FOO", "foo", true},
	{" FOO ", "foo", true},
	{"FOO,BAR", "foo", true},
	{"BAR,FOO", "foo", true},
	{"BAR, FOO", "foo", true},
	{"BAR,FOO, baz", "foo", true},
	{"BAR, FOO,BAZ", "foo", true},
	{"BAR,FOO, BAZ", "foo", true},
	{"BAR, FOO, BAZ", "foo", true},
	{"foobar", "foo", false},
	{"barfoo ", "foo", false},
}

func TestHasToken(t *testing.T) {
	for _, tt := range hasTokenTests {
		if hasToken(tt.header, tt.token) != tt.want {
			t.Errorf("hasToken(%q, %q) = %v; want %v", tt.header, tt.token, !tt.want, tt.want)
		}
	}
}

func TestNilHeaderClone(t *testing.T) {
	t1 := Header(nil)
	t2 := t1.Clone()
	if t2 != nil {
		t.Errorf("cloned header does not match original: got: %+v; want: %+v", t2, nil)
	}
}

var testHeader = Header{
	"Content-Length": {"123"},
	"Content-Type":   {"text/plain"},
	"Date":           {"some date at some time Z"},
	"Server":         {DefaultUserAgent},
}

var buf bytes.Buffer

func BenchmarkHeaderWriteSubset(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		testHeader.WriteSubset(&buf, nil)
	}
}

func TestHeaderWriteSubsetAllocs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping alloc test in short mode")
	}
	if race.Enabled {
		t.Skip("skipping test under race detector")
	}
	if runtime.GOMAXPROCS(0) > 1 {
		t.Skip("skipping; GOMAXPROCS>1")
	}
	n := testing.AllocsPerRun(100, func() {
		buf.Reset()
		testHeader.WriteSubset(&buf, nil)
	})
	if n > 0 {
		t.Errorf("allocs = %g; want 0", n)
	}
}

// Issue 34878: test that every call to
// cloneOrMakeHeader never returns a nil Header.
func TestCloneOrMakeHeader(t *testing.T) {
	tests := []struct {
		name     string
		in, want Header
	}{
		{"nil", nil, Header{}},
		{"empty", Header{}, Header{}},
		{
			name: "non-empty",
			in:   Header{"foo": {"bar"}},
			want: Header{"foo": {"bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cloneOrMakeHeader(tt.in)
			if got == nil {
				t.Fatal("unexpected nil Header")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Got:  %#v\nWant: %#v", got, tt.want)
			}
			got.Add("A", "B")
			got.Get("A")
		})
	}
}

// TestHTTP1HeaderOrder tests capitalized http1.1 header order written by request
func TestHTTP1HeaderOrder(t *testing.T) {
	req, err := NewRequest("GET", "https://prod.jdgroupmesh.cloud/stores/size/products/16069871?channel=android-app-phone&expand=variations,informationBlocks,customisations", nil)
	if err != nil {
		t.Fatalf(err.Error())
	}
	req.Header = Header{
		"X-NewRelic-ID":         {"12345"},
		"x-api-key":             {"ABCDE12345"},
		"MESH-Commerce-Channel": {"android-app-phone"},
		"mesh-version":          {"cart=4"},
		"User-Agent":            {"size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)"},
		"X-Request-Auth":        {"hawkHeader"},
		"X-acf-sensor-data":     {"3456"},
		"Content-Type":          {"application/json; charset=UTF-8"},
		"Accept":                {"application/json"},
		"Transfer-Encoding":     {"chunked"},
		"Host":                  {"prod.jdgroupmesh.cloud"},
		"Connection":            {"Keep-Alive"},
		"Accept-Encoding":       {"gzip"},
		HeaderOrderKey: {
			"x-newrelic-id",
			"x-api-key",
			"mesh-commerce-channel",
			"mesh-version",
			"user-agent",
			"x-request-auth",
			"x-acf-sensor-data",
			"transfer-encoding",
			"content-type",
			"accept",
			"host",
			"connection",
			"accept-encoding",
		},
		PHeaderOrderKey: {
			":method",
			":path",
			":authority",
			":scheme",
		},
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	err = req.Write(buf)
	if err != nil {
		t.Fatalf(err.Error())
	}
	expected := "GET /stores/size/products/16069871?channel=android-app-phone&expand=variations,informationBlocks,customisations HTTP/1.1\r\nX-NewRelic-ID: 12345\r\nx-api-key: ABCDE12345\r\nMESH-Commerce-Channel: android-app-phone\r\nmesh-version: cart=4\r\nUser-Agent: size/3.1.0.8355 (android-app-phone; Android 10; Build/CPH2185_11_A.28)\r\nX-Request-Auth: hawkHeader\r\nX-acf-sensor-data: 3456\r\nTransfer-Encoding: chunked\r\nContent-Type: application/json; charset=UTF-8\r\nAccept: application/json\r\nHost: prod.jdgroupmesh.cloud\r\nConnection: Keep-Alive\r\nAccept-Encoding: gzip\r\n\r\n"
	if expected != buf.String() {
		t.Fatalf("got:\n%swant:\n%s", buf.String(), expected)
	}
}
